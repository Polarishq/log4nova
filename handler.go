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
)

// Stats data structure
type NovaLogger struct {
    handler             http.Handler
    captureResponseBody bool
    readLock            sync.Mutex
    writeLock           sync.Mutex
    client              events.ClientInterface
    flushInterval       int
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
    return &NovaLogger{
        readLock:   sync.Mutex{},
        writeLock:  sync.Mutex{},
        client:     client.New(rtclient.New("", "", []string{}), strfmt.Default).Events,
        flushInterval: 1000,
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
                            time.Sleep(100 * time.Millisecond)
                        } else {
                            break
                        }
                    }
                    break
                } else {
                    nl.readLock.Lock()
                    event := models.Event{
                        //Entity: i.entity,
                        //Source: i.source,
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
        time.Sleep(time.Duration(nl.flushInterval) * time.Millisecond)
    }
}


//func (i *Input) Send(r io.Reader) {
//    mutex := sync.Mutex{}
//
//    events := inputModels.Events{}
//    log.Debugf("Starting scanner loop")
//
//    ch := make(chan string)
//
//    // Read from stdin, send all lines to a channel
//    go func(ch chan string) {
//        scanner := bufio.NewScanner(r)
//        for scanner.Scan() {
//            line := scanner.Text()
//            // log.Debugf("line: %s", line)
//            ch <- line
//        }
//        if err := scanner.Err(); err != nil {
//            log.WithError(err).Errorf("Error received from scanner")
//        }
//        close(ch)
//        return
//    }(ch)
//
//    // Read events slice, flush to Events endpoint
//    go func() {
//        for {
//            if len(events) > 0 {
//                mutex.Lock()
//                var ctx context.Context
//                var cancel context.CancelFunc
//                if i.ctx == nil {
//                    ctx, cancel = context.WithTimeout(context.Background(), 5000*time.Millisecond)
//                    defer cancel()
//                } else {
//                    ctx = i.ctx
//                }
//                params := &inputClientEvents.EventsParams{
//                    Events:  events,
//                    Input:   &i.input,
//                    Context: ctx,
//                }
//                eok, err := i.client.Events(params, i.creds.ClientAuth)
//                if eok != nil {
//                    log.Infof("Bytes: %d Events: %d", eok.Payload.Bytes, eok.Payload.Count)
//                }
//                if err != nil {
//                    log.Errorf("Error received from Events: %s", err)
//                }
//                events = inputModels.Events{}
//                mutex.Unlock()
//            }
//            time.Sleep(time.Duration(i.flushInterval) * time.Millisecond)
//        }
//    }()
//
//    // Read from channel, add to events Slice
//readloop:
//    for {
//        select {
//        case line, ok := <-ch:
//            if !ok {
//                for {
//                    if len(events) > 0 {
//                        log.Debugf("Waiting for events queue to empty")
//                        time.Sleep(100 * time.Millisecond)
//                    } else {
//                        break
//                    }
//                }
//                break readloop
//            } else {
//                mutex.Lock()
//                // log.Debugf("line: %s", line)
//                event := inputModels.Event{
//                    Entity: i.entity,
//                    Source: i.source,
//                    Event:  map[string]string{"raw": line},
//                }
//                events = append(events, &event)
//                mutex.Unlock()
//            }
//        }
//    }
//}

func stringify(r *http.Request) string {
    dump, _ := httputil.DumpRequest(r, true)
    return strings.Replace(string(dump), "\n", " ", -1)
}

func (nl *NovaLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // create logging
    buf := new(bytes.Buffer)
    logger.SetOutput(buf)
    logger.SetFormatter(&prefixed.TextFormatter{TimestampFormat: "Jan 02 03:04:05.000"})

    // Get your logs
    startTime := time.Now()
    logger.WithFields(logger.Fields{
        "request": stringify(r),
    }).Infof("Logging Request")

    //logger.Infof("LoggingHandler: request: %+v string_request %+v", r, stringify(r))
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

    //Debugf("LoggingHandler: response code=%d header='%v' string_body='%+v' time=%v",
    //    lwr.code, lwr.headers, string(lwr.data), endTime.Sub(startTime))

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