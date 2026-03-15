package service

import (
	"archive/zip"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/ledongthuc/pdf"

	"github.com/sashabaranov/go-openai"

	"github.com/lojes7/inquire/internal/model"
	"github.com/lojes7/inquire/pkg/infra"
	"github.com/lojes7/inquire/pkg/secure"
)

const (
	defaultEmbeddingModel = "text-embedding-v3"
	defaultSearchLimit    = 5
	maxSearchLimit        = 20
	maxEmbeddingTextBytes = 1 << 20 // 1MB
)

// EmbeddingProvider 抽象向量化服务，后续可切换到其它模型或远程服务。
type EmbeddingProvider interface {
	EmbedText(ctx context.Context, text string) (model.Vector, error)
}

// FileContentExtractor 抽象文件内容提取
type FileContentExtractor interface {
	Extract(filePath, originalName, fileType string) (string, error)
}

// FileIndexer 抽象文件索引流程
type FileIndexer interface {
	BuildIndex(ctx context.Context, filePath, originalName, fileType string) (content string, vector model.Vector, err error)
}

// FileVectorRepository 抽象向量检索存储
type FileVectorRepository interface {
	SearchByVector(ctx context.Context, userID uint64, queryVector model.Vector, limit int) ([]model.FileSemanticSearchItemResp, error)
}

// Qwen AI 结构体
type qwenEmbeddingProvider struct {
	client *openai.Client
	model  string
}

func (p *qwenEmbeddingProvider) EmbedText(ctx context.Context, text string) (model.Vector, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil, errors.New("embedding text is empty")
	}

	resp, err := p.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Model: openai.EmbeddingModel(p.model),
		Input: []string{trimmed},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 || len(resp.Data[0].Embedding) == 0 {
		return nil, errors.New("embedding response is empty")
	}
	return resp.Data[0].Embedding, nil
}

type localFileContentExtractor struct{}

func (e *localFileContentExtractor) Extract(filePath, originalName, fileType string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".pdf":
		return extractPdfContent(filePath)
	case ".docx":
		return extractDocxContent(filePath)
	case ".doc":
		return extractDocContent(filePath)
	case ".jpg", ".jpeg", ".png", ".bmp", ".tiff":
		return extractImageContent(filePath)
	}

	if isTextLikeFile(filePath, fileType) {
		f, err := os.Open(filePath)
		if err != nil {
			return "", err
		}
		defer f.Close()

		data, err := io.ReadAll(io.LimitReader(f, maxEmbeddingTextBytes))
		if err != nil {
			return "", err
		}
		content := strings.TrimSpace(strings.ToValidUTF8(string(data), " "))
		if content != "" {
			return content, nil
		}
	}

	return buildFileMetadataSummary(filePath, originalName, fileType), nil
}

func extractPdfContent(path string) (string, error) {
	// 1. 尝试使用纯 Go 库解析文本
	content, err := extractPdfTextNative(path)

	// 如果提取成功且内容长度可观，直接返回
	if err == nil && len(strings.TrimSpace(content)) > 50 {
		return content, nil
	}

	// 2. 如果纯 Go 提取失败或内容太少(可能是扫描版)，尝试使用 Tesseract OCR 直接处理 PDF
	// 前提：系统安装了 tesseract 且 tesseract 支持 PDF (通常需要 Ghostscript)
	ocrContent, ocrErr := extractContentViaTesseract(path)
	if ocrErr == nil && len(strings.TrimSpace(ocrContent)) > 0 {
		return ocrContent, nil
	}

	if err != nil {
		return "", err
	}
	return content, nil
}

func extractPdfTextNative(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var buf strings.Builder
	totalPage := r.NumPage()

	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}
		text, err := p.GetPlainText(nil)
		if err != nil {
			// 如果某一页读取失败，可以选择忽略或返回错误，这里选择继续尝试后面页面
			continue
		}
		if buf.Len()+len(text) > maxEmbeddingTextBytes {
			buf.WriteString(text[:maxEmbeddingTextBytes-buf.Len()])
			break
		}
		buf.WriteString(text)
		buf.WriteString("\n")
	}
	return strings.TrimSpace(buf.String()), nil
}

// extractDocContent 处理旧版 Word (.doc)
// 策略：优先使用 antiword 命令行工具；如果不可用，降级为提取二进制中的可打印字符（类似 strings 命令）
func extractDocContent(path string) (string, error) {
	// 1. 尝试 antiword
	if _, err := exec.LookPath("antiword"); err == nil {
		cmd := exec.Command("antiword", path)
		out, err := cmd.Output()
		if err == nil {
			return limitText(string(out)), nil
		}
	}

	// 2. 尝试 catdoc
	if _, err := exec.LookPath("catdoc"); err == nil {
		cmd := exec.Command("catdoc", path)
		out, err := cmd.Output()
		if err == nil {
			return limitText(string(out)), nil
		}
	}

	// 3. 降级策略：二进制滤光（提取连续的可打印字符串）
	// 虽然会丢失格式且包含乱码，但对于全文检索而言，能提取出关键词即可
	return extractPrintableChars(path)
}

// extractImageContent 使用 Tesseract OCR 提取图片内容
func extractImageContent(path string) (string, error) {
	return extractContentViaTesseract(path)
}

func extractContentViaTesseract(path string) (string, error) {
	// 检查 tesseract 是否存在
	if _, err := exec.LookPath("tesseract"); err != nil {
		return "", errors.New("tesseract not installed: skipping OCR")
	}

	// 执行 tesseract <input> stdout -l eng+chi_sim
	// 注意：这里假设用户安装了中英文语言包
	cmd := exec.Command("tesseract", path, "stdout", "-l", "chi_sim+eng")

	// 如果指定语言包失败，尝试默认语言
	if err := cmd.Start(); err != nil {
		cmd = exec.Command("tesseract", path, "stdout")
	} else {
		// 上面的 Start 只是检查能不能启动，这里我们需要 Wait，但 cmd 不能复用，所以重新构建
		cmd = exec.Command("tesseract", path, "stdout", "-l", "chi_sim+eng")
	}

	out, err := cmd.Output()
	if err != nil {
		// 再次降级到默认参数（不指定语言）
		cmd = exec.Command("tesseract", path, "stdout")
		out, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("ocr failed: %v", err)
		}
	}

	return limitText(string(out)), nil
}

// extractPrintableChars 从二进制文件中提取可打印字符
func extractPrintableChars(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// 限制读取大小
	data, err := io.ReadAll(io.LimitReader(f, maxEmbeddingTextBytes))
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	var lastIsPrintable bool

	for _, b := range data {
		r := rune(b)
		// 简单的 ASCII 可打印字符判断，或者 UTF8 序列（这里简化为 ASCII 过滤，对于中文二进制 .doc 支持较差，但这是最后的保底）
		// 改进：只保留 ASCII 图形字符和空白
		isPrintable := (r >= 32 && r <= 126) || r == '\n' || r == '\r' || r == '\t'

		if isPrintable {
			buf.WriteByte(b)
			lastIsPrintable = true
		} else {
			if lastIsPrintable {
				buf.WriteByte(' ') // 替换不可打印字符为空格
			}
			lastIsPrintable = false
		}
	}
	return limitText(buf.String()), nil
}

func limitText(s string) string {
	s = strings.TrimSpace(s)
	// 简单的 UTF-8 校验和清洗
	if !utf8.ValidString(s) {
		v := make([]rune, 0, len(s))
		for _, r := range s {
			if r == utf8.RuneError {
				continue
			}
			v = append(v, r)
		}
		s = string(v)
	}

	if len(s) > maxEmbeddingTextBytes {
		return s[:maxEmbeddingTextBytes]
	}
	return s
}

func extractDocxContent(path string) (string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer r.Close()

	var documentFile *zip.File
	for _, f := range r.File {
		if f.Name == "word/document.xml" {
			documentFile = f
			break
		}
	}

	if documentFile == nil {
		return "", errors.New("invalid docx file: missing word/document.xml")
	}

	rc, err := documentFile.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	decoder := xml.NewDecoder(rc)
	var buf strings.Builder

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "t" { // <w:t> contains text
				var text string
				if err := decoder.DecodeElement(&text, &t); err != nil {
					return "", err
				}
				if buf.Len()+len(text) > maxEmbeddingTextBytes {
					buf.WriteString(text[:maxEmbeddingTextBytes-buf.Len()])
					return strings.TrimSpace(buf.String()), nil
				}
				buf.WriteString(text)
			} else if t.Name.Local == "p" || t.Name.Local == "br" { // <w:p> paragraph or <w:br> break
				if buf.Len() > 0 && !strings.HasSuffix(buf.String(), "\n") {
					buf.WriteString("\n")
				}
			}
		}
	}

	return limitText(buf.String()), nil
}

type syncFileIndexer struct {
	extractor FileContentExtractor
	embedder  EmbeddingProvider
}

func (i *syncFileIndexer) BuildIndex(ctx context.Context, filePath, originalName, fileType string) (string, model.Vector, error) {
	content, err := i.extractor.Extract(filePath, originalName, fileType)
	if err != nil {
		return "", nil, err
	}
	vector, err := i.embedder.EmbedText(ctx, content)
	if err != nil {
		return "", nil, err
	}
	return content, vector, nil
}

type gormFileVectorRepository struct{}

func (r *gormFileVectorRepository) SearchByVector(ctx context.Context, userID uint64, queryVector model.Vector, limit int) ([]model.FileSemanticSearchItemResp, error) {
	db := infra.GetDB().WithContext(ctx)
	results := make([]model.FileSemanticSearchItemResp, 0)

	vectorLiteral, err := vectorToLiteral(queryVector)
	if err != nil {
		return nil, err
	}

	sql := `SELECT f.message_id,
			f.file_name,
			f.file_url,
			f.file_size,
			f.file_type,
			m.conversation_id,
			(f.content_vector <-> ?::vector) AS score
		FROM files f
		JOIN messages m ON m.id = f.message_id
		JOIN conversation_users cu ON cu.conversation_id = m.conversation_id
		WHERE cu.user_id = ? AND f.content_vector IS NOT NULL
		ORDER BY f.content_vector <-> ?::vector
		LIMIT ?`

	res := db.Raw(sql, vectorLiteral, userID, vectorLiteral, limit).Scan(&results)
	if res.Error != nil {
		return nil, secure.Wrap(500, "语义检索失败", res.Error)
	}
	return results, nil
}

var (
	semanticOnce    sync.Once
	semanticInitErr error

	semanticIndexer  FileIndexer
	semanticRepo     FileVectorRepository
	semanticEmbedder EmbeddingProvider
)

func initSemanticServices() error {
	semanticOnce.Do(func() {
		embedder, err := newQwenEmbeddingProviderFromEnv()
		if err != nil {
			semanticInitErr = err
			return
		}

		semanticEmbedder = embedder
		semanticIndexer = &syncFileIndexer{
			extractor: &localFileContentExtractor{},
			embedder:  embedder,
		}
		semanticRepo = &gormFileVectorRepository{}
	})

	if semanticInitErr != nil {
		return secure.Wrap(500, "AI 服务初始化失败", semanticInitErr)
	}
	return nil
}

func buildFileIndex(ctx context.Context, filePath, originalName, fileType string) (string, model.Vector, error) {
	if err := initSemanticServices(); err != nil {
		return "", nil, err
	}
	return semanticIndexer.BuildIndex(ctx, filePath, originalName, fileType)
}

// embedNaturalLanguage 将自然语言向量化
func embedNaturalLanguage(ctx context.Context, input string) (model.Vector, error) {
	if err := initSemanticServices(); err != nil {
		return nil, err
	}
	return semanticEmbedder.EmbedText(ctx, input)
}

// searchFilesByVector 根据向量检索文件，返回结果列表
func searchFilesByVector(ctx context.Context, userID uint64, queryVector model.Vector, limit int) ([]model.FileSemanticSearchItemResp, error) {
	if err := initSemanticServices(); err != nil {
		return nil, err
	}
	return semanticRepo.SearchByVector(ctx, userID, queryVector, limit)
}

func newQwenEmbeddingProviderFromEnv() (EmbeddingProvider, error) {
	apiKey := strings.TrimSpace(os.Getenv("QWEN_API_KEY"))
	baseURL := strings.TrimSpace(os.Getenv("QWEN_BASE_URL"))
	modelName := strings.TrimSpace(os.Getenv("QWEN_EMBEDDING_MODEL"))

	if apiKey == "" {
		return nil, errors.New("QWEN_API_KEY is empty")
	}
	if baseURL == "" {
		return nil, errors.New("QWEN_BASE_URL is empty")
	}
	if modelName == "" {
		modelName = defaultEmbeddingModel
	}

	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = strings.TrimRight(baseURL, "/")

	return &qwenEmbeddingProvider{
		client: openai.NewClientWithConfig(cfg),
		model:  modelName,
	}, nil
}

func SemanticSearchFiles(ctx context.Context, userID uint64, req model.FileSemanticSearchReq) ([]model.FileSemanticSearchItemResp, error) {
	query := strings.TrimSpace(req.Query)
	if query == "" {
		return nil, secure.Wrap(400, "query 不能为空", errors.New("empty query"))
	}

	limit := req.Limit
	if limit <= 0 {
		limit = defaultSearchLimit
	}
	if limit > maxSearchLimit {
		limit = maxSearchLimit
	}

	queryVector, err := embedNaturalLanguage(ctx, query)
	if err != nil {
		return nil, secure.Wrap(500, "语义向量化失败", err)
	}

	items, err := searchFilesByVector(ctx, userID, queryVector, limit)
	if err != nil {
		return nil, secure.Wrap(500, "根据 vector 寻找文件失败", err)
	}

	for idx := range items {
		distance := items[idx].Score
		items[idx].Score = 1 / (1 + distance)
	}

	return items, nil
}

func isTextLikeFile(filePath, fileType string) bool {
	if strings.HasPrefix(strings.ToLower(fileType), "text/") {
		return true
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".txt", ".md", ".csv", ".log", ".json", ".yaml", ".yml", ".xml":
		return true
	default:
		return false
	}
}

func buildFileMetadataSummary(filePath, originalName, fileType string) string {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(filePath)), ".")
	if ext == "" {
		ext = "unknown"
	}
	if strings.TrimSpace(fileType) == "" {
		fileType = "application/octet-stream"
	}
	return fmt.Sprintf("文件名:%s; 文件类型:%s; 扩展名:%s", strings.TrimSpace(originalName), fileType, ext)
}

func vectorToLiteral(v model.Vector) (string, error) {
	val, err := v.Value()
	if err != nil {
		return "", err
	}
	s, ok := val.(string)
	if !ok {
		return "", errors.New("vector literal conversion failed")
	}
	return s, nil
}
