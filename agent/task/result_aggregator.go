package task

import (
	"sync"
	"orange-agent/domain"
	"orange-agent/utils/logger"
)

// ResultAggregator 聚合所有子任务的结果
type ResultAggregator struct {
	subTasks   []*domain.SubTask
	results    map[uint]*domain.SubTask
	mu         sync.Mutex
	totalCount int
	completed  int
	failed     int
}

// NewResultAggregator 创建新的结果聚合器
func NewResultAggregator(subTasks []*domain.SubTask) *ResultAggregator {
	return &ResultAggregator{
		subTasks:   subTasks,
		results:    make(map[uint]*domain.SubTask),
		totalCount: len(subTasks),
		completed:  0,
		failed:     0,
	}
}

// AddResult 添加子任务结果
func (ra *ResultAggregator) AddResult(subTask *domain.SubTask) {
	ra.mu.Lock()
	defer ra.mu.Unlock()
	
	ra.results[subTask.ID] = subTask
	
	if subTask.Status == domain.StatusCompleted {
		ra.completed++
		logger.Info("子任务完成: %s (ID: %d)", subTask.Description, subTask.ID)
	} else if subTask.Status == domain.StatusFailed {
		ra.failed++
		logger.Warn("子任务失败: %s (ID: %d), 错误: %s", subTask.Description, subTask.ID, subTask.Error)
	}
	
	logger.Debug("任务进度: %d/%d 完成, %d/%d 失败", ra.completed, ra.totalCount, ra.failed, ra.totalCount)
}

// IsComplete 检查是否所有子任务都已处理
func (ra *ResultAggregator) IsComplete() bool {
	ra.mu.Lock()
	defer ra.mu.Unlock()
	return ra.completed+ra.failed >= ra.totalCount
}

// GetResults 获取所有结果
func (ra *ResultAggregator) GetResults() map[uint]*domain.SubTask {
	ra.mu.Lock()
	defer ra.mu.Unlock()
	
	// 返回副本
	results := make(map[uint]*domain.SubTask)
	for id, subTask := range ra.results {
		results[id] = subTask
	}
	return results
}

// GetSummary 获取聚合摘要
func (ra *ResultAggregator) GetSummary() *AggregationSummary {
	ra.mu.Lock()
	defer ra.mu.Unlock()
	
	return &AggregationSummary{
		Total:     ra.totalCount,
		Completed: ra.completed,
		Failed:    ra.failed,
		Results:   ra.getResultsList(),
	}
}

// getResultsList 获取结果列表
func (ra *ResultAggregator) getResultsList() []*SubTaskResult {
	var results []*SubTaskResult
	for _, subTask := range ra.results {
		results = append(results, &SubTaskResult{
			ID:          subTask.ID,
			Description: subTask.Description,
			Status:      string(subTask.Status),
			Output:      subTask.Output,
			Error:       subTask.Error,
		})
	}
	return results
}

// AggregationSummary 聚合摘要
type AggregationSummary struct {
	Total     int              `json:"total"`
	Completed int              `json:"completed"`
	Failed    int              `json:"failed"`
	Results   []*SubTaskResult `json:"results"`
}

// SubTaskResult 子任务结果
type SubTaskResult struct {
	ID          uint   `json:"id"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Output      string `json:"output"`
	Error       string `json:"error"`
}
