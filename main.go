package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/gordonklaus/portaudio"

	"whisper_local_faster_whsiper_go/internal/app"
	"whisper_local_faster_whsiper_go/internal/transcription"
)

func main() {
	// PortAudioを初期化
	if err := portaudio.Initialize(); err != nil {
		fmt.Printf("PortAudio初期化エラー: %v\n", err)
		os.Exit(1)
	}
	defer portaudio.Terminate()

	// アプリケーションインスタンスを作成
	myApp := app.NewApp()

	// ProcessAudio関数を設定
	myApp.SetProcessAudioFunc(transcription.ProcessAudio)

	myApp.PrintSystemInfo()

	// Whisper.cppが使用可能か確認
	if err := transcription.CheckWhisperAvailability(); err != nil {
		fmt.Printf("\nエラー: %v\n", err)
		os.Exit(1)
	}

	// Ollamaの確認
	if !myApp.CheckOllamaAvailability() {
		fmt.Printf("\nエラー: Ollamaサーバーに接続できません\n")
		fmt.Println("Ollamaを起動し、必要なモデルをダウンロードしてください")
		fmt.Println("詳細: https://ollama.com/")
		os.Exit(1)
	}

	fmt.Printf("\nシステム確認完了:\n")
	fmt.Printf("- 音声文字起こし: whisper.cppを直接使用\n")
	fmt.Printf("- テキスト分析: Ollama\n")
	fmt.Printf("- 使用モデル: %s\n", app.OllamaModel)
	fmt.Printf("- 全てローカル環境で動作します（インターネット不要）\n")

	// 利用可能なデバイス一覧表示
	devices, err := myApp.ListAudioDevices()
	if err != nil {
		fmt.Printf("オーディオデバイス一覧を取得できませんでした: %v\n", err)
		os.Exit(1)
	}

	if len(devices) == 0 {
		fmt.Println("利用可能な入力デバイスがありません")
		os.Exit(1)
	}

	// デバイス選択
	deviceIndex := 0
	if len(devices) > 1 {
		fmt.Print("\nデバイス番号を選択してください (デフォルト:0): ")
		input := ""
		fmt.Scanln(&input)

		if input != "" {
			if idx, err := strconv.Atoi(input); err == nil && idx >= 0 && idx < len(devices) {
				deviceIndex = idx
			} else {
				fmt.Println("入力が無効です。デフォルトのデバイスを使用します。")
			}
		}
	}

	selectedDevice := devices[deviceIndex]
	myApp.DeviceName = selectedDevice.Name

	// マークダウンファイルを初期化
	myApp.InitializeMarkdownFile()

	// コンテキスト作成
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// シグナルハンドラの設定
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 処理ワーカーの開始
	go myApp.ProcessingWorker(ctx)

	// 別のgoroutineでシグナルを待機
	go func() {
		<-sigChan
		fmt.Println("\n録音を停止中...")
		cancel() // コンテキストをキャンセル
	}()

	// 録音開始
	if err := myApp.StartRecording(ctx, &selectedDevice); err != nil {
		fmt.Printf("録音エラー: %v\n", err)
	}

	// 録音終了の処理
	myApp.SaveAudioSegment() // 残りのバッファを保存
	fmt.Println("処理中のファイルを完了中...")
	myApp.WaitForCompletion()
	myApp.AddRecordingEndNote()
	fmt.Println("録音を終了しました")
}
