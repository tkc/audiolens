package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gordonklaus/portaudio"
	
	"whisper_local_faster_whsiper_go/internal/audio"
)

// 定数定義
const (
	SampleRate       = 44100  // サンプリングレート
	RecordingSeconds = 30     // 録音間隔（秒）
	OllamaAPIURL     = "http://localhost:11434/api"
	OllamaModel      = "gemma3:4b"
)

// Appはアプリケーション全体を管理する構造体
type App struct {
	AudioBuffer      [][]float32     // 音声バッファ
	LastSaveTime     time.Time       // 最後の保存時刻
	IsRecording      bool            // 録音中フラグ
	RecordingDir     string          // 録音保存ディレクトリ
	TranscriptsDir   string          // 文字起こし保存ディレクトリ
	TranscribeScript string          // 文字起こしスクリプト
	AllTranscripts   []string        // すべての文字起こし
	PendingFiles     []string        // 処理待ちファイル
	MdFile           string          // マークダウンファイル
	DeviceName       string          // デバイス名
	SampleRate       int             // サンプリングレート
	RecordInterval   float64         // 録音間隔（秒）
	Mutex            sync.Mutex      // ミューテックス
	WG               sync.WaitGroup  // WaitGroup
	ProcessAudioFunc func(*App, string) // 音声処理関数
	animationStopCh  chan struct{}      // アニメーション停止用チャネル
}

// 新しいアプリケーションインスタンスを作成
func NewApp() *App {
	// カレントディレクトリを使用
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("%s\n", ErrorMessage("現在のディレクトリを取得できませんでした: "+err.Error()))
		os.Exit(1)
	}

	dataDir := filepath.Join(currentDir, "data")
	recordingsDir := filepath.Join(dataDir, "recordings")
	transcriptsDir := filepath.Join(dataDir, "transcripts")
	
	// カレントディレクトリからの相対パスでスクリプトを指定
	transcribeScript := filepath.Join(currentDir, "transcribe.sh")

	// ディレクトリの作成
	os.MkdirAll(recordingsDir, 0755)
	os.MkdirAll(transcriptsDir, 0755)

	timestamp := time.Now().Format("20060102_1504")
	mdFile := filepath.Join(transcriptsDir, fmt.Sprintf("%s_all_communication.md", timestamp))

	return &App{
		AudioBuffer:      make([][]float32, 0),
		LastSaveTime:     time.Now(),
		IsRecording:      true,
		RecordingDir:     recordingsDir,
		TranscriptsDir:   transcriptsDir,
		TranscribeScript: transcribeScript,
		AllTranscripts:   make([]string, 0),
		PendingFiles:     make([]string, 0),
		MdFile:           mdFile,
		SampleRate:       SampleRate,
		RecordInterval:   RecordingSeconds,
		ProcessAudioFunc: nil, // 後で設定
		animationStopCh:  make(chan struct{}),
	}
}

// システム情報と設定を表示
func (app *App) PrintSystemInfo() {
	fmt.Print(AsciiArt())
	fmt.Println(SectionHeader("ローカル環境マイク録音と文字起こしアプリケーション"))
	fmt.Println(InfoMessage("whisper.cppとOllamaを使用してAPIに依存せず動作します"))

	// システム情報の確認
	currentDir, _ := os.Getwd()
	homeDir, _ := os.UserHomeDir()
	fmt.Println(SectionHeader("システム情報"))
	fmt.Printf("%sカレントディレクトリ:%s %s\n", Bold, Reset, currentDir)
	fmt.Printf("%sホームディレクトリ:%s %s\n", Bold, Reset, homeDir)

	// システムパス情報
	fmt.Println(SectionHeader("システムパス情報"))
	fmt.Printf("%s- カレントディレクトリ:%s %s\n", Bold, Reset, currentDir)
	fmt.Printf("%s- データディレクトリ:%s %s\n", Bold, Reset, filepath.Dir(app.RecordingDir))
	fmt.Printf("%s- 録音ディレクトリ:%s %s\n", Bold, Reset, app.RecordingDir)
	fmt.Printf("%s- 文字起こしディレクトリ:%s %s\n", Bold, Reset, app.TranscriptsDir)
	fmt.Printf("%s- 文字起こしスクリプト:%s %s\n", Bold, Reset, app.TranscribeScript)
}

// オーディオデータを保存
func (app *App) SaveAudioSegment() {
	app.Mutex.Lock()
	defer app.Mutex.Unlock()

	if len(app.AudioBuffer) == 0 {
		return
	}

	// バッファの詳細情報を集計
	totalSamples := 0
	for _, buffer := range app.AudioBuffer {
		totalSamples += len(buffer)
	}

	// ファイル名と保存先の設定
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("recording_%s.wav", timestamp)
	filepath := filepath.Join(app.RecordingDir, filename)

	// WAV形式でオーディオデータを保存
	err := audio.SaveAsWav(filepath, app.AudioBuffer, app.SampleRate)
	if err != nil {
		fmt.Printf("%s\n", ErrorMessage("録音ファイル保存エラー: "+err.Error()))
		return
	}

	// ファイル存在確認
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		fmt.Printf("%s\n", ErrorMessage("録音ファイルが作成されませんでした: "+filepath))
	} else {
		duration := float64(totalSamples) / float64(app.SampleRate)
		fmt.Printf("\n%s\n", SuccessMessage(fmt.Sprintf("録音保存: %s (%.1f秒)", filepath, duration)))
	}

	// 処理待ちリストに追加
	app.PendingFiles = append(app.PendingFiles, filepath)

	// バッファをクリアして時間をリセット
	app.AudioBuffer = make([][]float32, 0)
	app.LastSaveTime = time.Now()
}

// 録音処理のメインループ - カッコいい表示付き
func (app *App) StartRecording(ctx context.Context, device *portaudio.DeviceInfo) error {
	fmt.Println(SectionHeader("録音開始"))
	fmt.Printf("%s選択デバイス:%s %s%s%s\n", Bold, Reset, Green, device.Name, Reset)
	fmt.Printf("%sサンプルレート:%s %s%d Hz%s\n", Bold, Reset, Yellow, app.SampleRate, Reset)
	fmt.Printf("%s録音間隔:%s %s%.1f秒%s\n", Bold, Reset, Yellow, app.RecordInterval, Reset)
	fmt.Printf("%s出力ファイル:%s %s\n", Bold, Reset, app.MdFile)
	fmt.Println(InfoMessage("Ctrl+C で録音を停止します"))

	// バッファサイズを調整（大きめに設定）
	framesPerBuffer := app.SampleRate / 4     // 0.25秒分
	// 録音間隔の4倍のバッファ数 - 未使用なので削除
	// numBuffers := int(app.RecordInterval * 4)

	// 録音データを格納するスライス
	recordedData := make([]float32, 0, app.SampleRate*int(app.RecordInterval))

	// 実際のストリームバッファ（一時的な録音用）
	buffer := make([]float32, framesPerBuffer)

	// デフォルトの入力ストリームを開く
	stream, err := portaudio.OpenDefaultStream(1, 0, float64(app.SampleRate), framesPerBuffer, buffer)
	if err != nil {
		return fmt.Errorf("ストリームを開けませんでした: %v", err)
	}
	defer stream.Close()

	// ストリームを開始
	err = stream.Start()
	if err != nil {
		return fmt.Errorf("ストリームを開始できませんでした: %v", err)
	}
	defer stream.Stop()

	// 録音中アニメーションを開始
	go RecordingAnimation(app.animationStopCh)
	
	// タイマーを設定
	ticker := time.NewTicker(250 * time.Millisecond) // 0.25秒ごとに読み取り
	defer ticker.Stop()
	
	// 録音の開始時間
	startTime := time.Now()
	
	// オーバーフローカウンター
	overflowCount := 0
	maxOverflows := 10 // 最大許容オーバーフロー回数

	// 録音ループ
	for {
		select {
		case <-ctx.Done():
			// アニメーションを停止
			close(app.animationStopCh)
			
			if len(recordedData) > 0 {
				// 残りのデータを保存
				app.Mutex.Lock()
				app.AudioBuffer = append(app.AudioBuffer, recordedData)
				app.Mutex.Unlock()
				app.SaveAudioSegment()
			}
			
			return nil
			
		case <-ticker.C:
			// データを読み取り
			err = stream.Read()
			
			if err != nil {
				// エラーが発生した場合
				if strings.Contains(err.Error(), "Input overflowed") {
					overflowCount++
					if overflowCount >= maxOverflows {
						// バッファをリセットして続行
						overflowCount = 0
						continue
					}
				} else {
					// アニメーションを一時的に停止して通常表示に戻す
					close(app.animationStopCh)
					app.animationStopCh = make(chan struct{})
					fmt.Printf("\r%s\n", ErrorMessage("読み取りエラー: "+err.Error()))
					go RecordingAnimation(app.animationStopCh) // アニメーション再開
				}
				continue
			}
			
			// バッファからデータをコピー
			bufferCopy := make([]float32, len(buffer))
			copy(bufferCopy, buffer)
			
			// 録音データに追加
			recordedData = append(recordedData, bufferCopy...)
			
			// 経過時間をチェック
			elapsed := time.Since(startTime).Seconds()
			
			if elapsed >= app.RecordInterval {
				// アニメーションを一時的に停止して通常表示に戻す
				close(app.animationStopCh)
				app.animationStopCh = make(chan struct{})
				
				// データをアプリケーションバッファに追加
				app.Mutex.Lock()
				// 複数のチャンクに分割して追加（より効率的な処理のため）
				chunkSize := app.SampleRate // 1秒ごとのチャンク
				numChunks := len(recordedData) / chunkSize
				
				for i := 0; i < numChunks; i++ {
					start := i * chunkSize
					end := start + chunkSize
					if end > len(recordedData) {
						end = len(recordedData)
					}
					
					chunk := make([]float32, end-start)
					copy(chunk, recordedData[start:end])
					app.AudioBuffer = append(app.AudioBuffer, chunk)
				}
				
				// 残りのデータがあれば追加
				if len(recordedData) % chunkSize > 0 {
					start := numChunks * chunkSize
					chunk := make([]float32, len(recordedData)-start)
					copy(chunk, recordedData[start:])
					app.AudioBuffer = append(app.AudioBuffer, chunk)
				}
				
				app.Mutex.Unlock()
				
				// セグメントを保存
				app.SaveAudioSegment()
				
				// 録音データをリセット
				recordedData = make([]float32, 0, app.SampleRate*int(app.RecordInterval))
				
				// 開始時間をリセット
				startTime = time.Now()
				
				// アニメーション再開
				go RecordingAnimation(app.animationStopCh)
			}
		}
	}
}

// 利用可能なオーディオデバイスを表示
func (app *App) ListAudioDevices() ([]portaudio.DeviceInfo, error) {
	devices, err := portaudio.Devices()
	if err != nil {
		return nil, fmt.Errorf("デバイス一覧の取得エラー: %v", err)
	}

	fmt.Println(SectionHeader("利用可能なオーディオデバイス"))
	inputDevices := make([]portaudio.DeviceInfo, 0)

	for i, device := range devices {
		if device.MaxInputChannels > 0 {
			fmt.Printf("  %s%d:%s %s%s%s (入力: %d)\n", 
				Bold+Green, i, Reset, 
				Cyan, device.Name, Reset, 
				device.MaxInputChannels)
			inputDevices = append(inputDevices, *device)
		}
	}

	return inputDevices, nil
}

// 処理ワーカーを実行
func (app *App) ProcessingWorker(ctx context.Context) {
	app.WG.Add(1)
	defer app.WG.Done()

	for {
		select {
		case <-ctx.Done():
			// もし残りのファイルがあれば処理
			for len(app.PendingFiles) > 0 {
				app.Mutex.Lock()
				if len(app.PendingFiles) > 0 {
					filepath := app.PendingFiles[0]
					app.PendingFiles = app.PendingFiles[1:]
					app.Mutex.Unlock()
					app.ProcessAudioFunc(app, filepath)
				} else {
					app.Mutex.Unlock()
					break
				}
			}
			return
		default:
			// 処理待ちファイルがあれば処理
			app.Mutex.Lock()
			if len(app.PendingFiles) > 0 {
				filepath := app.PendingFiles[0]
				app.PendingFiles = app.PendingFiles[1:]
				app.Mutex.Unlock()
				app.ProcessAudioFunc(app, filepath)
			} else {
				app.Mutex.Unlock()
				// 少し待機
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}

// 処理完了を待機
func (app *App) WaitForCompletion() {
	app.WG.Wait()
}

// ProcessAudioFuncを設定
func (app *App) SetProcessAudioFunc(fn func(*App, string)) {
	app.ProcessAudioFunc = fn
}
