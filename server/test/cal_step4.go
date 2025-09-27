package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// ========== 读取输入的结构（直接复用你前面定义的类型） ==========
// 这里假设 step2.go / step3.go 中已经定义了：
// - type CleanedDataset {...}  // Step2 输出（含 ValiditySection）
// - type QuotaDataset   {...}  // Step3 输出（含 StudentScores / ParentScores 等）
// - type ScoreResult    {...}  // {Mean float64, Sum int, Count int}

// ========== 输出给 AI 的 Payload 结构 ==========

type ReportPayload struct {
	Meta struct {
		RequestID   string `json:"request_id"`
		StudentID   string `json:"student_id"`
		Mode        string `json:"mode"`
		GeneratedAt string `json:"generated_at"`
		// 可扩展：Grade/Gender 如未来有来源可加
	} `json:"meta"`

	Student struct {
		RIASEC       map[string]ScoreResult `json:"riasec"`
		RIASECRank   RankView               `json:"riasec_rank"`
		OCEAN        map[string]ScoreResult `json:"ocean"`
		OCEANBand    map[string]string      `json:"ocean_band"` // 高/中/低
		Subjects     map[string]ScoreResult `json:"subjects"`
		SubjectsRank RankView               `json:"subjects_rank"`
		Career       map[string]int         `json:"career"` // 当前是每类一题 → int
	} `json:"student"`

	Parent struct {
		RIASEC map[string]ScoreResult `json:"riasec"`
		OCEAN  map[string]ScoreResult `json:"ocean"`
		Values map[string]ScoreResult `json:"values"` // 自主性/合作/坚持
	} `json:"parent"`

	Deltas struct {
		RIASEC map[string]float64 `json:"riasec"` // parent - student
		OCEAN  map[string]float64 `json:"ocean"`
		Values map[string]float64 `json:"values,omitempty"` // 目前可留空
	} `json:"deltas"`

	ModeSupport map[string]any `json:"mode_support,omitempty"` // 针对 3+1+2 / 3+3 的支持性证据

	ValiditySection struct {
		ValidityItems []ValidityItem `json:"validity_items"`
		Checks        struct {
			Student CheckResult `json:"student"`
			Parent  CheckResult `json:"parent"`
		} `json:"checks"`
		Note string `json:"note"`
	} `json:"validity_section"`
}

// RankView：Top/Bottom 排名视图
type RankView struct {
	Top    []string `json:"top"`
	Bottom []string `json:"bottom"`
}

// ========== Step4 主流程 ==========
func step4() error {
	// 1) 读取 Step3 与 Step2 的产物
	var quota QuotaDataset
	if err := readJSON("quota.json", &quota); err != nil {
		return fmt.Errorf("读取 quota.json 失败: %w", err)
	}
	var cleaned CleanedDataset
	if err := readJSON("items.cleaned.json", &cleaned); err != nil {
		return fmt.Errorf("读取 items.cleaned.json 失败: %w", err)
	}

	// 2) 组装 Payload
	payload := ReportPayload{}
	payload.Meta.RequestID = quota.Meta.RequestID
	payload.Meta.StudentID = quota.Meta.StudentID
	payload.Meta.Mode = quota.Meta.Mode
	payload.Meta.GeneratedAt = time.Now().Format(time.RFC3339)

	// 学生
	payload.Student.RIASEC = quota.StudentScores.RIASEC
	payload.Student.OCEAN = quota.StudentScores.OCEAN
	payload.Student.Subjects = quota.StudentScores.Subjects
	payload.Student.Career = quota.StudentScores.Career

	// 排名与区间
	payload.Student.RIASECRank = makeRank(quota.StudentScores.RIASEC, 2)
	payload.Student.OCEANBand = makeBand(quota.StudentScores.OCEAN)
	payload.Student.SubjectsRank = makeRank(quota.StudentScores.Subjects, 3)

	// 家长
	payload.Parent.RIASEC = quota.ParentScores.RIASEC
	payload.Parent.OCEAN = quota.ParentScores.OCEAN
	payload.Parent.Values = quota.ParentScores.Values

	// 差异（家长 - 学生）
	payload.Deltas.RIASEC = makeDelta(quota.ParentScores.RIASEC, quota.StudentScores.RIASEC)
	payload.Deltas.OCEAN = makeDelta(quota.ParentScores.OCEAN, quota.StudentScores.OCEAN)
	// payload.Deltas.Values = ... // 如需可加

	// 模式支持证据
	payload.ModeSupport = makeModeSupport(quota)

	// 效度板块
	payload.ValiditySection.ValidityItems = cleaned.ValiditySection.ValidityItems
	payload.ValiditySection.Checks = cleaned.ValiditySection.Checks
	payload.ValiditySection.Note = "请先基于本节事实，评价本次答卷的有效性（有效/存疑/无效），并在报告开头以一句提示说明理由。"

	// 3) 落盘
	if err := writeJSON("report_payload.json", payload); err != nil {
		return fmt.Errorf("写入 report_payload.json 失败: %w", err)
	}

	// 4) 控制台摘要
	printPayloadSummary(payload)

	return nil
}

// ========== 公共 I/O ==========
func readJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func writeJSON(path string, v any) error {
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0644)
}

// ========== 辅助：排名、区间、差异、模式支持 ==========

func makeRank(m map[string]ScoreResult, k int) RankView {
	type kv struct {
		K string
		V float64
	}
	var arr []kv
	for key, sr := range m {
		arr = append(arr, kv{key, sr.Mean})
	}
	sort.Slice(arr, func(i, j int) bool { return arr[i].V > arr[j].V })

	topN := min(k, len(arr))
	botN := min(k, len(arr))
	top := make([]string, 0, topN)
	bot := make([]string, 0, botN)

	for i := 0; i < topN; i++ {
		top = append(top, arr[i].K)
	}
	for i := 0; i < botN; i++ {
		bot = append(bot, arr[len(arr)-1-i].K)
	}
	return RankView{Top: top, Bottom: bot}
}

// 区间：≥4.0 高；3.0–3.9 中；≤2.9 低（仅标签，不是结论）
func makeBand(m map[string]ScoreResult) map[string]string {
	res := map[string]string{}
	for k, v := range m {
		switch {
		case v.Mean >= 4.0:
			res[k] = "高"
		case v.Mean >= 3.0:
			res[k] = "中"
		default:
			res[k] = "低"
		}
	}
	return res
}

func makeDelta(parent map[string]ScoreResult, student map[string]ScoreResult) map[string]float64 {
	res := map[string]float64{}
	for k, pv := range parent {
		if sv, ok := student[k]; ok && sv.Count > 0 {
			res[k] = pv.Mean - sv.Mean
		}
	}
	return res
}

func makeModeSupport(q QuotaDataset) map[string]any {
	mode := strings.TrimSpace(q.Meta.Mode)
	result := map[string]any{}
	switch mode {
	case "3+1+2":
		// 两个方向的简易证据：学科均值 + RIASEC 提示
		physicsSubjects := []string{"物理", "数学", "化学"}
		historySubjects := []string{"历史", "语文", "政治", "地理"}

		phyMean := avgSubjects(q.StudentScores.Subjects, physicsSubjects)
		hisMean := avgSubjects(q.StudentScores.Subjects, historySubjects)

		// RIASEC 提示（不加权、不下结论）
		riasec := q.StudentScores.RIASEC
		result["3+1+2"] = map[string]any{
			"physics_direction_support": map[string]any{
				"subjects_mean": phyMean,
				"riasec_hint":   pickMeans(riasec, []string{"R", "I"}),
				"weight_scheme": map[string]float64{"subjects": 0.7, "riasec": 0.3},
			},
			"history_direction_support": map[string]any{
				"subjects_mean": hisMean,
				"riasec_hint":   pickMeans(riasec, []string{"A", "S", "E"}),
				"weight_scheme": map[string]float64{"subjects": 0.7, "riasec": 0.3},
			},
		}
	case "3+3":
		// 给 AI 的提示：用学科均值排序 + RIASEC Top2 作为证据
		result["3+3"] = map[string]any{
			"subjects_rank_top": makeRank(q.StudentScores.Subjects, 3).Top,
			"riasec_top":        makeRank(q.StudentScores.RIASEC, 2).Top,
		}
	default:
		// 其他模式（如将来扩展），先不填
	}
	return result
}

func avgSubjects(m map[string]ScoreResult, keys []string) float64 {
	if len(keys) == 0 {
		return 0
	}
	sum := 0.0
	n := 0
	for _, k := range keys {
		if sr, ok := m[k]; ok && sr.Count > 0 {
			sum += sr.Mean
			n++
		}
	}
	if n == 0 {
		return 0
	}
	return sum / float64(n)
}

func pickMeans(m map[string]ScoreResult, keys []string) map[string]float64 {
	res := map[string]float64{}
	for _, k := range keys {
		if sr, ok := m[k]; ok && sr.Count > 0 {
			res[k] = sr.Mean
		}
	}
	return res
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ========== 控制台摘要 ==========
func printPayloadSummary(p ReportPayload) {
	fmt.Println("=== Step4 摘要（将写入 report_payload.json） ===")
	fmt.Printf("RequestID=%s StudentID=%s Mode=%s\n", p.Meta.RequestID, p.Meta.StudentID, p.Meta.Mode)

	fmt.Printf("[学生 RIASEC] Top=%v Bottom=%v\n", p.Student.RIASECRank.Top, p.Student.RIASECRank.Bottom)
	fmt.Printf("[学生 学科]   Top=%v Bottom=%v\n", p.Student.SubjectsRank.Top, p.Student.SubjectsRank.Bottom)

	// 模式支持简要
	if entry, ok := p.ModeSupport["3+1+2"]; ok {
		m := entry.(map[string]any)
		pm := m["physics_direction_support"].(map[string]any)["subjects_mean"]
		hm := m["history_direction_support"].(map[string]any)["subjects_mean"]
		fmt.Printf("[模式 3+1+2] 理科向学科均值=%.2f | 历史向学科均值=%.2f\n", pm, hm)
	}
	if entry, ok := p.ModeSupport["3+3"]; ok {
		m := entry.(map[string]any)
		fmt.Printf("[模式 3+3] 学科Top=%v | RIASEC Top=%v\n", m["subjects_rank_top"], m["riasec_top"])
	}

	// 效度检查简要
	fmt.Printf("[效度检查] Student: all_same=%v high_conc=%v | Parent: all_same=%v high_conc=%v\n",
		p.ValiditySection.Checks.Student.AllSameScore, p.ValiditySection.Checks.Student.HighConcentration,
		p.ValiditySection.Checks.Parent.AllSameScore, p.ValiditySection.Checks.Parent.HighConcentration,
	)
	fmt.Println("===========================================")
}

// ========== 测试入口 ==========

func TestStep4() {
	if err := step4(); err != nil {
		fmt.Println("Step4 执行失败:", err)
		return
	}
	fmt.Println("Step4 已生成 report_payload.json ✅")
}
