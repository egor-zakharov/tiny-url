package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Logger struct {
	log *zap.Logger
}

func NewLogger() *Logger {
	return &Logger{log: zap.NewNop()}
}

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func (l *Logger) Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	l.log = zl
	return nil
}

func (l *Logger) RequestLogger(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		logRw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		h(&logRw, r)

		duration := time.Since(start).Milliseconds()

		l.log.Sugar().With(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size,
		).Info("got HTTP request")

	})
}

func (l *Logger) GetLog() *zap.Logger {
	return l.log
}
