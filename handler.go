package log4nova

import (
    "net/http"
    "time"
    "github.com/satori/go.uuid"
    "strconv"
)

type NovaHandler struct {
    handler http.Handler
    logger  INovaLogger
}

//NewNovaHandler creates a new instance of the Nova Logging Handler
func NewNovaHandler (logger INovaLogger, handler http.Handler) *NovaHandler {
    logger.Start()
    return &NovaHandler{
        handler: handler,
        logger: logger,
    }
}

func (nl *NovaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Get the start time
    startTime := time.Now()

    // Capture the response data
    lwr := loggingResponseWriter{w: w, captureBody: false}
    nl.handler.ServeHTTP(&lwr, r)

    endTime := time.Now()

    //Send to log4nova
    nl.logger.WithFields(Fields{
        "path": r.URL.Path,
        "statusCode": strconv.Itoa(lwr.code),
        "requestURI" : r.RequestURI,
        "requestMethod": r.Method,
        "userAgent": r.UserAgent(),
        "logId": uuid.NewV1(),
        "responseTime": endTime.Sub(startTime).String(),
    }).Infof("Logging Response")
}