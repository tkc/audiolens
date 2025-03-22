package transcription

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// WhisperPath定数
const (
	WhisperPath = "/Users/takeshiiijima/github/whisper.cpp/build/bin/whisper-cli"
	ModelPath = "/Users/takeshiiijima/github/whisper.cpp/models/ggml-base.bin"
)

// 正規表現パターン
var outputTxtRegex = regexp.MustCompile(`output_txt: saving output to ['"](.*?)['"]`)

// CheckWhisperAvailabilityはWhisper.cppが使用可能か確認する
func CheckWhisperAvailability() error {
	// whisper-cliが存在するか確認
	if _, err := os.Stat(WhisperPath); os.IsNotExist(err) {
		return fmt.Errorf("whisper-cliが見つかりません: %s", WhisperPath)
	}

	// モデルファイルが存在するか確認
	if _, err := os.Stat(ModelPath); os.IsNotExist(err) {
		return fmt.Errorf("モデルファイルが見つかりません: %s", ModelPath)
	}

	// 実行権限の確認
	info, _ := os.Stat(WhisperPath)
	if info.Mode()&0111 == 0 {
		return fmt.Errorf("whisper-cliに実行権限がありません: %s", WhisperPath)
	}

	return nil
}

// 最新のテキストファイルを見つける
func findLatestTxtFile(dirPath string) string {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return ""
	}

	var latestFile string
	var latestTime time.Time

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".txt") {
			filePath := filepath.Join(dirPath, file.Name())
			fileInfo, err := file.Info()
			if err != nil {
				continue
			}

			// ファイルの更新時間を取得
			modTime := fileInfo.ModTime()

			// 最新のファイルを追跡
			if latestFile == "" || modTime.After(latestTime) {
				latestFile = filePath
				latestTime = modTime
			}
		}
	}

	if latestFile != "" {
		fmt.Printf("  最新のテキストファイルを使用: %s (%s)\n", 
			filepath.Base(latestFile), latestTime.Format("2006-01-02 15:04:05"))
	}

	return latestFile
}

// WhisperTranscribe は音声ファイルをWhisper.cppを使用して文字起こしします
func WhisperTranscribe(audioPath string) (string, error) {
	// 絶対パスに変換
	audioAbsPath, err := filepath.Abs(audioPath)
	if err != nil {
		return "", fmt.Errorf("絶対パスの取得に失敗: %v", err)
	}

	// オーディオファイルの存在チェック
	if _, err := os.Stat(audioAbsPath); os.IsNotExist(err) {
		return "", fmt.Errorf("オーディオファイルが見つかりません: %s", audioAbsPath)
	}

	fmt.Printf("  文字起こし中: %s...\n", filepath.Base(audioAbsPath))

	// whisper.cppを実行するコマンドを構築
	cmd := exec.Command(
		WhisperPath,
		"-m", ModelPath,
		"-f", audioAbsPath,
		"-l", "ja",
		"--output-txt",
		"--no-gpu",
	)

	// コマンドを実行
	var output []byte
	var cmdErr error
	output, cmdErr = cmd.CombinedOutput()
	if cmdErr != nil {
		return "", fmt.Errorf("文字起こし実行エラー: %v", cmdErr)
	}
	
	// コマンドは成功したが、デバッグのために出力を保存
	debugOutput := string(output)
	
	// 出力されたテキストファイルを読み込み
	// まず、標準のパスをチェック
	txtPath := audioAbsPath + ".txt"
	
	// デバッグ出力からファイルパスを探す
	var parsedPath string
	if matched := outputTxtRegex.FindStringSubmatch(debugOutput); len(matched) > 1 {
		parsedPath = matched[1]
		fmt.Printf("  出力からファイルパスを探索: %s\n", parsedPath)
		// ファイルが存在するか確認
		if _, err := os.Stat(parsedPath); err == nil {
			txtPath = parsedPath
		}
	}
	
	// ファイルがディレクトリに存在するか確認
	if _, err := os.Stat(txtPath); os.IsNotExist(err) {
		// 最新の.txtファイルを探す
		txtPath = findLatestTxtFile(filepath.Dir(audioAbsPath))
		if txtPath == "" {
			fmt.Printf("  文字起こしファイルが見つかりません\n")
			// デバッグ情報を表示
			fmt.Printf("  コマンド出力: %s\n", debugOutput)
			return "", fmt.Errorf("文字起こし結果が見つかりません")
		}
	}

	// テキストファイルの内容を読み込む
	content, err := os.ReadFile(txtPath)
	if err != nil {
		return "", fmt.Errorf("文字起こしファイル読み込みエラー: %v", err)
	}

	// 成功メッセージ
	fmt.Printf("  文字起こし完了: %s (%d文字)\n", filepath.Base(txtPath), len(content))

	return strings.TrimSpace(string(content)), nil
}
