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

// IndexDirectory 全量索引目录
func (idx *CodeIndexer) IndexDirectory(ctx context.Context, dirPath string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.log.Info("开始全量索引目录: %s", dirPath)

	// 清空现有索引
	if err := idx.store.Clear(ctx); err != nil {
		return err
	}

	indexed := 0
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if shouldSkipDir(path) {
				return filepath.SkipDir
			}
			return nil
		}

		if !isCodeFile(path) {
			return nil
		}

		if err := idx.indexFile(ctx, path, info.ModTime().Unix()); err != nil {
			idx.log.Warn("索引文件失败: %s, 错误: %v", path, err)
			return nil
		}
		indexed++
		return nil
	})

	if err != nil {
		return err
	}

	size, _ := idx.store.Size(ctx)
	idx.log.Info("全量索引完成，索引 %d 个文件，共 %d 个代码块", indexed, size)
	return nil
}

// IndexDirectoryIncremental 增量索引目录
func (idx *CodeIndexer) IndexDirectoryIncremental(ctx context.Context, dirPath string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.log.Info("开始增量索引目录: %s", dirPath)

	// 获取已索引文件列表
	indexedFiles, err := idx.store.GetAllIndexedFiles(ctx)
	if err != nil {
		idx.log.Warn("获取已索引文件失败: %v", err)
		indexedFiles = make(map[string]int64)
	}

	// 收集当前目录中的文件
	currentFiles := make(map[string]int64)

	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if shouldSkipDir(path) {
				return filepath.SkipDir
			}
			return nil
		}

		if !isCodeFile(path) {
			return nil
		}

		currentFiles[path] = info.ModTime().Unix()
		return nil
	})

	if err != nil {
		return err
	}

	// 统计
	added := 0
	updated := 0
	deleted := 0

	// 检查新增和修改的文件
	for filePath, modTime := range currentFiles {
		oldModTime, exists := indexedFiles[filePath]

		if !exists {
			// 新增文件
			if err := idx.indexFile(ctx, filePath, modTime); err != nil {
				idx.log.Warn("索引新文件失败: %s, 错误: %v", filePath, err)
			} else {
				added++
			}
		} else if modTime > oldModTime {
			// 修改的文件
			if err := idx.store.DeleteByFilePath(ctx, filePath); err != nil {
				idx.log.Warn("删除旧索引失败: %s, 错误: %v", filePath, err)
			}
			if err := idx.indexFile(ctx, filePath, modTime); err != nil {
				idx.log.Warn("重新索引文件失败: %s, 错误: %v", filePath, err)
			} else {
				updated++
			}
		}
		// 未修改的文件跳过
	}

	// 检查删除的文件
	for filePath := range indexedFiles {
		if _, exists := currentFiles[filePath]; !exists {
			// 文件已删除
			if err := idx.store.DeleteByFilePath(ctx, filePath); err != nil {
				idx.log.Warn("删除索引失败: %s, 错误: %v", filePath, err)
			} else {
				idx.store.DeleteFileMeta(ctx, filePath)
				deleted++
			}
		}
	}

	size, _ := idx.store.Size(ctx)
	idx.log.Info("增量索引完成: 新增 %d, 更新 %d, 删除 %d, 总计 %d 个代码块", added, updated, deleted, size)
	return nil
}

// indexFile 索引单个文件
func (idx *CodeIndexer) indexFile(ctx context.Context, filePath string, modTime int64) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	chunks := idx.chunker.SplitIntoChunks(filePath, string(content))

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

	// 保存文件元数据
	return idx.store.SetFileMeta(ctx, filePath, modTime)
}

// GetSize 获取索引大小
func (idx *CodeIndexer) GetSize(ctx context.Context) (int, error) {
	return idx.store.Size(ctx)
}

func shouldSkipDir(dirPath string) bool {
	skipDirs := []string{
		".git", "node_modules", "vendor", "__pycache__",
		".idea", ".vscode", "build", "dist", "target",
	}
	dirName := filepath.Base(dirPath)
	for _, skip := range skipDirs {
		if dirName == skip {
			return true
		}
	}
	return false
}

func isCodeFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	codeExts := map[string]bool{
		".go": true, ".py": true, ".js": true, ".ts": true,
		".java": true, ".c": true, ".cpp": true, ".h": true,
		".md": true, ".txt": true, ".yaml": true, ".yml": true,
		".json": true, ".xml": true, ".html": true, ".css": true,
		".sql": true, ".sh": true,
	}
	return codeExts[ext]
}
