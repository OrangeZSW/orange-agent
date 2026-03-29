package gorm

import (
	"errors"
	"orange-agent/domain"
	"orange-agent/repository"

	"gorm.io/gorm"
)

type agentConfigRepository struct {
	db *gorm.DB
}

// 构造函数：接收 db 参数，便于依赖注入
func NewAgentConfigRepository(db *gorm.DB) repository.AgentConfigRepository {
	return &agentConfigRepository{
		db: db,
	}
}

func (r *agentConfigRepository) DeleteAgentConfig(agentConfig *domain.AgentConfig) error {
	if agentConfig == nil {
		return errors.New("agentConfig cannot be nil")
	}
	return r.db.Delete(agentConfig).Error
}

// GetAgentConfigByName 根据名称获取 Agent 配置
func (r *agentConfigRepository) GetAgentConfigByName(name string) (*domain.AgentConfig, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	var agentConfig domain.AgentConfig
	err := r.db.Where("name = ?", name).First(&agentConfig).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &agentConfig, nil
}

// GetAgentConfigByModel 根据模型名称获取 Agent 配置
func (r *agentConfigRepository) GetAgentConfigByModel(modelName string) (*domain.AgentConfig, error) {
	if modelName == "" {
		return nil, errors.New("modelName cannot be empty")
	}

	var agentConfig domain.AgentConfig
	err := r.db.Where("models LIKE ?", "%"+modelName+"%").First(&agentConfig).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &agentConfig, nil
}

// CreateAgentConfig 创建 Agent 配置
func (r *agentConfigRepository) CreateAgentConfig(agentConfig *domain.AgentConfig) error {
	if agentConfig == nil {
		return errors.New("agentConfig cannot be nil")
	}
	return r.db.Create(agentConfig).Error
}

// UpdateAgentConfig 更新 Agent 配置
func (r *agentConfigRepository) UpdateAgentConfig(agentConfig *domain.AgentConfig) error {
	if agentConfig == nil {
		return errors.New("agentConfig cannot be nil")
	}
	return r.db.Save(agentConfig).Error
}

// GetAgentConfigById 根据 ID 获取 Agent 配置
func (r *agentConfigRepository) GetAgentConfigById(id uint) (*domain.AgentConfig, error) {
	if id == 0 {
		return nil, errors.New("id cannot be zero")
	}

	var agentConfig domain.AgentConfig
	err := r.db.Where("id = ?", id).First(&agentConfig).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &agentConfig, nil
}

// GetAllAgentConfig 获取所有 Agent 配置
func (r *agentConfigRepository) GetAllAgentConfig() ([]domain.AgentConfig, error) {
	var agentConfigs []domain.AgentConfig
	err := r.db.Find(&agentConfigs).Error
	if err != nil {
		return []domain.AgentConfig{}, err
	}
	return agentConfigs, nil
}
