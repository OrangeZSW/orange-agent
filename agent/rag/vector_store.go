package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"sync"

	"orange-agent/utils/logger"

	"github.com/redis/go-redis/v9"
)

const (
	// Redis向量存储的key前缀
	vectorPrefix   = "rag:vector:"
	vectorIndexKey = "rag:index:ids"
	vectorCountKey = "rag:index:count"
)

// VectorStore 向量存储接口
type VectorStore interface {
	// Add 添加向量和对应的文本块
	Add(ctx context.Context, id string, vector []float64, chunk Chunk) error
	// Search 搜索最相似的向量
	Search(ctx context.Context, queryVector []float64, topK int) ([]SearchResult, error)
	// Clear 清空存储
	Clear(ctx context.Context) error
	// Size 返回存储的向量数量
	Size(ctx context.Context) (int, error)
	// Close 关闭连接
	Close() error
}

// SearchResult 搜索结果
type SearchResult struct {
	Chunk Chunk
	Score float64
}

// VectorEntry 向量条目（用于存储）
type VectorEntry struct {
	ID     string    `json:"id"`
	Vector []float64 `json:"vector"`
	Chunk  Chunk     `json:"chunk"`
}

// RedisVectorStore Redis向量存储实现
type RedisVectorStore struct {
	client *redis.Client
	prefix string
	mu     sync.RWMutex
	log    *logger.Logger
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// NewRedisVectorStore 创建Redis向量存储
func NewRedisVectorStore(config *RedisConfig) (*RedisVectorStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("连接Redis失败: %v", err)
	}

	log := logger.GetLogger()
	log.Info("Redis向量存储连接成功: %s:%d", config.Host, config.Port)

	return &RedisVectorStore{
		client: client,
		prefix: vectorPrefix,
		log:    log,
	}, nil
}

// Add 添加向量
func (s *RedisVectorStore) Add(ctx context.Context, id string, vector []float64, chunk Chunk) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry := VectorEntry{
		ID:     id,
		Vector: vector,
		Chunk:  chunk,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("序列化向量失败: %v", err)
	}

	// 存储向量数据
	key := s.prefix + id
	if err := s.client.Set(ctx, key, data, 0).Err(); err != nil {
		return fmt.Errorf("存储向量失败: %v", err)
	}

	// 添加到索引集合
	if err := s.client.SAdd(ctx, vectorIndexKey, id).Err(); err != nil {
		return fmt.Errorf("添加索引失败: %v", err)
	}

	// 更新计数
	s.client.Incr(ctx, vectorCountKey)

	return nil
}

// Search 搜索最相似的向量（余弦相似度）
func (s *RedisVectorStore) Search(ctx context.Context, queryVector []float64, topK int) ([]SearchResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 获取所有向量ID
	ids, err := s.client.SMembers(ctx, vectorIndexKey).Result()
	if err != nil {
		return nil, fmt.Errorf("获取索引失败: %v", err)
	}

	if len(ids) == 0 {
		return nil, nil
	}

	// 批量获取向量数据
	type scoredEntry struct {
		entry VectorEntry
		score float64
	}

	scored := make([]scoredEntry, 0, len(ids))

	// 使用pipeline批量获取，提高性能
	pipe := s.client.Pipeline()
	cmds := make(map[string]*redis.StringCmd)

	for _, id := range ids {
		key := s.prefix + id
		cmds[id] = pipe.Get(ctx, key)
	}

	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, fmt.Errorf("批量获取向量失败: %v", err)
	}

	// 计算相似度
	for _, cmd := range cmds {
		data, err := cmd.Result()
		if err != nil {
			continue // 跳过不存在的向量
		}

		var entry VectorEntry
		if err := json.Unmarshal([]byte(data), &entry); err != nil {
			continue
		}

		score := cosineSimilarity(queryVector, entry.Vector)
		scored = append(scored, scoredEntry{entry: entry, score: score})
	}

	// 按相似度排序（降序）
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	// 返回前 topK 个结果
	count := topK
	if count > len(scored) {
		count = len(scored)
	}

	results := make([]SearchResult, count)
	for i := 0; i < count; i++ {
		results[i] = SearchResult{
			Chunk: scored[i].entry.Chunk,
			Score: scored[i].score,
		}
	}

	return results, nil
}

// Clear 清空存储
func (s *RedisVectorStore) Clear(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 获取所有向量ID
	ids, err := s.client.SMembers(ctx, vectorIndexKey).Result()
	if err != nil {
		return err
	}

	// 删除所有向量数据
	pipe := s.client.Pipeline()
	for _, id := range ids {
		key := s.prefix + id
		pipe.Del(ctx, key)
	}

	// 删除索引和计数
	pipe.Del(ctx, vectorIndexKey)
	pipe.Del(ctx, vectorCountKey)

	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	s.log.Info("已清空Redis向量存储")
	return nil
}

// Size 返回存储的向量数量
func (s *RedisVectorStore) Size(ctx context.Context) (int, error) {
	count, err := s.client.Get(ctx, vectorCountKey).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(count)
}

// Close 关闭连接
func (s *RedisVectorStore) Close() error {
	return s.client.Close()
}

// cosineSimilarity 计算余弦相似度
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
