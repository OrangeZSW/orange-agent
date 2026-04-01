package command

import (
	"context"
	"testing"
	
	"orange-agent/domain"
	
	"gopkg.in/telebot.v3"
)

type mockContext struct {
	telebot.Context
}

func TestCommandManager_Execute(t *testing.T) {
	// 创建测试用户
	user := &domain.User{
		ID:        1,
		TelegramID: 123456789,
		Name:      "testuser",
		ModelName: "gpt-4",
	}
	
	// 创建命令管理器（使用nil repository进行测试）
	cm := &CommandManager{
		handlers: make(map[string]CommandHandler),
		repo:     nil,
	}
	
	// 注册测试处理器
	cm.Register(&TestCommandHandler{})
	
	tests := []struct {
		name        string
		command     string
		wantContain string
		wantError   bool
	}{
		{
			name:        "有效命令",
			command:     "/test",
			wantContain: "测试命令响应",
			wantError:   false,
		},
		{
			name:        "带参数的有效命令",
			command:     "/test arg1 arg2",
			wantContain: "测试命令响应",
			wantError:   false,
		},
		{
			name:        "无效命令",
			command:     "/nonexistent",
			wantContain: "未知命令",
			wantError:   false,
		},
		{
			name:        "非命令消息",
			command:     "hello world",
			wantContain: "必须以 '/' 开头",
			wantError:   false,
		},
		{
			name:        "空命令",
			command:     "/",
			wantContain: "命令格式错误",
			wantError:   false,
		},
	}
	
	ctx := context.Background()
	mockCtx := &mockContext{}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cm.Execute(ctx, mockCtx, user, tt.command)
			
			if tt.wantError && result == "" {
				t.Errorf("期望错误但得到空结果")
			}
			
			if !contains(result, tt.wantContain) && tt.wantContain != "" {
				t.Errorf("Execute() = %v, 期望包含 %v", result, tt.wantContain)
			}
		})
	}
}

func TestCommandManager_GetAllCommands(t *testing.T) {
	cm := &CommandManager{
		handlers: make(map[string]CommandHandler),
	}
	
	// 注册多个处理器
	cm.Register(&TestCommandHandler{command: "cmd1"})
	cm.Register(&TestCommandHandler{command: "cmd2"})
	cm.Register(&TestCommandHandler{command: "cmd3"})
	
	commands := cm.GetAllCommands()
	
	if len(commands) != 3 {
		t.Errorf("期望 3 个命令，得到 %d 个", len(commands))
	}
}

func TestHelpCommand_Handle(t *testing.T) {
	cm := &CommandManager{
		handlers: make(map[string]CommandHandler),
	}
	
	helpCmd := &HelpCommand{cm: cm}
	
	// 注册一些命令用于测试
	cm.Register(&TestCommandHandler{command: "test", description: "测试命令"})
	cm.Register(&TestCommandHandler{command: "help", description: "帮助命令"})
	
	ctx := context.Background()
	mockCtx := &mockContext{}
	user := &domain.User{ID: 1}
	
	result := helpCmd.Handle(ctx, mockCtx, user, []string{})
	
	// 检查结果是否包含必要的部分
	expectedSections := []string{
		"Orange Agent 快捷命令",
		"/test - 测试命令",
		"/help - 帮助命令",
	}
	
	for _, section := range expectedSections {
		if !contains(result, section) {
			t.Errorf("帮助命令响应缺少: %s", section)
		}
	}
}

// 测试用的命令处理器
type TestCommandHandler struct {
	command     string
	description string
}

func (t *TestCommandHandler) Command() string {
	if t.command != "" {
		return t.command
	}
	return "test"
}

func (t *TestCommandHandler) Description() string {
	if t.description != "" {
		return t.description
	}
	return "测试命令"
}

func (t *TestCommandHandler) Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string {
	return "测试命令响应 - 参数: " + stringSliceToString(args)
}

// 辅助函数
func contains(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && (s == substr || containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func stringSliceToString(slice []string) string {
	result := ""
	for i, s := range slice {
		if i > 0 {
			result += " "
		}
		result += s
	}
	return result
}