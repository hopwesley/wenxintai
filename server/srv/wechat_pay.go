package srv

import (
	"net/http"
)

//const wechatForwardURL = "https://sharp-happy-grouse.ngrok-free.app/api/pay/"

func (s *HttpSrv) apiWeChatPayCallBack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if len(s.cfg.PaymentForward) > 0 {
		s.forwardCallback(w, r, s.cfg.PaymentForward)
		return
	}

	s.processWeChatPayment(w, r)
}

func (s *HttpSrv) processWeChatPayment(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, &CommonRes{Ok: true})
}
