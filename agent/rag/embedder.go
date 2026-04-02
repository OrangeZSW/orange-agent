package rag

import (
	"context"
	"math"
	"strings"
	"unicode"

	"orange-agent/utils/logger"
)

const (
	// VectorDim 向量维度
	VectorDim = 256
)

// TextEmbedder 文本嵌入接口
type TextEmbedder interface {
	// Embed 将文本转换为向量
	Embed(ctx context.Context, text string) ([]float64, error)
}

// SimpleEmbedder 简单的文本嵌入器（基于TF-IDF）
// 适用于小型代码库，无需额外API调用
type SimpleEmbedder struct {
	vocabulary map[string]int
	docFreq    map[string]int
	totalDocs  int
	log        *logger.Logger
}

// NewSimpleEmbedder 创建简单嵌入器
func NewSimpleEmbedder() *SimpleEmbedder {
	return &SimpleEmbedder{
		vocabulary: make(map[string]int),
		docFreq:    make(map[string]int),
		log:        logger.GetLogger(),
	}
}

// Embed 将文本转换为向量
func (e *SimpleEmbedder) Embed(ctx context.Context, text string) ([]float64, error) {
	keywords := extractKeywords(text)
	vector := make([]float64, VectorDim)

	for _, keyword := range keywords {
		if idx, exists := e.vocabulary[keyword]; exists {
			tf := float64(strings.Count(text, keyword)) / float64(len(text))
			df := float64(e.docFreq[keyword])
			idf := math.Log(float64(e.totalDocs+1) / (df + 1))
			vector[idx%VectorDim] += tf * idf
		}
	}

	normalizeVector(vector)
	return vector, nil
}

// BuildVocabulary 从文本列表构建词汇表
func (e *SimpleEmbedder) BuildVocabulary(texts []string) {
	e.log.Info("构建词汇表，文档数: %d", len(texts))

	keywordDocCount := make(map[string]int)
	for _, text := range texts {
		seen := make(map[string]bool)
		for _, keyword := range extractKeywords(text) {
			if !seen[keyword] {
				keywordDocCount[keyword]++
				seen[keyword] = true
			}
		}
	}

	idx := 0
	for keyword, count := range keywordDocCount {
		if count >= 2 {
			e.vocabulary[keyword] = idx
			e.docFreq[keyword] = count
			idx++
		}
	}

	e.totalDocs = len(texts)
	e.log.Info("词汇表构建完成，关键词数: %d", len(e.vocabulary))
}

// extractKeywords 提取文本中的关键词
func extractKeywords(text string) []string {
	text = strings.ToLower(text)
	var keywords []string
	var current strings.Builder

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			current.WriteRune(r)
		} else if current.Len() > 0 {
			word := current.String()
			if len(word) >= 2 && !isStopWord(word) {
				keywords = append(keywords, word)
			}
			current.Reset()
		}
	}

	if current.Len() > 0 {
		word := current.String()
		if len(word) >= 2 && !isStopWord(word) {
			keywords = append(keywords, word)
		}
	}

	return keywords
}

// normalizeVector 归一化向量
func normalizeVector(vector []float64) {
	var sum float64
	for _, v := range vector {
		sum += v * v
	}
	if sum == 0 {
		return
	}
	norm := math.Sqrt(sum)
	for i := range vector {
		vector[i] /= norm
	}
}

// isStopWord 判断是否是停用词
func isStopWord(word string) bool {
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "from": true,
		"is": true, "are": true, "was": true, "were": true, "be": true,
		"been": true, "being": true, "have": true, "has": true, "had": true,
		"do": true, "does": true, "did": true, "will": true, "would": true,
		"could": true, "should": true, "may": true, "might": true, "can": true,
		"this": true, "that": true, "these": true, "those": true,
		"i": true, "you": true, "he": true, "she": true, "it": true,
		"we": true, "they": true, "me": true, "him": true, "her": true,
		"us": true, "them": true, "my": true, "your": true, "his": true,
		"its": true, "our": true, "their": true,
	}
	return stopWords[word]
}
