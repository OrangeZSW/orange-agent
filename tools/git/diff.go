package git

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/common"
	"os/exec"
)

type GitDiff struct {
	common.BaseTool
}

func (g *GitDiff) Name() string {
	return "git_diff"
}

func (g *GitDiff) Description() string {
	return "Execute git diff command to show changes in the repository. Supports file-specific or staged/unstaged diffs."
}

func (g *GitDiff) Call(ctx context.Context, input string) (string, error) {
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

func (g *GitDiff) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file": map[string]interface{}{
				"type":        "string",
				"description": "Optional: specific file to diff. If not provided, shows all unstaged changes.",
			},
			"staged": map[string]interface{}{
				"type":        "boolean",
				"description": "Optional: if true, shows staged changes (equivalent to --cached). Default is false.",
			},
		},
		"required": []string{},
	}
}
