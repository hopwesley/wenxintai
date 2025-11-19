package srv

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/hopwesley/wenxintai/server/comm"
	"github.com/rs/zerolog"
)

const (
	apiHealthy       = "/api/health"
	apiLoadHobbies   = "/api/hobbies"
	apiInviteVerify  = "/api/invites/verify"
	apiTestFlow      = "/api/test_flow"
	apiTestBasicInfo = "/api/tests/basic_info"
	apiSSESubChannel = "/api/sub/"
	apiSubmitTest    = "/api/test_submit"
)

var (
	_srvOnce          = sync.Once{}
	_srvInst *HttpSrv = nil
)

type HttpSrv struct {
	log zerolog.Logger
	cfg *Config
	srv *http.Server
}

func Instance() *HttpSrv {
	_srvOnce.Do(func() {
		_srvInst = newBusinessService()
	})
	return _srvInst
}

func newBusinessService() *HttpSrv {
	return &HttpSrv{
		log: comm.LogInst().With().
			Str("model", "HttpSrv").
			Logger(),
	}
}

func (s *HttpSrv) Init(cfg *Config) error {
	s.cfg = cfg
	if err := s.initHobbies(); err != nil {
		s.log.Err(err).Msg("init hobbies failed")
		return err
	}

	if err := s.initRouter(); err != nil {
		s.log.Err(err).Msg("init router failed")
		return err
	}

	if err := s.initSSE(); err != nil {
		s.log.Err(err).Msg("init SSE failed")
		return err
	}

	if err := s.initWS(); err != nil {
		s.log.Err(err).Msg("init Websocket failed")
		return err
	}

	return nil
}

func (s *HttpSrv) initRouter() error {
	mux := http.NewServeMux()

	mux.HandleFunc(apiHealthy, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc(apiLoadHobbies, s.handleHobbies)
	mux.HandleFunc(apiInviteVerify, s.handleInviteVerify)
	mux.HandleFunc(apiTestFlow, s.handleTestFlow)
	mux.HandleFunc(apiTestBasicInfo, s.updateBasicInfo)
	mux.HandleFunc(apiSSESubChannel, s.handleSSEEvent)
	mux.HandleFunc(apiSubmitTest, s.handleTestSubmit)

	if stat, err := os.Stat(s.cfg.StaticDir); err != nil || !stat.IsDir() {
		if err == nil {
			return fmt.Errorf("%s is not a directory", s.cfg.StaticDir)
		}
		return err
	}

	fileServer := http.FileServer(http.Dir(s.cfg.StaticDir))
	mux.Handle("/", fileServer)

	srv := &http.Server{
		Addr:              s.cfg.srvAddr(),
		Handler:           mux,
		ReadTimeout:       time.Duration(s.cfg.ReadTimeout) * time.Second,
		ReadHeaderTimeout: time.Duration(s.cfg.ReadTimeout) * time.Second,
	}

	s.srv = srv
	return nil
}

func (s *HttpSrv) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.log.Debug().Str("method", r.Method).
			Str("path", r.URL.Path).
			Float64("time used:", time.Since(start).Seconds()).Send()
	})
}

func (s *HttpSrv) initWS() error {
	return nil
}

func (s *HttpSrv) StartServing() {
	go func() {
		s.log.Info().Msgf("HTTP server listening on %s", s.srv.Addr)
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.Fatal().Msgf("listen: %v", err)
		}
	}()
}

func (s *HttpSrv) Shutdown(ctx context.Context) error {
	if s.srv != nil {
		if err := s.srv.Shutdown(ctx); err != nil {
			return err
		}
		s.srv = nil
	}

	return nil
}
