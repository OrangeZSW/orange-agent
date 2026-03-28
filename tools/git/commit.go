package git

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/common"
	"os/exec"
)

type GitCommit struct {
	common.BaseTool
}

func (g *GitCommit) Name() string {
	return "git_commit"
}

func (g *GitCommit) Description() string {
	return "Create a new commit with the specified message. Automatically stages all changes before committing."
}

func (g *GitCommit) Call(ctx context.Context, input string) (string, error) {
	var params struct {
		Message string `json:"message"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %v", err)
	}

	if params.Message == "" {
		return "", fmt.Errorf("commit message is required")
	}

	// Stage all changes
	stageCmd := exec.Command("git", "add", "-A")
	if output, err := stageCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git add failed: %v\n%s", err, string(output))
	}

	// Commit changes
	commitCmd := exec.Command("git", "commit", "-m", params.Message)
	output, err := commitCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git commit failed: %v\n%s", err, string(output))
	}

	return string(output), nil
}

func (g *GitCommit) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Required: The commit message describing the changes.",
			},
		},
		"required": []string{"message"},
	}
}
