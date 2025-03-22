package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// マークダウンファイルを初期化
func (app *App) InitializeMarkdownFile() {
	f, err := os.Create(app.MdFile)
	if err != nil {
		fmt.Printf("エラー: マークダウンファイルを作成できませんでした: %v\n", err)
		return
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	writer.WriteString("# 会話記録\n\n")
	writer.WriteString("録音された会話の文字起こし、要約、キーワード、問題点を記録します。\n\n")
	writer.WriteString("※ ローカル環境のwhisper.cppとOllamaを使用しています\n\n")

	startTime := time.Now().Format("2006-01-02 15:04:05")
	writer.WriteString(fmt.Sprintf("**録音開始時刻**: %s\n\n", startTime))

	if app.DeviceName != "" {
		writer.WriteString(fmt.Sprintf("**録音デバイス**: %s\n\n", app.DeviceName))
	}

	writer.WriteString(fmt.Sprintf("**サンプリングレート**: %d Hz\n\n", app.SampleRate))
	writer.WriteString(fmt.Sprintf("**録音間隔**: %.1f 秒\n\n", app.RecordInterval))
	writer.WriteString(fmt.Sprintf("**Ollamaモデル**: %s\n\n", OllamaModel))
	writer.WriteString("---\n\n")
	writer.Flush()

	fmt.Printf("  ファイル初期化: %s\n", app.MdFile)
}

// 録音終了の記録を追加
func (app *App) AddRecordingEndNote() {
	// 終了時刻
	endTime := time.Now().Format("2006-01-02 15:04:05")

	file, err := os.OpenFile(app.MdFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("  マークダウンファイルを開けませんでした: %v\n", err)
		return
	}
	defer file.Close()

	var content strings.Builder
	content.WriteString(fmt.Sprintf("\n## 録音終了: %s\n\n", endTime))
	content.WriteString("### 録音セッション統計\n\n")
	content.WriteString(fmt.Sprintf("- 総セグメント数: %d\n", len(app.AllTranscripts)))

	totalChars := 0
	for _, t := range app.AllTranscripts {
		totalChars += len(t)
	}
	content.WriteString(fmt.Sprintf("- 合計文字数: %d\n", totalChars))
	content.WriteString("\n---\n\n")

	if _, err = file.WriteString(content.String()); err != nil {
		fmt.Printf("  マークダウンファイルへの書き込みエラー: %v\n", err)
		return
	}

	fmt.Printf("  録音終了記録: %s\n", app.MdFile)
}
