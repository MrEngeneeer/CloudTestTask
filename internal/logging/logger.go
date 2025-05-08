package logging

import (
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Логирование

type Logger struct {
	out *log.Logger
}

func New(w io.Writer) *Logger {
	return &Logger{
		out: log.New(w, "", log.LstdFlags|log.Lmicroseconds),
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.out.Printf("[INFO] "+format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.out.Printf("[ERROR] "+format, v...)
}

func (l *Logger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		l.Info("%s %s %s in %v", r.RemoteAddr, r.Method, r.RequestURI, time.Since(start))
	})
}

var Std = New(os.Stdout)
