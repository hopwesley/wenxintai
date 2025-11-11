package service

import (
        "context"
        "database/sql"
        "encoding/json"
        "sync"
        "testing"
        "time"

        "github.com/google/uuid"

        "github.com/hopwesley/wenxintai/server/internal/appconsts"
        "github.com/hopwesley/wenxintai/server/internal/store"
)

func TestSubmitFlowGeneratesReport(t *testing.T) {
        repo := newMemoryRepo()
        svc := NewSvc(repo)
        svc.generateQuestions = func(ctx context.Context, stage int16, mode string, userCtx map[string]string) (json.RawMessage, string, json.RawMessage, error) {
                payload := mustJSON([]map[string]any{{"question_id": uuid.NewString(), "stage": stage}})
                return payload, "prompt", payload, nil
        }
        svc.computeParams = func(answersS1, answersS2 json.RawMessage) (json.RawMessage, error) {
                return mustJSON(map[string]any{"stage1": string(answersS1), "stage2": string(answersS2)}), nil
        }
        svc.interpretReport = func(ctx context.Context, params json.RawMessage) (json.RawMessage, *string, error) {
                summary := "ok"
                return mustJSON(map[string]any{"params": string(params)}), &summary, nil
        }

        ctx := context.Background()
        invite := "ABC"
        assessmentID, qsetID, questions, err := svc.CreateAssessment(ctx, "standard", &invite, nil)
        if err != nil {
                t.Fatalf("CreateAssessment error: %v", err)
        }
        if assessmentID == "" || qsetID == "" {
                t.Fatalf("expected identifiers, got %q %q", assessmentID, qsetID)
        }
        if len(questions) == 0 {
                t.Fatalf("expected questions payload")
        }

        answerPayload := mustJSON([]map[string]any{{"question_id": "s1-q1", "value": 1}})
        next, reportID, err := svc.SubmitAnswersAndAdvance(ctx, qsetID, answerPayload)
        if err != nil {
                t.Fatalf("SubmitAnswersAndAdvance stage1 error: %v", err)
        }
        if reportID != nil {
                t.Fatalf("unexpected report id on stage1")
        }
        if next == nil || next.Stage != "S2" || next.QuestionSetID == "" {
                t.Fatalf("expected S2 payload: %#v", next)
        }
        stage2ID := next.QuestionSetID

        // Re-submit stage 1 to ensure idempotency returns same S2 id.
        nextAgain, reportIDAgain, err := svc.SubmitAnswersAndAdvance(ctx, qsetID, answerPayload)
        if err != nil {
                t.Fatalf("repeat SubmitAnswersAndAdvance stage1 error: %v", err)
        }
        if reportIDAgain != nil {
                t.Fatalf("unexpected report id on repeated stage1")
        }
        if nextAgain == nil || nextAgain.QuestionSetID != stage2ID {
                t.Fatalf("expected same S2 question set id, got %#v", nextAgain)
        }

        s2Payload := mustJSON([]map[string]any{{"question_id": "s2-q1", "value": 2}})
        nextStage, finalReportID, err := svc.SubmitAnswersAndAdvance(ctx, stage2ID, s2Payload)
        if err != nil {
                t.Fatalf("SubmitAnswersAndAdvance stage2 error: %v", err)
        }
        if nextStage != nil {
                t.Fatalf("unexpected next stage on completion")
        }
        if finalReportID == nil || *finalReportID == "" {
                t.Fatalf("expected report id")
        }

        // Re-submit stage2 to ensure idempotent report id.
        _, repeatedReportID, err := svc.SubmitAnswersAndAdvance(ctx, stage2ID, s2Payload)
        if err != nil {
                t.Fatalf("repeat SubmitAnswersAndAdvance stage2 error: %v", err)
        }
        if repeatedReportID == nil || *repeatedReportID != *finalReportID {
                t.Fatalf("expected same report id, got %v", repeatedReportID)
        }

        report, err := svc.GetReport(ctx, assessmentID)
        if err != nil {
                t.Fatalf("GetReport error: %v", err)
        }
        if report.ID != *finalReportID {
                t.Fatalf("report mismatch: %s vs %s", report.ID, *finalReportID)
        }

        progress, err := svc.GetProgress(ctx, assessmentID)
        if err != nil {
                t.Fatalf("GetProgress error: %v", err)
        }
        if progress.Status != appconsts.AReportReady {
                t.Fatalf("expected status AReportReady, got %d", progress.Status)
        }
        if progress.Label != "REPORT_READY" {
                t.Fatalf("unexpected label: %s", progress.Label)
        }
}

func mustJSON(v interface{}) json.RawMessage {
        data, err := json.Marshal(v)
        if err != nil {
                panic(err)
        }
        return data
}

type memoryRepo struct {
        mu            sync.Mutex
        assessments   map[string]*store.Assessment
        questionSets  map[string]*store.QuestionSet
        answers       map[string]*store.Answer
        computed      []*store.ComputedParams
        reports       []*store.Report
}

func newMemoryRepo() *memoryRepo {
        return &memoryRepo{
                assessments:  make(map[string]*store.Assessment),
                questionSets: make(map[string]*store.QuestionSet),
                answers:      make(map[string]*store.Answer),
        }
}

func cloneAssessment(a *store.Assessment) *store.Assessment {
        if a == nil {
                return nil
        }
        cloned := *a
        return &cloned
}

func cloneQuestionSet(qs *store.QuestionSet) *store.QuestionSet {
        if qs == nil {
                return nil
        }
        cloned := *qs
        if qs.QuestionsJSON != nil {
                cloned.QuestionsJSON = append(json.RawMessage{}, qs.QuestionsJSON...)
        }
        if qs.AIRawResponse != nil {
                cloned.AIRawResponse = append(json.RawMessage{}, qs.AIRawResponse...)
        }
        return &cloned
}

func cloneAnswer(ans *store.Answer) *store.Answer {
        if ans == nil {
                return nil
        }
        cloned := *ans
        if ans.AnswerJSON != nil {
                cloned.AnswerJSON = append(json.RawMessage{}, ans.AnswerJSON...)
        }
        return &cloned
}

func cloneReport(r *store.Report) *store.Report {
        if r == nil {
                return nil
        }
        cloned := *r
        if r.FullJSON != nil {
                cloned.FullJSON = append(json.RawMessage{}, r.FullJSON...)
        }
        return &cloned
}

func (m *memoryRepo) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
        return fn(nil)
}

func (m *memoryRepo) CreateAssessment(ctx context.Context, a *store.Assessment) error {
        m.mu.Lock()
        defer m.mu.Unlock()
        if a.ID == "" {
                a.ID = uuid.NewString()
        }
        now := time.Now()
        a.CreatedAt = now
        a.UpdatedAt = now
        m.assessments[a.ID] = cloneAssessment(a)
        return nil
}

func (m *memoryRepo) UpdateAssessmentStatus(ctx context.Context, id string, status int16) error {
        m.mu.Lock()
        defer m.mu.Unlock()
        existing, ok := m.assessments[id]
        if !ok {
                return store.ErrNotFound
        }
        existing.Status = status
        existing.UpdatedAt = time.Now()
        return nil
}

func (m *memoryRepo) GetAssessmentByID(ctx context.Context, id string) (*store.Assessment, error) {
        m.mu.Lock()
        defer m.mu.Unlock()
        existing, ok := m.assessments[id]
        if !ok {
                return nil, store.ErrNotFound
        }
        return cloneAssessment(existing), nil
}

func (m *memoryRepo) CreateQuestionSet(ctx context.Context, qs *store.QuestionSet) error {
        m.mu.Lock()
        defer m.mu.Unlock()
        if qs.ID == "" {
                qs.ID = uuid.NewString()
        }
        for _, existing := range m.questionSets {
                if existing.AssessmentID == qs.AssessmentID && existing.Stage == qs.Stage {
                        *qs = *cloneQuestionSet(existing)
                        return nil
                }
        }
        now := time.Now()
        qs.CreatedAt = now
        qs.UpdatedAt = now
        m.questionSets[qs.ID] = cloneQuestionSet(qs)
        return nil
}

func (m *memoryRepo) UpdateQuestionSetStatus(ctx context.Context, id string, status int16) error {
        m.mu.Lock()
        defer m.mu.Unlock()
        existing, ok := m.questionSets[id]
        if !ok {
                return store.ErrNotFound
        }
        existing.Status = status
        existing.UpdatedAt = time.Now()
        return nil
}

func (m *memoryRepo) GetQuestionSetByID(ctx context.Context, id string) (*store.QuestionSet, error) {
        m.mu.Lock()
        defer m.mu.Unlock()
        existing, ok := m.questionSets[id]
        if !ok {
                return nil, store.ErrNotFound
        }
        return cloneQuestionSet(existing), nil
}

func (m *memoryRepo) GetQuestionSetByAssessmentStage(ctx context.Context, assessmentID string, stage int16) (*store.QuestionSet, error) {
        m.mu.Lock()
        defer m.mu.Unlock()
        for _, qs := range m.questionSets {
                if qs.AssessmentID == assessmentID && qs.Stage == stage {
                        return cloneQuestionSet(qs), nil
                }
        }
        return nil, store.ErrNotFound
}

func (m *memoryRepo) UpsertAnswer(ctx context.Context, ans *store.Answer) (bool, error) {
        m.mu.Lock()
        defer m.mu.Unlock()
        if existing, ok := m.answers[ans.QuestionSetID]; ok {
                clone := cloneAnswer(existing)
                *ans = *clone
                return false, nil
        }
        if ans.ID == "" {
                ans.ID = uuid.NewString()
        }
        if ans.SubmittedAt.IsZero() {
                ans.SubmittedAt = time.Now()
        }
        m.answers[ans.QuestionSetID] = cloneAnswer(ans)
        return true, nil
}

func (m *memoryRepo) CreateComputedParams(ctx context.Context, cp *store.ComputedParams) error {
        m.mu.Lock()
        defer m.mu.Unlock()
        if cp.ID == "" {
                cp.ID = uuid.NewString()
        }
        if cp.CreatedAt.IsZero() {
                cp.CreatedAt = time.Now()
        }
        clone := *cp
        if cp.ParamsJSON != nil {
                        clone.ParamsJSON = append(json.RawMessage{}, cp.ParamsJSON...)
        }
        m.computed = append(m.computed, &clone)
        return nil
}

func (m *memoryRepo) CreateReport(ctx context.Context, r *store.Report) error {
        m.mu.Lock()
        defer m.mu.Unlock()
        if r.ID == "" {
                r.ID = uuid.NewString()
        }
        if r.CreatedAt.IsZero() {
                r.CreatedAt = time.Now()
        }
        clone := cloneReport(r)
        m.reports = append(m.reports, clone)
        return nil
}

func (m *memoryRepo) GetLatestReportByAssessment(ctx context.Context, assessmentID string) (*store.Report, error) {
        m.mu.Lock()
        defer m.mu.Unlock()
        var latest *store.Report
        for _, r := range m.reports {
                if r.AssessmentID != assessmentID {
                        continue
                }
                if latest == nil || r.CreatedAt.After(latest.CreatedAt) {
                        latest = r
                }
        }
        if latest == nil {
                return nil, store.ErrNotFound
        }
        return cloneReport(latest), nil
}

func (m *memoryRepo) GetAnswersByAssessment(ctx context.Context, assessmentID string) (json.RawMessage, json.RawMessage, error) {
        m.mu.Lock()
        defer m.mu.Unlock()
        var s1, s2 json.RawMessage
        for qsid, ans := range m.answers {
                qs := m.questionSets[qsid]
                if qs == nil || qs.AssessmentID != assessmentID {
                        continue
                }
                if qs.Stage == appconsts.StageS1 {
                        s1 = append(json.RawMessage{}, ans.AnswerJSON...)
                }
                if qs.Stage == appconsts.StageS2 {
                        s2 = append(json.RawMessage{}, ans.AnswerJSON...)
                }
        }
        return s1, s2, nil
}
