package rag

import "fmt"

// Chunk 代码块
type Chunk struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	FilePath  string `json:"file_path"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

// String 返回代码块的字符串表示
func (c Chunk) String() string {
	return fmt.Sprintf("%s:%d-%d", c.FilePath, c.StartLine, c.EndLine)
}
