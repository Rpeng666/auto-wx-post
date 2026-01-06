package markdown

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// Parser Markdown解析器
type Parser struct {
	htmlRenderer *html.Renderer
	parser       *parser.Parser
}

// Article 文章元数据
type Article struct {
	Title      string
	Subtitle   string
	Date       string
	Author     string
	GenCover   string
	Content    string
	Images     []string
}

// NewParser 创建Markdown解析器
func NewParser() *Parser {
	// HTML渲染选项
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{
		Flags: htmlFlags,
	}
	renderer := html.NewRenderer(opts)

	// 解析器扩展
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.Footnotes
	p := parser.NewWithExtensions(extensions)

	return &Parser{
		htmlRenderer: renderer,
		parser:       p,
	}
}

// ParseFile 解析Markdown文件
func (p *Parser) ParseFile(filePath string) (*Article, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	return p.Parse(string(content))
}

// Parse 解析Markdown内容
func (p *Parser) Parse(content string) (*Article, error) {
	article := &Article{}

	// 提取元数据 (YAML front matter)
	metadata, body := p.extractMetadata(content)
	article.Title = p.getMetadataField(metadata, "title")
	article.Subtitle = p.getMetadataField(metadata, "subtitle")
	article.Date = p.getMetadataField(metadata, "date")
	article.Author = p.getMetadataField(metadata, "author")
	article.GenCover = p.getMetadataField(metadata, "gen_cover")
	article.Content = body

	// 提取图片
	article.Images = p.extractImages(body)

	return article, nil
}

// ToHTML 转换为HTML
func (p *Parser) ToHTML(content string) string {
	md := []byte(content)
	htmlBytes := markdown.ToHTML(md, p.parser, p.htmlRenderer)
	return string(htmlBytes)
}

// extractMetadata 提取元数据
func (p *Parser) extractMetadata(content string) (map[string]string, string) {
	metadata := make(map[string]string)
	
	// 查找 YAML front matter (--- 包围的部分)
	parts := strings.Split(content, "---\n")
	if len(parts) < 3 {
		return metadata, content
	}

	// 解析元数据
	scanner := bufio.NewScanner(strings.NewReader(parts[1]))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, ":") {
			kv := strings.SplitN(line, ":", 2)
			if len(kv) == 2 {
				key := strings.TrimSpace(kv[0])
				value := strings.TrimSpace(kv[1])
				value = strings.Trim(value, `"'`)
				metadata[key] = value
			}
		}
	}

	// 返回正文
	body := strings.Join(parts[2:], "---\n")
	return metadata, body
}

// getMetadataField 获取元数据字段
func (p *Parser) getMetadataField(metadata map[string]string, key string) string {
	if val, ok := metadata[key]; ok {
		return val
	}
	return ""
}

// extractImages 提取图片链接
func (p *Parser) extractImages(content string) []string {
	var images []string
	
	// 匹配 ![alt](url) 格式
	re := regexp.MustCompile(`!\[.*?\]\((.*?)\)`)
	matches := re.FindAllStringSubmatch(content, -1)
	
	for _, match := range matches {
		if len(match) > 1 {
			images = append(images, match[1])
		}
	}
	
	return images
}

// UpdateImageURLs 更新图片URL
func (p *Parser) UpdateImageURLs(content string, urlMap map[string]string) string {
	result := content
	for oldURL, newURL := range urlMap {
		result = strings.ReplaceAll(result, fmt.Sprintf("(%s)", oldURL), fmt.Sprintf("(%s)", newURL))
	}
	return result
}
