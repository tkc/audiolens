package transcription

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func TranscribeWithShellScript(transcribeScript string, audioPath string) (string, error) {
	// 絶対パスに変換
	audioAbsPath, err := filepath.Abs(audioPath)
	if err != nil {
		return "", fmt.Errorf("絶対パスの取得に失敗: %v", err)
	}

	// オーディオファイルの存在チェック
	if _, err := os.Stat(audioAbsPath); os.IsNotExist(err) {
		return "", fmt.Errorf("オーディオファイルが見つかりません: %s", audioAbsPath)
	}

	// シェルスクリプトの実行権限確認と付与
	info, err := os.Stat(transcribeScript)
	if err != nil {
		return "", fmt.Errorf("文字起こしスクリプトがありません: %v", err)
	}

	// 実行権限を確認し、必要なら追加
	if info.Mode()&0111 == 0 {
		err = os.Chmod(transcribeScript, 0755)
		if err != nil {
			return "", fmt.Errorf("文字起こしスクリプトに実行権限を付与できませんでした: %v", err)
		}
	}

	// シェルスクリプトを実行
	cmd := exec.Command(transcribeScript, audioAbsPath)
	fmt.Printf("  文字起こしコマンド: %s %s\n", transcribeScript, audioAbsPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("文字起こし実行エラー: %v\n出力: %s", err, string(output))
	}

	fmt.Printf("  出力: %s\n", string(output))

	// 出力テキストファイルを読み取る
	txtPath := audioAbsPath + ".txt"
	if _, err := os.Stat(txtPath); err == nil {
		fmt.Printf("  文字起こしファイル読み込み: %s\n", txtPath)
		content, err := os.ReadFile(txtPath)
		if err != nil {
			return "", fmt.Errorf("文字起こしファイル読み込みエラー: %v", err)
		}
		return strings.TrimSpace(string(content)), nil
	} else {
		fmt.Printf("  文字起こしファイルが見つかりません: %s\n", txtPath)

		// ディレクトリ内のテキストファイルを探す
		dirPath := filepath.Dir(audioAbsPath)
		files, err := os.ReadDir(dirPath)
		if err == nil {
			var txtFiles []string
			for _, file := range files {
				if strings.HasSuffix(file.Name(), ".txt") {
					txtFiles = append(txtFiles, file.Name())
				}
			}
			if len(txtFiles) > 0 {
				fmt.Printf("  ディレクトリ内のテキストファイル: %v\n", txtFiles)
			}
		}
		return "", fmt.Errorf("文字起こし結果が見つかりません")
	}
}
