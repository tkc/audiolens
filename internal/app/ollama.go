package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Ollama API用の構造体
type OllamaRequest struct {
	Model   string `json:"model"`
	Prompt  string `json:"prompt"`
	System  string `json:"system"`
	Stream  bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

// Ollamaの利用可能性チェック
func (app *App) CheckOllamaAvailability() bool {
	// サーバー設定を確認
	resp, err := http.Get(fmt.Sprintf("%s/tags", OllamaAPIURL))
	if err != nil {
		fmt.Printf("Ollama接続エラー: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	// レスポンスを解析
	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Ollamaレスポンス解析エラー: %v\n", err)
		return false
	}

	// モデルが存在するか確認
	modelExists := false
	for _, model := range result.Models {
		if model.Name == OllamaModel {
			modelExists = true
			break
		}
	}

	if !modelExists {
		fmt.Printf("\n注意: %s モデルが見つかりません。\n", OllamaModel)
		fmt.Printf("   次のコマンドでインストールしてください: ollama pull %s\n", OllamaModel)
		return false
	}

	return true
}

// Ollamaローカルモデルに問い合わせ
func QueryOllama(prompt string, systemPrompt string) (string, error) {
	url := fmt.Sprintf("%s/generate", OllamaAPIURL)
	
	// リクエストボディを構築
	reqBody := OllamaRequest{
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
	
	var result OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("APIレスポンス解析エラー: %v", err)
	}
	
	return result.Response, nil
}
