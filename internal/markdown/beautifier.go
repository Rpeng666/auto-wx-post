package markdown

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Beautifier HTML美化器
type Beautifier struct {
	cssTemplates map[string]string
}

// NewBeautifier 创建HTML美化器
func NewBeautifier(templateDir string) (*Beautifier, error) {
	b := &Beautifier{
		cssTemplates: make(map[string]string),
	}

	// 加载CSS模板
	if err := b.loadTemplates(templateDir); err != nil {
		return nil, err
	}

	return b, nil
}

// Beautify 美化HTML
func (b *Beautifier) Beautify(htmlContent string) (string, error) {
	// 包装段落
	htmlContent = b.replaceParagraphs(htmlContent)

	// 格式化标题
	htmlContent = b.replaceHeaders(htmlContent)

	// 转换链接为脚注
	htmlContent = b.replaceLinks(htmlContent)

	// 格式化图片
	htmlContent = b.formatImages(htmlContent)

	// 其他格式修复
	htmlContent = b.formatFix(htmlContent)

	// 添加头部和尾部
	htmlContent = b.wrapWithTemplate(htmlContent)

	return htmlContent, nil
}

// replaceParagraphs 替换段落样式
func (b *Beautifier) replaceParagraphs(content string) string {
	paraStyle := b.getTemplate("para")
	if paraStyle == "" {
		paraStyle = `<p style="margin: 10px 0; line-height: 1.75em;">`
	}
	return strings.ReplaceAll(content, "<p>", paraStyle)
}

// replaceHeaders 替换标题样式
func (b *Beautifier) replaceHeaders(content string) string {
	re := regexp.MustCompile(`<h(\d)>(.*?)</h(\d)>`)
	return re.ReplaceAllStringFunc(content, func(match string) string {
		matches := re.FindStringSubmatch(match)
		if len(matches) < 4 {
			return match
		}

		level := matches[1]
		text := matches[2]

		// 计算字体大小
		fontSize := 18
		if l := level[0] - '0'; l >= 1 && l <= 6 {
			fontSize = 18 + (4-int(l))*2
		}

		template := b.getTemplate("sub")
		if template == "" {
			return fmt.Sprintf(`<h%s style="font-size: %dpx; font-weight: bold; margin: 20px 0 10px;">%s</h%s>`,
				level, fontSize, text, level)
		}

		return fmt.Sprintf(template, level, fontSize, text, level)
	})
}

// replaceLinks 替换链接为脚注
func (b *Beautifier) replaceLinks(content string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return content
	}

	links := make([]struct {
		href string
		text string
	}, 0)

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		text := s.Text()
		links = append(links, struct {
			href string
			text string
		}{href, text})
	})

	if len(links) == 0 {
		return content
	}

	// 替换链接为脚注引用
	for i, link := range links {
		oldLink := fmt.Sprintf(`<a href="%s">%s</a>`, link.href, link.text)
		newLink := fmt.Sprintf(`%s<sup>[%d]</sup>`, link.text, i+1)
		content = strings.ReplaceAll(content, oldLink, newLink)
	}

	// 添加脚注区域
	refHeader := b.getTemplate("ref_header")
	if refHeader == "" {
		refHeader = `<hr style="margin: 30px 0;"/><h4>参考链接</h4>`
	}
	content += "\n" + refHeader
	content += `<section class="footnotes">`

	for i, link := range links {
		refLink := b.getTemplate("ref_link")
		if refLink == "" {
			refLink = `<p>[%d] %s: <a href="%s">%s</a></p>`
		}
		content += fmt.Sprintf(refLink, i+1, link.text, link.href, link.href)
	}

	content += "</section>"
	return content
}

// formatImages 格式化图片
func (b *Beautifier) formatImages(content string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return content
	}

	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		alt, _ := s.Attr("alt")
		src, _ := s.Attr("src")

		oldImg := fmt.Sprintf(`<img alt="%s" src="%s" />`, alt, src)

		figureTemplate := b.getTemplate("figure")
		if figureTemplate == "" {
			figureTemplate = `<figure style="text-align: center; margin: 20px 0;">
				<img alt="%s" src="%s" style="max-width: 100%%; border-radius: 8px;" />
				<figcaption style="margin-top: 10px; color: #666; font-size: 14px;">%s</figcaption>
			</figure>`
		}

		newImg := fmt.Sprintf(figureTemplate, alt, src, alt)
		content = strings.ReplaceAll(content, oldImg, newImg)
	})

	return content
}

// formatFix 其他格式修复
func (b *Beautifier) formatFix(content string) string {
	// 列表项之间添加间距
	content = strings.ReplaceAll(content, "</li>", "</li>\n<p></p>")

	// 代码块样式
	codeStyle := b.getTemplate("code")
	if codeStyle == "" {
		codeStyle = `background: #272822; padding: 15px; border-radius: 5px; overflow-x: auto;`
	}
	content = strings.ReplaceAll(content, `background: #272822`, codeStyle)

	// 预格式化文本样式
	content = strings.ReplaceAll(content,
		`<pre style="line-height: 125%">`,
		`<pre style="line-height: 125%; color: white; font-size: 11px; margin: 10px 0;">`)

	return content
}

// wrapWithTemplate 用模板包装内容
func (b *Beautifier) wrapWithTemplate(content string) string {
	header := b.getTemplate("header")
	if header == "" {
		header = `<section style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; 
			font-size: 16px; color: #333; padding: 20px; max-width: 800px; margin: 0 auto;">`
	}
	return header + content + "</section>"
}

// loadTemplates 加载CSS模板
func (b *Beautifier) loadTemplates(templateDir string) error {
	if templateDir == "" || !fileExists(templateDir) {
		// 使用默认模板
		return nil
	}

	templates := []string{"para", "sub", "link", "ref_header", "ref_link", "figure", "code", "header"}

	for _, name := range templates {
		path := filepath.Join(templateDir, name+".tmpl")
		if fileExists(path) {
			content, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			b.cssTemplates[name] = string(content)
		}
	}

	return nil
}

// getTemplate 获取模板
func (b *Beautifier) getTemplate(name string) string {
	if tmpl, ok := b.cssTemplates[name]; ok {
		return tmpl
	}
	return ""
}

// fileExists 检查文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
