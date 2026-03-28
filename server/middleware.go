package server

import (
	"bytes"
	"crypto/subtle"
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"go.uber.org/zap"
)

func withAuthMiddleware(next http.Handler, authToken string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authToken == "" {
			next.ServeHTTP(w, r)
			return
		}

		token := extractBearerToken(r)
		if subtle.ConstantTimeCompare([]byte(token), []byte(authToken)) != 1 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func withOriginValidation(next http.Handler, allowedOrigins []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if len(allowedOrigins) > 0 && origin != "" && !slices.Contains(allowedOrigins, origin) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func extractBearerToken(r *http.Request) string {
	token, _ := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	return token
}

func withRequestLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rpc := peekJSONRPCRequest(r)

		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)

		fields := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"status", sw.status,
			"latency_ms", time.Since(start).Milliseconds(),
		}
		if rpc.Method != "" {
			fields = append(fields, "rpc_method", rpc.Method)
		}
		if rpc.ToolName != "" {
			fields = append(fields, "tool", rpc.ToolName)
		}

		if len(rpc.Params) > 0 {
			fields = append(fields, "params_bytes", len(rpc.Params))
		}
		switch {
		case sw.status >= 500:
			zap.S().Errorw("request", fields...)
		case sw.status >= 400:
			zap.S().Warnw("request", fields...)
		default:
			zap.S().Infow("request", fields...)
		}
	})
}

type jsonRPCInfo struct {
	Method   string
	ToolName string
	Params   json.RawMessage
}

const maxPeekSize = 8 * 1024

func peekJSONRPCRequest(r *http.Request) jsonRPCInfo {
	if r.Body == nil {
		return jsonRPCInfo{}
	}
	peeked, err := io.ReadAll(io.LimitReader(r.Body, maxPeekSize))
	r.Body = io.NopCloser(io.MultiReader(bytes.NewReader(peeked), r.Body))
	if err != nil || len(peeked) == 0 {
		return jsonRPCInfo{}
	}

	var req struct {
		Method string          `json:"method"`
		Params json.RawMessage `json:"params"`
	}
	if json.Unmarshal(peeked, &req) != nil {
		return jsonRPCInfo{}
	}

	info := jsonRPCInfo{
		Method: req.Method,
		Params: req.Params,
	}

	if req.Method == "tools/call" {
		var params struct {
			Name string `json:"name"`
		}
		if json.Unmarshal(req.Params, &params) == nil {
			info.ToolName = params.Name
		}
	}
	return info
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *statusWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
