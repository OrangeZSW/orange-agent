package git

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/common"
	"os/exec"
)

type GitPush struct {
	common.BaseTool
}

func (g *GitPush) Name() string {
	return "git_push"
}

func (g *GitPush) Description() string {
	return "Push local commits to the remote repository. Supports pushing to a specific branch or all branches."
}

func (g *GitPush) Call(ctx context.Context, input string) (string, error) {
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

func (g *GitPush) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"branch": map[string]interface{}{
				"type":        "string",
				"description": "Optional: specific branch to push. If not provided, pushes all tracked branches.",
			},
			"force": map[string]interface{}{
				"type":        "boolean",
				"description": "Optional: if true, performs a force push. Use with caution.",
			},
		},
		"required": []string{},
	}
}
