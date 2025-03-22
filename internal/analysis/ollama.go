package analysis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	OllamaAPIURL = "http://localhost:11434/api"
	OllamaModel  = "gemma3:4b"
)

// Ollama API用の構造体
type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	System string `json:"system"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
}

// Ollamaローカルモデルに問い合わせ（内部用）
func queryOllama(prompt string, systemPrompt string) (string, error) {
	url := fmt.Sprintf("%s/generate", OllamaAPIURL)

	// リクエストボディを構築
	reqBody := ollamaRequest{
		Model:  OllamaModel,
		Prompt: prompt,
		System: systemPrompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("JSONエンコードエラー: %v", err)
	}

	// HTTPリクエストを作成
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("HTTPリクエスト作成エラー: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// タイムアウト付きのHTTPクライアントを作成
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	// リクエストを送信
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("APIリクエストエラー: %v", err)
	}
	defer resp.Body.Close()

	// レスポンスを解析
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("APIエラー: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var result ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("APIレスポンス解析エラー: %v", err)
	}

	return result.Response, nil
}
