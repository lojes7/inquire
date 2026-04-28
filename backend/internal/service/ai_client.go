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
