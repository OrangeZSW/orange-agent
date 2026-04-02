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
	globalRetriever *CodeRetriever
	once            sync.Once
	mu              sync.RWMutex
	log             = logger.GetLogger()
)

// GetRetriever 获取全局代码检索器实例
func GetRetriever() *CodeRetriever {
	return globalRetriever
}

// InitializeWithRedis 使用Redis初始化代码检索器
func InitializeWithRedis(config *RedisConfig) error {
	var initErr error
	once.Do(func() {
		// 创建Redis向量存储
		store, err := NewRedisVectorStore(config)
		if err != nil {
			initErr = err
			return
		}

		// 创建嵌入器
		embedder := NewSimpleEmbedder()

		// 创建索引器
		indexer := NewCodeIndexer(store, embedder)

		// 创建检索器
		globalRetriever = NewCodeRetriever(indexer, embedder)

		log.Info("Redis向量存储初始化成功")
	})

	return initErr
}

// InitializeIndex 初始化代码索引
func InitializeIndex(ctx context.Context, projectRoot string) error {
	mu.Lock()
	defer mu.Unlock()

	if globalRetriever == nil {
		return fmt.Errorf("向量存储未初始化，请先调用 InitializeWithRedis")
	}

	log.Info("开始初始化代码索引，项目根目录: %s", projectRoot)

	// 收集所有代码文件用于构建词汇表
	var allTexts []string
	err := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
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

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		allTexts = append(allTexts, string(content))
		return nil
	})

	if err != nil {
		return err
	}

	// 构建词汇表
	if embedder, ok := globalRetriever.embedder.(*SimpleEmbedder); ok {
		embedder.BuildVocabulary(allTexts)
	}

	// 索引目录
	if err := globalRetriever.IndexDirectory(ctx, projectRoot); err != nil {
		return err
	}

	size, _ := globalRetriever.GetIndexSize(ctx)
	log.Info("代码索引初始化完成，共 %d 个代码块", size)
	return nil
}

// RefreshIndex 刷新代码索引
func RefreshIndex(ctx context.Context, projectRoot string) error {
	return InitializeIndex(ctx, projectRoot)
}
