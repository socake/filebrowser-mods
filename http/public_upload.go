package fbhttp

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/filebrowser/filebrowser/v2/share"
)

const uploadDir = "/uploads"

type publicUploadResponse struct {
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	Path        string `json:"path"`
	ShareHash   string `json:"share_hash"`
	DownloadURL string `json:"download_url"`
	BrowseURL   string `json:"browse_url"`
}

var publicUploadHandler handleFunc = func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
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
