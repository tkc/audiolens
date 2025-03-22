package audio

import (
	"encoding/binary"
	"fmt"
	"os"
)

// WAVファイルとして保存
func SaveAsWav(filepath string, audioBuffer [][]float32, sampleRate int) error {
	if len(audioBuffer) == 0 {
		return fmt.Errorf("保存するオーディオバッファが空です")
	}

	// バッファのサイズを計算
	totalSamples := 0
	for _, buffer := range audioBuffer {
		totalSamples += len(buffer)
	}

	// バッファを1つの配列に結合
	combined := make([]float32, 0, totalSamples)
	for _, buffer := range audioBuffer {
		combined = append(combined, buffer...)
	}

	// float32からint16に変換（WAVファイル形式用）
	samples := make([]int16, len(combined))
	for i, sample := range combined {
		// float32（-1.0〜1.0）からint16（-32768〜32767）への変換
		samples[i] = int16(sample * 32767.0)
	}

	// ファイルを作成
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("WAVファイルを作成できませんでした: %v", err)
	}
	defer file.Close()

	// WAVヘッダーを書き込む
	writeWavHeader(file, uint32(totalSamples), 1, uint32(sampleRate), 16)

	// サンプルデータを書き込む
	for _, sample := range samples {
		err := binary.Write(file, binary.LittleEndian, sample)
		if err != nil {
			return fmt.Errorf("サンプルデータ書き込みエラー: %v", err)
		}
	}
	
	return nil
}

// WAVヘッダーを書き込む関数
func writeWavHeader(file *os.File, dataSize uint32, numChannels uint16, sampleRate uint32, bitsPerSample uint16) {
	// 計算値
	bytesPerSample := bitsPerSample / 8
	dataBytes := dataSize * uint32(bytesPerSample)

	// RIFFヘッダー
	file.WriteString("RIFF")
	binary.Write(file, binary.LittleEndian, uint32(36+dataBytes))
	file.WriteString("WAVE")

	// fmtチャンク
	file.WriteString("fmt ")
	binary.Write(file, binary.LittleEndian, uint32(16)) // fmtチャンクのサイズ
	binary.Write(file, binary.LittleEndian, uint16(1))  // フォーマットタグ（PCM = 1）
	binary.Write(file, binary.LittleEndian, numChannels)
	binary.Write(file, binary.LittleEndian, sampleRate)
	binary.Write(file, binary.LittleEndian, uint32(sampleRate*uint32(numChannels)*uint32(bytesPerSample))) // バイト/秒
	binary.Write(file, binary.LittleEndian, uint16(numChannels*bytesPerSample))                           // ブロックサイズ
	binary.Write(file, binary.LittleEndian, bitsPerSample)

	// dataチャンク
	file.WriteString("data")
	binary.Write(file, binary.LittleEndian, dataBytes)
}
