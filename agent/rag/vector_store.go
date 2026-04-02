package rag

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"sync"

	"orange-agent/utils/logger"

	"github.com/redis/go-redis/v9"
)

const (
	// Redis索引名称
	vectorIndexName = "rag_vector_idx"
	// Redis Hash前缀
	vectorHashPrefix = "rag:doc:"
)

// VectorStore 向量存储接口
type VectorStore interface {
	Add(ctx context.Context, id string, vector []float64, chunk Chunk) error
	Search(ctx context.Context, queryVector []float64, topK int) ([]SearchResult, error)
	Clear(ctx context.Context) error
	Size(ctx context.Context) (int, error)
	Close() error
}

// SearchResult 搜索结果
type SearchResult struct {
	Chunk Chunk
	Score float64
}

// RedisVectorStore Redis向量存储（使用RediSearch）
type RedisVectorStore struct {
	client    *redis.Client
	vectorDim int
	mu        sync.RWMutex
	log       *logger.Logger
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// NewRedisVectorStore 创建Redis向量存储
func NewRedisVectorStore(config *RedisConfig, vectorDim int) (*RedisVectorStore, error) {
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
	log.Info("Redis Stack连接成功: %s:%d", config.Host, config.Port)

	store := &RedisVectorStore{
		client:    client,
		vectorDim: vectorDim,
		log:       log,
	}

	// 创建向量索引
	if err := store.createIndex(ctx); err != nil {
		log.Warn("创建向量索引失败（可能已存在）: %v", err)
	}

	return store, nil
}

// createIndex 创建RediSearch向量索引
func (s *RedisVectorStore) createIndex(ctx context.Context) error {
	// 先尝试删除旧索引
	s.client.Do(ctx, "FT.DROPINDEX", vectorIndexName, "DD")

	// 创建向量索引
	// FT.CREATE rag_vector_idx ON HASH PREFIX 1 rag:doc: SCHEMA content TEXT path TAG start NUMERIC end NUMERIC vector VECTOR FLAT 6 TYPE FLOAT32 DIM 256 DISTANCE_METRIC COSINE
	cmd := []interface{}{
		"FT.CREATE", vectorIndexName,
		"ON", "HASH",
		"PREFIX", "1", vectorHashPrefix,
		"SCHEMA",
		"content", "TEXT",
		"path", "TAG",
		"start", "NUMERIC",
		"end", "NUMERIC",
		"chunk_id", "TAG",
		"vector", "VECTOR", "FLAT", "6",
		"TYPE", "FLOAT32",
		"DIM", s.vectorDim,
		"DISTANCE_METRIC", "COSINE",
	}

	if err := s.client.Do(ctx, cmd...).Err(); err != nil {
		return fmt.Errorf("创建向量索引失败: %v", err)
	}

	s.log.Info("RediSearch向量索引创建成功，维度: %d", s.vectorDim)
	return nil
}

// Add 添加向量
func (s *RedisVectorStore) Add(ctx context.Context, id string, vector []float64, chunk Chunk) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 将float64向量转为float32字节（RediSearch要求）
	vectorBytes := float64ToFloat32Bytes(vector)

	// 存储为Hash
	key := vectorHashPrefix + id
	cmd := []interface{}{
		"HSET", key,
		"chunk_id", chunk.ID,
		"content", chunk.Content,
		"path", chunk.FilePath,
		"start", chunk.StartLine,
		"end", chunk.EndLine,
		"vector", vectorBytes,
	}

	return s.client.Do(ctx, cmd...).Err()
}

// Search 使用RediSearch进行向量搜索
func (s *RedisVectorStore) Search(ctx context.Context, queryVector []float64, topK int) ([]SearchResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 将查询向量转为字节
	queryBytes := float64ToFloat32Bytes(queryVector)

	// FT.SEARCH rag_vector_idx "*=>[KNN 5 @vector $query_vec]" PARAMS 2 query_vec <bytes> DIALECT 2
	cmd := []interface{}{
		"FT.SEARCH", vectorIndexName,
		fmt.Sprintf("*=>[KNN %d @vector $query_vec]", topK),
		"PARAMS", "2", "query_vec", queryBytes,
		"DIALECT", "2",
		"RETURN", "6", "chunk_id", "content", "path", "start", "end", "__vector_score",
	}

	result, err := s.client.Do(ctx, cmd...).Result()
	if err != nil {
		return nil, fmt.Errorf("向量搜索失败: %v", err)
	}

	return s.parseSearchResult(result)
}

// parseSearchResult 解析RediSearch搜索结果
func (s *RedisVectorStore) parseSearchResult(result interface{}) ([]SearchResult, error) {
	results, ok := result.([]interface{})
	if !ok || len(results) < 2 {
		return nil, nil
	}

	totalCount := results[0].(int64)
	if totalCount == 0 {
		return nil, nil
	}

	var searchResults []SearchResult

	// 结果格式: [总数, key1, [field1, val1, ...], key2, [field2, val2, ...], ...]
	for i := 1; i < len(results); i += 2 {
		if i+1 >= len(results) {
			break
		}

		fields, ok := results[i+1].([]interface{})
		if !ok {
			continue
		}

		chunk := Chunk{}
		var score float64

		// 解析字段
		for j := 0; j < len(fields); j += 2 {
			if j+1 >= len(fields) {
				break
			}
			fieldName := fields[j].(string)
			fieldValue := fields[j+1].(string)

			switch fieldName {
			case "chunk_id":
				chunk.ID = fieldValue
			case "content":
				chunk.Content = fieldValue
			case "path":
				chunk.FilePath = fieldValue
			case "start":
				chunk.StartLine, _ = strconv.Atoi(fieldValue)
			case "end":
				chunk.EndLine, _ = strconv.Atoi(fieldValue)
			case "__vector_score":
				score, _ = strconv.ParseFloat(fieldValue, 64)
				// 余弦距离转相似度 (1 - distance)
				score = 1 - score
			}
		}

		searchResults = append(searchResults, SearchResult{
			Chunk: chunk,
			Score: score,
		})
	}

	return searchResults, nil
}

// Clear 清空所有向量数据
func (s *RedisVectorStore) Clear(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 删除索引（同时删除关联的Hash）
	if err := s.client.Do(ctx, "FT.DROPINDEX", vectorIndexName, "DD").Err(); err != nil {
		// 索引不存在不算错误
		s.log.Warn("删除索引失败: %v", err)
	}

	// 重新创建索引
	return s.createIndex(ctx)
}

// Size 返回向量数量
func (s *RedisVectorStore) Size(ctx context.Context) (int, error) {
	// FT.INFO rag_vector_idx
	result, err := s.client.Do(ctx, "FT.INFO", vectorIndexName).Result()
	if err != nil {
		return 0, err
	}

	info, ok := result.([]interface{})
	if !ok {
		return 0, fmt.Errorf("解析索引信息失败")
	}

	// 查找 num_docs 字段
	for i := 0; i < len(info)-1; i += 2 {
		if info[i].(string) == "num_docs" {
			return int(info[i+1].(int64)), nil
		}
	}

	return 0, nil
}

// Close 关闭连接
func (s *RedisVectorStore) Close() error {
	return s.client.Close()
}

// float64ToFloat32Bytes 将float64向量转为float32字节数组
func float64ToFloat32Bytes(vector []float64) []byte {
	buf := make([]byte, len(vector)*4)
	for i, v := range vector {
		bits := math.Float32bits(float32(v))
		buf[i*4] = byte(bits)
		buf[i*4+1] = byte(bits >> 8)
		buf[i*4+2] = byte(bits >> 16)
		buf[i*4+3] = byte(bits >> 24)
	}
	return buf
}
