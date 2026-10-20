package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workspaceapp "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

func (h *handler) registerOrcaCLI(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/local-projects/{id}/orca/commands", func(w http.ResponseWriter, r *http.Request) {
		if h.services.OrcaExecution == nil {
			writeError(w, r, http.StatusServiceUnavailable, &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: "the product Orca runtime is not configured", RequiredAction: "start_installed_swa_service"})
			return
		}
		var command workspaceapp.OrcaCommand
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64*1024))
		decoder.DisallowUnknownFields()
		err := decoder.Decode(&command)
		if err == nil {
			var extra any
			if e := decoder.Decode(&extra); e != io.EOF {
				err = errors.New("exactly one JSON command is required")
			}
		}
		if err != nil || (command.ProjectID != "" && command.ProjectID != r.PathValue("id")) || (command.RequestID != "" && r.Header.Get("Idempotency-Key") != command.RequestID) {
			writeError(w, r, http.StatusBadRequest, &apierrors.ServiceError{Code: apierrors.ValidationFailed, Message: "invalid command, project or request header", RequiredAction: "correct_command_fields"})
			return
		}
		// project 形式只用于 open 和无 selector 的 request 查询，其余操作由当前令牌限定。
		if command.Action == "controller.open" || command.Action == "request.show" && command.Controller == "" {
			command.ProjectID = r.PathValue("id")
		}
		caller := localCaller(r)
		if caller.Kind == "human" {
			caller.ProjectID = r.PathValue("id")
		}
		reply, err := h.services.OrcaExecution.ExecuteOrca(r.Context(), caller, command)
		if err == nil {
			writeJSON(w, http.StatusOK, reply)
			return
		}
		var native *workspaceapp.OrcaNativeError
		if errors.As(err, &native) {
			reply.OK, reply.Error = false, native
			if reply.RequestID == "" {
				reply.RequestID = command.RequestID
			}
			writeJSON(w, http.StatusConflict, reply)
			return
		}
		h.attentionResult(w, r, http.StatusOK, nil, err)
	})
}
