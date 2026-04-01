package utils

import (
	"testing"
)

func TestToolMessageFormatter(t *testing.T) {
	formatter := NewToolMessageFormatter()

	tests := []struct {
		name     string
		toolName string
		args     string
		result   string
		err      string
	}{
		{
			name:     "文件读取工具调用",
			toolName: "file_read",
			args:     `{"file_path": "agent/task/orchestrator.go"}`,
			result:   "文件内容...",
		},
		{
			name:     "构建工具调用",
			toolName: "build_tools",
			args:     "{}",
			result:   "构建成功",
		},
		{
			name:     "带复杂参数的工具调用",
			toolName: "database_query",
			args:     `{"query": "SELECT * FROM users WHERE age > 18", "args": ["18"]}`,
			result:   `[{"id": 1, "name": "Alice", "age": 25}, {"id": 2, "name": "Bob", "age": 30}]`,
		},
		{
			name:     "工具调用失败",
			toolName: "file_read",
			args:     `{"file_path": "nonexistent.txt"}`,
			err:      "文件不存在",
		},
		{
			name:     "空参数工具",
			toolName: "test_run",
			args:     "",
			result:   "测试通过",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试工具调用消息格式化
			callMsg := formatter.FormatToolCallMessage(tt.toolName, tt.args)
			t.Logf("工具调用消息:\n%s\n", callMsg)

			if tt.err != "" {
				// 测试错误消息格式化
				errorMsg := formatter.FormatToolErrorMessage(tt.toolName, tt.args, tt.err)
				t.Logf("工具调用错误消息:\n%s\n", errorMsg)
			} else {
				// 测试成功消息格式化
				successMsg := formatter.FormatToolSuccessMessage(tt.toolName, tt.args, tt.result)
				t.Logf("工具调用成功消息:\n%s\n", successMsg)
			}
		})
	}
}

func TestPrettifyToolName(t *testing.T) {
	formatter := NewToolMessageFormatter()
	
	tests := []struct {
		input    string
		expected string
	}{
		{"file_read", "File Read"},
		{"build_tools", "Build Tools"},
		{"database_query", "Database Query"},
		{"test_run", "Test Run"},
		{"api_tester", "Api Tester"},
		{"single", "Single"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// 使用反射访问私有方法
			result := formatter.prettifyToolName(tt.input)
			if result != tt.expected {
				t.Errorf("prettifyToolName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPrettifyArguments(t *testing.T) {
	formatter := NewToolMessageFormatter()
	
	tests := []struct {
		name     string
		args     string
		contains []string
	}{
		{
			name: "空参数",
			args: "",
			contains: []string{"无参数"},
		},
		{
			name: "空对象参数",
			args: "{}",
			contains: []string{"无参数"},
		},
		{
			name: "JSON参数",
			args: `{"file_path": "test.txt", "mode": "read"}`,
			contains: []string{"```json", "file_path", "test.txt", "mode", "read"},
		},
		{
			name: "复杂JSON参数",
			args: `{"query": "SELECT * FROM users", "args": ["1", "2"], "limit": 10}`,
			contains: []string{"```json", "query", "SELECT", "args", "limit"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.prettifyArguments(tt.args)
			for _, contain := range tt.contains {
				if !containsString(result, contain) {
					t.Errorf("prettifyArguments(%q) 应该包含 %q, 实际结果:\n%s", tt.args, contain, result)
				}
			}
		})
	}
}

func TestPrettifyResult(t *testing.T) {
	formatter := NewToolMessageFormatter()
	
	tests := []struct {
		name     string
		result   string
		contains []string
	}{
		{
			name:     "空结果",
			result:   "",
			contains: []string{"无输出"},
		},
		{
			name:     "短文本结果",
			result:   "操作成功",
			contains: []string{"```", "操作成功"},
		},
		{
			name:     "JSON结果",
			result:   `{"status": "success", "data": {"id": 1}}`,
			contains: []string{"```json", "status", "success", "data"},
		},
		{
			name:     "长文本截断",
			result:   "这是一个非常长的结果" + repeatString("测试", 200),
			contains: []string{"...", "```"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.prettifyResult(tt.result)
			for _, contain := range tt.contains {
				if !containsString(result, contain) {
					t.Errorf("prettifyResult(%q) 应该包含 %q, 实际结果:\n%s", tt.result, contain, result)
				}
			}
		})
	}
}

// 辅助函数
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[0:len(substr)] == substr || containsString(s[1:], substr)))
}

func repeatString(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}