package ai_api

import "math"

type MetricDef struct {
	RawMin         float64 // 原始分理论/经验最小值
	RawMax         float64 // 原始分理论/经验最大值
	HigherIsBetter bool    // true=越高越好，false=越低越好（例如“错误率”这种）
}

var metricRegistry = map[string]MetricDef{
	"subjects.fit": {
		RawMin:         -0.6,
		RawMax:         0.8,
		HigherIsBetter: true,
	},
	"combo33.score": {
		RawMin:         -0.1,
		RawMax:         0.7,
		HigherIsBetter: true,
	},
	"combo312.score": {
		RawMin:         0.0,
		RawMax:         0.80,
		HigherIsBetter: true,
	},
	"common.global_cosine": {
		RawMin:         -1.0,
		RawMax:         1.0,
		HigherIsBetter: true,
	},
	"common.quality_score": {
		RawMin:         0.0,
		RawMax:         1.0,
		HigherIsBetter: true,
	},
}

func NormalizeMetric(key string, raw float64) float64 {
	cfg, ok := metricRegistry[key]
	if !ok || cfg.RawMax == cfg.RawMin {
		// 未注册的指标，保守处理：直接返回 0
		return 0
	}

	ratio := (raw - cfg.RawMin) / (cfg.RawMax - cfg.RawMin)
	score := ratio * 100

	// 越低越好（目前你的 P0 里都不是，不过保留这个能力）
	if !cfg.HigherIsBetter {
		score = 100 - score
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	// 保留一位小数就足够展示
	return math.Round(score*10) / 10
}
