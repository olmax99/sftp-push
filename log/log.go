// ------------------------ Logger Setup------------------------
// Default:
// - use output file with location ~/.sftppush/sftppush.log
// - log level: debug
// - format text
//
// Define through config:
// - log level
// - custom log location
// - format json
//--------------------------------------------------------------
package log

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Logger defines a set of methods for writing application logs. Derived from and
// inspired by logrus.Entry.
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Debugln(args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Errorln(args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Fatalln(args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Infoln(args ...interface{})
	Panic(args ...interface{})
	Panicf(format string, args ...interface{})
	Panicln(args ...interface{})
	Print(args ...interface{})
	Printf(format string, args ...interface{})
	Println(args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Warning(args ...interface{})
	Warningf(format string, args ...interface{})
	Warningln(args ...interface{})
	Warnln(args ...interface{})
}

// DefaultLog returns the global logrus struct
type DefaultLog struct {
	Logger *logrus.Logger
}

// func init() {
// 	DefaultLog.Logger = newLogrusLogger(config.Config())
// }

// NewLogger returns a configured logrus instance
func NewLogger(cfg *viper.Viper) (*logrus.Logger, string) {
	l, msg := newLogrusLogger(cfg)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "create NewLogger.")
	// }
	d := DefaultLog{Logger: l}
	return d.Logger, msg
}

func newLogrusLogger(cfg *viper.Viper) (*logrus.Logger, string) {

	l := logrus.New()
	msg := "Stderr"

	if form := cfg.GetString("defaults.log.format"); form == "json" {
		l.Formatter = new(logrus.JSONFormatter)
	}

	log_loc := cfg.GetString("defaults.log.location")
	_, err := os.Stat(filepath.Dir(log_loc))
	if err != nil {
		l.Out = os.Stderr
	}
	file, err := os.OpenFile(log_loc, os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		l.Out = file
		msg = log_loc
	} else {
		l.Out = os.Stderr
	}

	switch cfg.GetString("defaults.log.level") {
	case "debug":
		l.Level = logrus.DebugLevel
		l.SetReportCaller(true) // Produces 20-40% overhead
	case "warning":
		l.Level = logrus.WarnLevel
	case "info":
		l.Level = logrus.InfoLevel
	default:
		l.Level = logrus.DebugLevel
	}

	return l, msg
}

// // Fields is a map string interface to define fields in the structured log
// type Fields map[string]interface{}

// // With allow us to define fields in out structured logs
// func (f Fields) With(k string, v interface{}) Fields {
// 	f[k] = v
// 	return f
// }

// // WithFields allow us to define fields in out structured logs
// func (f Fields) WithFields(f2 Fields) Fields {
// 	for k, v := range f2 {
// 		f[k] = v
// 	}
// 	return f
// }

// // WithFields allow us to define fields in out structured logs
// func (l *DefaultLog) WithFields(fields Fields) *logrus.Entry {
// 	return l.Logger.WithFields(logrus.Fields(fields))
// }

// // Debug package-level convenience method.
// func (l *DefaultLog) Debug(args ...interface{}) {
// 	l.Debug(args...)
// }

// // Debugf package-level convenience method.
// func (l *DefaultLog) Debugf(format string, args ...interface{}) {
// 	l.Debugf(format, args...)
// }

// // Debugln package-level convenience method.
// func Debugln(args ...interface{}) {
// 	Logger.Debugln(args...)
// }

// // Error package-level convenience method.
// func Error(args ...interface{}) {
// 	Logger.Error(args...)
// }

// // Errorf package-level convenience method.
// func Errorf(format string, args ...interface{}) {
// 	Logger.Errorf(format, args...)
// }

// // Errorln package-level convenience method.
// func Errorln(args ...interface{}) {
// 	Logger.Errorln(args...)
// }

// // Fatal package-level convenience method.
// func Fatal(args ...interface{}) {
// 	Logger.Fatal(args...)
// }

// // Fatalf package-level convenience method.
// func Fatalf(format string, args ...interface{}) {
// 	Logger.Fatalf(format, args...)
// }

// // Fatalln package-level convenience method.
// func Fatalln(args ...interface{}) {
// 	Logger.Fatalln(args...)
// }

// // Info package-level convenience method.
// func Info(args ...interface{}) {
// 	Logger.Info(args...)
// }

// // Infof package-level convenience method.
// func Infof(format string, args ...interface{}) {
// 	Logger.Infof(format, args...)
// }

// // Infoln package-level convenience method.
// func Infoln(args ...interface{}) {
// 	Logger.Infoln(args...)
// }

// // Panic package-level convenience method.
// func Panic(args ...interface{}) {
// 	Logger.Panic(args...)
// }

// // Panicf package-level convenience method.
// func Panicf(format string, args ...interface{}) {
// 	Logger.Panicf(format, args...)
// }

// // Panicln package-level convenience method.
// func Panicln(args ...interface{}) {
// 	Logger.Panicln(args...)
// }

// // Print package-level convenience method.
// func Print(args ...interface{}) {
// 	Logger.Print(args...)
// }

// // Printf package-level convenience method.
// func Printf(format string, args ...interface{}) {
// 	Logger.Printf(format, args...)
// }

// // Println package-level convenience method.
// func Println(args ...interface{}) {
// 	Logger.Println(args...)
// }

// // Warn package-level convenience method.
// func Warn(args ...interface{}) {
// 	Logger.Warn(args...)
// }

// // Warnf package-level convenience method.
// func Warnf(format string, args ...interface{}) {
// 	Logger.Warnf(format, args...)
// }

// // Warning package-level convenience method.
// func Warning(args ...interface{}) {
// 	Logger.Warning(args...)
// }

// // Warningf package-level convenience method.
// func Warningf(format string, args ...interface{}) {
// 	Logger.Warningf(format, args...)
// }

// // Warningln package-level convenience method.
// func Warningln(args ...interface{}) {
// 	Logger.Warningln(args...)
// }

// // Warnln package-level convenience method.
// func Warnln(args ...interface{}) {
// 	Logger.Warnln(args...)
// }
