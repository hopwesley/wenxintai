package srv

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hopwesley/wenxintai/server/comm"
	"github.com/rs/zerolog"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

const (
	apiHealthy        = "/api/health"
	apiLoadHobbies    = "/api/hobbies"
	apiLoadProducts   = "/api/products"
	apiLoadCurProduct = "/api/current_product"

	apiTestFlow = "/api/test_flow"

	apiTestBasicInfo = "/api/tests/basic_info"

	apiSSEQuestionSub = "/api/sub/question/"
	apiSSEReportSub   = "/api/sub/report/"
	apiSubmitTest     = "/api/test_submit"
	apiGenerateReport = "/api/generate_report"
	apiFinishReport   = "/api/finish_report"

	apiWeChatSignIn         = "/api/auth/wx/status"
	apiWeChatSignInCallBack = "/api/wechat_signin"
	apiWeChatLogOut         = "/api/auth/logout"
	apiWeChatUpdateProfile  = "/api/user/update_profile"
	apiWeChatMyProfile      = "/api/auth/profile"

	apiInvitePayment           = "/api/pay/use_invite"
	apiWeChatPayment           = "/api/pay/"
	apiWeChatCreateNativeOrder = "/api/pay/wechat/native/create"
	apiWeChatNativeOrderStatus = "/api/pay/wechat/order-status"
)

var (
	_srvOnce          = sync.Once{}
	_srvInst *HttpSrv = nil
)

type route struct {
	pattern      string
	method       string
	handler      http.HandlerFunc
	requireLogin bool
}

type HttpSrv struct {
	log        zerolog.Logger
	cfg        *Config
	payment    *WeChatPayConfig
	srv        *http.Server
	httpClient *http.Client

	wxClient        *core.Client
	wxNativeService *native.NativeApiService
	wxNotifyHandler *notify.Handler
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

func (s *HttpSrv) Init(cfg *Config, payment *WeChatPayConfig) error {
	s.cfg = cfg
	s.payment = payment

	if err := s.initWeChatPay(); err != nil {
		s.log.Err(err).Msg("init wechat pay failed")
		return err
	}

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
func (s *HttpSrv) initWeChatPay() error {
	ctx := context.Background()

	// 商户私钥（apiclient_key.pem）
	mchPrivateKey, err := utils.LoadPrivateKey(s.payment.privateKeyPEM)
	if err != nil {
		return fmt.Errorf("load merchant private key failed: %w", err)
	}

	// 微信支付公钥（pub_key.pem）
	wechatPayPubKey, err := utils.LoadPublicKey(s.payment.publicKeyPEM)
	if err != nil {
		return fmt.Errorf("load wechatpay public key failed: %w", err)
	}

	opts := []core.ClientOption{
		option.WithWechatPayPublicKeyAuthCipher(
			s.payment.MchID,
			s.payment.MchSerial,
			mchPrivateKey,
			s.payment.PublicKeyID,
			wechatPayPubKey,
		),
	}

	client, err := core.NewClient(ctx, opts...)
	if err != nil {
		return fmt.Errorf("new wechat pay client failed: %w", err)
	}
	s.wxClient = client
	s.wxNativeService = &native.NativeApiService{Client: client}

	h, err := notify.NewRSANotifyHandler(
		s.payment.APIV3Key,
		verifiers.NewSHA256WithRSAPubkeyVerifier(
			s.payment.PublicKeyID,
			*wechatPayPubKey,
		),
	)
	if err != nil {
		return fmt.Errorf("new rsa notify handler failed: %w", err)
	}
	s.wxNotifyHandler = h

	s.log.Info().Msg("init WeChat pay success")
	return nil
}

func (s *HttpSrv) initRouter() error {
	mux := http.NewServeMux()

	mux.HandleFunc(apiHealthy, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	routes := []route{
		{apiLoadHobbies, http.MethodGet, s.handleHobbies, false},
		{apiLoadProducts, http.MethodGet, s.handleProducts, false},
		{apiLoadCurProduct, http.MethodPost, s.handleCurrentProduct, false},

		{apiWeChatSignIn, http.MethodGet, s.wechatSignStatus, false},
		{apiWeChatSignInCallBack, http.MethodGet, s.wechatSignInCallBack, false},
		{apiWeChatPayment, http.MethodPost, s.apiWeChatPayCallBack, false},
		{apiWeChatLogOut, http.MethodPost, s.wechatLogout, false},

		{apiTestFlow, http.MethodPost, s.handleTestFlow, true},
		{apiTestBasicInfo, http.MethodPost, s.updateBasicInfo, true},

		{apiInvitePayment, http.MethodPost, s.apiPayByInvite, false},

		{apiSSEQuestionSub, http.MethodGet, s.handleQuestionSSEEvent, true},
		{apiSubmitTest, http.MethodPost, s.handleTestSubmit, true},

		{apiSSEReportSub, http.MethodGet, s.handleReportSSEEvent, true},
		{apiGenerateReport, http.MethodPost, s.generateReport, true},
		{apiFinishReport, http.MethodPost, s.finishReport, true},

		{apiWeChatUpdateProfile, http.MethodPost, s.apiWeChatUpdateProfile, true},
		{apiWeChatMyProfile, http.MethodGet, s.apiWeChatMyProfile, true},
		{apiWeChatCreateNativeOrder, http.MethodPost, s.apiWeChatCreateNativeOrder, true},
		{apiWeChatNativeOrderStatus, http.MethodGet, s.apiWeChatOrderStatus, true},
	}

	for _, rt := range routes {
		r := rt
		mux.HandleFunc(r.pattern, func(w http.ResponseWriter, req *http.Request) {
			s.wrapApi(r, w, req)
		})
	}

	//if err := s.registerSpaStatic(mux); err != nil {
	//	return err
	//}

	handler := s.loggingMiddleware(mux)
	srv := &http.Server{
		Addr:              s.cfg.srvAddr(),
		Handler:           handler,
		ReadTimeout:       time.Duration(s.cfg.ReadTimeout) * time.Second,
		ReadHeaderTimeout: time.Duration(s.cfg.ReadTimeout) * time.Second,
	}

	s.srv = srv
	return nil
}

func (s *HttpSrv) wrapApi(r route, w http.ResponseWriter, req *http.Request) {
	if req.Method != r.method {
		w.Header().Set("Allow", r.method)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	if r.requireLogin {
		uid, err := s.currentUserFromCookie(req)
		if err != nil {
			s.log.Err(err).Msg("parse wx_user cookie failed")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if uid == "" {
			http.Error(w, "unauthorized: please login first", http.StatusUnauthorized)
			return
		}

		req = req.WithContext(context.WithValue(req.Context(), ctxKeyUserID, uid))
	}

	r.handler(w, req)
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

func (s *HttpSrv) registerSpaStatic(mux *http.ServeMux) error {
	staticDir := s.cfg.StaticDir
	if stat, err := os.Stat(staticDir); err != nil || !stat.IsDir() {
		if err == nil {
			return fmt.Errorf("%s is not a directory", staticDir)
		}
		return err
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 清理 path，避免 "../" 等异常路径
		cleanPath := filepath.Clean(r.URL.Path)

		// 拼出静态文件路径，例如 /assets/xxx.js -> {staticDir}/assets/xxx.js
		p := filepath.Join(staticDir, cleanPath)

		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			// 找到文件，直接返回
			http.ServeFile(w, r, p)
			return
		}

		// 找不到对应文件（或是目录），统一回退到 index.html
		indexPath := filepath.Join(staticDir, "index.html")
		http.ServeFile(w, r, indexPath)
	})

	return nil
}
