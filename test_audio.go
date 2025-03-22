// 録音テスト用ファイル
package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gordonklaus/portaudio"
)

func main() {
	// PortAudioの初期化
	portaudio.Initialize()
	defer portaudio.Terminate()

	// 録音パラメータ
	sampleRate := 44100
	bufferSize := sampleRate * 5 // 5秒分
	
	// バッファを作成
	buffer := make([]float32, bufferSize)
	
	// 出力ディレクトリを作成
	outputDir, _ := os.Getwd()
	outputDir = filepath.Join(outputDir, "test_output")
	os.MkdirAll(outputDir, 0755)
	
	// 出力ファイル名
	outputFile := filepath.Join(outputDir, "test_recording.raw")
	
	fmt.Println("録音テストを開始します...")
	fmt.Printf("サンプルレート: %d Hz\n", sampleRate)
	fmt.Printf("バッファサイズ: %d サンプル（%d秒）\n", bufferSize, bufferSize/sampleRate)
	fmt.Printf("出力ファイル: %s\n", outputFile)
	
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
	
	fmt.Println("録音中... 5秒後に停止します")
	
	// 読み取り
	err = stream.Read()
	if err != nil {
		fmt.Printf("録音エラー: %v\n", err)
		return
	}
	
	// 時間を待つ
	time.Sleep(5 * time.Second)
	
	// ストリームを停止
	stream.Stop()
	
	fmt.Println("録音が完了しました。データを保存しています...")
	
	// RAWファイルとして保存
	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("ファイルを作成できませんでした: %v\n", err)
		return
	}
	defer file.Close()
	
	// float32をバイナリとして書き込み
	for _, sample := range buffer {
		// 簡易的な方法でデータを保存（実際のプロジェクトではWAVエンコードが必要）
		binary.Write(file, binary.LittleEndian, sample)
	}
	
	fileInfo, _ := os.Stat(outputFile)
	fmt.Printf("ファイルを保存しました: %s (%d バイト)\n", outputFile, fileInfo.Size())
	
	fmt.Println("テスト完了")
}
