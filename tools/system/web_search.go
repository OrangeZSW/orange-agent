package system

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"orange-agent/common"
	"regexp"
	"strings"
	"time"
)

// WebSearchTool 网络搜索工具
var WebSearchTool = common.BaseTool{
	Name:        "web_search",
	Description: "联网搜索功能，支持搜索引擎查询和网页内容抓取",
	Parameters: map[string]interface{}{
		"query": map[string]interface{}{
			"type":        "string",
			"description": "搜索关键词或URL",
		},
		"search_type": map[string]interface{}{
			"type":        "string",
			"description": "搜索类型：search(搜索引擎搜索)、fetch(抓取网页内容)",
			"enum":        []interface{}{"search", "fetch"},
		},
		"engine": map[string]interface{}{
			"type":        "string",
			"description": "搜索引擎：duckduckgo(默认)、google、bing",
			"enum":        []interface{}{"duckduckgo", "google", "bing"},
		},
		"num_results": map[string]interface{}{
			"type":        "integer",
			"description": "返回结果数量（默认5条，最多10条）",
		},
		"required": []string{"query", "search_type"},
	},
	Call: handleWebSearch,
}

// SearchResult 搜索结果结构
type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// WebContent 网页内容结构
type WebContent struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Content     string `json:"content"`
	Description string `json:"description"`
}

func handleWebSearch(ctx context.Context, input string) (string, error) {
	// 解析参数
	var params struct {
		Query      string `json:"query"`
		SearchType string `json:"search_type"`
		Engine     string `json:"engine"`
		NumResults int    `json:"num_results"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	if params.Query == "" {
		return "", fmt.Errorf("搜索关键词不能为空")
	}

	// 设置默认值
	if params.Engine == "" {
		params.Engine = "duckduckgo"
	}
	if params.NumResults == 0 {
		params.NumResults = 5
	}
	if params.NumResults > 10 {
		params.NumResults = 10
	}

	switch params.SearchType {
	case "search":
		return searchWeb(params.Query, params.Engine, params.NumResults)
	case "fetch":
		return fetchWebContent(params.Query)
	default:
		return "", fmt.Errorf("不支持的搜索类型: %s", params.SearchType)
	}
}

// searchWeb 使用搜索引擎搜索
func searchWeb(query, engine string, numResults int) (string, error) {
	var results []SearchResult
	var err error

	switch engine {
	case "duckduckgo":
		results, err = searchDuckDuckGo(query, numResults)
	case "google":
		results, err = searchGoogle(query, numResults)
	case "bing":
		results, err = searchBing(query, numResults)
	default:
		results, err = searchDuckDuckGo(query, numResults)
	}

	if err != nil {
		return "", fmt.Errorf("搜索失败: %v", err)
	}

	if len(results) == 0 {
		return "未找到相关结果", nil
	}

	// 格式化输出
	var output strings.Builder
	output.WriteString(fmt.Sprintf("🔍 搜索结果 (共 %d 条):\n\n", len(results)))
	for i, result := range results {
		output.WriteString(fmt.Sprintf("%d. **%s**\n", i+1, result.Title))
		output.WriteString(fmt.Sprintf("   📎 %s\n", result.URL))
		output.WriteString(fmt.Sprintf("   📝 %s\n\n", result.Snippet))
	}

	return output.String(), nil
}

// searchDuckDuckGo 使用 DuckDuckGo 搜索
func searchDuckDuckGo(query string, numResults int) ([]SearchResult, error) {
	// DuckDuckGo Instant Answer API
	apiURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1", url.QueryEscape(query))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解析 DuckDuckGo 响应
	var ddgResponse struct {
		AbstractText   string `json:"AbstractText"`
		AbstractURL    string `json:"AbstractURL"`
		AbstractSource string `json:"AbstractSource"`
		RelatedTopics  []struct {
			Text string `json:"Text"`
			URL  string `json:"FirstURL"`
		} `json:"RelatedTopics"`
	}

	if err := json.Unmarshal(body, &ddgResponse); err != nil {
		return nil, err
	}

	var results []SearchResult

	// 添加摘要结果
	if ddgResponse.AbstractText != "" {
		results = append(results, SearchResult{
			Title:   ddgResponse.AbstractSource,
			URL:     ddgResponse.AbstractURL,
			Snippet: ddgResponse.AbstractText,
		})
	}

	// 添加相关主题
	for i, topic := range ddgResponse.RelatedTopics {
		if i >= numResults-1 {
			break
		}
		if topic.Text != "" && topic.URL != "" {
			results = append(results, SearchResult{
				Title:   extractTitle(topic.Text),
				URL:     topic.URL,
				Snippet: topic.Text,
			})
		}
	}

	// 如果 DuckDuckGo API 没有返回结果，尝试使用 HTML 抓取
	if len(results) == 0 {
		return searchDuckDuckGoHTML(query, numResults)
	}

	return results, nil
}

// searchDuckDuckGoHTML 通过 HTML 页面抓取 DuckDuckGo 搜索结果
func searchDuckDuckGoHTML(query string, numResults int) ([]SearchResult, error) {
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	html := string(body)

	// 解析搜索结果
	var results []SearchResult

	// 提取结果链接和标题
	linkRegex := regexp.MustCompile(`<a rel="nofollow" class="result__a" href="([^"]+)">([^<]+)</a>`)
	snippetRegex := regexp.MustCompile(`<a class="result__snippet"[^>]*>([^<]+)</a>`)

	links := linkRegex.FindAllStringSubmatch(html, -1)
	snippets := snippetRegex.FindAllStringSubmatch(html, -1)

	for i := 0; i < len(links) && i < numResults; i++ {
		if len(links[i]) >= 3 {
			resultURL := links[i][1]
			title := strings.TrimSpace(links[i][2])
			snippet := ""

			if i < len(snippets) && len(snippets[i]) >= 2 {
				snippet = strings.TrimSpace(snippets[i][1])
			}

			// DuckDuckGo 使用重定向 URL，需要提取实际 URL
			if strings.Contains(resultURL, "uddg=") {
				if u, err := url.Parse(resultURL); err == nil {
					if actualURL := u.Query().Get("uddg"); actualURL != "" {
						resultURL = actualURL
					}
				}
			}

			results = append(results, SearchResult{
				Title:   title,
				URL:     resultURL,
				Snippet: snippet,
			})
		}
	}

	return results, nil
}

// searchGoogle 使用 Google 搜索 (需要 API Key)
func searchGoogle(query string, numResults int) ([]SearchResult, error) {
	// 注意：Google Custom Search API 需要 API Key
	// 这里提供一个模拟实现，实际使用需要配置 API Key
	return []SearchResult{
		{
			Title:   "Google 搜索提示",
			URL:     "https://www.google.com/search?q=" + url.QueryEscape(query),
			Snippet: "Google 搜索需要配置 Custom Search API Key。请使用 DuckDuckGo 搜索引擎，它不需要 API Key。",
		},
	}, nil
}

// searchBing 使用 Bing 搜索 (需要 API Key)
func searchBing(query string, numResults int) ([]SearchResult, error) {
	// 注意：Bing Search API 需要 API Key
	// 这里提供一个模拟实现，实际使用需要配置 API Key
	return []SearchResult{
		{
			Title:   "Bing 搜索提示",
			URL:     "https://www.bing.com/search?q=" + url.QueryEscape(query),
			Snippet: "Bing 搜索需要配置 API Key。请使用 DuckDuckGo 搜索引擎，它不需要 API Key。",
		},
	}, nil
}

// fetchWebContent 抓取网页内容
func fetchWebContent(urlStr string) (string, error) {
	// 验证 URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("无效的 URL: %v", err)
	}

	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
		urlStr = parsedURL.String()
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP 错误: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	html := string(body)

	// 提取标题
	title := extractHTMLContent(html, "<title>", "</title>")
	if title == "" {
		title = "无标题"
	}

	// 提取描述
	description := extractMetaContent(html, "description")

	// 提取正文内容
	content := extractMainContent(html)

	// 限制内容长度
	if len(content) > 3000 {
		content = content[:3000] + "...\n(内容已截断，完整内容请访问原网页)"
	}

	// 格式化输出
	var output strings.Builder
	output.WriteString(fmt.Sprintf("📄 网页内容:\n\n"))
	output.WriteString(fmt.Sprintf("📌 标题: %s\n", title))
	output.WriteString(fmt.Sprintf("📎 URL: %s\n", urlStr))
	if description != "" {
		output.WriteString(fmt.Sprintf("📝 描述: %s\n", description))
	}
	output.WriteString(fmt.Sprintf("\n📖 正文内容:\n%s\n", content))

	return output.String(), nil
}

// extractHTMLContent 提取 HTML 标签内容
func extractHTMLContent(html, startTag, endTag string) string {
	start := strings.Index(html, startTag)
	if start == -1 {
		return ""
	}
	start += len(startTag)

	end := strings.Index(html[start:], endTag)
	if end == -1 {
		return ""
	}

	return strings.TrimSpace(html[start : start+end])
}

// extractMetaContent 提取 meta 标签内容
func extractMetaContent(html, name string) string {
	// 匹配 <meta name="description" content="...">
	pattern := fmt.Sprintf(`<meta[^>]+name=["']%s["'][^>]+content=["']([^"']+)["']`, name)
	regex := regexp.MustCompile(pattern)
	matches := regex.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}

	// 匹配 <meta content="..." name="description">
	pattern2 := fmt.Sprintf(`<meta[^>]+content=["']([^"']+)["'][^>]+name=["']%s["']`, name)
	regex2 := regexp.MustCompile(pattern2)
	matches2 := regex2.FindStringSubmatch(html)
	if len(matches2) > 1 {
		return matches2[1]
	}

	return ""
}

// extractMainContent 提取网页主要内容
func extractMainContent(html string) string {
	// 移除 script 和 style 标签
	scriptRegex := regexp.MustCompile(`<script[^>]*>[\s\S]*?</script>`)
	html = scriptRegex.ReplaceAllString(html, "")

	styleRegex := regexp.MustCompile(`<style[^>]*>[\s\S]*?</style>`)
	html = styleRegex.ReplaceAllString(html, "")

	// 移除 HTML 注释
	commentRegex := regexp.MustCompile(`<!--[\s\S]*?-->`)
	html = commentRegex.ReplaceAllString(html, "")

	// 移除所有 HTML 标签
	tagRegex := regexp.MustCompile(`<[^>]+>`)
	text := tagRegex.ReplaceAllString(html, " ")

	// 清理空白字符
	spaceRegex := regexp.MustCompile(`\s+`)
	text = spaceRegex.ReplaceAllString(text, " ")

	// 清理特殊字符
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")

	return strings.TrimSpace(text)
}

// extractTitle 从文本中提取标题
func extractTitle(text string) string {
	// 如果文本中有 " - "，取第一部分作为标题
	if idx := strings.Index(text, " - "); idx > 0 {
		return strings.TrimSpace(text[:idx])
	}
	// 如果文本太长，截取前 100 个字符
	if len(text) > 100 {
		return strings.TrimSpace(text[:100]) + "..."
	}
	return strings.TrimSpace(text)
}