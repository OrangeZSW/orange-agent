package system

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"orange-agent/common"
	"strings"
)

type ApiTesterTools struct {
	common.BaseTool
}

func (a *ApiTesterTools) Name() string {
	return "api_tester"
}

func (a *ApiTesterTools) Description() string {
	return "测试API接口（支持GET、POST等方法）"
}

func (a *ApiTesterTools) Call(ctx context.Context, input string) (string, error) {
	var params struct {
		URL    string `json:"url"`
		Method string `json:"method"`
		Data   string `json:"data"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %v", err)
	}

	if params.URL == "" {
		return "", fmt.Errorf("url is required")
	}

	if params.Method == "" {
		params.Method = "GET"
	}

	var req *http.Request
	var err error

	if params.Data != "" && (params.Method == "POST" || params.Method == "PUT") {
		req, err = http.NewRequest(params.Method, params.URL, strings.NewReader(params.Data))
	} else {
		req, err = http.NewRequest(params.Method, params.URL, nil)
	}

	if err != nil {
		return "", err
	}

	if params.Data != "" && (params.Method == "POST" || params.Method == "PUT") {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("Status: %s\nHeaders: %v\nBody: %s", resp.Status, resp.Header, string(body))
	return result, nil
}

func (a *ApiTesterTools) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"url": map[string]interface{}{
				"type":        "string",
				"description": "API接口URL",
			},
			"method": map[string]interface{}{
				"type":        "string",
				"description": "HTTP方法：GET、POST、PUT、DELETE等",
				"enum":        []interface{}{"GET", "POST", "PUT", "DELETE"},
			},
			"data": map[string]interface{}{
				"type":        "string",
				"description": "请求数据（JSON格式，可选）",
			},
		},
		"required": []string{"url", "method"},
	}
}
