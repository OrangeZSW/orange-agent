package system

import (
	"orange-agent/common"
)

var PerformanceMonitorTool = common.BaseTool{
	Name:        "performance_monitor",
	Description: "监控系统性能指标（CPU、内存、磁盘等）",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"metric": map[string]interface{}{
				"type":        "string",
				"description": "要监控的指标：cpu、memory、disk、all",
				"enum":        []interface{}{"cpu", "memory", "disk", "all"},
			},
			"interval": map[string]interface{}{
				"type":        "integer",
				"description": "采样间隔（秒，可选，默认为1秒）",
			},
		},
		"required": []interface{}{"metric"},
	},
}

func MonitorPerformance(metric string, interval int) (string, error) {
	if interval <= 0 {
		interval = 1
	}

	// 这里简化实现，实际应该调用系统命令或API
	var result string
	switch metric {
	case "cpu":
		result = "CPU使用率监控中..."
	case "memory":
		result = "内存使用率监控中..."
	case "disk":
		result = "磁盘使用率监控中..."
	case "all":
		result = "所有性能指标监控中..."
	default:
		return "无效的监控指标", nil
	}

	return result, nil
}
