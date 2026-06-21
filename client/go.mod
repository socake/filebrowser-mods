module github.com/socake/filebrowser-mods/client

go 1.25.0

// 第一版 CLI 刻意保持零外部依赖（仅标准库），编出来的 lumen 是最小单二进制。
// 将来命令变多需要补全/子命令树时，再视情况引入 spf13/cobra。
