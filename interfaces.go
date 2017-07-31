package log4nova

import "github.com/sirupsen/logrus"

type INovaLogger interface {
    Start()
    Write(p []byte) (n int, err error)
    WithField(key string, value interface{}) *logrus.Entry
    WithFields(fields Fields) *logrus.Entry
    WithError(err error) *logrus.Entry
    Debugf(format string, v ...interface{})
    Infof(format string, v ...interface{})
    Warningf(format string, v ...interface{})
    Errorf(format string, v ...interface{})
    Error(v ...interface{})
    Warning(v ...interface{})
    Info(v ...interface{})
    Debug(v ...interface{})
    SetDebug()
    SetInfo()
    SetWarn()
    SetError()
}