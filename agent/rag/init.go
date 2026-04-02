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
	initMu          sync.RWMutex
	log             = logger.GetLogger()
)

// GetRetriever 获取全局代码检索器实例
func GetRetriever() *CodeRetriever {
	initMu.RLock()
	defer initMu.RUnlock()
	return globalRetriever
}

// InitializeWithRedis 使用Redis初始化代码检索器
func InitializeWithRedis(config *RedisConfig) error {
	initMu.Lock()
	defer initMu.Unlock()

	// 如果已经初始化成功，跳过
	if globalRetriever != nil {
		log.Info("向量存储已初始化，跳过")
		return nil
	}

	log.Info("正在初始化Redis向量存储: %s:%d", config.Host, config.Port)

	// 创建Redis向量存储
	store, err := NewRedisVectorStore(config, VectorDim)
	if err != nil {
		return fmt.Errorf("Redis连接失败: %v", err)
	}

	// 创建嵌入器
	embedder := NewSimpleEmbedder()

	// 创建索引器
	indexer := NewCodeIndexer(store, embedder)

	// 创建检索器
	globalRetriever = NewCodeRetriever(indexer, embedder)

	log.Info("Redis向量存储初始化成功")
	return nil
}

// InitializeIndex 初始化代码索引
func InitializeIndex(ctx context.Context, projectRoot string) error {
	initMu.RLock()
	retriever := globalRetriever
	initMu.RUnlock()

	if retriever == nil {
		return fmt.Errorf("向量存储未初始化，请检查Redis连接配置是否正确")
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
		return fmt.Errorf("扫描文件失败: %v", err)
	}

	log.Info("扫描到 %d 个代码文件", len(allTexts))

	// 构建词汇表
	if embedder, ok := retriever.embedder.(*SimpleEmbedder); ok {
		embedder.BuildVocabulary(allTexts)
	}

	// 索引目录
	if err := retriever.IndexDirectory(ctx, projectRoot); err != nil {
		return fmt.Errorf("索引失败: %v", err)
	}

	size, err := retriever.GetIndexSize(ctx)
	if err != nil {
		log.Warn("获取索引大小失败: %v", err)
	}
	log.Info("代码索引初始化完成，共 %d 个代码块", size)
	return nil
}

// RefreshIndex 刷新代码索引
func RefreshIndex(ctx context.Context, projectRoot string) error {
	return InitializeIndex(ctx, projectRoot)
}
