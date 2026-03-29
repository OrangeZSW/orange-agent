package system

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/common"
)

var PerformanceMonitorTool = common.BaseTool{
	Name:        "performance_monitor",
	Description: "监控系统性能指标（CPU、内存、磁盘等）",
	Parameters: map[string]interface{}{
		"metric": map[string]interface{}{
			"type":        "string",
			"description": "要监控的指标：cpu、memory、disk、all",
			"enum":        []interface{}{"cpu", "memory", "disk", "all"},
		},
		"interval": map[string]interface{}{
			"type":        "integer",
			"description": "采样间隔（秒，可选，默认为1秒）",
		},
		"required": []string{"metric"},
	},
	Call: handlerPerformanceMonitor,
}

func handlerPerformanceMonitor(ctx context.Context, input string) (string, error) {
	// 解析JSON参数
	var params struct {
		Metric   string `json:"metric"`
		Interval int    `json:"interval"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %v", err)
	}

	if params.Metric == "" {
		return "", fmt.Errorf("metric is required")
	}

	interval := params.Interval
	if interval <= 0 {
		interval = 1
	}

	var result string
	switch params.Metric {
	case "cpu":
		result = "CPU使用率监控中..."
	case "memory":
		result = "内存使用率监控中..."
	case "disk":
		result = "磁盘使用率监控中..."
	case "all":
		result = "所有性能指标监控中..."
	default:
		return "", fmt.Errorf("无效的监控指标: %s", params.Metric)
	}

	return result, nil
}
