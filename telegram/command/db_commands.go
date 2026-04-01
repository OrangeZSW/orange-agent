package command

import (
	"context"
	"fmt"
	"orange-agent/agent/tools/database"
	"orange-agent/domain"
	"strings"

	"gopkg.in/telebot.v3"
)

// DbQueryCommand 数据库查询命令
type DbQueryCommand struct{}

func (d *DbQueryCommand) Command() string {
	return "db"
}

func (d *DbQueryCommand) Description() string {
	return "执行数据库查询"
}

func (d *DbQueryCommand) Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string {
	if len(args) == 0 {
		return "❌ 请提供SQL查询语句\n📝 用法: `/db <SQL查询>`\n示例: `/db SELECT * FROM users`"
	}
	
	// 合并所有参数作为SQL查询
	sqlQuery := strings.Join(args, " ")
	
	// 检查是否为SELECT查询
	upperQuery := strings.ToUpper(strings.TrimSpace(sqlQuery))
	if !strings.HasPrefix(upperQuery, "SELECT") {
		return "⚠️ *安全提示*\n\n目前只支持SELECT查询，请使用安全的查询语句。\n如需执行写操作，请使用AI助手。"
	}
	
	// 执行数据库查询操作
	result, err := database.DatabaseQuery(sqlQuery, []string{})
	if err != nil {
		return fmt.Sprintf("❌ 数据库查询失败: %v", err)
	}
	
	// 如果结果太长，进行截断
	maxLength := 1500
	if len(result) > maxLength {
		result = result[:maxLength] + "\n\n... (结果过长，已截断)"
	}
	
	var response strings.Builder
	response.WriteString("🗄️ *数据库查询结果*\n\n")
	response.WriteString(fmt.Sprintf("🔍 *查询语句:*\n```sql\n%s\n```\n\n", sqlQuery))
	
	// 检查是否有结果
	if strings.TrimSpace(result) == "" {
		response.WriteString("📭 查询成功，但未返回任何数据")
	} else {
		response.WriteString("📊 *查询结果:*\n```\n")
		response.WriteString(result)
		response.WriteString("\n```")
	}
	
	return response.String()
}

// DbExecuteCommand 数据库执行命令（需要谨慎使用）
type DbExecuteCommand struct{}

func (d *DbExecuteCommand) Command() string {
	return "dbe"
}

func (d *DbExecuteCommand) Description() string {
	return "执行数据库写操作（INSERT/UPDATE/DELETE）- 谨慎使用"
}

func (d *DbExecuteCommand) Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string {
	if len(args) == 0 {
		return "❌ 请提供SQL执行语句\n📝 用法: `/dbe <SQL语句>`\n示例: `/dbe UPDATE users SET status=1 WHERE id=1`"
	}
	
	// 合并所有参数作为SQL语句
	sqlQuery := strings.Join(args, " ")
	
	// 检查SQL类型
	upperQuery := strings.ToUpper(strings.TrimSpace(sqlQuery))
	allowedCommands := []string{"INSERT", "UPDATE", "DELETE"}
	
	isAllowed := false
	for _, cmd := range allowedCommands {
		if strings.HasPrefix(upperQuery, cmd) {
			isAllowed = true
			break
		}
	}
	
	if !isAllowed {
		return fmt.Sprintf("❌ 不支持的SQL命令: %s\n\n📋 只支持: %s", 
			strings.Split(upperQuery, " ")[0], 
			strings.Join(allowedCommands, ", "))
	}
	
	// 确认执行（实际使用中可能需要更严格的权限控制）
	response := fmt.Sprintf("⚠️ *危险操作确认*\n\n您将要执行以下SQL语句:\n```sql\n%s\n```\n\n请确保您知道自己在做什么！\n\n回复 `确认执行` 来继续。", sqlQuery)
	
	// 这里可以添加确认机制
	// 实际实现中，可以存储状态等待用户确认
	
	return response
}