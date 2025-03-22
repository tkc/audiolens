// 連続録音テスト用ファイル
package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/gordonklaus/portaudio"
)

const (
	sampleRate   = 44100        // サンプルレート
	bufferSize   = sampleRate   // 1秒分のバッファ
	recordPeriod = 5 * time.Second // 5秒ごとに保存
)

func main() {
	// PortAudioの初期化
	portaudio.Initialize()
	defer portaudio.Terminate()

	// 出力ディレクトリを作成
	outputDir, _ := os.Getwd()
	outputDir = filepath.Join(outputDir, "test_continuous")
	os.MkdirAll(outputDir, 0755)

	fmt.Println("連続録音テストを開始します...")
	fmt.Printf("サンプルレート: %d Hz\n", sampleRate)
	fmt.Printf("バッファサイズ: %d サンプル（1秒）\n", bufferSize)
	fmt.Printf("保存間隔: %v\n", recordPeriod)
	fmt.Printf("出力ディレクトリ: %s\n", outputDir)
	fmt.Println("Ctrl+C で停止します")

	// バッファを作成
	buffer := make([]float32, bufferSize)

	// デフォルトの入力ストリームを開く
	stream, err := portaudio.OpenDefaultStream(1, 0, float64(sampleRate), bufferSize, buffer)
	if err != nil {
		fmt.Printf("ストリームを開けませんでした: %v\n", err)
		return
	}
	defer stream.Close()

	// ストリームを開始
	err = stream.Start()
	if err != nil {
		fmt.Printf("ストリームを開始できませんでした: %v\n", err)
		return
	}
	defer stream.Stop()

	// シグナル処理
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// 録音タイマー
	ticker := time.NewTicker(recordPeriod)
	defer ticker.Stop()

	// 連続録音ループ
	fileCounter := 1
	for {
		select {
		case <-interrupt:
			fmt.Println("録音を停止します...")
			return
		case <-ticker.C:
			// ここで定期的に保存
			outputFile := filepath.Join(outputDir, fmt.Sprintf("recording_%03d.raw", fileCounter))
			
			// データ読み取り
			err = stream.Read()
			if err != nil {
				fmt.Printf("録音エラー: %v\n", err)
				continue
			}
			
			// データをファイルに保存
			saveBuffer(outputFile, buffer)
			
			fmt.Printf("保存完了: %s\n", outputFile)
			fileCounter++
		default:
			// 定期的に読み取り（ストリームを維持）
			err = stream.Read()
			if err != nil {
				fmt.Printf("エラー: %v\n", err)
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// バッファをファイルに保存
func saveBuffer(filename string, buffer []float32) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("ファイルを作成できませんでした: %v\n", err)
		return
	}
	defer file.Close()

	// バッファデータを書き込み
	for _, sample := range buffer {
		binary.Write(file, binary.LittleEndian, sample)
	}

	fileInfo, _ := os.Stat(filename)
	fmt.Printf("ファイル保存成功: %s (%d バイト)\n", filename, fileInfo.Size())
}
