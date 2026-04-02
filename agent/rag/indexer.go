package rag

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"orange-agent/utils/logger"
)

// CodeIndexer 代码索引器
type CodeIndexer struct {
	store    VectorStore
	chunker  *Chunker
	embedder TextEmbedder
	log      *logger.Logger
	mu       sync.RWMutex
}

// TextEmbedder 文本嵌入接口
type TextEmbedder interface {
	// Embed 将文本转换为向量
	Embed(ctx context.Context, text string) ([]float64, error)
}

// NewCodeIndexer 创建代码索引器
func NewCodeIndexer(store VectorStore, embedder TextEmbedder) *CodeIndexer {
	return &CodeIndexer{
		store:    store,
		chunker:  NewChunker(),
		embedder: embedder,
		log:      logger.GetLogger(),
	}
}

// IndexDirectory 索引目录中的所有代码文件
func (idx *CodeIndexer) IndexDirectory(ctx context.Context, dirPath string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.log.Info("开始索引目录: %s", dirPath)

	// 清空现有索引
	if err := idx.store.Clear(ctx); err != nil {
		return err
	}

	// 遍历目录
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			// 跳过不需要索引的目录
			if shouldSkipDir(path) {
				return filepath.SkipDir
			}
			return nil
		}

		// 只处理代码文件
		if !isCodeFile(path) {
			return nil
		}

		// 索引文件
		if err := idx.indexFile(ctx, path); err != nil {
			idx.log.Warn("索引文件失败: %s, 错误: %v", path, err)
			// 继续处理其他文件
			return nil
		}

		return nil
	})

	if err != nil {
		return err
	}

	size, _ := idx.store.Size(ctx)
	idx.log.Info("索引完成，共 %d 个代码块", size)
	return nil
}

// indexFile 索引单个文件
func (idx *CodeIndexer) indexFile(ctx context.Context, filePath string) error {
	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// 分块
	chunks := idx.chunker.SplitIntoChunks(filePath, string(content))

	// 为每个块生成向量并存储
	for _, chunk := range chunks {
		vector, err := idx.embedder.Embed(ctx, chunk.Content)
		if err != nil {
			idx.log.Warn("生成向量失败: %s, 错误: %v", chunk.ID, err)
			continue
		}

		if err := idx.store.Add(ctx, chunk.ID, vector, chunk); err != nil {
			idx.log.Warn("存储向量失败: %s, 错误: %v", chunk.ID, err)
			continue
		}
	}

	return nil
}

// GetSize 获取索引大小
func (idx *CodeIndexer) GetSize(ctx context.Context) (int, error) {
	return idx.store.Size(ctx)
}

// shouldSkipDir 判断是否应该跳过目录
func shouldSkipDir(dirPath string) bool {
	skipDirs := []string{
		".git",
		"node_modules",
		"vendor",
		"__pycache__",
		".idea",
		".vscode",
		"build",
		"dist",
		"target",
	}

	dirName := filepath.Base(dirPath)
	for _, skip := range skipDirs {
		if dirName == skip {
			return true
		}
	}
	return false
}

// isCodeFile 判断是否是代码文件
func isCodeFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	codeExts := map[string]bool{
		".go":    true,
		".py":    true,
		".js":    true,
		".ts":    true,
		".java":  true,
		".c":     true,
		".cpp":   true,
		".h":     true,
		".hpp":   true,
		".cs":    true,
		".php":   true,
		".rb":    true,
		".rs":    true,
		".swift": true,
		".kt":    true,
		".scala": true,
		".md":    true,
		".txt":   true,
		".yaml":  true,
		".yml":   true,
		".json":  true,
		".xml":   true,
		".html":  true,
		".css":   true,
		".sql":   true,
		".sh":    true,
		".bash":  true,
	}
	return codeExts[ext]
}
