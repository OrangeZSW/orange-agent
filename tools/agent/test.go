package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"orange-agent/common"
	"orange-agent/domain"
	"orange-agent/mysql"
	"time"
)

var AgentTestTool = common.BaseTool{
	Name:        "agent_test",
	Description: "测试Agent连接状态",
	Parameters: map[string]interface{}{
		"name": map[string]interface{}{
			"type":        "string",
			"description": "Agent名称",
		},
		"required": []string{"name"},
	},
	Call: handlerAgentTest,
}

func handlerAgentTest(ctx context.Context, input string) (string, error) {
	// 解析JSON参数
	var params struct {
		Name string `json:"name"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %v", err)
	}

	if params.Name == "" {
		return "", fmt.Errorf("name is required")
	}

	var agent domain.AgentConfig
	if err := mysql.GetDB().WithContext(ctx).Where("name = ?", params.Name).First(&agent).Error; err != nil {
		return "", fmt.Errorf("Agent %s 不存在", params.Name)
	}

	// 检查是否有模型
	if len(agent.Models) == 0 {
		result := map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Agent %s 没有配置模型", params.Name),
			"name":    params.Name,
		}
		jsonResult, _ := json.MarshalIndent(result, "", "  ")
		return string(jsonResult), nil
	}

	// 创建测试请求
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 尝试发送简单的测试请求
	testBody := map[string]interface{}{
		"model": agent.Models[0],
		"messages": []map[string]string{
			{"role": "user", "content": "Hello"},
		},
		"max_tokens": 5,
	}

	jsonBody, _ := json.Marshal(testBody)
	req, err := http.NewRequest("POST", agent.BaseUrl+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		result := map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("创建请求失败: %v", err),
			"name":    params.Name,
		}
		jsonResult, _ := json.MarshalIndent(result, "", "  ")
		return string(jsonResult), nil
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+agent.Token)

	startTime := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(startTime)

	if err != nil {
		result := map[string]interface{}{
			"status":     "error",
			"message":    fmt.Sprintf("连接失败: %v", err),
			"name":       params.Name,
			"latency_ms": elapsed.Milliseconds(),
		}
		jsonResult, _ := json.MarshalIndent(result, "", "  ")
		return string(jsonResult), nil
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result := map[string]interface{}{
			"status":      "success",
			"message":     "连接测试成功",
			"name":        params.Name,
			"status_code": resp.StatusCode,
			"latency_ms":  elapsed.Milliseconds(),
		}
		jsonResult, _ := json.MarshalIndent(result, "", "  ")
		return string(jsonResult), nil
	}

	result := map[string]interface{}{
		"status":      "error",
		"message":     fmt.Sprintf("连接测试失败，状态码: %d", resp.StatusCode),
		"name":        params.Name,
		"status_code": resp.StatusCode,
		"response":    string(body),
		"latency_ms":  elapsed.Milliseconds(),
	}
	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonResult), nil
}
