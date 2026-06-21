package fbhttp

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/filebrowser/filebrowser/v2/share"
)

const uploadDir = "/uploads"

// publicUploadRateLimit 控制每个来源 IP 在 publicUploadWindow 时间窗口内
// 允许的最大上传次数（固定窗口计数器，进程内 map+mutex 实现）。
// 这是为了防滥用（刷盘/刷分享表），不影响"免登录投递"这个产品卖点。
const (
	publicUploadRateLimit = 10              // 每个 IP 每个窗口允许的次数
	publicUploadWindow    = time.Minute     // 时间窗口
	publicUploadTokenEnv  = "FB_PUBLIC_UPLOAD_TOKEN"
)

// ipUploadCounter 是某个 IP 在当前固定窗口内的计数。
type ipUploadCounter struct {
	count       int
	windowStart time.Time
}

var (
	publicUploadMu       sync.Mutex
	publicUploadCounters = map[string]*ipUploadCounter{}
)

// validUploadToken 校验上传口令。从 X-Upload-Token 头、Authorization: Bearer
// 头或 ?token= 查询参数中取值，与期望值做常量时间比较。
// 注意：不从 multipart 表单字段读取，以免提前消费请求体导致后面的
// ParseMultipartForm(1GB) 上限失效。
func validUploadToken(r *http.Request, expected string) bool {
	got := r.Header.Get("X-Upload-Token")
	if got == "" {
		if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
			got = strings.TrimPrefix(auth, "Bearer ")
		}
	}
	if got == "" {
		got = r.URL.Query().Get("token")
	}
	if got == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(expected)) == 1
}

// clientIP 从请求中提取来源 IP（仅取 host 部分，去掉端口）。
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// allowPublicUpload 实现固定窗口限速：在 publicUploadWindow 内，同一 IP
// 最多 publicUploadRateLimit 次。超限返回 false。
func allowPublicUpload(ip string) bool {
	now := time.Now()
	publicUploadMu.Lock()
	defer publicUploadMu.Unlock()

	c, ok := publicUploadCounters[ip]
	if !ok || now.Sub(c.windowStart) >= publicUploadWindow {
		publicUploadCounters[ip] = &ipUploadCounter{count: 1, windowStart: now}
		return true
	}
	if c.count >= publicUploadRateLimit {
		return false
	}
	c.count++
	return true
}

type publicUploadResponse struct {
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	Path        string `json:"path"`
	ShareHash   string `json:"share_hash"`
	DownloadURL string `json:"download_url"`
	BrowseURL   string `json:"browse_url"`
}

var publicUploadHandler handleFunc = func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	// 可选鉴权开关：若设置了环境变量 FB_PUBLIC_UPLOAD_TOKEN，则要求请求带上
	// 该口令才能上传（请求头 X-Upload-Token、Authorization: Bearer <token>
	// 或表单字段 token 任一即可）；不设置则维持"免登录投递"。
	if expected := os.Getenv(publicUploadTokenEnv); expected != "" {
		if !validUploadToken(r, expected) {
			return http.StatusUnauthorized, fmt.Errorf("upload token required")
		}
	}

	// 基础限速：按来源 IP 做固定窗口计数，防止匿名投递箱被刷盘/刷分享表。
	if !allowPublicUpload(clientIP(r)) {
		return http.StatusTooManyRequests, fmt.Errorf("rate limit exceeded, try again later")
	}

	// Get admin user (ID=1) for filesystem access
	user, err := d.store.Users.Get(d.server.Root, uint(1))
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to get admin user: %w", err)
	}

	// Ensure upload directory exists
	if err := user.Fs.MkdirAll(uploadDir, 0755); err != nil { //nolint:gomnd
		return http.StatusInternalServerError, fmt.Errorf("failed to create upload dir: %w", err)
	}

	// Parse multipart form (max 1GB)
	const maxUploadSize = 1 << 30
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		return http.StatusBadRequest, fmt.Errorf("failed to parse form: %w", err)
	}
	defer func() {
		if r.MultipartForm != nil {
			r.MultipartForm.RemoveAll() //nolint:errcheck
		}
	}()

	file, header, err := r.FormFile("file")
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("missing 'file' field: %w", err)
	}
	defer file.Close()

	// Sanitize filename
	filename := filepath.Base(header.Filename)
	if filename == "." || filename == "/" {
		filename = "upload"
	}

	// Build destination path with timestamp to avoid collision
	ts := time.Now().Format("20060102-150405")
	dstPath := path.Join(uploadDir, ts+"_"+filename)

	// Check if file already exists, add suffix if needed
	if _, statErr := user.Fs.Stat(dstPath); statErr == nil {
		ext := filepath.Ext(filename)
		base := strings.TrimSuffix(filename, ext)
		dstPath = path.Join(uploadDir, fmt.Sprintf("%s_%s_1%s", ts, base, ext))
	}

	// Write file
	dir, _ := path.Split(dstPath)
	if mkErr := user.Fs.MkdirAll(dir, 0755); mkErr != nil { //nolint:gomnd
		return http.StatusInternalServerError, mkErr
	}

	dst, err := user.Fs.OpenFile(dstPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644) //nolint:gomnd
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	written, err := io.Copy(dst, file)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to write file: %w", err)
	}

	// Create permanent share link
	// 与 http/share.go 保持一致的 24 字节熵，避免分享链接被遍历/猜测。
	hashBytes := make([]byte, 24)
	if _, err := rand.Read(hashBytes); err != nil {
		return http.StatusInternalServerError, err
	}
	hashStr := base64.URLEncoding.EncodeToString(hashBytes)

	shareLink := &share.Link{
		Path:   dstPath,
		Hash:   hashStr,
		Expire: 0, // permanent
		UserID: user.ID,
	}

	if err := d.store.Share.Save(shareLink); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to create share: %w", err)
	}

	resp := publicUploadResponse{
		Filename:    filename,
		Size:        written,
		Path:        dstPath,
		ShareHash:   hashStr,
		DownloadURL: fmt.Sprintf("/api/public/dl/%s", hashStr),
		BrowseURL:   fmt.Sprintf("/share/%s", hashStr),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return 0, json.NewEncoder(w).Encode(resp)
}
