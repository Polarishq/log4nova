package log4nova

import (
    "net/http"
)

// Stats data structure
type NovaLogger struct {
    handler             http.Handler
    captureResponseBody bool
    streamName          string
}

type loggingResponseWriter struct {
    headers     http.Header
    w           http.ResponseWriter
    data        []byte
    code        int
    captureBody bool
}

//NewNovaHandler creates a new instance of the Nova Logging Handler
func NewNovaHandler(handler http.Handler, captureResponseBody bool) *NovaLogger {
    return &NovaLogger{}
}

func (nl *NovaLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    //start := time.Now()
    lwr := &loggingResponseWriter{w: w, captureBody: nl.captureResponseBody}
    nl.handler.ServeHTTP(lwr, r)
    //end := time.Now()
    //key := r.RequestURI
}

func (lw *loggingResponseWriter) Write(b []byte) (int, error) {
    if lw.captureBody {
        lw.data = append(lw.data, b...)
    }
    return lw.w.Write(b)
}

func (lw *loggingResponseWriter) WriteHeader(code int) {
    lw.headers = lw.Header()
    lw.code = code
    lw.w.WriteHeader(code)
}

func (lw *loggingResponseWriter) Header() http.Header {
    return lw.w.Header()
}