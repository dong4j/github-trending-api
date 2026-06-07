// Package utils 提供通用工具函数
package utils

import (
	"regexp"
	"strconv"
	"strings"
)

// getListNum 从字符串列表中提取数字
// 例如: ["6,338"] -> 6338
// 这是为了解析 GitHub 上显示的数字格式，如 "6,338" 转换为 6338
func GetListNum(arr []string) int {
	if len(arr) == 0 {
		return 0
	}
	joined := strings.Join(arr, "")
	// 移除非数字字符，保留数字和逗号
	re := regexp.MustCompile(`[\d,]+`)
	match := re.FindString(joined)
	if match == "" {
		return 0
	}
	// 移除逗号
	cleaned := strings.ReplaceAll(match, ",", "")
	num, err := strconv.Atoi(cleaned)
	if err != nil {
		return 0
	}
	return num
}
