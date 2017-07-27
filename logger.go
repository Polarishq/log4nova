package log4nova

import (
    "net/http"
    "github.com/Polarishq/bouncer/models"
    "time"
    "context"
    "github.com/Sirupsen/logrus"
    "github.com/Polarishq/bouncer/client/events"
    rtclient "github.com/go-openapi/runtime/client"
    "github.com/Polarishq/bouncer/client"
    "github.com/go-openapi/strfmt"
    "bytes"
    "bufio"
    "errors"
    "fmt"
    "encoding/json"
    "sync"
    "io"
)

// Stats data structure
type NovaLogger struct {
    logrusLogger        *logrus.Logger
    client              events.ClientInterface
    SendInterval        int
    clientID            string
    clientSecret        string
    host                string
    inStream            chan string
    writeLock           sync.Mutex
}

type Nova struct {}


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

//NewNovaLogger creates a new instance of the NovaLogger
func NewNovaLogger(customLogger *logrus.Logger, clientID, clientSecret, host string) *NovaLogger {
    // Configure default values
    var novaHost string
    var logger *logrus.Logger
    if clientID == "" || clientSecret == "" {
        panic(errors.New("NOVA_CLIENT_ID or NOVA_CLIENT_SECRET not set properly on env"))
    }
    if host == "" {

        novaHost = client.DefaultHost
    } else {
        novaHost = host
    }
    if customLogger != nil {
        logger = customLogger
    } else {
        logger = logrus.New()
    }

    // Create log-input client
    transCfg := client.DefaultTransportConfig()
    auth := rtclient.BasicAuth(clientID, clientSecret)
    httpCl := &http.Client{}
    transportWithClient := rtclient.NewWithClient(novaHost, client.DefaultBasePath, transCfg.Schemes, httpCl)
    transportWithClient.Transport = httpCl.Transport
    transportWithClient.DefaultAuthentication = auth

    // Return new logger
    return &NovaLogger{
        client:     client.New(transportWithClient, strfmt.Default).Events,
        SendInterval: 1000,
        clientID: clientID,
        clientSecret: clientSecret,
        logrusLogger: logger,
        inStream: make(chan string),
        writeLock: sync.Mutex{},
    }
}


func (nl *NovaLogger) Start(){
    nl.logrusLogger.Out = nl
    //nl.logrusLogger.Formatter = &logrus.JSONFormatter{}
    // Begin the output process
    out := nl.formatLogs(nl.inStream)
    go nl.flushFromOutputChannel(out)
}

func (nl *NovaLogger) Write(p []byte) (n int, err error) {
    go nl.writeLogsToChannel(string(p))
    fmt.Println(string(p))
    return len(p), nil
}

//Send logs to the out channel
func (nl *NovaLogger) writeLogsToChannel(log string) {
    nl.inStream <- log
    //close(ch)
    return
}

//Format logs from strings to the out channel event format
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
                    //nlValue := reflect.ValueOf(&novaLog).Elem()
                    //typeOfnl := nlValue.Type()
                    //for i := 0; i < nlValue.NumField(); i++ {
                    //    field := nlValue.Field(i)
                    //    //Filter out non string fields
                    //    if field.Type() != reflect.TypeOf("") {
                    //        continue
                    //    }
                    //    nlValue := field.Interface().(string)
                    //    nlType := typeOfnl.Field(i).Name
                    //    event.Event[nlType] = &nlValue
                    //}

                    fmt.Println("Pushed formatted log to channel")
                    logrus.Infoln("Pushed formatted log to channel")
                    out <- &event
                }
            }
        }
        //fmt.Println("Closing out channel")
        //logrus.Infoln("Closing out channel")
        //close(out)
    }()

    return out
}

//Flush logs to the log-input endpoint
func (nl *NovaLogger) flushFromOutputChannel(out <-chan *models.Event) error {
    fmt.Println("Flushing from output channel")
    logrus.Infoln("Flushing from output channel")
    time.Sleep(time.Duration(nl.SendInterval)/2 * time.Millisecond)
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
                logrus.Infoln("Sending events")
                auth := rtclient.BasicAuth(nl.clientID, nl.clientSecret)
                _, err := nl.client.Events(params, auth)

                if err != nil {
                    logrus.Errorf("Error sending to log-store: %v\n", err)
                } else {

                }

            }
        }
    }
}


// Fields allows passing key value pairs to Logrus
type Fields map[string]interface{}

// WithField adds a field to the logrus entry
func (nl *NovaLogger) WithField(key string, value interface{}) *logrus.Entry {
    return nl.logrusLogger.WithField(key, value)
}

// WithFields add fields to the logrus entry
func (nl *NovaLogger) WithFields(fields Fields) *logrus.Entry {
    sendfields := make(logrus.Fields)
    for k, v := range fields {
        sendfields[k] = v
    }
    return nl.logrusLogger.WithFields(sendfields)
}

// WithError adds an error field to the logrus entry
func (nl *NovaLogger) WithError(err error) *logrus.Entry {
    return nl.logrusLogger.WithError(err)
}

// Debugf logs a message at level Debug on the standard logger.
func (nl *NovaLogger) Debugf(format string, v ...interface{}) {
    nl.logrusLogger.Debugf(format, v...)
}

// Infof logs a message at level Info on the standard logger.
func (nl *NovaLogger) Infof(format string, v ...interface{}) {
    nl.logrusLogger.Infof(format, v...)
}

// Warningf logs a message at level Warn on the standard logger.
func (nl *NovaLogger) Warningf(format string, v ...interface{}) {
    nl.logrusLogger.Warningf(format, v...)
}

// Errorf logs a message at level Error on the standard logger.
func (nl *NovaLogger) Errorf(format string, v ...interface{}) {
    nl.logrusLogger.Errorf(format, v...)
}

// Error logs a message at level Error on the standard logger.
func (nl *NovaLogger) Error(v ...interface{}) {
    nl.logrusLogger.Error(v...)
}

// Warning logs a message at level Warn on the standard logger.
func (nl *NovaLogger) Warning(v ...interface{}) {
    nl.logrusLogger.Warning(v...)
}

// Info logs a message at level Info on the standard logger.
func (nl *NovaLogger) Info(v ...interface{}) {
    nl.logrusLogger.Info(v...)
}

// Debug logs a message at level Debug on the standard logger.
func (nl *NovaLogger) Debug(v ...interface{}) {
    nl.logrusLogger.Debug(v...)
}

// SetDebug sets the log level to debug
func (nl *NovaLogger) SetDebug() {
    nl.logrusLogger.SetLevel(logrus.DebugLevel)
}

// SetDebug sets the log level to debug
func (nl *NovaLogger) SetInfo() {
    nl.logrusLogger.SetLevel(logrus.InfoLevel)
}

// SetWarn sets the log level to warn
func (nl *NovaLogger) SetWarn() {
    nl.logrusLogger.SetLevel(logrus.WarnLevel)
}

// SetError sets the log level to error
func (nl *NovaLogger) SetError() {
    nl.logrusLogger.SetLevel(logrus.ErrorLevel)
}
