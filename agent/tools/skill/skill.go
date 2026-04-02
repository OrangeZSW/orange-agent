package skill

import (
	"context"
	"orange-agent/common"
	"orange-agent/utils"
	"orange-agent/utils/file"
	"orange-agent/utils/logger"
	"strings"
	"sync"

	"go.yaml.in/yaml/v3"
)

type Skill struct {
	Name        string
	Description string
	Content     string
}

var (
	skills    []Skill
	once      sync.Once
	SkillTool = common.BaseTool{
		Name:        "skill",
		Description: "获取技能详细信息",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type": "string",
				},
			},
			"required": []string{"name"},
		},
		Call: SkillCall,
	}
)

func SkillCall(ctx context.Context, input string) (string, error) {
	params, err := utils.StrToMap(input)
	if err != nil {
		return "", err
	}
	for _, skill := range initSkills() {
		if skill.Name == params["name"].(string) {
			res, _ := yaml.Marshal(skill)
			return string(res), nil
		}
	}
	return "", nil
}

func initSkills() []Skill {
	once.Do(func() {
		log := logger.GetLogger()
		paths := "."
		list, err := file.GetFileList(paths)
		if err != nil {
			log.Error("获取文件列表失败.路径:%s,err:%v", paths, err)
		}
		for _, item := range list {
			var skill Skill
			if item.Name == "SKILL.md" {
				content, _ := file.ReadFile(item.Path)
				parts := strings.SplitN(string(content), "---\n", 3)
				if len(parts) < 3 {
					continue
				}
				if err := yaml.Unmarshal([]byte(parts[1]), &skill); err != nil {
					log.Error("解析yaml失败.路径:%s,err:%v", item.Path, err)
					continue
				}
				skill.Content = parts[2]
				skills = append(skills, skill)
			}
		}
	})
	return skills
}

func GetSkills() []Skill {
	return initSkills()
}
