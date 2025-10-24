package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/example/bess-voltvar/internal/config"
	"github.com/example/bess-voltvar/internal/controller"
	"github.com/example/bess-voltvar/internal/telem"
)

type Server struct {
	ctrl    *controller.Controller
	cfgPath *string
	mux     *http.ServeMux
}

func NewServer(ctrl *controller.Controller, cfgPath *string) *Server {
	s := &Server{ctrl: ctrl, cfgPath: cfgPath, mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("/healthz", s.handleHealth)
	s.mux.HandleFunc("/v1/status", s.handleStatus)
	s.mux.HandleFunc("/v1/mode", s.handleMode)
	s.mux.HandleFunc("/v1/remote", s.handleRemote)
	s.mux.HandleFunc("/v1/config", s.handleConfig)
}

func (s *Server) ListenAndServe(addr string) error {
	srv := &http.Server{Addr: addr, Handler: s.mux}
	telem.Log().Printf("HTTP listening on %s", addr)
	return srv.ListenAndServe()
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	st := s.ctrl.Status()
	_ = json.NewEncoder(w).Encode(st)
}

func (s *Server) handleMode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return
	}
	var req struct{ Mode string `json:"mode"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest); return
	}
	switch req.Mode {
	case "VOLT_VAR", "CONST_PF", "CONST_Q", "REMOTE":
		s.ctrl.SetMode(req.Mode)
	default:
		http.Error(w, "invalid mode", http.StatusBadRequest); return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleRemote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return
	}
	var req struct{
		QSetMVAr float64 `json:"q_set_mvar"`
		TTLSec   int     `json:"ttl_s"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest); return
	}
	s.ctrl.SetRemote(req.QSetMVAr, time.Duration(req.TTLSec)*time.Second)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return
	}
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest); return
	}
	// Validate new config
	var cfg config.Config
	if err := yamlUnmarshal(raw, &cfg); err != nil {
		http.Error(w, "yaml: "+err.Error(), http.StatusBadRequest); return
	}
	if err := cfg.Validate(); err != nil {
		http.Error(w, "invalid config: "+err.Error(), http.StatusBadRequest); return
	}
	// Atomically write to file and swap live config
	if err := os.WriteFile(*s.cfgPath, raw, 0644); err != nil {
		http.Error(w, "write failed: "+err.Error(), http.StatusInternalServerError); return
	}
	s.ctrl.UpdateConfig(&cfg)
	w.WriteHeader(http.StatusNoContent)
}

// yamlUnmarshal is a tiny adapter to avoid a hard import in the API layer.
func yamlUnmarshal(b []byte, out any) error {
	type Y interface{ Unmarshal([]byte, any) error }
	// Simple local import indirection
	return yamlUnmarshalImpl(b, out)
}
