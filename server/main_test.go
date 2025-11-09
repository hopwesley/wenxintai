package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hopwesley/wenxintai/server/assessment"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func TestPipelineFlow(t *testing.T) {
	originalCaller := assessment.DeepSeekCaller
	defer func() { assessment.DeepSeekCaller = originalCaller }()

	var callCount int
	var usedKeys []string
	assessment.DeepSeekCaller = func(apiKey string, reqBody interface{}) (string, error) {
		usedKeys = append(usedKeys, apiKey)
		callCount++
		switch callCount {
		case 1:
			return `{"items":[{"module":"OCEAN"}]}`, nil
		case 2:
			return `{"items":[{"module":"RIASEC"}]}`, nil
		case 3:
			return `{"items":[{"module":"ASC"}]}`, nil
		default:
			return `{"final_report":{"mode":"3+3","report_validity":"ok"}}`, nil
		}
	}

	srv := newPipelineServer("test-default-key", nil)
	ts := httptest.NewServer(srv.routes())
	defer ts.Close()

	// Step 1: login and start session
	loginReq := loginRequest{WeChatID: "wx-user", Nickname: "小明", AvatarURL: "avatar.png"}
	var loginResp loginResponse
	doJSONRequest(t, ts.Client(), ts.URL+"/api/login", http.MethodPost, loginReq, &loginResp)
	if loginResp.SessionID == "" {
		t.Fatalf("expected session id")
	}

	// Step 2: fetch hobbies list
	resp, err := ts.Client().Get(ts.URL + "/api/hobbies")
	if err != nil {
		t.Fatalf("fetch hobbies: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected hobbies status: %d", resp.StatusCode)
	}

	// Step 3: generate questions
	qReq := questionsRequest{
		SessionID: loginResp.SessionID,
		Mode:      string(assessment.Mode33),
		Gender:    "男",
		Grade:     "高一",
		Hobby:     "篮球",
	}
	var qResp questionsResponse
	doJSONRequest(t, ts.Client(), ts.URL+"/api/questions", http.MethodPost, qReq, &qResp)
	if qResp.Questions == nil || len(qResp.Questions.Modules) != 3 {
		t.Fatalf("expected 3 modules, got %#v", qResp.Questions)
	}

	// Step 4: submit answers and score
	riasec := assessment.AllRIASECCombos[assessment.ComboPHY_CHE_BIO]
	asc := assessment.AllASCCombos[assessment.ComboPHY_CHE_BIO]["aligned"]
	aReq := answersRequest{
		SessionID:     loginResp.SessionID,
		Mode:          string(assessment.Mode33),
		RIASECAnswers: riasec,
		ASCAnswers:    asc,
	}
	var aResp answersResponse
	doJSONRequest(t, ts.Client(), ts.URL+"/api/answers", http.MethodPost, aReq, &aResp)
	if len(aResp.SubjectScores) == 0 {
		t.Fatalf("expected subject scores")
	}

	// Step 5: request report
	rReq := reportRequest{
		SessionID: loginResp.SessionID,
		Mode:      string(assessment.Mode33),
	}
	var rResp reportResponse
	doJSONRequest(t, ts.Client(), ts.URL+"/api/report", http.MethodPost, rReq, &rResp)
	if len(rResp.Report) == 0 {
		t.Fatalf("expected report payload")
	}

	if callCount != 4 {
		t.Fatalf("expected 4 deepseek calls, got %d", callCount)
	}
	for i, key := range usedKeys {
		if key != "test-default-key" {
			t.Fatalf("call %d used unexpected api key %q", i, key)
		}
	}

	srv.mu.RLock()
	sess := srv.sessions[loginResp.SessionID]
	srv.mu.RUnlock()
	if sess == nil {
		t.Fatalf("session not stored")
	}
	if sess.Param == nil || sess.Report == nil {
		t.Fatalf("session missing param/report")
	}
}

func doJSONRequest(t *testing.T, client httpClient, url, method string, payload interface{}, out interface{}) {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %d", resp.StatusCode)
	}
	if out != nil {
		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(out); err != nil {
			t.Fatalf("decode response: %v", err)
		}
	}
}
