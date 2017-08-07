package log4nova

import (
    "net/http"
    "time"
    "context"
    "github.com/sirupsen/logrus"
    "github.com/Polarishq/logface-sdk-go/client/events"
    rtclient "github.com/go-openapi/runtime/client"
    "github.com/go-openapi/strfmt"
    "github.com/cenkalti/backoff"
    "errors"
    "fmt"
    "encoding/json"
    "sync"
    "github.com/Polarishq/logface-sdk-go/client"
    "github.com/Polarishq/logface-sdk-go/models"
)

const (
    MaxBufferSize = 400
    DefaultFlushInterval = 2000
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
    isRunning           bool
    isStopped           bool
    SendToStdOut        bool
}

//NewNovaLoggerWithHost creates a Nova Logging instance with the default host
func NewNovaLogger(clientID, clientSecret string) *NovaLogger {
    if clientID == "" || clientSecret == "" {
        panic(errors.New("Nova client ID or client secret not set properly"))
    }

    novaHost := client.DefaultHost
    // Return new logger
    return NewNovaLoggerWithHost(clientID, clientSecret, novaHost)
}

func NewNovaLoggerWithHost(clientID, clientSecret, host string) *NovaLogger {
    logger := logrus.New()
    transCfg := client.DefaultTransportConfig()
    auth := rtclient.BasicAuth(clientID, clientSecret)
    httpCl := &http.Client{}
    transportWithClient := rtclient.NewWithClient(host, client.DefaultBasePath, transCfg.Schemes, httpCl)
    transportWithClient.Transport = httpCl.Transport
    transportWithClient.DefaultAuthentication = auth
    eventsClient := client.New(transportWithClient, strfmt.Default).Events
    return NewNovaLoggerWithCustom(eventsClient, logger, clientID, clientSecret, host)

}

//NewNovaLogger creates a new instance of the NovaLogger
func NewNovaLoggerWithCustom(eventsClient events.ClientInterface, logger *logrus.Logger, clientID, clientSecret, host string) *NovaLogger {
    // Return new logger
    return &NovaLogger{
        client:     eventsClient,
        SendInterval: DefaultFlushInterval,
        SendToStdOut: false,
        clientID: clientID,
        clientSecret: clientSecret,
        logrusLogger: logger,
        inStream: make(chan string),
    }
}

//Start kicks off the logger to feed data off to nova as available
func (nl *NovaLogger) Start() {
    if nl.isRunning {
        return
    }

    nl.isRunning = true
    nl.isStopped = false
    nl.logrusLogger.Out = nl
    nl.logrusLogger.Formatter = &logrus.JSONFormatter{}
    // Begin the formatting process
    go nl.flushFromOutputChannel(nl.formatLogs(nl.inStream))
    return
}

//Stop ends the logging
func (nl *NovaLogger) Stop() {
    if !nl.isRunning {
        return
    }
    nl.isStopped = true
    close(nl.inStream)
}

//Write sends all writes to the input channel
func (nl *NovaLogger) Write(p []byte) (n int, err error) {
    go nl.writeLogsToChannel(string(p))
    if nl.SendToStdOut {
        //Tee to stdout
        fmt.Println(string(p))
    }
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
            case log, ok := <-in:
                //Break when channel is closed
                if !ok {
                    break
                }
                //Marshal the data out to iterate over and set on the event
                logMap := make(map[string]interface{})
                err := json.Unmarshal([]byte(log), &logMap)
                if err != nil {
                    panic(err)
                }

                event := models.Event{
                    Event: map[string]string{
                        "raw": log,
                    },
                }
                for k, v := range logMap {
                    stringVal := fmt.Sprintf("%s", v)
                    event.Event[k] = stringVal
                }

                //Block and insert a new event
                lock.Lock()
                if len(out) == MaxBufferSize {
                    //Shift and drop oldest event
                    fmt.Printf("Dropping event: %s\n", out[0].Event["raw"])
                    out = append(out[1:], &event)
                } else {
                    out = append(out, &event)
                }
                lock.Unlock()
            }
        }
    }()
    return &out, lock
}

//Flush logs to the nova endpoint
func (nl *NovaLogger) flushFromOutputChannel(out *[]*models.Event, lock sync.Mutex) {
    for !nl.isStopped {
        time.Sleep(time.Duration(nl.SendInterval) * time.Millisecond)
        retryBackoff := backoff.NewExponentialBackOff()
        auth := rtclient.BasicAuth(nl.clientID, nl.clientSecret)

        //If we have logs, spawn a new thread to flush logs out
        if len(*out) > 0 {
            go func() {
                // Make a copy of the array and block while doing so
                lock.Lock()
                tmp := make([]*models.Event, len(*out))
                copy(tmp, *out)
                *out = make([]*models.Event, 0)
                lock.Unlock()

                //Push events to nova
                ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
                defer cancel()
                params := &events.EventsParams{
                    Events:  models.Events(tmp),
                    Context: ctx,
                }

                //Setup retry func
                operation := func() error {
                    _, err := nl.client.Events(params, auth)
                    return err
                }

                //If retry with backoff fails, then send the error to stdout
                err := backoff.Retry(operation, retryBackoff)
                if err != nil {
                    fmt.Printf("Error sending to log-store: %v\n", err)
                    //Push events back onto the output array
                    lock.Lock()
                    *out = append(*out, tmp...)
                    lock.Unlock()
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
