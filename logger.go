package log4nova

import (
    "net/http"
    "github.com/Polarishq/bouncer/models"
    "time"
    "context"
    "github.com/sirupsen/logrus"
    "github.com/Polarishq/bouncer/client/events"
    rtclient "github.com/go-openapi/runtime/client"
    "github.com/Polarishq/bouncer/client"
    "github.com/go-openapi/strfmt"
    "errors"
    "fmt"
    "encoding/json"
    "sync"
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
}

//NewNovaLogger creates a new instance of the NovaLogger
func NewNovaLogger(customClient events.ClientInterface, customLogger *logrus.Logger, clientID, clientSecret, host string) *NovaLogger {
    // Configure default values
    var novaHost string
    var logger *logrus.Logger
    if clientID == "" || clientSecret == "" {
        panic(errors.New("Nova client ID or client secret not set properly"))
    }
    if host == "" {
        novaHost = client.DefaultHost
    } else {
        novaHost = host
    }

    //Should be used mainly just for test
    if customLogger != nil {
        logger = customLogger
    } else {
        logger = logrus.New()
    }

    // Set up events client
    var eventsClient events.ClientInterface
    if customClient != nil {
        eventsClient = customClient
    } else {
        // Create log-input client
        transCfg := client.DefaultTransportConfig()
        auth := rtclient.BasicAuth(clientID, clientSecret)
        httpCl := &http.Client{}
        transportWithClient := rtclient.NewWithClient(novaHost, client.DefaultBasePath, transCfg.Schemes, httpCl)
        transportWithClient.Transport = httpCl.Transport
        transportWithClient.DefaultAuthentication = auth
        eventsClient = client.New(transportWithClient, strfmt.Default).Events
    }


    // Return new logger
    return &NovaLogger{
        client:     eventsClient,
        SendInterval: 2000,
        clientID: clientID,
        clientSecret: clientSecret,
        logrusLogger: logger,
        inStream: make(chan string),
    }
}

//Start kicks off the logger to feed data off to splunk as available
func (nl *NovaLogger) Start() {
    nl.logrusLogger.Out = nl
    nl.logrusLogger.Formatter = &logrus.JSONFormatter{}
    // Begin the formatting process
    go nl.flushFromOutputChannel(nl.formatLogs(nl.inStream))
    return
}

//Write sends all writes to the input channel
func (nl *NovaLogger) Write(p []byte) (n int, err error) {
    go nl.writeLogsToChannel(string(p))
    fmt.Println(string(p))
    return len(p), nil
}

//Send logs to the out channel
func (nl *NovaLogger) writeLogsToChannel(log string) {
    nl.inStream <- log
    return
}

//Format logs from strings to the out channel event format
func (nl *NovaLogger) formatLogs(in <-chan string) (*[]*models.Event, sync.Mutex) {
    // Create the output channel and the lock
    out := make([]*models.Event,0)
    lock := sync.Mutex{}

    //Spawn new thread to wait for data on the input channel
    go func() {
        for {
            select {
            case log := <-in:
                //Marshal the data out to iterate over and set on the event
                logMap := make(map[string]interface{})
                err := json.Unmarshal([]byte(log), &logMap)
                if err != nil {
                    panic(err)
                }

                event := models.Event{
                    Event: map[string]*string{
                        "raw": &log,
                    },
                }
                for k, v := range logMap {
                    stringVal := fmt.Sprintf("%s", v)
                    event.Event[k] = &stringVal
                }

                fmt.Println("Pushed formatted log to channel")

                //Block and insert a new event
                lock.Lock()
                out = append(out, &event)
                lock.Unlock()
            }
        }
    }()
    return &out, lock
}

//Flush logs to the log-input endpoint
func (nl *NovaLogger) flushFromOutputChannel(out *[]*models.Event, lock sync.Mutex) {
    for {
        time.Sleep(time.Duration(nl.SendInterval) * time.Millisecond)
        fmt.Println("Waking from sleep to flush logs")

        //If we have logs, spawn a new thread to flush logs out
        if len(*out) > 0 {
            go func() {
                // Make a copy of the array and block while doing so
                lock.Lock()
                tmp := make([]*models.Event, len(*out))
                copy(tmp, *out)
                *out = make([]*models.Event, 0)
                lock.Unlock()

                //Iterate over to push events into log-input
                for _, event := range tmp {
                    ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
                    defer cancel()
                    params := &events.EventsParams{
                        Events:  models.Events{event},
                        Context: ctx,
                    }
                    fmt.Println("Sending events")
                    auth := rtclient.BasicAuth(nl.clientID, nl.clientSecret)
                    _, err := nl.client.Events(params, auth)

                    //Should this panic? how to surface this
                    if err != nil {
                        fmt.Printf("Error sending to log-store: %v\n", err)
                    }
                }
            }()
        }
    }
    return
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
    nl.logrusLogger.Level = logrus.DebugLevel
}

// SetDebug sets the log level to debug
func (nl *NovaLogger) SetInfo() {
    nl.logrusLogger.Level = logrus.InfoLevel
}

// SetWarn sets the log level to warn
func (nl *NovaLogger) SetWarn() {
    nl.logrusLogger.Level = logrus.WarnLevel
}

// SetError sets the log level to error
func (nl *NovaLogger) SetError() {
    nl.logrusLogger.Level = logrus.ErrorLevel
}
