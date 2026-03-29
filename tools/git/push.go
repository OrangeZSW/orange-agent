package git

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/common"
	"os/exec"
)

var GitPushTool = common.BaseTool{
	Name:        "git_push",
	Description: "推送本地提交到远程仓库，支持推送到指定分支",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"branch": map[string]interface{}{
				"type":        "string",
				"description": "可选：要推送的具体分支，不提供则推送所有跟踪的分支",
			},
			"force": map[string]interface{}{
				"type":        "boolean",
				"description": "可选：如果为 true，执行强制推送，请谨慎使用",
			},
		},
	},
	Call: handlerGitPush,
}

func handlerGitPush(ctx context.Context, input string) (string, error) {
	// 解析JSON参数
	var params struct {
		Branch string `json:"branch"`
		Force  bool   `json:"force"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %v", err)
	}

	cmdArgs := []string{"push"}
	if params.Force {
		cmdArgs = append(cmdArgs, "-f")
	}
	if params.Branch != "" {
		cmdArgs = append(cmdArgs, "origin", params.Branch)
	} else {
		cmdArgs = append(cmdArgs, "origin")
	}

	cmd := exec.Command("git", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git push failed: %v\n%s", err, string(output))
	}

	return string(output), nil
}
