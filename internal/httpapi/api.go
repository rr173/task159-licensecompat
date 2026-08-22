package httpapi

import (
	"encoding/json"
	"errors"
	"licensecompat.local/internal/model"
	"licensecompat.local/internal/service"
	"net/http"
	"strings"
	"time"
)

type API struct {
	service *service.Service
	mux     *http.ServeMux
}

func New(s *service.Service) *API {
	api := &API{service: s, mux: http.NewServeMux()}
	api.routes()
	return api
}
func (a *API) Handler() http.Handler { return a.mux }
func (a *API) routes() {
	a.mux.HandleFunc("GET /healthz", a.health)
	a.mux.HandleFunc("GET /v1/self-check", a.selfCheck)
	a.mux.HandleFunc("GET /v1/recovery", a.recovery)
	a.mux.HandleFunc("POST /v1/recovery", a.recover)
	a.mux.HandleFunc("GET /v1/policies", a.policies)
	a.mux.HandleFunc("POST /v1/policies", a.createPolicy)
	a.mux.HandleFunc("POST /v1/policies/{id}/activate", a.activatePolicy)
	a.mux.HandleFunc("GET /v1/submissions", a.submissions)
	a.mux.HandleFunc("POST /v1/submissions", a.submit)
	a.mux.HandleFunc("GET /v1/submissions/{id}", a.submission)
	a.mux.HandleFunc("POST /v1/submissions/{id}/analyze", a.analyze)
	a.mux.HandleFunc("GET /v1/analyses/{id}", a.analysis)
	a.mux.HandleFunc("GET /v1/analyses/{id}/findings", a.findings)
	a.mux.HandleFunc("GET /v1/analyses/{id}/summary", a.summary)
	a.mux.HandleFunc("GET /v1/analyses/{id}/explain", a.explain)
	a.mux.HandleFunc("GET /v1/analyses/{id}/waivers", a.waivers)
	a.mux.HandleFunc("POST /v1/analyses/{id}/waivers", a.requestWaiver)
	a.mux.HandleFunc("POST /v1/analyses/{id}/publish", a.publish)
	a.mux.HandleFunc("POST /v1/waivers/{id}/approve", a.approveWaiver)
	a.mux.HandleFunc("GET /v1/compare", a.compare)
	a.mux.HandleFunc("GET /v1/audit/{type}/{id}", a.audit)
}
func (a *API) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
func (a *API) selfCheck(w http.ResponseWriter, r *http.Request) {
	value, err := a.service.SelfCheck(r.Context())
	respond(w, value, err, http.StatusOK)
}
func (a *API) recovery(w http.ResponseWriter, r *http.Request) {
	value, err := a.service.Store().RecoverableAnalyses(r.Context())
	respond(w, value, err, http.StatusOK)
}
func (a *API) recover(w http.ResponseWriter, r *http.Request) {
	value, err := a.service.Recover(r.Context())
	respond(w, map[string]any{"recovered": value}, err, http.StatusOK)
}
func (a *API) policies(w http.ResponseWriter, r *http.Request) {
	value, err := a.service.Policies(r.Context())
	respond(w, value, err, http.StatusOK)
}
func (a *API) createPolicy(w http.ResponseWriter, r *http.Request) {
	var value model.Policy
	if err := readJSON(r, &value); err != nil {
		respond(w, nil, err, http.StatusBadRequest)
		return
	}
	created, err := a.service.CreatePolicy(r.Context(), value)
	respond(w, created, err, http.StatusCreated)
}
func (a *API) activatePolicy(w http.ResponseWriter, r *http.Request) {
	value, err := a.service.ActivatePolicy(r.Context(), r.PathValue("id"))
	respond(w, value, err, http.StatusOK)
}
func (a *API) submissions(w http.ResponseWriter, r *http.Request) {
	value, err := a.service.Submissions(r.Context())
	respond(w, value, err, http.StatusOK)
}
func (a *API) submit(w http.ResponseWriter, r *http.Request) {
	var value model.Submission
	if err := readJSON(r, &value); err != nil {
		respond(w, nil, err, http.StatusBadRequest)
		return
	}
	stored, duplicate, err := a.service.Submit(r.Context(), value)
	if duplicate {
		writeJSON(w, http.StatusOK, map[string]any{"duplicate": true, "submission": stored})
		return
	}
	respond(w, map[string]any{"duplicate": false, "submission": stored}, err, http.StatusCreated)
}
func (a *API) submission(w http.ResponseWriter, r *http.Request) {
	value, err := a.service.Submission(r.Context(), r.PathValue("id"))
	respond(w, value, err, http.StatusOK)
}
func (a *API) analyze(w http.ResponseWriter, r *http.Request) {
	value, err := a.service.Analyze(r.Context(), r.PathValue("id"))
	respond(w, value, err, http.StatusCreated)
}
func (a *API) analysis(w http.ResponseWriter, r *http.Request) {
	value, err := a.service.Decision(r.Context(), r.PathValue("id"))
	respond(w, value, err, http.StatusOK)
}
func (a *API) findings(w http.ResponseWriter, r *http.Request) {
	value, err := a.service.Findings(r.Context(), r.PathValue("id"))
	respond(w, value, err, http.StatusOK)
}
func (a *API) summary(w http.ResponseWriter, r *http.Request) {
	value, err := a.service.Report(r.Context(), r.PathValue("id"))
	respond(w, value, err, http.StatusOK)
}
func (a *API) explain(w http.ResponseWriter, r *http.Request) {
	value, err := a.service.Explain(r.Context(), r.PathValue("id"))
	respond(w, value, err, http.StatusOK)
}
func (a *API) waivers(w http.ResponseWriter, r *http.Request) {
	value, err := a.service.Waivers(r.Context(), r.PathValue("id"))
	respond(w, value, err, http.StatusOK)
}
func (a *API) requestWaiver(w http.ResponseWriter, r *http.Request) {
	var body struct {
		FindingID string     `json:"finding_id"`
		Reason    string     `json:"reason"`
		ExpiresAt *time.Time `json:"expires_at"`
	}
	if err := readJSON(r, &body); err != nil {
		respond(w, nil, err, http.StatusBadRequest)
		return
	}
	value, err := a.service.RequestWaiver(r.Context(), r.PathValue("id"), body.FindingID, body.Reason, body.ExpiresAt)
	respond(w, value, err, http.StatusCreated)
}
func (a *API) approveWaiver(w http.ResponseWriter, r *http.Request) {
	value, err := a.service.ApproveWaiver(r.Context(), r.PathValue("id"))
	respond(w, value, err, http.StatusOK)
}
func (a *API) publish(w http.ResponseWriter, r *http.Request) {
	value, err := a.service.Publish(r.Context(), r.PathValue("id"))
	respond(w, value, err, http.StatusOK)
}
func (a *API) compare(w http.ResponseWriter, r *http.Request) {
	value, err := a.service.Compare(r.Context(), r.URL.Query().Get("left"), r.URL.Query().Get("right"))
	respond(w, value, err, http.StatusOK)
}
func (a *API) audit(w http.ResponseWriter, r *http.Request) {
	value, err := a.service.Audit(r.Context(), r.PathValue("type"), r.PathValue("id"))
	respond(w, value, err, http.StatusOK)
}
func readJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func respond(w http.ResponseWriter, value any, err error, success int) {
	if err == nil {
		writeJSON(w, success, value)
		return
	}
	status := http.StatusInternalServerError
	if model.IsInvalid(err) {
		status = http.StatusBadRequest
	}
	if model.IsNotFound(err) {
		status = http.StatusNotFound
	}
	if model.IsConflict(err) || model.IsImmutable(err) || errors.Is(err, model.ErrInvalidState) {
		status = http.StatusConflict
	}
	if strings.Contains(err.Error(), "circular") {
		status = http.StatusBadRequest
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
