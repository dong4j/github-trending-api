package spider

import (
	"fmt"
	"log"
	"math"
	"time"
)

// InferYear 从 MM/DD 格式推断真实的年份
// 逻辑：假设 Zread 榜单的日期通常在当前日期之前（不会预知未来很久）。
// 如果月份大于当前月份，则很有可能是去年的（比如 1 月初看 12 月底的榜单）。
func InferYear(mmdd string, now time.Time) (int, error) {
	if len(mmdd) < 5 {
		return 0, fmt.Errorf("invalid format: %s", mmdd)
	}

	var m int
	if _, err := fmt.Sscanf(mmdd[0:2], "%d", &m); err != nil {
		return 0, fmt.Errorf("parse month: %w", err)
	}

	inferred := now.Year()
	if m > int(now.Month()) {
		inferred = now.Year() - 1
	}

	// 异常告警：如果推断的年份与当前年份相差 > 1 年，可能是极端数据或逻辑漏洞，打出 Warning。
	// 这对 2027 年元旦的观察窗口至关重要。
	if math.Abs(float64(inferred-now.Year())) > 1 {
		log.Printf("[zread_infer] WARN: inferred year %d differs from current year %d by > 1 for input %q", inferred, now.Year(), mmdd)
	}

	return inferred, nil
}
