package rag

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"orange-agent/utils/logger"
)

// CodeRetriever 代码检索器
type CodeRetriever struct {
	indexer  *CodeIndexer
	embedder TextEmbedder
	log      *logger.Logger
	mu       sync.RWMutex
}

// NewCodeRetriever 创建代码检索器
func NewCodeRetriever(indexer *CodeIndexer, embedder TextEmbedder) *CodeRetriever {
	return &CodeRetriever{
		indexer:  indexer,
		embedder: embedder,
		log:      logger.GetLogger(),
	}
}

// Retrieve 根据查询检索相关代码
func (r *CodeRetriever) Retrieve(ctx context.Context, query string, topK int) ([]Chunk, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 生成查询向量
	queryVector, err := r.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("生成查询向量失败: %v", err)
	}

	// 搜索相似的代码块
	results, err := r.indexer.store.Search(ctx, queryVector, topK)
	if err != nil {
		return nil, fmt.Errorf("搜索失败: %v", err)
	}

	// 提取代码块
	chunks := make([]Chunk, len(results))
	for i, result := range results {
		chunks[i] = result.Chunk
	}

	return chunks, nil
}

// BuildContext 构建检索到的代码上下文
func (r *CodeRetriever) BuildContext(chunks []Chunk) string {
	if len(chunks) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("【相关代码上下文】\n\n")

	for i, chunk := range chunks {
		builder.WriteString(fmt.Sprintf("--- 文件: %s (行 %d-%d) ---\n",
			chunk.FilePath, chunk.StartLine, chunk.EndLine))
		builder.WriteString(chunk.Content)
		builder.WriteString("\n")

		if i < len(chunks)-1 {
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

// IndexDirectory 索引目录
func (r *CodeRetriever) IndexDirectory(ctx context.Context, dirPath string) error {
	return r.indexer.IndexDirectory(ctx, dirPath)
}

// GetIndexSize 获取索引大小
func (r *CodeRetriever) GetIndexSize(ctx context.Context) (int, error) {
	return r.indexer.GetSize(ctx)
}
