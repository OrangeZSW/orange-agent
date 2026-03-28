package git

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/common"
	"os/exec"
)

var GitCommitTool = common.BaseTool{
	Name:        "git_commit",
	Description: "创建新的提交，自动暂存所有更改",
	Parameters: map[string]interface{}{
		"message": map[string]interface{}{
			"type":        "string",
			"description": "提交信息，描述更改内容",
		},
		"required": []string{"message"},
	},
	Call: handlerGitCommit,
}

func handlerGitCommit(ctx context.Context, input string) (string, error) {
	// 解析JSON参数
	var params struct {
		Message string `json:"message"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %v", err)
	}

	if params.Message == "" {
		return "", fmt.Errorf("commit message is required")
	}

	// 暂存所有更改
	stageCmd := exec.Command("git", "add", "-A")
	if output, err := stageCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git add failed: %v\n%s", err, string(output))
	}

	// 提交更改
	commitCmd := exec.Command("git", "commit", "-m", params.Message)
	output, err := commitCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git commit failed: %v\n%s", err, string(output))
	}

	return string(output), nil
}
