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
	"strconv"
	"strings"
	"time"

	"github.com/socake/filebrowser-mods/client/internal/progress"
)

// tus 分块上传参数。
const (
	// tusChunkSize 是每个 PATCH 块的大小。
	tusChunkSize = int64(8 << 20) // 8 MiB
	// tusThreshold 是启用 tus 断点续传的文件大小阈值；小于它直接整文件 PUT。
	tusThreshold = int64(8 << 20) // 8 MiB
	// tusVersion 是握手时声明的 tus 协议版本（服务端目前不校验，带上以示规范）。
	tusVersion = "1.0.0"
)

// countingReader 包装 io.Reader，每读到字节就回调进度（用于流式上传/下载计数）。
type countingReader struct {
	r  io.Reader
	cb func(int64)
}

func (cr *countingReader) Read(p []byte) (int, error) {
	n, err := cr.r.Read(p)
	if n > 0 && cr.cb != nil {
		cr.cb(int64(n))
	}
	return n, err
}

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

// Put 上传本地文件到远程路径。
//
// 大文件（≥ tusThreshold）走 tus 断点续传：POST 创建 + 分块 PATCH，中断后重跑会
// 通过 HEAD 查已传 offset 自动续上。小文件回退到整文件直传
// （POST /api/resources?override=true，请求体即原始字节）。
//
// bar 为可选进度条（nil 表示不显示）。remote 末尾不应带斜杠（那是建目录语义）。
func (c *Client) Put(remote, local string, bar *progress.Bar) error {
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
	size := info.Size()

	if size >= tusThreshold {
		return c.putTus(remote, f, size, bar)
	}
	return c.putDirect(remote, f, size, bar)
}

// putDirect 整文件直传，作为小文件路径与 tus 的回退。
func (c *Client) putDirect(remote string, f *os.File, size int64, bar *progress.Bar) error {
	bar.Start(size)
	u := c.Server + "/api/resources" + encodePath(remote) + "?override=true"
	body := io.Reader(f)
	if bar != nil {
		body = &countingReader{r: f, cb: bar.Add}
	}
	req, err := http.NewRequest(http.MethodPost, u, body)
	if err != nil {
		return err
	}
	req.ContentLength = size
	req.Header.Set("Content-Type", "application/octet-stream")
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	rb, _ := io.ReadAll(resp.Body)
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		bar.Done()
		return nil
	case http.StatusUnauthorized:
		return errUnauthorized
	default:
		return fmt.Errorf("上传失败 (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(rb)))
	}
}

// putTus 用 tus 协议分块上传，支持断点续传。
//
// 握手（服务端 http/tus_handlers.go）：
//   - HEAD  /api/tus<path>          → 返回 Upload-Offset(=当前文件大小) / Upload-Length；
//     无活动上传(缓存缺失/文件不存在)返回 404。
//   - POST  /api/tus<path>?override → 头 Upload-Length 创建上传(并清空已有文件)，返回 201。
//   - PATCH /api/tus<path>          → 头 Upload-Offset + Content-Type:
//     application/offset+octet-stream，分块追加，返回 204 + 新 Upload-Offset。
func (c *Client) putTus(remote string, f *os.File, size int64, bar *progress.Bar) error {
	tusURL := c.Server + "/api/tus" + encodePath(remote)

	// 先 HEAD 探测是否有可续传的活动上传。
	offset, resumable, err := c.tusHead(tusURL, size)
	if err != nil {
		return err
	}
	if !resumable {
		// 没有匹配的活动上传：新建（override=true 覆盖可能存在的旧文件）。
		if err := c.tusPost(tusURL, size); err != nil {
			return err
		}
		offset = 0
	}

	if resumable && offset > 0 {
		fmt.Fprintf(os.Stderr, "断点续传：服务端已有 %d/%d 字节，从该处继续\n", offset, size)
	}
	bar.Start(size)
	if offset > 0 {
		// 续传：把已在服务端的字节计入进度，并把本地文件指针移到该位置。
		bar.Add(offset)
		if _, err := f.Seek(offset, io.SeekStart); err != nil {
			return err
		}
	}

	for offset < size {
		chunk := tusChunkSize
		if remaining := size - offset; remaining < chunk {
			chunk = remaining
		}
		newOffset, err := c.tusPatch(tusURL, offset, f, chunk, bar)
		if err != nil {
			return err
		}
		if newOffset <= offset {
			return fmt.Errorf("上传未推进：服务端 offset 停在 %d", newOffset)
		}
		offset = newOffset
	}

	bar.Done()
	return nil
}

// tusHead 查询活动上传的当前 offset。
// 返回 (offset, resumable, err)：resumable=false 表示需要重新 POST 创建。
func (c *Client) tusHead(tusURL string, size int64) (int64, bool, error) {
	req, err := http.NewRequest(http.MethodHead, tusURL, nil)
	if err != nil {
		return 0, false, err
	}
	req.Header.Set("Tus-Resumable", tusVersion)
	resp, err := c.do(req)
	if err != nil {
		return 0, false, err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		// 仅当声明的总长与本地一致才续传，否则当作需要重新创建。
		if ul, err := strconv.ParseInt(resp.Header.Get("Upload-Length"), 10, 64); err == nil && ul != size {
			return 0, false, nil
		}
		off, err := strconv.ParseInt(resp.Header.Get("Upload-Offset"), 10, 64)
		if err != nil {
			return 0, false, nil
		}
		return off, true, nil
	case http.StatusNotFound, http.StatusForbidden:
		return 0, false, nil
	case http.StatusUnauthorized:
		return 0, false, errUnauthorized
	default:
		return 0, false, fmt.Errorf("tus HEAD 失败 (HTTP %d)", resp.StatusCode)
	}
}

// tusPost 创建一个 tus 上传（会清空可能已存在的同名文件）。
func (c *Client) tusPost(tusURL string, size int64) error {
	req, err := http.NewRequest(http.MethodPost, tusURL+"?override=true", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Tus-Resumable", tusVersion)
	req.Header.Set("Upload-Length", strconv.FormatInt(size, 10))
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	switch resp.StatusCode {
	case http.StatusCreated:
		return nil
	case http.StatusUnauthorized:
		return errUnauthorized
	case http.StatusConflict:
		return fmt.Errorf("远程已存在同名文件且无法覆盖")
	default:
		return fmt.Errorf("tus 创建失败 (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
}

// tusPatch 从 offset 起追加一个 chunk 字节的块，返回服务端报告的新 offset。
func (c *Client) tusPatch(tusURL string, offset int64, f io.Reader, chunk int64, bar *progress.Bar) (int64, error) {
	var body io.Reader = io.LimitReader(f, chunk)
	if bar != nil {
		body = &countingReader{r: body, cb: bar.Add}
	}
	req, err := http.NewRequest(http.MethodPatch, tusURL, body)
	if err != nil {
		return 0, err
	}
	req.ContentLength = chunk
	req.Header.Set("Tus-Resumable", tusVersion)
	req.Header.Set("Content-Type", "application/offset+octet-stream")
	req.Header.Set("Upload-Offset", strconv.FormatInt(offset, 10))
	resp, err := c.do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	rb, _ := io.ReadAll(resp.Body)
	switch resp.StatusCode {
	case http.StatusNoContent:
		newOffset, err := strconv.ParseInt(resp.Header.Get("Upload-Offset"), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("tus 响应缺少有效 Upload-Offset")
		}
		return newOffset, nil
	case http.StatusUnauthorized:
		return 0, errUnauthorized
	case http.StatusConflict:
		return 0, fmt.Errorf("tus offset 冲突 (HTTP 409): %s", strings.TrimSpace(string(rb)))
	default:
		return 0, fmt.Errorf("tus 上传块失败 (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(rb)))
	}
}

// Get 下载远程文件到本地路径。背后接口：GET /api/raw/<path>，响应体即文件字节。
// bar 为可选进度条（nil 表示不显示），总大小取自响应 Content-Length。
func (c *Client) Get(remote, local string, bar *progress.Bar) (int64, error) {
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

	bar.Start(resp.ContentLength)
	var src io.Reader = resp.Body
	if bar != nil {
		src = &countingReader{r: resp.Body, cb: bar.Add}
	}
	n, err := io.Copy(out, src)
	if err != nil {
		return 0, err
	}
	bar.Done()
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
