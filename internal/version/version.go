// Package version 提供当前服务的语义化版本号常量。
//
// 版本号在每次发版时手工更新 (v1.0.0, v1.1.0 ...),
// 运行时通过 version.Version 读取, 可用于 /healthz、/version 接口或启动日志。
package version

// Version 当前服务的语义化版本号, 遵循 https://semver.org/lang/zh-CN/
const Version = "1.0.0"
