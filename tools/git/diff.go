package git

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/common"
	"os/exec"
)

var GitDiffTool = common.BaseTool{
	Name:        "git_diff",
	Description: "执行 git diff 命令，显示仓库中的更改，支持指定文件或暂存区/未暂存区的差异",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file": map[string]interface{}{
				"type":        "string",
				"description": "可选：要显示差异的具体文件，不提供则显示所有未暂存的更改",
			},
			"staged": map[string]interface{}{
				"type":        "boolean",
				"description": "可选：如果为 true，显示已暂存的更改（相当于 --cached），默认为 false",
			},
		},
	},
	Call: handlerGitDiff,
}

func handlerGitDiff(ctx context.Context, input string) (string, error) {
	// 解析JSON参数
	var params struct {
		File   string `json:"file"`
		Staged bool   `json:"staged"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %v", err)
	}

	cmdArgs := []string{"diff"}
	if params.Staged {
		cmdArgs = append(cmdArgs, "--cached")
	}
	if params.File != "" {
		cmdArgs = append(cmdArgs, params.File)
	}

	cmd := exec.Command("git", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git diff failed: %v\n%s", err, string(output))
	}

	return string(output), nil
}
