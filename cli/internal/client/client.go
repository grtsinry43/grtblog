// Package client 是 grtblog API (/api/v2) 的 HTTP 客户端。
package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// APIError 表示服务端返回的业务错误（envelope 中 code != 0）或 HTTP 层错误。
type APIError struct {
	HTTPStatus int
	Code       int
	BizErr     string
	Msg        string
}

func (e *APIError) Error() string {
	if e.BizErr != "" {
		return fmt.Sprintf("%s: %s", e.BizErr, e.Msg)
	}
	return fmt.Sprintf("HTTP %d: %s", e.HTTPStatus, e.Msg)
}

// IsAuth 表示该错误是否为认证失败（401）。
func (e *APIError) IsAuth() bool {
	return e.HTTPStatus == http.StatusUnauthorized || e.Code == 401
}

// Client 封装对 /api/v2 的访问。
type Client struct {
	rc *resty.Client
}

// envelope 是服务端统一响应结构。
type envelope struct {
	Code   int             `json:"code"`
	BizErr string          `json:"bizErr"`
	Msg    string          `json:"msg"`
	Data   json.RawMessage `json:"data"`
}

// New 创建一个客户端。token 以 gt_ 开头时按 admin token 原样发送，否则按 Bearer JWT 发送。
func New(server, token, version string) *Client {
	rc := resty.New().
		SetBaseURL(strings.TrimRight(server, "/")+"/api/v2").
		SetTimeout(60*time.Second).
		SetHeader("Accept", "application/json").
		SetHeader("User-Agent", "grtblog-cli/"+version)
	if token != "" {
		if strings.HasPrefix(token, "gt_") {
			rc.SetHeader("Authorization", token)
		} else {
			rc.SetAuthToken(token)
		}
	}
	return &Client{rc: rc}
}

// Get 发起 GET 请求并解包 envelope 到 out。
func (c *Client) Get(path string, query map[string]string, out any) error {
	req := c.rc.R()
	if len(query) > 0 {
		req.SetQueryParams(query)
	}
	return c.do(req, http.MethodGet, path, out)
}

// Post 发起 JSON POST 请求。
func (c *Client) Post(path string, body any, out any) error {
	req := c.rc.R()
	if body != nil {
		req.SetBody(body)
	}
	return c.do(req, http.MethodPost, path, out)
}

// Put 发起 JSON PUT 请求。
func (c *Client) Put(path string, body any, out any) error {
	req := c.rc.R()
	if body != nil {
		req.SetBody(body)
	}
	return c.do(req, http.MethodPut, path, out)
}

// Delete 发起 DELETE 请求。
func (c *Client) Delete(path string, out any) error {
	return c.do(c.rc.R(), http.MethodDelete, path, out)
}

// Upload 以 multipart/form-data 上传文件，field 为表单字段名。
func (c *Client) Upload(path, field, filePath string, out any) error {
	req := c.rc.R().SetFile(field, filePath)
	return c.do(req, http.MethodPost, path, out)
}

// Download 下载文件到 dest，返回服务端 Content-Disposition 中的文件名（可能为空）。
func (c *Client) Download(path, dest string) (string, error) {
	resp, err := c.rc.R().SetDoNotParseResponse(true).Get(path)
	if err != nil {
		return "", fmt.Errorf("无法连接服务器: %w", err)
	}
	defer resp.RawBody().Close()
	if resp.StatusCode() != http.StatusOK {
		return "", &APIError{HTTPStatus: resp.StatusCode(), Msg: http.StatusText(resp.StatusCode())}
	}
	f, err := os.Create(dest)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, resp.RawBody()); err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}
	return parseContentDisposition(resp.Header().Get("Content-Disposition")), nil
}

// RawGet 发起 GET 请求但不做 envelope 解包（用于探测等）。
func (c *Client) RawGet(path string) (int, []byte, error) {
	resp, err := c.rc.R().Get(path)
	if err != nil {
		return 0, nil, fmt.Errorf("无法连接服务器: %w", err)
	}
	return resp.StatusCode(), resp.Body(), nil
}

func (c *Client) do(req *resty.Request, method, path string, out any) error {
	resp, err := req.Execute(method, path)
	if err != nil {
		return fmt.Errorf("无法连接服务器: %w", err)
	}
	body := resp.Body()
	var env envelope
	if err := json.Unmarshal(body, &env); err != nil {
		if resp.IsError() {
			return &APIError{HTTPStatus: resp.StatusCode(), Msg: strings.TrimSpace(string(body))}
		}
		return fmt.Errorf("响应解析失败（非 JSON）: %w", err)
	}
	if env.Code != 0 {
		return &APIError{HTTPStatus: resp.StatusCode(), Code: env.Code, BizErr: env.BizErr, Msg: env.Msg}
	}
	if out != nil && len(env.Data) > 0 && string(env.Data) != "null" {
		if err := json.Unmarshal(env.Data, out); err != nil {
			return fmt.Errorf("响应数据解析失败: %w", err)
		}
	}
	return nil
}

// parseContentDisposition 从 Content-Disposition 中提取文件名。
func parseContentDisposition(v string) string {
	if v == "" {
		return ""
	}
	for _, part := range strings.Split(v, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "filename*=") {
			// RFC 5987: filename*=UTF-8''xxx
			if idx := strings.Index(part, "''"); idx >= 0 {
				return strings.Trim(part[idx+2:], `"`)
			}
		}
		if strings.HasPrefix(part, "filename=") {
			return strings.Trim(strings.TrimPrefix(part, "filename="), `"`)
		}
	}
	return ""
}
