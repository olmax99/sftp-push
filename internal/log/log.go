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
package internal

import (
	log1 "log"
	"log/syslog"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	lSyslog "github.com/sirupsen/logrus/hooks/syslog"

	"github.com/spf13/viper"
)

// DefaultLog returns the global logrus struct
type DefaultLog struct {
	Logger *logrus.Logger
}

// NewLogger returns a configured logrus instance
func NewLogger(cfg *viper.Viper) (*logrus.Logger, string) {
	l, msg := newLogrusLogger(cfg)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "create NewLogger.")
	// }
	d := DefaultLog{Logger: l}
	return d.Logger, msg
}

// Create a new logrus instance with custom setup
func newLogrusLogger(cfg *viper.Viper) (*logrus.Logger, string) {

	l := logrus.New()
	msg := "Stderr"

	// log destination Sdterr | Syslog | custom
	if form := cfg.GetString("defaults.log.format"); form == "json" {
		l.Formatter = new(logrus.JSONFormatter)
	}

	switch log_loc := cfg.GetString("defaults.log.location"); {
	case log_loc == "syslog":
		msg = log_loc
		hook, err := lSyslog.NewSyslogHook("udp", "localhost:514", syslog.LOG_INFO, "")
		if err == nil {
			l.Hooks.Add(hook)
		} else {
			log1.Printf("WARNING[-] newLogrusLogger: %s", err)
		}
	default:
		_, err := os.Stat(filepath.Dir(log_loc))
		if err != nil {
			log1.Printf("WARNING[-] newLogrusLogger: %s", err)
			l.Out = os.Stderr
		}
		file, err := os.OpenFile(log_loc, os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			log1.Printf("WARNING[-] newLogrusLogger: %s", err)
			l.Out = file
			msg = log_loc
		} else {
			l.Out = os.Stderr
		}
	}

	// log level
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

	// TODO add high level service paramaters, e.g. env: stage | prod
	// l.AddHook(NewExtraFieldHook(service, config.Env))

	return l, msg
}

// type ExtraFieldHook struct {
// 	service string
// 	env     string
// 	pid     int
// }

// func NewExtraFieldHook(service string, env string) *ExtraFieldHook {
// 	return &ExtraFieldHook{
// 		service: service,
// 		env:     env,
// 		pid:     os.Getpid(),
// 	}
// }

// func (h *ExtraFieldHook) Levels() []logrus.Level {
// 	return logrus.AllLevels
// }

// func (h *ExtraFieldHook) Fire(entry *logrus.Entry) error {
// 	entry.Data["service"] = h.service
// 	entry.Data["env"] = h.env
// 	entry.Data["pid"] = h.pid
// 	return nil
// }
