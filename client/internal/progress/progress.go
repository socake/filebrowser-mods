// Package progress 提供一个零依赖的单行进度条（标准库实现）。
//
// 它面向流式传输场景：put / get 在拷贝字节时不断调用 Add，进度条按时间节流
// 刷新到 stderr（用 \r 回到行首覆盖），传输结束调 Done 收尾换行。
//
// 仅在 stderr 连接到终端（CharDevice）时渲染动画行；被重定向到管道/文件时静默，
// 避免在日志里塞满回车控制符——这种情况下由命令层打印最终结果摘要即可。
package progress

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// Bar 是一个并发安全的单行进度条。零值不可用，请用 New 构造；
// 所有方法对 nil 接收者安全（传 nil 即关闭进度显示）。
type Bar struct {
	label string
	out   *os.File
	tty   bool

	mu    sync.Mutex
	total int64
	n     int64
	start time.Time
	last  time.Time
}

// New 创建一个进度条，label 是行首标签（如“上传 foo.bin”）。
// 还未调用 Start 之前不会渲染任何内容。
func New(label string) *Bar {
	out := os.Stderr
	tty := false
	if fi, err := out.Stat(); err == nil && fi.Mode()&os.ModeCharDevice != 0 {
		tty = true
	}
	return &Bar{label: label, out: out, tty: tty}
}

// Start 标记传输开始并设置总字节数（用于计算百分比）。
// total<=0 时只显示已传字节与速率，不显示百分比。
func (b *Bar) Start(total int64) {
	if b == nil {
		return
	}
	b.mu.Lock()
	b.total = total
	b.start = time.Now()
	b.last = time.Time{}
	b.mu.Unlock()
	b.render(true)
}

// Add 累加 n 个已传字节并按节流刷新进度行。可从传输 goroutine 调用。
func (b *Bar) Add(n int64) {
	if b == nil || n <= 0 {
		return
	}
	b.mu.Lock()
	b.n += n
	b.mu.Unlock()
	b.render(false)
}

// Done 渲染最终状态并换行收尾。
func (b *Bar) Done() {
	if b == nil {
		return
	}
	b.render(true)
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.tty && !b.start.IsZero() {
		fmt.Fprintln(b.out)
	}
}

func (b *Bar) render(force bool) {
	if b == nil || !b.tty {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.start.IsZero() {
		return
	}
	now := time.Now()
	if !force && now.Sub(b.last) < 100*time.Millisecond {
		return
	}
	b.last = now

	elapsed := now.Sub(b.start).Seconds()
	var rate float64
	if elapsed > 0 {
		rate = float64(b.n) / elapsed
	}
	if b.total > 0 {
		pct := float64(b.n) / float64(b.total) * 100
		fmt.Fprintf(b.out, "\r\033[K%s %5.1f%% %s/%s %s/s",
			b.label, pct, human(b.n), human(b.total), human(int64(rate)))
	} else {
		fmt.Fprintf(b.out, "\r\033[K%s %s %s/s", b.label, human(b.n), human(int64(rate)))
	}
}

// human 把字节数格式化为人类可读单位（1024 进制）。
func human(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%dB", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(n)/float64(div), "KMGTPE"[exp])
}
