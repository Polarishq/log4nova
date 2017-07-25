package log4nova

import (
    "net/http"
    "github.com/Polarishq/bouncer/models"
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
    "encoding/json"
    "reflect"
    "strconv"
)

// Stats data structure
type NovaLogger struct {
    handler             http.Handler
    captureResponseBody bool
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
    Api             string  `json:"api,omitempty"`
    StatusCode      string  `json:"status_code,omitempty"`
    RequestURL      string  `json:"RequestURL,omitempty"`
    RequestMethod   string  `json:"RequestMethod,omitempty"`
    UserAgent       string  `json:"UserAgent,omitempty"`
    Time            string  `json:"time,omitempty"`
    ResponseTime    string  `json:"response_time,omitempty"`
    Entity          string  `json:"entity,omitempty"`
    Source          string  `json:"source,omitempty"`
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

func (nl *NovaLogger) formatLogs(in <-chan string) <-chan *models.Event {
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
                    novaLog := novaLogFormat{}
                    err := json.Unmarshal([]byte(log), &novaLog)
                    if err != nil {
                        panic(err)
                    }
                    event := models.Event{
                        Event:  map[string]*string{
                            //"raw": &log,
                        },
                    }

                    // Append nova log to map
                    nlValue := reflect.ValueOf(&novaLog).Elem()
                    typeOfnl := nlValue.Type()
                    for i := 0; i < nlValue.NumField(); i++ {
                        field := nlValue.Field(i)
                        // Ignore fields that don't have the same type as a string
                        if field.Type() != reflect.TypeOf("") {
                            continue
                        }
                        nlValue := field.Interface().(string)
                        nlType := typeOfnl.Field(i).Name
                        event.Event[nlType] = &nlValue
                    }

                    fmt.Println("Pushed formatted log to channel")
                    out <- &event
                }
            }
        }
        fmt.Println("Closing out channel")
        close(out)
    }()

    return out
}

func (nl *NovaLogger) flushFromOutputChannel(out <-chan *models.Event) error {
    time.Sleep(time.Duration(nl.sendInterval) * time.Millisecond)
    fmt.Println("Flushing from output channel")
    for {
        select {
        case event, ok := <- out:
            if !ok {
                //for {
                //    if len(out) > 0 {
                //        time.Sleep(time.Duration(nl.sendInterval)/2 * time.Millisecond)
                //    } else {
                //        break
                //    }
                //}
                break
            } else {
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
            }
        }
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
    //logger.WithFields(logger.Fields{
    //    "request": stringify(r),
    //}).Infof("Logging Request")

    lwr := loggingResponseWriter{w: w, captureBody: nl.captureResponseBody}
    nl.handler.ServeHTTP(&lwr, r)
    endTime := time.Now()
    uuid_evt := uuid.NewV1()
    fmt.Println(uuid_evt)
    logger.WithFields(logger.Fields{
        "api": r.URL.Path,
        "status_code": strconv.Itoa(lwr.code),
        "RequestURL" : r.RequestURI,
        "RequestMethod": r.Method,
        "UserAgent": r.UserAgent(),
        "time": startTime,
        "entity": nl.host,
        "source": "rest_access",
        "log_id": uuid_evt,
        //"headers": lwr.headers,
        //"body": string(lwr.data),
        "response_time": endTime.Sub(startTime).String(),
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