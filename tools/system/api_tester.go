package system

import (
	"orange-agent/common"
)

var ApiTesterTool = common.BaseTool{
	Name:        "api_tester",
	Description: "测试API接口（支持GET、POST等方法）",
	Parameters: map[string]interface{}{
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
		"required": []interface{}{"url", "method"},
	},
}

func TestApi(url, method string, data string) (string, error) {
	// 这里简化实现，实际应该使用http.Client发送请求
	result := "API测试中...\nURL: " + url + "\nMethod: " + method
	if data != "" {
		result += "\nData: " + data
	}
	return result, nil
}
