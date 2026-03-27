package mysql

import (
	"orange-agent/domain"

	"gorm.io/gorm"
)

type AgentConfigSql struct {
	db *gorm.DB
}

func NewAgentConfigSql() *AgentConfigSql {
	return &AgentConfigSql{
		db: GetDB(),
	}
}

func (a *AgentConfigSql) GetAgentConfigByName(name string) (*domain.AgentConfig, error) {
	var agentConfig domain.AgentConfig
	err := a.db.Where("name = ?", name).First(&agentConfig).Error
	if err != nil {
		return nil, err
	}
	return &agentConfig, nil
}

// getBymodel
func (a *AgentConfigSql) GetAgentConfigByModel(model string) (*domain.AgentConfig, error) {
	var agentConfig domain.AgentConfig
	err := a.db.Where("models LIKE ?", "%"+model+"%").First(&agentConfig).Error
	if err != nil {
		return nil, err
	}
	return &agentConfig, nil
}

// create
func (a *AgentConfigSql) CreateAgentConfig(agentConfig *domain.AgentConfig) error {
	return a.db.Create(agentConfig).Error
}

// update
func (a *AgentConfigSql) UpdateAgentConfig(agentConfig *domain.AgentConfig) error {
	return a.db.Save(agentConfig).Error
}

// getById
func (a *AgentConfigSql) GetAgentConfigById(id uint) (*domain.AgentConfig, error) {
	var agentConfig domain.AgentConfig
	err := a.db.Where("id = ?", id).First(&agentConfig).Error
	if err != nil {
		return nil, err
	}
	return &agentConfig, nil
}

// get All
func (a *AgentConfigSql) GetAllAgentConfig() (*[]domain.AgentConfig, error) {
	var agentConfigs []domain.AgentConfig
	err := a.db.Find(&agentConfigs).Error
	if err != nil {
		return nil, err
	}
	return &agentConfigs, nil
}
