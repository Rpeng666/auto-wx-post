package wechat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

// MediaType 素材类型
type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeVoice MediaType = "voice"
	MediaTypeVideo MediaType = "video"
	MediaTypeThumb MediaType = "thumb"
)

// MediaUploadResult 素材上传结果
type MediaUploadResult struct {
	MediaID string `json:"media_id"`
	URL     string `json:"url"`
}

// ArticleRequest 图文素材
type ArticleRequest struct {
	Articles []Article `json:"articles"`
}

// Article 文章
type Article struct {
	Title            string `json:"title"`
	ThumbMediaID     string `json:"thumb_media_id"`
	Author           string `json:"author"`
	Digest           string `json:"digest"`
	ShowCoverPic     int    `json:"show_cover_pic"`
	Content          string `json:"content"`
	ContentSourceURL string `json:"content_source_url"`
}

// DraftResponse 草稿箱响应
type DraftResponse struct {
	MediaID string `json:"media_id"`
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// UploadPermanentMedia 上传永久素材
func (c *Client) UploadPermanentMedia(ctx context.Context, mediaType MediaType, filePath string) (*MediaUploadResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("media", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("copy file: %w", err)
	}

	contentType := writer.FormDataContentType()
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close writer: %w", err)
	}

	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf(
		"https://api.weixin.qq.com/cgi-bin/material/add_material?access_token=%s&type=%s",
		token, mediaType,
	)

	req, err := c.httpClient.Post(url, contentType, body)
	if err != nil {
		return nil, fmt.Errorf("upload media: %w", err)
	}
	defer req.Body.Close()

	var result struct {
		MediaUploadResult
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}

	if err := json.NewDecoder(req.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat error: %d - %s", result.ErrCode, result.ErrMsg)
	}

	return &result.MediaUploadResult, nil
}

// AddDraft 添加草稿
func (c *Client) AddDraft(ctx context.Context, articles []Article) (string, error) {
	reqBody := ArticleRequest{Articles: articles}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal articles: %w", err)
	}

	endpoint := "https://api.weixin.qq.com/cgi-bin/draft/add"

	var resp DraftResponse
	if err := c.DoRequest(ctx, "POST", endpoint, bytes.NewReader(data), &resp); err != nil {
		return "", err
	}

	if resp.ErrCode != 0 {
		return "", fmt.Errorf("add draft error: %d - %s", resp.ErrCode, resp.ErrMsg)
	}

	return resp.MediaID, nil
}
