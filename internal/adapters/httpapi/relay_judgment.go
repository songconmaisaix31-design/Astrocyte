package httpapi

import "net/http"

// 独立子 mux 保持原 S5 路径，避免与已注册项目的交叉通配符冲突。
func (h *handler) registerRelayJudgment(mux *http.ServeMux, services JudgmentServices) {
	judgments := http.NewServeMux()
	h.registerJudgment(judgments, services)
	judgments.HandleFunc("/api/v1/", h.handleAPINotFound)
	mux.Handle("/api/v1/local-projects/", judgments)
	mux.Handle("/api/v1/attention-taste-profile", judgments)
}
