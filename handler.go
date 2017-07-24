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
    "os"
    "errors"
    "fmt"
    "github.com/satori/go.uuid"
)

// Stats data structure
type NovaLogger struct {
    handler             http.Handler
    captureResponseBody bool
    inLock              sync.Mutex
    outLock             sync.Mutex
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
    transCfg := client.DefaultTransportConfig()
    auth := rtclient.BasicAuth(clientID, clientSecret)
    httpCl := &http.Client{}
    transportWithClient := rtclient.NewWithClient("api-integ.logface.io", client.DefaultBasePath, transCfg.Schemes, httpCl)
    transportWithClient.Transport = httpCl.Transport

    transportWithClient.DefaultAuthentication = auth

    return &NovaLogger{
        inLock:   sync.Mutex{},
        outLock:  sync.Mutex{},
        client:     client.New(transportWithClient, strfmt.Default).Events,
        sendInterval: 1000,
        clientID: clientID,
        clientSecret: clientSecret,
        captureResponseBody: captureResponseBody,
        handler: handler,
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
    go func() {
        for {
            select {
            case log, ok := <-in:
                if !ok {
                    //for {
                    //    if len(out) > 0 {
                    //        fmt.Println("Sleeping again")
                    //        time.Sleep(time.Duration(nl.sendInterval) / 2 * time.Millisecond)
                    //    } else {
                    //        fmt.Println("Breaking out")
                    //        break
                    //    }
                    //}
                    break
                } else {
                    fmt.Println("Getting read lock")
                    //nl.outLock.Lock()
                    event := models.Event{
                        Entity: "log4nova",
                        Source: nl.host,
                        Event:  map[string]*string{
                            "raw": &log,
                        },
                    }
                    fmt.Println("Pushed formatted log to channel")
                    out <- &event
                    //nl.outLock.Unlock()
                    fmt.Println("Released read lock")
                }
            }
        }
        fmt.Println("Closing out channel")
        close(out)
    }()

    return out
}

func (nl *NovaLogger) flushFromOutputChannel(out chan *models.Event) error {
    fmt.Println("Flushing from output channel")
    for {
        select {
        case event, ok := <- out:
            if !ok {
                for {
                    if len(out) > 0 {
                        time.Sleep(time.Duration(nl.sendInterval)/2 * time.Millisecond)
                    } else {
                        break
                    }
                }
            } else {
                fmt.Println("Getting write lock")
                //nl.writeLock.Lock()
                ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
                defer cancel()
                params := &events.EventsParams{
                    Events:  models.Events{event},
                    Context: ctx,
                }
                fmt.Println("Sending events")
                auth := rtclient.BasicAuth(nl.clientID, nl.clientSecret)
                _, err := nl.client.Events(params, auth)

                if err != nil {
                    fmt.Println(fmt.Errorf("Error sending to log-store: %v", err))
                } else {

                }

                fmt.Println("Releasing write lock")
                //nl.writeLock.Unlock()
            }
        }
        //time.Sleep(time.Duration(nl.sendInterval) * time.Millisecond)
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
    logger.SetFormatter(&logger.JSONFormatter{})

    // Get your logs
    startTime := time.Now()
    logger.WithFields(logger.Fields{
        "request": stringify(r),
    }).Infof("Logging Request")

    lwr := loggingResponseWriter{w: w, captureBody: nl.captureResponseBody}
    nl.handler.ServeHTTP(&lwr, r)
    endTime := time.Now()
    uuid_evt := uuid.NewV1()
    fmt.Println(uuid_evt)
    logger.WithFields(logger.Fields{
        "log_id": uuid_evt,
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