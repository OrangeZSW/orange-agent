package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"orange-agent/common"
	"orange-agent/config/config"
	"orange-agent/domain"
	"time"
)

var AgentTestTool = common.BaseTool{
	Name:        "agent_test",
	Description: "测试Agent连接状态",
	Parameters: map[string]string{
		"name": "Agent名称",
	},
	Required: []string{"name"},
	Handler:  handleAgentTest,
}

func handleAgentTest(params map[string]interface{}) (string, error) {
	name, ok := params["name"].(string)
	if !ok || name == "" {
		return "", fmt.Errorf("name 参数不能为空")
	}

	var agent domain.AgentConfig
	if err := config.DB.Where("name = ?", name).First(&agent).Error; err != nil {
		return "", fmt.Errorf("Agent %s 不存在", name)
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
			"name":    name,
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
			"status":    "error",
			"message":   fmt.Sprintf("连接失败: %v", err),
			"name":      name,
			"latency_ms": elapsed.Milliseconds(),
		}
		jsonResult, _ := json.MarshalIndent(result, "", "  ")
		return string(jsonResult), nil
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result := map[string]interface{}{
			"status":     "success",
			"message":    "连接测试成功",
			"name":       name,
			"status_code": resp.StatusCode,
			"latency_ms": elapsed.Milliseconds(),
		}
		jsonResult, _ := json.MarshalIndent(result, "", "  ")
		return string(jsonResult), nil
	}

	result := map[string]interface{}{
		"status":      "error",
		"message":     fmt.Sprintf("连接测试失败，状态码: %d", resp.StatusCode),
		"name":        name,
		"status_code": resp.StatusCode,
		"response":    string(body),
		"latency_ms":  elapsed.Milliseconds(),
	}
	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonResult), nil
}
