// Package api 是 LumenBrowser 服务端 REST 接口的最小封装（标准库实现）。
//
// 鉴权：登录返回纯文本 JWT，后续请求通过 HTTP 头 X-Auth 携带（见服务端 http/auth.go）。
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Client 是带 token 的已登录客户端。
type Client struct {
	Server string
	Token  string
	HTTP   *http.Client
}

// New 构造一个客户端；server 末尾的斜杠会被去掉。
func New(server, token string) *Client {
	return &Client{
		Server: strings.TrimRight(server, "/"),
		Token:  token,
		HTTP:   &http.Client{},
	}
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	if c.Token != "" {
		req.Header.Set("X-Auth", c.Token)
	}
	return c.HTTP.Do(req)
}

// Login 调 POST /api/login（body: {username,password}），成功返回纯文本 JWT。
func Login(server, username, password string) (string, error) {
	server = strings.TrimRight(server, "/")
	payload, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	resp, err := http.Post(server+"/api/login", "application/json", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	switch resp.StatusCode {
	case http.StatusOK:
		return strings.TrimSpace(string(body)), nil
	case http.StatusForbidden:
		return "", fmt.Errorf("用户名或密码错误")
	default:
		return "", fmt.Errorf("登录失败 (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
}

// FileItem 是目录项（宽松解析 /api/resources 响应中的 items）。
type FileItem struct {
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	IsDir    bool      `json:"isDir"`
	Modified time.Time `json:"modified"`
	Type     string    `json:"type"`
}

type resourceResponse struct {
	IsDir bool       `json:"isDir"`
	Items []FileItem `json:"items"`
}

// List 调 GET /api/resources/<path> 列目录，返回目录项。
func (c *Client) List(path string) ([]FileItem, error) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	req, err := http.NewRequest(http.MethodGet, c.Server+"/api/resources"+encodePath(path), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	switch resp.StatusCode {
	case http.StatusOK:
		// fall through
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("未授权：token 可能已过期，请重新 lumen login")
	default:
		return nil, fmt.Errorf("列目录失败 (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
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

// UploadResult 是 /api/public/upload 的返回。
type UploadResult struct {
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	Path        string `json:"path"`
	ShareHash   string `json:"share_hash"`
	DownloadURL string `json:"download_url"`
	BrowseURL   string `json:"browse_url"`
}

// PublicUpload 调 POST /api/public/upload（免登录），multipart 字段名 file，返回分享信息。
//
// 注意：第一版把文件读入内存缓冲后发送，超大文件（接近服务端 1GB 上限）会占用相应内存，
// 后续可改 io.Pipe 流式上传。
func PublicUpload(server, filePath string) (*UploadResult, error) {
	server = strings.TrimRight(server, "/")
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(fw, f); err != nil {
		return nil, err
	}
	if err := mw.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, server+"/api/public/upload", &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("投递失败 (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var ur UploadResult
	if err := json.Unmarshal(body, &ur); err != nil {
		return nil, err
	}
	return &ur, nil
}

// Put 上传本地文件到远程路径。背后接口：POST /api/resources/<path>?override=true，
// 请求体即文件原始字节。第一版把文件流直接交给 http.Client（不读入内存）。
//
// remote 末尾不应带斜杠（那是建目录语义）；override=true 表示已存在则覆盖。
func (c *Client) Put(remote, local string) error {
	if !strings.HasPrefix(remote, "/") {
		remote = "/" + remote
	}
	f, err := os.Open(local)
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}

	u := c.Server + "/api/resources" + encodePath(remote) + "?override=true"
	req, err := http.NewRequest(http.MethodPost, u, f)
	if err != nil {
		return err
	}
	req.ContentLength = info.Size()
	req.Header.Set("Content-Type", "application/octet-stream")
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		return nil
	case http.StatusUnauthorized:
		return errUnauthorized
	default:
		return fmt.Errorf("上传失败 (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
}

// Get 下载远程文件到本地路径。背后接口：GET /api/raw/<path>，响应体即文件字节。
func (c *Client) Get(remote, local string) (int64, error) {
	if !strings.HasPrefix(remote, "/") {
		remote = "/" + remote
	}
	req, err := http.NewRequest(http.MethodGet, c.Server+"/api/raw"+encodePath(remote), nil)
	if err != nil {
		return 0, err
	}
	resp, err := c.do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		// fall through
	case http.StatusUnauthorized:
		return 0, errUnauthorized
	default:
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("下载失败 (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	out, err := os.Create(local)
	if err != nil {
		return 0, err
	}
	defer out.Close()
	n, err := io.Copy(out, resp.Body)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// Remove 删除远程文件或目录。背后接口：DELETE /api/resources/<path>，成功返回 204。
func (c *Client) Remove(remote string) error {
	if !strings.HasPrefix(remote, "/") {
		remote = "/" + remote
	}
	req, err := http.NewRequest(http.MethodDelete, c.Server+"/api/resources"+encodePath(remote), nil)
	if err != nil {
		return err
	}
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	switch resp.StatusCode {
	case http.StatusOK, http.StatusNoContent:
		return nil
	case http.StatusUnauthorized:
		return errUnauthorized
	default:
		return fmt.Errorf("删除失败 (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
}

// Move 重命名/移动远程文件。背后接口：
// PATCH /api/resources/<src>?action=rename&destination=<dst>。
// 服务端会对 destination 做一次 url.QueryUnescape，故这里按 query value 编码目标路径。
func (c *Client) Move(src, dst string) error {
	if !strings.HasPrefix(src, "/") {
		src = "/" + src
	}
	if !strings.HasPrefix(dst, "/") {
		dst = "/" + dst
	}
	q := url.Values{}
	q.Set("action", "rename")
	q.Set("destination", dst)
	u := c.Server + "/api/resources" + encodePath(src) + "?" + q.Encode()
	req, err := http.NewRequest(http.MethodPatch, u, nil)
	if err != nil {
		return err
	}
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	switch resp.StatusCode {
	case http.StatusOK, http.StatusNoContent:
		return nil
	case http.StatusUnauthorized:
		return errUnauthorized
	case http.StatusConflict:
		return fmt.Errorf("目标已存在: %s", dst)
	default:
		return fmt.Errorf("移动失败 (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
}

// Mkdir 新建远程目录。背后接口：POST /api/resources/<path>/（结尾斜杠表示目录）。
func (c *Client) Mkdir(remote string) error {
	if !strings.HasPrefix(remote, "/") {
		remote = "/" + remote
	}
	remote = strings.TrimRight(remote, "/")
	u := c.Server + "/api/resources" + encodePath(remote) + "/"
	req, err := http.NewRequest(http.MethodPost, u, nil)
	if err != nil {
		return err
	}
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusNoContent:
		return nil
	case http.StatusUnauthorized:
		return errUnauthorized
	case http.StatusConflict:
		return fmt.Errorf("目录已存在: %s", remote)
	default:
		return fmt.Errorf("建目录失败 (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
}

// ShareLink 是 /api/share、/api/shares 返回的分享记录（对应服务端 share.Link）。
type ShareLink struct {
	Hash   string `json:"hash"`
	Path   string `json:"path"`
	UserID uint   `json:"userID"`
	Expire int64  `json:"expire"`
}

// shareCreateBody 对应服务端 share.CreateBody。
type shareCreateBody struct {
	Password string `json:"password,omitempty"`
	Expires  string `json:"expires,omitempty"`
	Unit     string `json:"unit,omitempty"`
}

// CreateShare 为远程路径创建分享。背后接口：POST /api/share/<path>，
// body {password, expires, unit}（expires 为数字字符串，unit ∈ seconds/minutes/hours/days，
// 缺省按 hours 计；三者皆空表示永久）。返回创建的分享记录。
func (c *Client) CreateShare(remote, password, expires, unit string) (*ShareLink, error) {
	if !strings.HasPrefix(remote, "/") {
		remote = "/" + remote
	}
	payload, _ := json.Marshal(shareCreateBody{Password: password, Expires: expires, Unit: unit})
	req, err := http.NewRequest(http.MethodPost, c.Server+"/api/share"+encodePath(remote), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		// fall through
	case http.StatusUnauthorized:
		return nil, errUnauthorized
	default:
		return nil, fmt.Errorf("创建分享失败 (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var sl ShareLink
	if err := json.Unmarshal(body, &sl); err != nil {
		return nil, err
	}
	return &sl, nil
}

// ListShares 列出当前用户的分享（管理员看全部）。背后接口：GET /api/shares。
func (c *Client) ListShares() ([]ShareLink, error) {
	req, err := http.NewRequest(http.MethodGet, c.Server+"/api/shares", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	switch resp.StatusCode {
	case http.StatusOK:
		// fall through
	case http.StatusUnauthorized:
		return nil, errUnauthorized
	default:
		return nil, fmt.Errorf("列分享失败 (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var ls []ShareLink
	if err := json.Unmarshal(body, &ls); err != nil {
		return nil, err
	}
	return ls, nil
}

// SearchResult 是搜索命中的一项（path 相对搜索根目录）。
type SearchResult struct {
	Dir  bool   `json:"dir"`
	Path string `json:"path"`
}

// Search 在远程目录下按关键词搜索。背后接口：GET /api/search/<path>?query=。
// 响应是流式 NDJSON：每行一个 JSON 对象，服务端每 5 秒可能插入一个空行作为心跳，
// 解析时需跳过空行。
func (c *Client) Search(remote, query string) ([]SearchResult, error) {
	if !strings.HasPrefix(remote, "/") {
		remote = "/" + remote
	}
	u := c.Server + "/api/search" + encodePath(remote) + "?query=" + url.QueryEscape(query)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errUnauthorized
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("搜索失败 (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var results []SearchResult
	dec := json.NewDecoder(resp.Body)
	for {
		var sr SearchResult
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

// errUnauthorized 统一表达 token 失效，便于命令层提示重新登录。
var errUnauthorized = fmt.Errorf("未授权：token 可能已过期，请重新 lumen login")

// encodePath 对路径逐段 URL 编码，保留分隔斜杠。
func encodePath(p string) string {
	parts := strings.Split(p, "/")
	for i, s := range parts {
		parts[i] = url.PathEscape(s)
	}
	return strings.Join(parts, "/")
}
