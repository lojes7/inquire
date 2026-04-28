package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// aiServiceURL 默认值，可通过环境变量 AI_SERVICE_URL 覆盖
func getAIServiceURL() string {
	if url := os.Getenv("AI_SERVICE_URL"); url != "" {
		return url
	}
	return "http://localhost:8001"
}

// embedMaxRetry AI 服务调用的最大重试次数
const embedMaxRetry = 2

// embedRequestTimeout AI 服务的请求超时时间
const embedRequestTimeout = 120 * time.Second

// EmbedReq 发送给 AI 服务的嵌入请求
type EmbedReq struct {
	FilePath string `json:"file_path"`
	FileID   uint64 `json:"file_id"`
	FileType string `json:"file_type"`
}

// EmbedVectorItem AI 服务返回的单个向量块
type EmbedVectorItem struct {
	Number int       `json:"number"`
	Vector []float32 `json:"vector"`
}

// EmbedResp AI 服务返回的嵌入响应
type EmbedResp struct {
	Status     string            `json:"status"`
	FileID     uint64            `json:"file_id"`
	Vectors    []EmbedVectorItem `json:"vectors"`
	ChunkCount int               `json:"chunk_count"`
	Answer     string            `json:"answer"` // 错误时的描述信息
}

var embedHTTPClient = &http.Client{
	Timeout: embedRequestTimeout,
}

// callAIEmbed 调用 AI 服务的 /embed 端点，返回向量列表。
// 支持自动重试。
func callAIEmbed(filePath string, fileID uint64, fileType string) (*EmbedResp, error) {
	reqBody := EmbedReq{
		FilePath: filePath,
		FileID:   fileID,
		FileType: fileType,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal embed request: %w", err)
	}

	baseURL := getAIServiceURL()
	endpoint := baseURL + "/embed"
	log.Printf("[ai-client] Calling %s for file_id=%d path=%s type=%s", endpoint, fileID, filePath, fileType)

	var lastErr error
	for i := 0; i <= embedMaxRetry; i++ {
		if i > 0 {
			log.Printf("[ai-client] Retry %d/%d for file_id=%d", i, embedMaxRetry, fileID)
			time.Sleep(time.Duration(i) * time.Second)
		}

		resp, err := callEmbedOnce(endpoint, payload)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		log.Printf("[ai-client] call failed for file_id=%d (attempt %d/%d): %v",
			fileID, i+1, embedMaxRetry+1, err)
	}

	return nil, fmt.Errorf("callAIEmbed failed after %d retries: %w", embedMaxRetry+1, lastErr)
}

func callEmbedOnce(endpoint string, payload []byte) (*EmbedResp, error) {
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := embedHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("[ai-client] HTTP %d from AI service, body: %s", resp.StatusCode, string(body))
	}

	var embedResp EmbedResp
	if err := json.Unmarshal(body, &embedResp); err != nil {
		return nil, fmt.Errorf("unmarshal response (HTTP %d): %w (body=%s)", resp.StatusCode, err, string(body))
	}

	if resp.StatusCode != http.StatusOK || embedResp.Status != "success" {
		errMsg := embedResp.Answer
		if errMsg == "" {
			errMsg = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("AI service error (HTTP %d): %s", resp.StatusCode, errMsg)
	}

	if len(embedResp.Vectors) == 0 {
		return nil, fmt.Errorf("AI service returned empty vectors for file_id=%d", embedResp.FileID)
	}

	return &embedResp, nil
}

// askReq 发送给 AI 服务 /ask 的文本查询请求
type askReq struct {
	InputData []map[string]string `json:"input_data"`
}

// callAIAsk 调用 AI 服务的 /ask 端点，获取查询文本的嵌入向量。
func callAIAsk(query string) ([]float32, error) {
	reqBody := askReq{
		InputData: []map[string]string{{"text": query}},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal ask request: %w", err)
	}

	endpoint := getAIServiceURL() + "/ask"
	log.Printf("[ai-client] Calling /ask, query_len=%d", len(query))

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create ask request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	httpResp, err := embedHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http ask request: %w", err)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read ask response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		log.Printf("[ai-client] /ask HTTP %d, body: %s", httpResp.StatusCode, string(body))
		return nil, fmt.Errorf("AI /ask returned HTTP %d", httpResp.StatusCode)
	}

	// 解析 /ask 响应结构: {"status":"success","answer":{"output":{"embeddings":[{"embedding":[...]}]}}}
	var raw struct {
		Status string          `json:"status"`
		Answer json.RawMessage `json:"answer"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse ask response: %w", err)
	}
	if raw.Status != "success" {
		return nil, fmt.Errorf("AI /ask error: %s", string(raw.Answer))
	}

	var answer struct {
		Output map[string]any `json:"output"`
	}
	if err := json.Unmarshal(raw.Answer, &answer); err != nil {
		return nil, fmt.Errorf("parse ask answer: %w", err)
	}

	// 从 output 中提取向量，结构与 _extract_embedding 一致：
	// output["embeddings"] → [] → [0]["embedding"] → []float32
	vec, err := extractEmbeddingFromOutput(answer.Output)
	if err != nil {
		return nil, err
	}
	if len(vec) != 1024 {
		return nil, fmt.Errorf("query embedding dimension mismatch: got %d, expected 1024", len(vec))
	}

	return vec, nil
}

// extractEmbeddingFromOutput 从 AI 服务返回的 output dict 中提取向量。
// 与 Python 端 _extract_embedding 对应，处理 {"embeddings": [{"embedding": [...]}]} 结构。
func extractEmbeddingFromOutput(output map[string]any) ([]float32, error) {
	// 尝试 "embeddings" key（DashScope 多模态融合返回）
	if raw, ok := output["embeddings"]; ok {
		if arr, ok := raw.([]any); ok && len(arr) > 0 {
			if first, ok := arr[0].(map[string]any); ok {
				if embRaw, ok := first["embedding"]; ok {
					return floatSliceFromInterface(embRaw)
				}
			}
			// 备份: 如果第一个元素是向量列表而非 dict
			return floatSliceFromInterface(arr[0])
		}
	}

	// 尝试 "embedding" key
	if raw, ok := output["embedding"]; ok {
		return floatSliceFromInterface(raw)
	}

	return nil, fmt.Errorf("no embedding found in output")
}

// floatSliceFromInterface 把 []any 或 []float64 转为 []float32。
func floatSliceFromInterface(raw any) ([]float32, error) {
	if arr, ok := raw.([]any); ok {
		vec := make([]float32, len(arr))
		for i, v := range arr {
			switch val := v.(type) {
			case float64:
				vec[i] = float32(val)
			case float32:
				vec[i] = val
			default:
				return nil, fmt.Errorf("unexpected embedding element type %T at index %d", v, i)
			}
		}
		return vec, nil
	}
	if arr, ok := raw.([]float64); ok {
		vec := make([]float32, len(arr))
		for i, v := range arr {
			vec[i] = float32(v)
		}
		return vec, nil
	}
	return nil, fmt.Errorf("embedding value is not a slice, got %T", raw)
}
