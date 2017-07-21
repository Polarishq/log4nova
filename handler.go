package log4nova

import (
    "net/http"
    "github.com/Polarishq/bouncer/models"
    "sync"
    "time"
    "context"
    logger "github.com/Sirupsen/logrus"
    "github.com/Polarishq/bouncer/client/events"
    rtclient "github.com/go-openapi/runtime/client"
    "github.com/Polarishq/bouncer/client"
    "github.com/go-openapi/strfmt"
    "net/http/httputil"
    "strings"
    "bytes"
    "bufio"
    "github.com/x-cray/logrus-prefixed-formatter"
    "os"
    "errors"
)

// Stats data structure
type NovaLogger struct {
    handler             http.Handler
    captureResponseBody bool
    readLock            sync.Mutex
    writeLock           sync.Mutex
    client              events.ClientInterface
    sendInterval        int
    clientID            string
    clientSecret        string
    host                string
}

type loggingResponseWriter struct {
    headers     http.Header
    w           http.ResponseWriter
    data        []byte
    code        int
    captureBody bool
}


type novaLogFormat struct {

}

//NewNovaHandler creates a new instance of the Nova Logging Handler
func NewNovaHandler(handler http.Handler, captureResponseBody bool) *NovaLogger {
    clientID := os.Getenv("NOVA_CLIENT_ID")
    clientSecret := os.Getenv("NOVA_CLIENT_SECRET")
    if clientID == "" || clientSecret == "" {
        panic(errors.New("Client ID or Client Secret not set properly on env"))
    }

    return &NovaLogger{
        readLock:   sync.Mutex{},
        writeLock:  sync.Mutex{},
        client:     client.New(rtclient.New(client.DefaultHost, client.DefaultBasePath, client.DefaultSchemes), strfmt.Default).Events,
        sendInterval: 1000,
        clientID: clientID,
        clientSecret: clientSecret,
    }
}


func (nl *NovaLogger) writeLogsToChannel(logs *bytes.Buffer, ch chan string) {
    scanner := bufio.NewScanner(logs)
    for scanner.Scan() {
        log := scanner.Text()

        ch <- log
    }
    close(ch)
    return
}

func (nl *NovaLogger) formatLogs(in chan string) chan *models.Event {
    //Format logs for splunk
    out := make(chan *models.Event)
        for {
            select {
            case log, ok := <-in:
                if !ok {
                    for {
                        if len(out) > 0 {
                            time.Sleep(time.Duration(nl.sendInterval)/2 * time.Millisecond)
                        } else {
                            break
                        }
                    }
                    break
                } else {
                    nl.readLock.Lock()
                    event := models.Event{
                        Entity: "log4nova",
                        Source: nl.host,
                        Event:  map[string]*string{"raw": &log},
                    }
                    out <- &event
                    nl.readLock.Unlock()
                }
            }
        }
    return out
}

func (nl *NovaLogger) flushFromOutputChannel(out chan *models.Event) error {
    for {
        select {
            case event, ok := <- out:
                if !ok {

                } else {
                    nl.writeLock.Lock()
                    ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
                    defer cancel()
                    params := &events.EventsParams{
                        Events:  models.Events{event},
                        Input:   &"input type here",
                        Context: ctx,
                    }
                    auth := rtclient.BasicAuth("", "")
                    _, err := nl.client.Events(params, auth)

                    if err != nil {

                    } else {

                    }

                    nl.writeLock.Unlock()
                }
        }
        time.Sleep(time.Duration(nl.sendInterval) * time.Millisecond)
    }
}

func stringify(r *http.Request) string {
    dump, _ := httputil.DumpRequest(r, true)
    return strings.Replace(string(dump), "\n", " ", -1)
}

func (nl *NovaLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    nl.host = r.Host
    // create logging
    buf := new(bytes.Buffer)
    logger.SetOutput(buf)
    logger.SetFormatter(&prefixed.TextFormatter{TimestampFormat: "Jan 02 03:04:05.000"})

    // Get your logs
    startTime := time.Now()
    logger.WithFields(logger.Fields{
        "request": stringify(r),
    }).Infof("Logging Request")

    lwr := loggingResponseWriter{w: w, captureBody: nl.captureResponseBody}
    nl.handler.ServeHTTP(&lwr, r)
    endTime := time.Now()
    logger.WithFields(logger.Fields{
        "response_code": lwr.code,
        "headers": lwr.headers,
        "body": string(lwr.data),
        "response_start": startTime,
        "response_end": endTime,
        "response_time": endTime.Sub(startTime),
    }).Infof("Logging Response")

    // Begin the output process
    in := make(chan string)
    go nl.writeLogsToChannel(buf, in)
    out := nl.formatLogs(in)
    go nl.flushFromOutputChannel(out)
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