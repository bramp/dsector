package ufwb

import (
	"bytes"
	log "github.com/Sirupsen/logrus"
	"time"
)

func init() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&SimpleFormatter{})
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
