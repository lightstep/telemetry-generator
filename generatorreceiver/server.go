package generatorreceiver

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/flags"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

type httpServer struct {
	serverType string
	server     *http.Server
	logger     *zap.Logger
	config     *Config
}

type flagHttpResponse struct {
	Name       string `json:"name"`
	Enabled    bool   `json:"enabled"`
	DurationNs int64  `json:"duration"`
}

func (h *httpServer) getFlags(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(w, "bad request")
	}

	jsonFlags := make([]flagHttpResponse, 0)
	for flagName, flagVal := range flags.Manager.GetFlags() {
		jsonFlags = append(jsonFlags, flagHttpResponse{
			Name:       flagName,
			Enabled:    flagVal.Active(),
			DurationNs: int64(flagVal.CurrentDuration()),
		})
	}
	resp, err := json.MarshalIndent(jsonFlags, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "internal error: could not set attr proc: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "%s", string(resp))
}

func (h *httpServer) setFlag(w http.ResponseWriter, r *http.Request) {
	f := r.URL.Query().Get("flag")
	v := r.URL.Query().Get("enabled")
	if f == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(w, "bad request: expected flag param")
		return
	}

	if v == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(w, "bad request: expected enabled param")
		return
	}
	reqFlag := flags.Manager.GetFlag(f)

	if reqFlag == nil {
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintf(w, "flag %s not found", f)
		return
	}

	if v == "true" || v == "1" {
		reqFlag.Enable()
	} else if v == "false" || v == "0" {
		reqFlag.Disable()
	} else {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(w, "bad request: enabled must be true or false")
		return
	}
	w.WriteHeader(http.StatusAccepted)
	_, _ = fmt.Fprintf(w, "flag %s updated", f)
}

func (h *httpServer) Start(_ context.Context, host component.Host) error {
	handler := http.NewServeMux()
	handler.HandleFunc("/api/v1/flags", h.getFlags)
	handler.HandleFunc("/api/v1/flag", h.setFlag)

	var listener net.Listener
	var err error
	h.server = h.config.ApiIngress.ToServer(handler)
	listener, err = h.config.ApiIngress.ToListener()
	if err != nil {
		h.logger.Fatal("failed to bind to address %s: %w", zap.String("endpoint", h.config.ApiIngress.Endpoint), zap.Error(err))
		return err
	}
	h.logger.Info("starting api server")
	go func() {
		if err := h.server.Serve(listener); err != http.ErrServerClosed {
			host.ReportFatalError(err)
		}
	}()

	return nil
}

func (h *httpServer) Shutdown(ctx context.Context) error {
	return h.server.Shutdown(ctx)
}

func newHTTPServer(config *Config, logger *zap.Logger) (*httpServer, error) {
	h := &httpServer{
		config: config,
		logger: logger,
	}

	return h, nil
}
