package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// client 是 LumenBrowser 服务端 REST 接口的最小封装。
//
// 鉴权：所有请求带 HTTP 头 X-Auth: <token>（服务端 http/auth.go）。
// 这里刻意只依赖标准库，不复用 client/internal/api，避免跨模块耦合（mcp 是独立模块）。
type client struct {
	server string
	token  string
	http   *http.Client
}

func newClient(server, token string) *client {
	return &client{
		server: strings.TrimRight(server, "/"),
		token:  token,
		http:   &http.Client{Timeout: 120 * time.Second},
	}
}

func (c *client) do(req *http.Request) (*http.Response, error) {
	if c.token != "" {
		req.Header.Set("X-Auth", c.token)
	}
	return c.http.Do(req)
}

// fileItem 是目录项（宽松解析 /api/resources 响应里的 items）。
type fileItem struct {
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	IsDir    bool      `json:"isDir"`
	Modified time.Time `json:"modified"`
	Type     string    `json:"type"`
}

type resourceResponse struct {
	IsDir bool       `json:"isDir"`
	Items []fileItem `json:"items"`
}

// list 调 GET /api/resources/<path> 列目录。
func (c *client) list(path string) ([]fileItem, error) {
	path = ensureLeadingSlash(path)
	req, err := http.NewRequest(http.MethodGet, c.server+"/api/resources"+encodePath(path), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if err := checkStatus(resp.StatusCode, body, "列目录"); err != nil {
		return nil, err
	}
	var rr resourceResponse
	if err := json.Unmarshal(body, &rr); err != nil {
		return nil, err
	}
	if !rr.IsDir {
		return nil, fmt.Errorf("%s 不是目录", path)
	}
	return rr.Items, nil
}

// searchResult 是搜索命中的一项（path 相对搜索根目录）。
type searchResult struct {
	Dir  bool   `json:"dir"`
	Path string `json:"path"`
}

// search 调 GET /api/search/<path>?query=。响应是流式 NDJSON：每行一个 JSON 对象，
// 服务端可能插入空行作为心跳，json.Decoder 会自动跳过空白。
func (c *client) search(path, query string) ([]searchResult, error) {
	path = ensureLeadingSlash(path)
	u := c.server + "/api/search" + encodePath(path) + "?query=" + url.QueryEscape(query)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, checkStatus(resp.StatusCode, body, "搜索")
	}
	var results []searchResult
	dec := json.NewDecoder(resp.Body)
	for {
		var sr searchResult
		err := dec.Decode(&sr)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		results = append(results, sr)
	}
	return results, nil
}

// readFile 调 GET /api/raw/<path>，返回文件原始字节。
func (c *client) readFile(path string) ([]byte, error) {
	path = ensureLeadingSlash(path)
	req, err := http.NewRequest(http.MethodGet, c.server+"/api/raw"+encodePath(path), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if err := checkStatus(resp.StatusCode, body, "读文件"); err != nil {
		return nil, err
	}
	return body, nil
}

// upload 把本地文件整体上传到远程路径。
// 背后接口：POST /api/resources/<path>?override=true，请求体即原始文件字节。
func (c *client) upload(localPath, remotePath string) (int64, error) {
	remotePath = ensureLeadingSlash(remotePath)
	f, err := os.Open(localPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return 0, err
	}
	if info.IsDir() {
		return 0, fmt.Errorf("%s 是目录，不能上传", localPath)
	}
	u := c.server + "/api/resources" + encodePath(remotePath) + "?override=true"
	req, err := http.NewRequest(http.MethodPost, u, f)
	if err != nil {
		return 0, err
	}
	req.ContentLength = info.Size()
	req.Header.Set("Content-Type", "application/octet-stream")
	resp, err := c.do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if err := checkStatus(resp.StatusCode, body, "上传"); err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// shareLink 是 POST /api/share 的返回（对应服务端 share.Link）。
type shareLink struct {
	Hash   string `json:"hash"`
	Path   string `json:"path"`
	Type   string `json:"type"`
	Expire int64  `json:"expire"`
}

type shareCreateBody struct {
	Password string `json:"password,omitempty"`
	Expires  string `json:"expires,omitempty"`
	Unit     string `json:"unit,omitempty"`
	Type     string `json:"type,omitempty"`
}

// createShare 为远程路径创建分享。背后接口：POST /api/share/<path>，
// body {password,expires,unit,type}（全空表示永久公开）。type ∈ preview/download。
// 返回分享记录；分享链接为 <server>/share/<hash>。
func (c *client) createShare(path, shareType string) (*shareLink, string, error) {
	path = ensureLeadingSlash(path)
	payload, _ := json.Marshal(shareCreateBody{Type: shareType})
	req, err := http.NewRequest(http.MethodPost, c.server+"/api/share"+encodePath(path), bytes.NewReader(payload))
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if err := checkStatus(resp.StatusCode, body, "创建分享"); err != nil {
		return nil, "", err
	}
	var sl shareLink
	if err := json.Unmarshal(body, &sl); err != nil {
		return nil, "", err
	}
	shareURL := c.server + "/share/" + sl.Hash
	return &sl, shareURL, nil
}

// checkStatus 把非 2xx 响应翻译成带提示的错误；401 特别提示 token 失效。
func checkStatus(code int, body []byte, action string) error {
	switch {
	case code >= 200 && code < 300:
		return nil
	case code == http.StatusUnauthorized:
		return fmt.Errorf("%s失败：未授权，token 可能已过期，请重新 lumen login（或更新 LUMEN_TOKEN）", action)
	default:
		return fmt.Errorf("%s失败 (HTTP %d): %s", action, code, strings.TrimSpace(string(body)))
	}
}

// ensureLeadingSlash 保证路径以 / 开头。
func ensureLeadingSlash(p string) string {
	if p == "" {
		return "/"
	}
	if !strings.HasPrefix(p, "/") {
		return "/" + p
	}
	return p
}

// encodePath 对路径逐段 URL 编码，保留分隔斜杠。
func encodePath(p string) string {
	parts := strings.Split(p, "/")
	for i, s := range parts {
		parts[i] = url.PathEscape(s)
	}
	return strings.Join(parts, "/")
}
