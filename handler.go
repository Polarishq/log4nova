package log4nova

import (
"net/http"
    "time"
    "github.com/aws/aws-sdk-go/service/kinesis"
    "github.com/aws/aws-sdk-go/aws/session"
    "encoding/json"
)

// Stats data structure
type NovaLogger struct {
    handler             http.Handler
    svc                 *kinesis.Kinesis
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
func NewNovaHandler(handler http.Handler, captureResponseBody bool, streamName string) *NovaLogger {
    sess := session.Must(session.NewSession())
    svc := kinesis.New(sess)
    return &NovaLogger{handler, svc, captureResponseBody, streamName}
}

func (nl *NovaLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    lwr := &loggingResponseWriter{w: w, captureBody: nl.captureResponseBody}
    nl.handler.ServeHTTP(lwr, r)
    end := time.Now()
    key := r.RequestURI
    _, err := nl.sendToKinesis(start, end, key, lwr)
    if err != nil {
        panic(err)
    }
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

func (nl *NovaLogger) sendToKinesis(start time.Time, end time.Time, partitionKey string, writer *loggingResponseWriter) (*kinesis.PutRecordOutput, error) {
    jsonBytes, err := json.Marshal(writer)
    if err != nil {
        return nil, err
    }

    input := &kinesis.PutRecordInput{
        StreamName: &nl.streamName,
        PartitionKey: &partitionKey,
        Data: jsonBytes,
    }
    return nl.svc.PutRecord(input)
}