package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/hopwesley/wenxintai/server/internal/ai"
	"github.com/hopwesley/wenxintai/server/internal/appconsts"
	"github.com/hopwesley/wenxintai/server/internal/logic"
	"github.com/hopwesley/wenxintai/server/internal/store"
)

type GenerateQuestionsFunc func(ctx context.Context, stage int16, mode string, userCtx map[string]string) (json.RawMessage, string, json.RawMessage, error)
type InterpretReportFunc func(ctx context.Context, params json.RawMessage) (json.RawMessage, *string, error)
type ComputeParamsFunc func(answersS1, answersS2 json.RawMessage) (json.RawMessage, error)

type Svc struct {
	repo              store.Repo
	generateQuestions GenerateQuestionsFunc
	interpretReport   InterpretReportFunc
	computeParams     ComputeParamsFunc
}

func NewSvc(repo store.Repo) *Svc {
	return &Svc{
		repo:              repo,
		generateQuestions: ai.GenerateQuestions,
		interpretReport:   ai.InterpretReport,
		computeParams:     logic.ComputeParams,
	}
}

func (s *Svc) withTx(ctx context.Context, fn func(context.Context, *sql.Tx) error) error {
	return s.repo.WithTx(ctx, func(tx *sql.Tx) error {
		txCtx := store.ContextWithTx(ctx, tx)
		return fn(txCtx, tx)
	})
}

func normalizeOptional(input *string) *string {
	if input == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*input)
	if trimmed == "" {
		return nil
	}
	v := trimmed
	return &v
}

func (s *Svc) CreateAssessment(ctx context.Context, mode string, inviteCode, wechatOpenID *string) (assessmentID, questionSetID string, questions json.RawMessage, err error) {
	if strings.TrimSpace(mode) == "" {
		return "", "", nil, newError(ErrorCodeBadRequest, "mode is required", nil)
	}
	inviteCode = normalizeOptional(inviteCode)
	wechatOpenID = normalizeOptional(wechatOpenID)
	if inviteCode == nil && wechatOpenID == nil {
		return "", "", nil, newError(ErrorCodeBadRequest, "invite_code or wechat_openid required", nil)
	}

	questionsJSON, prompt, raw, err := s.generateQuestions(ctx, appconsts.StageS1, mode, map[string]string{
		"stage": "S1",
	})
	if err != nil {
		return "", "", nil, fmt.Errorf("generate questions: %w", err)
	}

	assessment := &store.Assessment{
		ID:           uuid.NewString(),
		InviteCode:   inviteCode,
		WechatOpenID: wechatOpenID,
		Mode:         mode,
		Status:       appconsts.AS1Pending,
	}
	questionSet := &store.QuestionSet{
		ID:            uuid.NewString(),
		AssessmentID:  assessment.ID,
		Stage:         appconsts.StageS1,
		QuestionsJSON: questionsJSON,
		AIPrompt:      &prompt,
		AIRawResponse: raw,
		Status:        appconsts.QSetIssued,
	}

	err = s.withTx(ctx, func(txCtx context.Context, tx *sql.Tx) error {
		if err := s.repo.CreateAssessment(txCtx, assessment); err != nil {
			return err
		}
		if err := s.repo.CreateQuestionSet(txCtx, questionSet); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return "", "", nil, err
	}
	return assessment.ID, questionSet.ID, questionsJSON, nil
}

type NextStage struct {
	QuestionSetID string
	Stage         string
	Questions     json.RawMessage
}

func (s *Svc) SubmitAnswersAndAdvance(ctx context.Context, questionSetID string, answers json.RawMessage) (next *NextStage, reportID *string, err error) {
	var nextStage *NextStage
	var finalReportID *string

	err = s.withTx(ctx, func(txCtx context.Context, tx *sql.Tx) error {
		qs, err := s.repo.GetQuestionSetByID(txCtx, questionSetID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				return newError(ErrorCodeNotFound, "resource not found", err)
			}
			return err
		}

		ans := &store.Answer{
			QuestionSetID: qs.ID,
			AnswerJSON:    answers,
		}
		if _, err := s.repo.UpsertAnswer(txCtx, ans); err != nil {
			return err
		}
		if err := s.repo.UpdateQuestionSetStatus(txCtx, qs.ID, appconsts.QSetAnswered); err != nil {
			return err
		}

		assessment, err := s.repo.GetAssessmentByID(txCtx, qs.AssessmentID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				return newError(ErrorCodeNotFound, "resource not found", err)
			}
			return err
		}

		switch qs.Stage {
		case appconsts.StageS1:
			if assessment.Status < appconsts.AS1Submitted {
				if err := s.repo.UpdateAssessmentStatus(txCtx, assessment.ID, appconsts.AS1Submitted); err != nil {
					return err
				}
				assessment.Status = appconsts.AS1Submitted
			}

			var stageTwo *store.QuestionSet
			stageTwo, err = s.repo.GetQuestionSetByAssessmentStage(txCtx, assessment.ID, appconsts.StageS2)
			if err != nil {
				if errors.Is(err, store.ErrNotFound) {
					questionsJSON, prompt, raw, genErr := s.generateQuestions(ctx, appconsts.StageS2, assessment.Mode, map[string]string{
						"stage":         "S2",
						"assessment_id": assessment.ID,
					})
					if genErr != nil {
						return genErr
					}
					stageTwo = &store.QuestionSet{
						ID:            uuid.NewString(),
						AssessmentID:  assessment.ID,
						Stage:         appconsts.StageS2,
						QuestionsJSON: questionsJSON,
						AIPrompt:      &prompt,
						AIRawResponse: raw,
						Status:        appconsts.QSetIssued,
					}
					if err := s.repo.CreateQuestionSet(txCtx, stageTwo); err != nil {
						return err
					}
				} else {
					return err
				}
			}

			if assessment.Status < appconsts.AS2Pending {
				if err := s.repo.UpdateAssessmentStatus(txCtx, assessment.ID, appconsts.AS2Pending); err != nil {
					return err
				}
				assessment.Status = appconsts.AS2Pending
			}

			nextStage = &NextStage{
				QuestionSetID: stageTwo.ID,
				Stage:         "S2",
				Questions:     stageTwo.QuestionsJSON,
			}

		case appconsts.StageS2:
			if assessment.Status < appconsts.AS2Submitted {
				if err := s.repo.UpdateAssessmentStatus(txCtx, assessment.ID, appconsts.AS2Submitted); err != nil {
					return err
				}
				assessment.Status = appconsts.AS2Submitted
			}

			existingReport, reportErr := s.repo.GetLatestReportByAssessment(txCtx, assessment.ID)
			if reportErr == nil {
				finalReportID = &existingReport.ID
				if assessment.Status < appconsts.AReportReady {
					if err := s.repo.UpdateAssessmentStatus(txCtx, assessment.ID, appconsts.AReportReady); err != nil {
						return err
					}
				}
				return nil
			}
			if reportErr != nil && !errors.Is(reportErr, store.ErrNotFound) {
				return reportErr
			}

			answersS1, answersS2, err := s.repo.GetAnswersByAssessment(txCtx, assessment.ID)
			if err != nil {
				return err
			}
			if len(answersS1) == 0 {
				return newError(ErrorCodeConflict, "stage S1 answers missing", nil)
			}
			params, err := s.computeParams(answersS1, answersS2)
			if err != nil {
				return err
			}
			computed := &store.ComputedParams{
				AssessmentID: assessment.ID,
				Stage:        0,
				ParamsJSON:   params,
			}
			if err := s.repo.CreateComputedParams(txCtx, computed); err != nil {
				return err
			}
			full, summary, err := s.interpretReport(ctx, params)
			if err != nil {
				return err
			}
			report := &store.Report{
				AssessmentID: assessment.ID,
				ReportType:   appconsts.ReportAIInterpretation,
				Summary:      summary,
				FullJSON:     full,
			}
			if err := s.repo.CreateReport(txCtx, report); err != nil {
				return err
			}
			if err := s.repo.UpdateAssessmentStatus(txCtx, assessment.ID, appconsts.AReportReady); err != nil {
				return err
			}
			rid := report.ID
			finalReportID = &rid

		default:
			return newError(ErrorCodeInternal, fmt.Sprintf("unknown question set stage %d", qs.Stage), nil)
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}
	return nextStage, finalReportID, nil
}

func (s *Svc) GetReport(ctx context.Context, assessmentID string) (*store.Report, error) {
	report, err := s.repo.GetLatestReportByAssessment(ctx, assessmentID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, newError(ErrorCodeNotFound, "resource not found", err)
		}
		return nil, err
	}
	return report, nil
}

type Progress struct {
	Status int16
	Label  string
}

func (s *Svc) GetProgress(ctx context.Context, assessmentID string) (*Progress, error) {
	assessment, err := s.repo.GetAssessmentByID(ctx, assessmentID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, newError(ErrorCodeNotFound, "resource not found", err)
		}
		return nil, err
	}
	return &Progress{Status: assessment.Status, Label: statusLabel(assessment.Status)}, nil
}

func (s *Svc) GetQuestionSet(ctx context.Context, questionSetID string) (*store.QuestionSet, error) {
	qs, err := s.repo.GetQuestionSetByID(ctx, questionSetID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, newError(ErrorCodeNotFound, "resource not found", err)
		}
		return nil, err
	}
	return qs, nil
}

func statusLabel(status int16) string {
	switch status {
	case appconsts.ACreated:
		return "CREATED"
	case appconsts.AS1Pending:
		return "S1_PENDING"
	case appconsts.AS1Submitted:
		return "S1_SUBMITTED"
	case appconsts.AS2Pending:
		return "S2_PENDING"
	case appconsts.AS2Submitted:
		return "S2_SUBMITTED"
	case appconsts.AReportReady:
		return "REPORT_READY"
	case appconsts.ACancelled:
		return "CANCELLED"
	default:
		return "UNKNOWN"
	}
}
