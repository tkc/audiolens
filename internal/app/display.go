package app

import (
	"fmt"
	"strings"
	"time"
	
	"github.com/schollz/progressbar/v3"
)

// ANSI エスケープコード
const (
	Reset      = "\033[0m"
	Bold       = "\033[1m"
	Italic     = "\033[3m"
	Underline  = "\033[4m"
	
	// 文字色
	Black     = "\033[30m"
	Red       = "\033[31m"
	Green     = "\033[32m"
	Yellow    = "\033[33m"
	Blue      = "\033[34m"
	Purple    = "\033[35m"
	Cyan      = "\033[36m"
	White     = "\033[37m"
	
	// 背景色
	BlackBg   = "\033[40m"
	RedBg     = "\033[41m"
	GreenBg   = "\033[42m"
	YellowBg  = "\033[43m"
	BlueBg    = "\033[44m"
	PurpleBg  = "\033[45m"
	CyanBg    = "\033[46m"
	WhiteBg   = "\033[47m"
)

// 録音中のアニメーション
func RecordingAnimation(stopCh <-chan struct{}) {
	spinner := "*"
	for {
		select {
		case <-stopCh:
			fmt.Print("\r" + strings.Repeat(" ", 50) + "\r") // クリア
			return
		default:
			// 録音中の表示
			fmt.Printf("\r%s %s録音中...%s ", spinner, Green, Reset)
			
			// 経過時間などを追加
			elapsed := time.Now().Format("15:04:05")
			fmt.Printf("%s[%s]%s", Yellow, elapsed, Reset)
			
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// プログレスバーを生成
func CreateProgressBar(total int, description string) *progressbar.ProgressBar {
	return progressbar.NewOptions(total,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(false),
		progressbar.OptionSetWidth(30),
		progressbar.OptionSetDescription(fmt.Sprintf("%s[%s%s%s]%s", Blue, Reset, description, Blue, Reset)),
		progressbar.OptionShowCount(),
		progressbar.OptionSetRenderBlankState(true),
	)
}

// プログレスバーをレンダリング
func RenderProgressBar(percent int, width int) string {
	// 古いスタイルのプログレスバーをシミュレート
	bar := strings.Builder{}
	bar.WriteString("[")
	
	completed := width * percent / 100
	for i := 0; i < width; i++ {
		if i < completed {
			bar.WriteString(Green + "=" + Reset)
		} else if i == completed {
			bar.WriteString(Yellow + ">" + Reset)
		} else {
			bar.WriteString("-")
		}
	}
	
	bar.WriteString("] ")
	bar.WriteString(fmt.Sprintf("%3d%%", percent))
	return bar.String()
}

// プログレスバーの互換性用インターフェース
func ProgressBar(percent int, width int) string {
	// 互換性のために新しい実装を呼び出す
	return RenderProgressBar(percent, width)
}

// セクションヘッダー
func SectionHeader(title string) string {
	return fmt.Sprintf("\n%s%s=== %s ===%s\n", Bold, Cyan, title, Reset)
}

// 成功メッセージ
func SuccessMessage(message string) string {
	return fmt.Sprintf("%s%s✓ %s%s", Bold, Green, message, Reset)
}

// エラーメッセージ
func ErrorMessage(message string) string {
	return fmt.Sprintf("%s%s✗ %s%s", Bold, Red, message, Reset)
}

// 警告メッセージ
func WarningMessage(message string) string {
	return fmt.Sprintf("%s%s⚠ %s%s", Bold, Yellow, message, Reset)
}

// 情報メッセージ
func InfoMessage(message string) string {
	return fmt.Sprintf("%s%sℹ %s%s", Bold, Blue, message, Reset)
}

// 分析結果のヘッダー
func AnalysisHeader(title string) string {
	// シンプルな横線のみのスタイル
	return fmt.Sprintf("\n%s%s%s\n%s", 
		Blue, title, Reset,
		Blue + strings.Repeat("─", 80) + Reset)
}

// テキストをボックスで囲む
func TextBox(text string, title string) string {
	// シンプルな横線のみのスタイル
	return text + "\n" + Blue + strings.Repeat("─", 80) + Reset
}

// アスキーアート
func AsciiArt() string {
	art := `
   __        ___     _                       
   \ \      / / |__ (_)___ _ __   ___ _ __  
    \ \ /\ / /| '_ \| / __| '_ \ / _ \ '__| 
     \ V  V / | | | | \__ \ |_) |  __/ |    
      \_/\_/  |_| |_|_|___/ .__/ \___|_|    
                          |_|                
  `
	
	return Blue + Bold + art + Reset
}
