package system

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/common"
	"os/exec"
	"strings"
	"time"
)

var DependencyCheckTool = common.BaseTool{
	Name:        "dependency_check",
	Description: "检查 Go 项目的依赖包版本、更新情况和安全漏洞",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"check_outdated": map[string]interface{}{
				"type":        "boolean",
				"description": "是否检查过时的依赖（需要联网）",
				"default":     false,
			},
			"check_vulns": map[string]interface{}{
				"type":        "boolean",
				"description": "是否检查安全漏洞（需要安装 govulncheck）",
				"default":     false,
			},
		},
		"required": []string{}, // 至少需要一个为 true
	},
	Call: handlerDependencyCheck,
}

func handlerDependencyCheck(ctx context.Context, input string) (string, error) {
	// 解析JSON参数
	var params struct {
		CheckOutdated bool `json:"check_outdated"`
		CheckVulns    bool `json:"check_vulns"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %v", err)
	}

	// 至少需要执行一项检查
	if !params.CheckOutdated && !params.CheckVulns {
		return "", fmt.Errorf("至少需要指定 check_outdated 或 check_vulns 为 true")
	}

	var output strings.Builder
	var hasError bool

	// 检查过时依赖
	if params.CheckOutdated {
		if err := checkOutdatedDependencies(ctx, &output); err != nil {
			hasError = true
			output.WriteString(fmt.Sprintf("❌ 过时依赖检查失败: %v\n", err))
		}
	}

	// 检查安全漏洞
	if params.CheckVulns {
		if err := checkVulnerabilities(ctx, &output); err != nil {
			hasError = true
			output.WriteString(fmt.Sprintf("❌ 安全漏洞检查失败: %v\n", err))
		}
	}

	result := output.String()

	// 如果所有检查都失败，返回错误
	if hasError && result == "" {
		return "", fmt.Errorf("所有检查都失败了")
	}

	// 如果部分成功，返回结果并附带警告
	if hasError {
		return result, fmt.Errorf("部分检查失败，请查看详细信息")
	}

	return result, nil
}

// checkOutdatedDependencies 检查过时的依赖
func checkOutdatedDependencies(ctx context.Context, output *strings.Builder) error {
	// 设置超时时间
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "list", "-m", "-u", "all")
	out, err := cmd.CombinedOutput()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("命令执行超时（30秒）")
		}
		return fmt.Errorf("执行命令失败: %v, 输出: %s", err, string(out))
	}

	// 解析输出，只显示有更新的依赖
	lines := strings.Split(string(out), "\n")
	var outdated []string

	for _, line := range lines {
		// 格式: "module version [update available]"
		if strings.Contains(line, "[") && strings.Contains(line, "]") {
			outdated = append(outdated, line)
		}
	}

	if len(outdated) > 0 {
		output.WriteString("📦 过时依赖（有更新可用）:\n")
		for _, dep := range outdated {
			output.WriteString(fmt.Sprintf("  %s\n", dep))
		}
	} else {
		output.WriteString("✅ 所有依赖都是最新版本\n")
	}

	output.WriteString("\n")
	return nil
}

// checkVulnerabilities 检查安全漏洞
func checkVulnerabilities(ctx context.Context, output *strings.Builder) error {
	// 检查 govulncheck 是否可用
	if _, err := exec.LookPath("govulncheck"); err != nil {
		return fmt.Errorf("govulncheck 未安装，请执行: go install golang.org/x/vuln/cmd/govulncheck@latest")
	}

	// 设置超时时间（govulncheck 可能需要更长时间）
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "govulncheck", "./...")
	out, err := cmd.CombinedOutput()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("命令执行超时（60秒）")
		}
		// govulncheck 发现漏洞时会返回非0退出码，但输出仍然有用
		if len(out) > 0 {
			output.WriteString("🔒 安全漏洞检查结果:\n")
			output.WriteString(string(out))
			output.WriteString("\n")
			return nil // 这不是系统错误，只是发现了漏洞
		}
		return fmt.Errorf("执行命令失败: %v", err)
	}

	// 检查是否发现漏洞
	if strings.Contains(string(out), "No vulnerabilities found") {
		output.WriteString("✅ 未发现已知安全漏洞\n")
	} else if len(out) > 0 {
		output.WriteString("🔒 安全漏洞检查结果:\n")
		output.WriteString(string(out))
	} else {
		output.WriteString("✅ 安全漏洞检查完成，未发现问题\n")
	}

	output.WriteString("\n")
	return nil
}
