package rag

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"orange-agent/utils/logger"
)

var (
	retriever *CodeRetriever
	mu        sync.RWMutex
	log       = logger.GetLogger()
)

// GetRetriever 获取代码检索器
func GetRetriever() *CodeRetriever {
	mu.RLock()
	defer mu.RUnlock()
	return retriever
}

// Init 初始化RAG模块
func Init(config *RedisConfig) error {
	mu.Lock()
	defer mu.Unlock()

	if retriever != nil {
		log.Info("RAG模块已初始化，跳过")
		return nil
	}

	store, err := NewRedisVectorStore(config, VectorDim)
	if err != nil {
		return fmt.Errorf("Redis连接失败: %v", err)
	}

	embedder := NewSimpleEmbedder()
	indexer := NewCodeIndexer(store, embedder)
	retriever = NewCodeRetriever(indexer, embedder)

	log.Info("RAG模块初始化成功")
	return nil
}

// IndexFull 全量索引
func IndexFull(ctx context.Context, projectRoot string) error {
	r := GetRetriever()
	if r == nil {
		return fmt.Errorf("RAG模块未初始化")
	}

	log.Info("开始全量索引: %s", projectRoot)

	texts := collectCodeFiles(projectRoot)
	log.Info("扫描到 %d 个代码文件", len(texts))

	if embedder, ok := r.embedder.(*SimpleEmbedder); ok {
		embedder.BuildVocabulary(texts)
	}

	if err := r.IndexDirectory(ctx, projectRoot); err != nil {
		return err
	}

	size, _ := r.GetIndexSize(ctx)
	log.Info("全量索引完成，共 %d 个代码块", size)
	return nil
}

// IndexIncremental 增量索引
func IndexIncremental(ctx context.Context, projectRoot string) error {
	r := GetRetriever()
	if r == nil {
		return fmt.Errorf("RAG模块未初始化")
	}

	log.Info("开始增量索引: %s", projectRoot)

	texts := collectCodeFiles(projectRoot)
	log.Info("扫描到 %d 个代码文件", len(texts))

	if embedder, ok := r.embedder.(*SimpleEmbedder); ok {
		embedder.BuildVocabulary(texts)
	}

	if err := r.IndexDirectoryIncremental(ctx, projectRoot); err != nil {
		return err
	}

	size, _ := r.GetIndexSize(ctx)
	log.Info("增量索引完成，共 %d 个代码块", size)
	return nil
}

// Search 搜索代码
func Search(ctx context.Context, query string, topK int) ([]Chunk, error) {
	r := GetRetriever()
	if r == nil {
		return nil, fmt.Errorf("RAG模块未初始化")
	}
	return r.Retrieve(ctx, query, topK)
}

// BuildContext 构建上下文
func BuildContext(chunks []Chunk) string {
	r := GetRetriever()
	if r == nil {
		return ""
	}
	return r.BuildContext(chunks)
}

// GetSize 获取索引大小
func GetSize(ctx context.Context) (int, error) {
	r := GetRetriever()
	if r == nil {
		return 0, fmt.Errorf("RAG模块未初始化")
	}
	return r.GetIndexSize(ctx)
}

// collectCodeFiles 收集代码文件内容
func collectCodeFiles(root string) []string {
	var texts []string
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !isCodeFile(path) {
			return nil
		}
		if shouldSkipDir(path) {
			return filepath.SkipDir
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		texts = append(texts, string(content))
		return nil
	})
	return texts
}
