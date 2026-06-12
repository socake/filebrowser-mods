package fbhttp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type docsFileEntry struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	IsDir   bool   `json:"isDir"`
	Size    int64  `json:"size"`
	ModTime string `json:"modTime"`
	Ext     string `json:"ext"`
}

type docsListResponse struct {
	Path    string          `json:"path"`
	Entries []docsFileEntry `json:"entries"`
}

// hiddenOrSensitive returns true for files that should not be exposed.
func hiddenOrSensitive(name string) bool {
	lower := strings.ToLower(name)
	if strings.HasPrefix(name, ".") {
		return true
	}
	sensitiveNames := []string{"credentials", "secrets", "password", ".env", "id_rsa", "id_ed25519", "token"}
	for _, s := range sensitiveNames {
		if strings.Contains(lower, s) {
			return true
		}
	}
	return false
}

// docsPageHandler serves the HTML document portal.
var docsPageHandler handleFunc = func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(docsHTML))
	return 0, err
}

// docsListHandler returns a JSON directory listing.
var docsListHandler handleFunc = func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	user, err := d.store.Users.Get(d.server.Root, uint(1))
	if err != nil {
		return http.StatusInternalServerError, err
	}

	reqPath := r.URL.Query().Get("path")
	if reqPath == "" {
		reqPath = "/"
	}
	reqPath = path.Clean("/" + reqPath)

	// Resolve real path and verify it's within root
	realRoot := d.server.Root
	realPath := filepath.Join(realRoot, reqPath)
	realPath, err = filepath.Abs(realPath)
	if err != nil || !strings.HasPrefix(realPath, realRoot) {
		return http.StatusForbidden, fmt.Errorf("access denied")
	}

	info, err := user.Fs.Stat(reqPath)
	if err != nil {
		return http.StatusNotFound, err
	}
	if !info.IsDir() {
		return http.StatusBadRequest, fmt.Errorf("not a directory")
	}

	f, err := user.Fs.Open(reqPath)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer f.Close()

	dirFile, ok := f.(interface {
		Readdir(int) ([]os.FileInfo, error)
	})
	if !ok {
		return http.StatusInternalServerError, fmt.Errorf("cannot read directory")
	}

	entries, err := dirFile.Readdir(-1)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	var result []docsFileEntry
	for _, entry := range entries {
		name := entry.Name()
		if hiddenOrSensitive(name) {
			continue
		}
		result = append(result, docsFileEntry{
			Name:    name,
			Path:    path.Join(reqPath, name),
			IsDir:   entry.IsDir(),
			Size:    entry.Size(),
			ModTime: entry.ModTime().Format(time.RFC3339),
			Ext:     strings.ToLower(filepath.Ext(name)),
		})
	}

	// Sort: dirs first, then by name
	sort.Slice(result, func(i, j int) bool {
		if result[i].IsDir != result[j].IsDir {
			return result[i].IsDir
		}
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})

	resp := docsListResponse{Path: reqPath, Entries: result}
	w.Header().Set("Content-Type", "application/json")
	return 0, json.NewEncoder(w).Encode(resp)
}

// docsContentHandler returns raw file content for preview.
var docsContentHandler handleFunc = func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	user, err := d.store.Users.Get(d.server.Root, uint(1))
	if err != nil {
		return http.StatusInternalServerError, err
	}

	reqPath := r.URL.Query().Get("path")
	if reqPath == "" {
		return http.StatusBadRequest, fmt.Errorf("path required")
	}
	reqPath = path.Clean("/" + reqPath)

	realRoot := d.server.Root
	realPath := filepath.Join(realRoot, reqPath)
	realPath, err = filepath.Abs(realPath)
	if err != nil || !strings.HasPrefix(realPath, realRoot) {
		return http.StatusForbidden, fmt.Errorf("access denied")
	}

	if hiddenOrSensitive(filepath.Base(reqPath)) {
		return http.StatusForbidden, fmt.Errorf("access denied")
	}

	info, err := user.Fs.Stat(reqPath)
	if err != nil {
		return http.StatusNotFound, err
	}
	if info.IsDir() {
		return http.StatusBadRequest, fmt.Errorf("not a file")
	}
	// Limit to 10MB
	if info.Size() > 10*1024*1024 {
		return http.StatusRequestEntityTooLarge, fmt.Errorf("file too large for preview")
	}

	f, err := user.Fs.Open(reqPath)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer f.Close()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-File-Name", filepath.Base(reqPath))
	w.Header().Set("X-File-Size", fmt.Sprintf("%d", info.Size()))
	w.Header().Set("X-File-ModTime", info.ModTime().Format(time.RFC3339))
	http.ServeContent(w, r, filepath.Base(reqPath), info.ModTime(), f)
	return 0, nil
}

// docsDownloadHandler triggers a file download.
var docsDownloadHandler handleFunc = func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	user, err := d.store.Users.Get(d.server.Root, uint(1))
	if err != nil {
		return http.StatusInternalServerError, err
	}

	reqPath := r.URL.Query().Get("path")
	if reqPath == "" {
		return http.StatusBadRequest, fmt.Errorf("path required")
	}
	reqPath = path.Clean("/" + reqPath)

	realRoot := d.server.Root
	realPath := filepath.Join(realRoot, reqPath)
	realPath, err = filepath.Abs(realPath)
	if err != nil || !strings.HasPrefix(realPath, realRoot) {
		return http.StatusForbidden, fmt.Errorf("access denied")
	}

	info, err := user.Fs.Stat(reqPath)
	if err != nil {
		return http.StatusNotFound, err
	}
	if info.IsDir() {
		return http.StatusBadRequest, fmt.Errorf("cannot download directory")
	}

	f, err := user.Fs.Open(reqPath)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer f.Close()

	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filepath.Base(reqPath)))
	http.ServeContent(w, r, filepath.Base(reqPath), info.ModTime(), f)
	return 0, nil
}
