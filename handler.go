package log4nova

import (
    "net/http"
    "time"
    "fmt"
    "github.com/satori/go.uuid"
    "strconv"
)

type NovaHandler struct {
    handler http.Handler
    logger  *NovaLogger
}

//NewNovaHandler creates a new instance of the Nova Logging Handler
func NewNovaHandler (handler http.Handler, logger *NovaLogger) *NovaHandler {
    return &NovaHandler{
        handler: handler,
        logger: logger,
    }
}

func (nl *NovaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    logger := nl.logger

    // Get your logs
    startTime := time.Now()

    lwr := loggingResponseWriter{w: w, captureBody: false}
    nl.handler.ServeHTTP(&lwr, r)
    endTime := time.Now()
    uuid_evt := uuid.NewV1()
    fmt.Println(uuid_evt)
    logger.WithFields(Fields{
        "api": r.URL.Path,
        "status_code": strconv.Itoa(lwr.code),
        "RequestURL" : r.RequestURI,
        "RequestMethod": r.Method,
        "UserAgent": r.UserAgent(),
        "log_id": uuid_evt,
        "response_time": endTime.Sub(startTime).String(),
    }).Infof("Logging Response")
}