package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/hopwesley/wenxintai/server/service"
)

func main() {
	http.HandleFunc("/api/start-session", withCommon(service.StartReqHandler))
	http.HandleFunc("/api/hello", withCommon(service.HelloHandler))

	// 启动 HTTP 服务器，监听 80 端口
	log.Println("🚀 Server running on http://localhost:80")
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Fatal(err)
	}
}

// 公共处理：接受一个具体的 Handler，再返回一个带公共逻辑的 Handler
func withCommon(handler http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			resp := service.ApiRes{
				Success: false,
				Message: "只支持 POST 请求",
			}
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(resp)
			return
		}

		handler(w, r)
		log.Printf("[DONE] %s in %v", r.URL.Path, time.Since(start))
	}
}
