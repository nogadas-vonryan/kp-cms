package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

var logSink = func(msg string) {
	log.Print(msg)
}

// SetLogSink configures the log sink used by the API package.
// If nil is provided, it falls back to the standard logger.
func SetLogSink(sink func(string)) {
	if sink == nil {
		logSink = func(msg string) {
			log.Print(msg)
		}
		return
	}
	logSink = sink
}

// AccessLogMiddleware logs HTTP requests through the configured log sink.
func AccessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		duration := time.Since(start)
		requestID := middleware.GetReqID(r.Context())
		remoteAddr := r.RemoteAddr
		if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
			remoteAddr = realIP
		}
		logSink(fmt.Sprintf("%s %s %d %dB %s request_id=%s remote_ip=%s ua=%q", r.Method, r.URL.Path, ww.Status(), ww.BytesWritten(), duration, requestID, remoteAddr, r.UserAgent()))
	})
}
