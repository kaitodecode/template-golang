package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Init inisialisasi global logger
// mode = "dev" -> pretty console, mode = "prod" -> JSON structured log
func Init(mode string) {
	log = logrus.New()
	log.Out = os.Stdout

	// Set level default
	log.SetLevel(logrus.DebugLevel)

	if mode == "prod" {
		// JSON log untuk production (mudah di-parse monitoring)
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05Z07:00",
			PrettyPrint:     false,
		})
	} else {
		// Text log untuk development (warna-warni, lebih terbaca)
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     true,
		})
	}
}

// L untuk ambil instance global logger
func L() *logrus.Logger {
	if log == nil {
		Init("dev") // default fallback
	}
	return log
}

// Fields untuk logging dengan field tambahan (structured log)
func Fields(fields logrus.Fields) *logrus.Entry {
	return L().WithFields(fields)
}
