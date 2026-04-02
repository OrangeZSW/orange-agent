package rag

import (
	"fmt"
	"strings"
)

const (
	// ChunkSize 每个块的最大字符数
	ChunkSize = 500
	// ChunkOverlap 块之间的重叠字符数
	ChunkOverlap = 50
)

// Chunker 代码分块器
type Chunker struct {
	chunkSize    int
	chunkOverlap int
}

// NewChunker 创建代码分块器
func NewChunker() *Chunker {
	return &Chunker{
		chunkSize:    ChunkSize,
		chunkOverlap: ChunkOverlap,
	}
}

// SplitIntoChunks 将文件内容分块
func (c *Chunker) SplitIntoChunks(filePath string, content string) []Chunk {
	var chunks []Chunk
	lines := strings.Split(content, "\n")

	currentChunk := strings.Builder{}
	startLine := 1
	chunkID := 0

	for i, line := range lines {
		currentChunk.WriteString(line)
		currentChunk.WriteString("\n")

		// 当块达到指定大小时，创建新块
		if currentChunk.Len() >= c.chunkSize {
			chunks = append(chunks, Chunk{
				ID:        fmt.Sprintf("%s_%d", filePath, chunkID),
				Content:   currentChunk.String(),
				FilePath:  filePath,
				StartLine: startLine,
				EndLine:   i + 1,
			})

			chunkID++

			// 保留重叠部分
			overlapLines := c.getOverlapLines(lines, i, c.chunkOverlap)
			currentChunk.Reset()
			currentChunk.WriteString(overlapLines)

			// 计算新的起始行
			overlapLineCount := strings.Count(overlapLines, "\n")
			if overlapLineCount > 0 {
				startLine = i + 1 - overlapLineCount + 1
			} else {
				startLine = i + 1
			}
		}
	}

	// 处理最后一个块
	if currentChunk.Len() > 0 {
		chunks = append(chunks, Chunk{
			ID:        fmt.Sprintf("%s_%d", filePath, chunkID),
			Content:   currentChunk.String(),
			FilePath:  filePath,
			StartLine: startLine,
			EndLine:   len(lines),
		})
	}

	return chunks
}

// getOverlapLines 获取重叠的行
func (c *Chunker) getOverlapLines(lines []string, currentIndex int, overlapChars int) string {
	if currentIndex <= 0 || overlapChars <= 0 {
		return ""
	}

	var overlap strings.Builder
	charCount := 0

	// 从当前行往前找重叠内容
	for i := currentIndex; i >= 0 && charCount < overlapChars; i-- {
		line := lines[i]
		overlap.WriteString(line)
		overlap.WriteString("\n")
		charCount += len(line)

		if charCount >= overlapChars {
			break
		}
	}

	// 反转字符串（因为我们是从后往前构建的）
	result := overlap.String()
	return reverseLines(result)
}

// reverseLines 反转行的顺序
func reverseLines(s string) string {
	lines := strings.Split(s, "\n")
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}
	return strings.Join(lines, "\n")
}
