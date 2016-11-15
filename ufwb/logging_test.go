package ufwb

// logging_test doesn't contain any tests, it instead is used to change the logging behaviour inside tests.

import (
	"bytes"
	log "github.com/Sirupsen/logrus"
	"os"
	"time"
)

func init() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&SimpleFormatter{})

	if level := os.Getenv("LOGLEVEL"); level != "" {
		def, err := log.ParseLevel(level)
		if err == nil {
			log.SetLevel(def)
		} else {
			log.Infof("Failed to set log level: %s", err)
		}
	}
}

type SimpleFormatter struct{}

func (f *SimpleFormatter) Format(entry *log.Entry) ([]byte, error) {
	var b *bytes.Buffer

	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	b.WriteString(entry.Time.Format(time.RFC3339))
	b.WriteString(" [")
	b.WriteString(entry.Level.String())
	b.WriteString("] ")
	b.WriteString(entry.Message)
	b.WriteByte('\n')

	return b.Bytes(), nil
}
