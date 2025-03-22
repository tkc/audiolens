# audiolens

このアプリケーションは、マイクからの音声をリアルタイムで録音し、whisper.cpp を使用して文字起こしを行います。さらに、Ollama を使用してテキスト分析（要約、キーワード抽出、問題点検出、進行状況評価）を行います。すべての処理はローカル環境で完結し、インターネット接続は不要です。

## 前提条件

- Go 1.16 以上
- [PortAudio](http://www.portaudio.com/)
- [whisper.cpp](https://github.com/ggerganov/whisper.cpp)
- [Ollama](https://ollama.ai/)

## セットアップ

### 依存関係のインストール

#### macOS

```bash
# Homebrewを使用してPortAudioをインストール
brew install portaudio

# whisper.cppをインストール
brew install whisper-cpp

# Ollamaをインストール
brew install ollama
```

#### Ollama モデルのダウンロード

```bash
# アプリケーションで使用するGemma 3モデルをダウンロード
ollama pull gemma3:4b
```

### アプリケーションのビルドとインストール

```bash
# プロジェクトディレクトリに移動
cd whisper_local_faster_whsiper_go

# 依存関係をインストール
go mod download

# アプリケーションをビルド
go build -o bin/whisper_recorder

# 実行権限を付与
chmod +x bin/whisper_recorder
```

## 使用方法

1. Serve ollama

```bash
ollama serve
```

2. 別のターミナルでアプリケーションを実行します

```bash
./bin/whisper_recorder
```

3. 利用可能なマイクデバイスのリストから使用するデバイスを選択します

4. Ctrl+C を押して録音を停止します

## 機能

- マイクからのリアルタイム録音
- 定期的な録音セグメントの保存
- whisper.cpp を使用した音声文字起こし
- Ollama を使用したテキスト分析
  - 要約生成
  - キーワード抽出
  - 問題点抽出
  - 議論の進行状況評価
- 結果をマークダウンファイルに保存

## プロジェクト構成

```
.
├── main.go                         # メインアプリケーション
├── transcribe.sh                   # 文字起こし用シェルスクリプト
├── internal/
│   ├── app/                        # アプリケーション基本構造
│   │   ├── app.go
│   │   ├── ollama.go
│   │   └── markdown.go
│   ├── audio/                      # オーディオ処理
│   │   └── wav.go
│   ├── transcription/              # 文字起こし処理
│   │   ├── transcribe.go
│   │   └── process.go
│   └── analysis/                   # テキスト分析
│       ├── analysis.go
│       └── ollama.go
├── data/                           # 録音と文字起こし結果
│   ├── recordings/
│   └── transcripts/
├── go.mod
└── go.sum
```

## 注意事項

- 文字起こしの精度は whisper.cpp のモデルに依存します
- Ollama のモデルによって分析結果の質が変わります
- アプリケーションの処理速度はご使用のコンピュータの性能に依存します
