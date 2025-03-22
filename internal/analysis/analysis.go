package analysis

import (
	"fmt"
	"strings"
)

// テキストの要約生成
func GenerateSummary(text string) (string, error) {
	systemPrompt := "あなたは優秀な要約者です。与えられたテキストを30字程度で要約してください。"
	// 直接インポートすると循環参照になるため、app.QueryOllamaを直接使用せず
	// 代わりにqueryOllamaヘルパー関数を使う
	summary, err := queryOllama(text, systemPrompt)
	if err != nil {
		return "", err
	}

	fmt.Printf("  要約生成完了: %s\n", summary)
	return summary, nil
}

// キーワード抽出
func ExtractKeywords(text string) ([]string, error) {
	systemPrompt := "次の文から最も重要なキーワードを3〜5つ抽出し、カンマ区切りのリストで返してください。"
	keywordsText, err := queryOllama(text, systemPrompt)
	if err != nil {
		return nil, err
	}

	keywords := strings.Split(keywordsText, ",")
	for i, kw := range keywords {
		keywords[i] = strings.TrimSpace(kw)
	}

	fmt.Printf("  キーワード抽出完了: %v\n", keywords)
	return keywords, nil
}

// 問題点抽出
func IdentifyIssues(text string) (string, error) {
	systemPrompt := "次の文から言及されている問題点や課題を短く抽出してください。問題が見つからない場合は「特に問題点はありません」と返してください。"
	issues, err := queryOllama(text, systemPrompt)
	if err != nil {
		return "", err
	}

	fmt.Printf("  問題点抽出完了: %s\n", issues)
	return issues, nil
}

// 議論の順調さを評価
func EvaluateProgress(text string) (string, error) {
	systemPrompt := "次の会話を分析し、議論が順調に進んでいるかどうかを0から5の評価で返してください。0は全く順調でない、5は非常に順調である、ということを意味します。評価理由も簡潔に添えてください。「評価: [数字]、理由: [説明]」というフォーマットで回答してください。"
	progressEvaluation, err := queryOllama(text, systemPrompt)
	if err != nil {
		return "", err
	}

	fmt.Printf("  進行状況評価完了: %s\n", progressEvaluation)
	return progressEvaluation, nil
}

// 会話内の攻撃的な言葉をチェック
func CheckAggressiveLanguage(text string) (string, error) {
	systemPrompt := "下記の会話テキストに攻撃的な言葉や非友好的な表現が含まれているか分析してください。以下の点に注目して判断してください：価値を負かす発言、直接的な人格批判、危害や威嚇、話還しにつながる言葉、底意や当てこすり、価値を否定する言葉過剰な価値判断、厄介、他者の尊厳を傷つける発言。"
	systemPrompt += "結果は以下のフォーマットで返してください：\n「攻撃性評価: [低/中/高]\n検出された表現: [具体的な表現や言葉（あれば）]\n理由: [簡潔な説明]」\n攻撃的な表現が見つからない場合は、「攻撃性評価: 低\n検出された表現: なし\n理由: 会話内に攻撃的表現は見つかりませんでした。」と返してください。"
	
	aggressiveCheck, err := queryOllama(text, systemPrompt)
	if err != nil {
		return "", err
	}

	fmt.Printf("  攻撃的言葉チェック完了: %s\n", aggressiveCheck)
	return aggressiveCheck, nil
}
