package command

import (
	"context"
	"fmt"
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
	
	query := strings.Join(args, " ")
	
	// 执行数据库查询操作
	result, err := executeTool("database_query", map[string]interface{}{
		"query": query,
	})
	if err != nil {
		return fmt.Sprintf("❌ 数据库查询失败: %v", err)
	}
	
	return fmt.Sprintf("🗄️ *数据库查询结果*\n\n```json\n%s\n```", result)
}

// DbExecuteCommand 数据库执行命令
type DbExecuteCommand struct{}

func (d *DbExecuteCommand) Command() string {
	return "dbe"
}

func (d *DbExecuteCommand) Description() string {
	return "执行数据库写操作（INSERT/UPDATE/DELETE）"
}

func (d *DbExecuteCommand) Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string {
	if len(args) == 0 {
		return "❌ 请提供SQL执行语句\n📝 用法: `/dbe <SQL语句>`\n示例: `/dbe INSERT INTO users (name) VALUES ('test')`"
	}
	
	query := strings.Join(args, " ")
	
	// 确认执行（危险操作）
	response := fmt.Sprintf("⚠️ *危险操作确认*\n\n您将要执行以下SQL语句：\n\n```sql\n%s\n```\n\n此操作可能修改数据，请确认！\n\n回复 `确认执行` 来继续。", query)
	
	return response
}