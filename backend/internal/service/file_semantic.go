package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

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
