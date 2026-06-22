package fbhttp

import (
	"embed"
	"net/http"
)

// mcpBinFS 嵌入交叉编译好的 lumen-mcp 各平台二进制。
// 二进制由 build 脚本(scripts/build-mcp.sh)交叉编译生成，不入 git。
// 若 clone 后未构建，目录为空，下载接口返回 404、targets 返回空数组。
//
//go:embed mcpbin
var mcpBinFS embed.FS

// mcpTargets 支持的平台白名单 → 文件扩展名
var mcpTargets = map[string]string{
	"windows-amd64": ".exe",
	"darwin-arm64":  "",
	"darwin-amd64":  "",
	"linux-amd64":   "",
}

// mcpDownloadHandler 按 os/arch 下载对应平台的 lumen-mcp 二进制(需登录)
var mcpDownloadHandler = withUser(func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
	key := r.URL.Query().Get("os") + "-" + r.URL.Query().Get("arch")
	ext, ok := mcpTargets[key]
	if !ok {
		return http.StatusBadRequest, nil
	}
	content, err := mcpBinFS.ReadFile("mcpbin/lumen-mcp-" + key + ext)
	if err != nil {
		return http.StatusNotFound, nil
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="lumen-mcp`+ext+`"`)
	w.Header().Set("Content-Length", itoa(len(content)))
	_, _ = w.Write(content)
	return 0, nil
})

// mcpTargetsHandler 返回当前已构建可下载的平台列表(供前端判断)
var mcpTargetsHandler = withUser(func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
	avail := []string{}
	for key, ext := range mcpTargets {
		if f, err := mcpBinFS.Open("mcpbin/lumen-mcp-" + key + ext); err == nil {
			_ = f.Close()
			avail = append(avail, key)
		}
	}
	return renderJSON(w, r, avail)
})

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
