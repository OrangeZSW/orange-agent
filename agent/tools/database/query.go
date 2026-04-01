package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"orange-agent/common"
	"orange-agent/config"
	"orange-agent/utils/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DatabaseQueryTool = common.BaseTool{
	Name:        "database_query",
	Description: "执行数据库查询操作，支持SELECT语句",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "SQL查询语句",
			},
			"args": map[string]interface{}{
				"type":        "array",
				"description": "查询参数（可选）",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
		},
		"required": []string{"query"},
	},
	Call: handlerDatabaseQuery,
}

var DatabaseExecuteTool = common.BaseTool{
	Name:        "database_execute",
	Description: "执行数据库写操作，支持INSERT、UPDATE、DELETE语句",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "SQL执行语句",
			},
			"args": map[string]interface{}{
				"type":        "array",
				"description": "执行参数（可选）",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
		},
		"required": []string{"query"},
	},
	Call: handlerDatabaseExecute,
}

// 获取数据库连接
func getDB() (*gorm.DB, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %v", err)
	}

	dsn := cfg.Database.Username + ":" + cfg.Database.Password + "@tcp(" + cfg.Database.Host + ":" + cfg.Database.Port + ")/" + cfg.Database.Database + "?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	return db, nil
}

func handlerDatabaseQuery(ctx context.Context, input string) (string, error) {
	log := logger.GetLogger()
	
	// 解析JSON参数
	var params struct {
		Query string   `json:"query"`
		Args  []string `json:"args,omitempty"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %v", err)
	}

	if params.Query == "" {
		return "", fmt.Errorf("query is required")
	}

	// 获取数据库连接
	db, err := getDB()
	if err != nil {
		return "", err
	}

	// 转换参数为[]interface{}
	args := make([]interface{}, len(params.Args))
	for i, arg := range params.Args {
		args[i] = arg
	}

	// 执行查询
	rows, err := db.Raw(params.Query, args...).Rows()
	if err != nil {
		log.Error("Database query failed: %v", err)
		return "", fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()

	// 获取列名
	columns, err := rows.Columns()
	if err != nil {
		log.Error("Failed to get columns: %v", err)
		return "", fmt.Errorf("failed to get columns: %v", err)
	}

	// 准备结果
	var results []map[string]interface{}
	for rows.Next() {
		// 创建切片来存储列值
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// 扫描行数据
		if err := rows.Scan(valuePtrs...); err != nil {
			log.Error("Failed to scan row: %v", err)
			return "", fmt.Errorf("failed to scan row: %v", err)
		}

		// 创建行数据映射
		rowData := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			
			// 处理[]byte类型
			if b, ok := val.([]byte); ok {
				rowData[col] = string(b)
			} else {
				rowData[col] = val
			}
		}
		results = append(results, rowData)
	}

	// 检查行遍历错误
	if err := rows.Err(); err != nil {
		log.Error("Row iteration error: %v", err)
		return "", fmt.Errorf("row iteration error: %v", err)
	}

	// 转换结果为JSON
	jsonResult, err := json.Marshal(results)
	if err != nil {
		log.Error("Failed to marshal results: %v", err)
		return "", fmt.Errorf("failed to marshal results: %v", err)
	}

	return string(jsonResult), nil
}

func handlerDatabaseExecute(ctx context.Context, input string) (string, error) {
	log := logger.GetLogger()
	
	// 解析JSON参数
	var params struct {
		Query string   `json:"query"`
		Args  []string `json:"args,omitempty"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %v", err)
	}

	if params.Query == "" {
		return "", fmt.Errorf("query is required")
	}

	// 获取数据库连接
	db, err := getDB()
	if err != nil {
		return "", err
	}

	// 转换参数为[]interface{}
	args := make([]interface{}, len(params.Args))
	for i, arg := range params.Args {
		args[i] = arg
	}

	// 执行SQL
	result := db.Exec(params.Query, args...)
	if result.Error != nil {
		log.Error("Database execute failed: %v", result.Error)
		return "", fmt.Errorf("execute failed: %v", result.Error)
	}

	// 获取影响的行数
	rowsAffected := result.RowsAffected
	
	// 尝试获取最后插入ID（如果是INSERT操作）
	var lastInsertId int64
	// 注意：gorm的Exec结果不直接提供LastInsertId，我们需要使用原生数据库连接
	sqlDB, err := db.DB()
	if err == nil {
		// 对于MySQL，我们可以尝试使用原生查询获取最后插入ID
		// 但这需要在同一事务中，这里我们简单处理
		lastInsertId = 0 // 简化处理，实际使用可能需要更复杂的逻辑
	}

	// 构建结果
	executeResult := map[string]interface{}{
		"rows_affected": rowsAffected,
		"last_insert_id": lastInsertId,
		"success": true,
	}

	// 转换结果为JSON
	jsonResult, err := json.Marshal(executeResult)
	if err != nil {
		log.Error("Failed to marshal result: %v", err)
		return "", fmt.Errorf("failed to marshal result: %v", err)
	}

	return string(jsonResult), nil
}
