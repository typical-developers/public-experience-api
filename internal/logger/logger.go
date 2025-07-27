package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Formatter struct{}

func (f *Formatter) Format(entry *log.Entry) ([]byte, error) {
	var buff bytes.Buffer

	fmt.Fprintf(
		&buff, "[%s] %s: %s\n",
		entry.Time.Format("2006-01-02T15:04:05.000Z"),
		strings.ToUpper(entry.Level.String()), entry.Message,
	)

	if len(entry.Data) > 0 {
		fields, err := json.MarshalIndent(entry.Data, "", "  ")
		if err != nil {
			return nil, err
		}

		fmt.Fprintf(&buff, "%s\n", fields)
	}

	return buff.Bytes(), nil
}

func init() {
	log.SetFormatter(&Formatter{})
	log.SetOutput(os.Stdout)

	if os.Getenv("LOG_LEVEL") == "debug" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}
