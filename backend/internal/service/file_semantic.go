package service

import (
	"fmt"
	"log"
	"strings"

	"github.com/lojes7/inquire/internal/model"
	"github.com/lojes7/inquire/pkg/infra"
	"github.com/lojes7/inquire/pkg/secure"
)

const (
	// defaultSearchLimit 语义搜索默认返回数量
	defaultSearchLimit = 5
	// defaultSearchThreshold 默认余弦距离阈值（小于此值才返回）
	defaultSearchThreshold = 0.5
)

// SearchFiles 对用户工作区文件执行语义搜索。
//
// query: 用户输入的自然语言查询
// limit: 最多返回数量（默认 5，最大 20）
// threshold: 余弦距离上限，用 pgvector 的 <=> 运算符（默认 0.5）
func SearchFiles(userID uint64, query string, limit int, threshold float64) ([]model.SemanticSearchFileItem, error) {
	if strings.TrimSpace(query) == "" {
		return nil, secure.Wrap(400, "查询内容不能为空", fmt.Errorf("empty query"))
	}

	// 1) 调用 AI 服务获取查询文本的嵌入向量
	queryVec, err := callAIAsk(query)
	if err != nil {
		log.Printf("[search] AI ask failed: %v", err)
		return nil, secure.Wrap(500, "生成查询向量失败", err)
	}

	// 2) 参数默认值
	if limit <= 0 || limit > 20 {
		limit = defaultSearchLimit
	}
	if threshold <= 0 || threshold > 2 {
		threshold = defaultSearchThreshold
	}

	// 3) 构建 pgvector 格式的向量字符串
	vecStr := formatVector(queryVec)

	// 4) pgvector 余弦距离搜索，仅限用户可访问的工作区文件
	db := infra.GetDB()

	type rawRow struct {
		ID       uint64  `gorm:"column:id"`
		FileName string  `gorm:"column:file_name"`
		FileSize int64   `gorm:"column:file_size"`
		FileType string  `gorm:"column:file_type"`
		Distance float64 `gorm:"column:distance"`
	}

	sql := `
		SELECT f.id, f.file_name, f.file_size, f.file_type,
		       MIN(fv.vector <=> ?::vector) AS distance
		FROM file_vectors fv
		JOIN files f ON f.id = fv.file_id AND f.deleted_at IS NULL
		JOIN user_files uf ON uf.file_id = f.id
		     AND uf.user_id = ? AND uf.deleted_at IS NULL
		WHERE fv.deleted_at IS NULL
		  AND fv.vector <=> ?::vector < ?
		GROUP BY f.id, f.file_name, f.file_size, f.file_type
		ORDER BY distance ASC
		LIMIT ?
	`

	var rows []rawRow
	if err := db.Raw(sql, vecStr, userID, vecStr, threshold, limit).Scan(&rows).Error; err != nil {
		log.Printf("[search] pgvector query failed: %v", err)
		return nil, secure.Wrap(500, "向量搜索失败", err)
	}

	// 5) 构建响应
	results := make([]model.SemanticSearchFileItem, 0, len(rows))
	for _, row := range rows {
		results = append(results, model.SemanticSearchFileItem{
			FileID:   row.ID,
			FileName: row.FileName,
			FileSize: row.FileSize,
			FileType: row.FileType,
			Score:    row.Distance,
		})
	}

	return results, nil
}

// formatVector 将 []float32 转为 pgvector 所需的 [x,y,z] 格式字符串。
func formatVector(vec []float32) string {
	parts := make([]string, len(vec))
	for i, v := range vec {
		parts[i] = fmt.Sprintf("%f", v)
	}
	return "[" + strings.Join(parts, ",") + "]"
}
