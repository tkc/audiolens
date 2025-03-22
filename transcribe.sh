#!/bin/bash
# 音声ファイルの文字起こしを行うシェルスクリプト

# 入力チェック
if [ "$#" -ne 1 ]; then
    echo "使用方法: $0 入力ファイル.wav"
    exit 1
fi

INPUT_FILE="$1"
WHISPER_PATH="/Users/takeshiiijima/github/whisper.cpp/build/bin/whisper-cli"
MODEL_PATH="/Users/takeshiiijima/github/whisper.cpp/models/ggml-base.bin"

# ファイルの存在チェック
if [ ! -f "$INPUT_FILE" ]; then
    echo "エラー: 入力ファイルが見つかりません: $INPUT_FILE"
    exit 1
fi

if [ ! -f "$WHISPER_PATH" ]; then
    echo "エラー: whisper-cliが見つかりません: $WHISPER_PATH"
    exit 1
fi

if [ ! -f "$MODEL_PATH" ]; then
    echo "エラー: モデルファイルが見つかりません: $MODEL_PATH"
    exit 1
fi

echo "文字起こし開始: $INPUT_FILE"

# whisper.cppを実行
"$WHISPER_PATH" -m "$MODEL_PATH" -f "$INPUT_FILE" -l ja --output-txt --no-gpu

# 実行結果の確認
if [ $? -eq 0 ]; then
    echo "文字起こし完了: ${INPUT_FILE}.txt"
    exit 0
else
    echo "文字起こしに失敗しました"
    exit 1
fi
