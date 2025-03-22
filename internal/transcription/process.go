package transcription

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"whisper_local_faster_whsiper_go/internal/analysis"
	"whisper_local_faster_whsiper_go/internal/app"
)

// 音声ファイルを文字起こしして分析
func ProcessAudio(application *app.App, audioPath string) {
	fmt.Print(app.SectionHeader("文字起こし処理開始"))
	fmt.Printf("%s対象音声ファイル:%s %s%s%s\n", app.Bold, app.Reset, app.Cyan, audioPath, app.Reset)

	// 文字起こし
	transcriptText, err := WhisperTranscribe(audioPath)
	if err != nil {
		fmt.Printf("%s\n", app.ErrorMessage("文字起こし失敗: "+err.Error()))
		return
	}

	// 文字起こし結果を表示
	fmt.Println(app.AnalysisHeader("文字起こし結果 (" + fmt.Sprintf("%d", len(transcriptText)) + "文字)"))
	fmt.Println(app.TextBox(transcriptText, "文字起こし"))

	// 文字起こし結果を全体のリストに追加
	application.Mutex.Lock()
	application.AllTranscripts = append(application.AllTranscripts, transcriptText)
	combinedText := strings.Join(application.AllTranscripts, " ")
	application.Mutex.Unlock()

	// プログレスバー表示用のカウンター
	totalTasks := 5 // タスク数を5に増やす（攻撃的言葉チェックを追加）
	completedTasks := 0

	fmt.Println(app.SectionHeader("テキスト分析"))

	// 一連の分析を並行処理
	var wg sync.WaitGroup
	var summary, issues, progressScore, aggressiveCheck string
	var keywords []string

	// 実際のプログレスバー作成
	bar := app.CreateProgressBar(totalTasks, "分析しています")

	// アニメーションチャネル
	animDone := make(chan struct{})

	// プログレスバー更新スレッド
	go func() {
		for {
			select {
			case <-animDone:
				// 完了時は100%に設定
				_ = bar.Add(totalTasks - completedTasks) // 残りを一気に追加して完了させる
				return
			default:
				// 現在の大きさを取得
				current := int(bar.State().CurrentBytes)
				if completedTasks > current {
					// 差分を追加
					_ = bar.Add(completedTasks - current)
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// 要約生成
	wg.Add(1)
	go func() {
		defer wg.Done()
		result, err := analysis.GenerateSummary(combinedText)
		if err != nil {
			fmt.Printf("\r%s\n", app.ErrorMessage("要約生成エラー: "+err.Error()))
		} else {
			summary = result
		}
		completedTasks++
	}()

	// キーワード抽出
	wg.Add(1)
	go func() {
		defer wg.Done()
		result, err := analysis.ExtractKeywords(combinedText)
		if err != nil {
			fmt.Printf("\r%s\n", app.ErrorMessage("キーワード抽出エラー: "+err.Error()))
		} else {
			keywords = result
		}
		completedTasks++
	}()

	// 問題点抽出
	wg.Add(1)
	go func() {
		defer wg.Done()
		result, err := analysis.IdentifyIssues(combinedText)
		if err != nil {
			fmt.Printf("\r%s\n", app.ErrorMessage("問題点抽出エラー: "+err.Error()))
		} else {
			issues = result
		}
		completedTasks++
	}()

	// 進行状況評価
	wg.Add(1)
	go func() {
		defer wg.Done()
		result, err := analysis.EvaluateProgress(combinedText)
		if err != nil {
			fmt.Printf("\r%s\n", app.ErrorMessage("進行状況評価エラー: "+err.Error()))
		} else {
			progressScore = result
		}
		completedTasks++
	}()

	// 攻撃的言葉チェック
	wg.Add(1)
	go func() {
		defer wg.Done()
		result, err := analysis.CheckAggressiveLanguage(combinedText)
		if err != nil {
			fmt.Printf("\r%s\n", app.ErrorMessage("攻撃的言葉チェックエラー: "+err.Error()))
		} else {
			aggressiveCheck = result
		}
		completedTasks++
	}()

	// すべての分析が完了するまで待機
	wg.Wait()
	close(animDone) // アニメーション停止

	// 完了表示
	fmt.Println() // 改行を入れて表示を整える
	fmt.Printf("%s\n", app.SuccessMessage("分析完了！"))

	// 分析結果を表示
	fmt.Println(app.SectionHeader("分析結果"))

	if summary != "" {
		fmt.Println(app.AnalysisHeader("要約"))
		fmt.Println(app.TextBox(summary, ""))
		fmt.Println()
	}

	if len(keywords) > 0 {
		fmt.Println(app.AnalysisHeader("キーワード"))
		fmt.Println(app.TextBox(strings.Join(keywords, ", "), ""))
		fmt.Println()
	}

	if issues != "" {
		fmt.Println(app.AnalysisHeader("問題点"))
		fmt.Println(app.TextBox(issues, ""))
		fmt.Println()
	}

	if progressScore != "" {
		fmt.Println(app.AnalysisHeader("進行状況評価"))
		fmt.Println(app.TextBox(progressScore, ""))
		fmt.Println()
	}

	if aggressiveCheck != "" {
		fmt.Println(app.AnalysisHeader("攻撃的言葉チェック"))
		fmt.Println(app.TextBox(aggressiveCheck, ""))
		fmt.Println()
	}

	// マークダウンに保存
	saveMarkdown(application, transcriptText, combinedText, summary, keywords, issues, progressScore, aggressiveCheck)

	fmt.Printf("%s\n", app.SuccessMessage("文字起こしと分析が完了しました"))
	fmt.Printf("%s結果は以下に保存されました:%s %s\n", app.Bold, app.Reset, application.MdFile)
	fmt.Println(app.SectionHeader("処理完了"))
}

// マークダウンに保存
func saveMarkdown(application *app.App, currentTranscript, combinedText, summary string, keywords []string, issues, progressScore, aggressiveCheck string) {
	// ファイルが存在しない場合は初期化
	if _, err := os.Stat(application.MdFile); os.IsNotExist(err) {
		application.InitializeMarkdownFile()
	}

	// 追記用の内容
	timestampDisplay := time.Now().Format("2006-01-02 15:04:05")

	// 追記内容の作成
	var content strings.Builder
	content.WriteString(fmt.Sprintf("\n## %s\n\n%s\n\n", timestampDisplay, currentTranscript))
	content.WriteString(fmt.Sprintf("### 全体テキスト\n\n%s\n\n", combinedText))

	if summary != "" {
		content.WriteString(fmt.Sprintf("### 全体要約\n\n%s\n\n", summary))
	}

	if progressScore != "" {
		content.WriteString(fmt.Sprintf("### 議論の進行状況評価\n\n%s\n\n", progressScore))
	}

	if len(keywords) > 0 {
		content.WriteString(fmt.Sprintf("### 全体キーワード\n\n%s\n\n", strings.Join(keywords, ", ")))
	}

	if issues != "" {
		content.WriteString(fmt.Sprintf("### 全体問題点\n\n%s\n\n", issues))
	}

	if aggressiveCheck != "" {
		content.WriteString(fmt.Sprintf("### 攻撃的言葉チェック\n\n%s\n\n", aggressiveCheck))
	}

	content.WriteString("---\n")

	// 追記
	file, err := os.OpenFile(application.MdFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("%s\n", app.ErrorMessage("マークダウンファイルを開けませんでした: "+err.Error()))
		return
	}
	defer file.Close()

	if _, err = file.WriteString(content.String()); err != nil {
		fmt.Printf("%s\n", app.ErrorMessage("マークダウンファイルへの書き込みエラー: "+err.Error()))
		return
	}
}
