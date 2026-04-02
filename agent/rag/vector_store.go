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
	// Redis文件元数据前缀
	fileMetaPrefix = "rag:file:"
)

// VectorStore 向量存储接口
type VectorStore interface {
	// Add 添加向量
	Add(ctx context.Context, id string, vector []float64, chunk Chunk) error
	// Search 搜索相似向量
	Search(ctx context.Context, queryVector []float64, topK int) ([]SearchResult, error)
	// Clear 清空所有数据
	Clear(ctx context.Context) error
	// Size 返回向量数量
	Size(ctx context.Context) (int, error)
	// Close 关闭连接
	Close() error
	// DeleteByFilePath 删除指定文件的所有向量
	DeleteByFilePath(ctx context.Context, filePath string) error
	// SetFileMeta 设置文件元数据（修改时间）
	SetFileMeta(ctx context.Context, filePath string, modTime int64) error
	// GetFileMeta 获取文件元数据
	GetFileMeta(ctx context.Context, filePath string) (int64, error)
	// GetAllIndexedFiles 获取所有已索引文件
	GetAllIndexedFiles(ctx context.Context) (map[string]int64, error)
	// DeleteFileMeta 删除文件元数据
	DeleteFileMeta(ctx context.Context, filePath string) error
}

// SearchResult 搜索结果
type SearchResult struct {
	Chunk Chunk
	Score float64
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// RedisVectorStore Redis向量存储
type RedisVectorStore struct {
	client    *redis.Client
	vectorDim int
	mu        sync.RWMutex
	log       *logger.Logger
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

	// 创建向量索引（如果不存在）
	if err := store.ensureIndex(ctx); err != nil {
		log.Warn("创建向量索引失败: %v", err)
	}

	return store, nil
}

// ensureIndex 确保索引存在
func (s *RedisVectorStore) ensureIndex(ctx context.Context) error {
	// 检查索引是否存在
	_, err := s.client.Do(ctx, "FT.INFO", vectorIndexName).Result()
	if err == nil {
		return nil // 索引已存在
	}

	// 创建索引
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

	s.log.Info("RediSearch向量索引创建成功")
	return nil
}

// Add 添加向量
func (s *RedisVectorStore) Add(ctx context.Context, id string, vector []float64, chunk Chunk) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vectorBytes := float64ToFloat32Bytes(vector)
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

	queryBytes := float64ToFloat32Bytes(queryVector)

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

// parseSearchResult 解析搜索结果
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

	// 删除索引
	s.client.Do(ctx, "FT.DROPINDEX", vectorIndexName, "DD")

	// 删除所有rag:doc:* 和 rag:file:* 的key
	iter := s.client.Scan(ctx, 0, "rag:*", 0).Iterator()
	for iter.Next(ctx) {
		s.client.Del(ctx, iter.Val())
	}

	// 重新创建索引
	return s.ensureIndex(ctx)
}

// DeleteByFilePath 删除指定文件的所有向量
func (s *RedisVectorStore) DeleteByFilePath(ctx context.Context, filePath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 使用FT.SEARCH查找该文件的所有chunk
	cmd := []interface{}{
		"FT.SEARCH", vectorIndexName,
		fmt.Sprintf("@path:{%s}", filePath),
		"RETURN", "1", "chunk_id",
		"NOCONTENT",
	}

	result, err := s.client.Do(ctx, cmd...).Result()
	if err != nil {
		return err
	}

	results, ok := result.([]interface{})
	if !ok || len(results) < 2 {
		return nil
	}

	// 删除所有找到的chunk
	for i := 1; i < len(results); i++ {
		key := results[i].(string)
		s.client.Del(ctx, key)
	}

	return nil
}

// SetFileMeta 设置文件元数据
func (s *RedisVectorStore) SetFileMeta(ctx context.Context, filePath string, modTime int64) error {
	key := fileMetaPrefix + filePath
	return s.client.Set(ctx, key, modTime, 0).Err()
}

// GetFileMeta 获取文件元数据
func (s *RedisVectorStore) GetFileMeta(ctx context.Context, filePath string) (int64, error) {
	key := fileMetaPrefix + filePath
	val, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(val, 10, 64)
}

// DeleteFileMeta 删除文件元数据
func (s *RedisVectorStore) DeleteFileMeta(ctx context.Context, filePath string) error {
	key := fileMetaPrefix + filePath
	return s.client.Del(ctx, key).Err()
}

// GetAllIndexedFiles 获取所有已索引文件及其修改时间
func (s *RedisVectorStore) GetAllIndexedFiles(ctx context.Context) (map[string]int64, error) {
	result := make(map[string]int64)

	iter := s.client.Scan(ctx, 0, fileMetaPrefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		filePath := key[len(fileMetaPrefix):]

		val, err := s.client.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		modTime, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			continue
		}

		result[filePath] = modTime
	}

	return result, iter.Err()
}

// Size 返回向量数量
func (s *RedisVectorStore) Size(ctx context.Context) (int, error) {
	result, err := s.client.Do(ctx, "FT.INFO", vectorIndexName).Result()
	if err != nil {
		return 0, err
	}

	info, ok := result.([]interface{})
	if !ok {
		return 0, fmt.Errorf("解析索引信息失败")
	}

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
