package database

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/common"
	"orange-agent/repository/resource"
	"orange-agent/utils/logger"
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

func handlerDatabaseQuery(ctx context.Context, input string) (string, error) {
	log := logger.GetLogger()
	repo := resource.GetRepositories()

	// 修复：检查数据库实例是否初始化，避免空指针panic
	if repo == nil || repo.SqlQuery == nil {
		return "", fmt.Errorf("database repository not initialized")
	}

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

	// 转换参数为[]interface{}
	args := make([]interface{}, len(params.Args))
	for i, arg := range params.Args {
		args[i] = arg
	}

	// 执行查询
	rows, err := repo.SqlQuery.ExecuteRows(params.Query, args...)
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
	repo := resource.GetRepositories()

	// 修复：检查数据库实例是否初始化，避免空指针panic
	if repo == nil || repo.SqlQuery == nil {
		return "", fmt.Errorf("database repository not initialized")
	}

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

	// 转换参数为[]interface{}
	args := make([]interface{}, len(params.Args))
	for i, arg := range params.Args {
		args[i] = arg
	}

	// 执行SQL
	result := repo.SqlQuery.Execute(params.Query, args...)
	// 修复：检查返回结果是否为空
	if result == nil {
		return "", fmt.Errorf("execute returned nil result")
	}
	if result.Error != nil {
		log.Error("Database execute failed: %v", result.Error)
		return "", fmt.Errorf("execute failed: %v", result.Error)
	}

	// 获取影响的行数
	rowsAffected := result.RowsAffected

	// 尝试获取最后插入ID（如果是INSERT操作）
	var lastInsertId int64
	// 可选：如果result支持LastInsertId方法则添加调用
	// lastInsertId, err = result.LastInsertId()
	// if err != nil {
	//     log.Debug("Failed to get last insert id: %v", err)
	//     lastInsertId = 0
	// }

	// 构建结果
	executeResult := map[string]interface{}{
		"rows_affected":  rowsAffected,
		"last_insert_id": lastInsertId,
		"success":        true,
	}

	// 转换结果为JSON
	jsonResult, err := json.Marshal(executeResult)
	if err != nil {
		log.Error("Failed to marshal result: %v", err)
		return "", fmt.Errorf("failed to marshal result: %v", err)
	}

	return string(jsonResult), nil
}