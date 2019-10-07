package minecraft

import (
	"fmt"
	"regexp"
	"time"
)

var (
	logRe = regexp.MustCompile(`\A\[(\d{2}:\d{2}:\d{2})\] \[(.*)/(.*)\]: (.*)\z`)
)

type LogMessage struct {
	Timestamp time.Time
	Source    string
	Level     string
	Message   string
}

func ParseLogMessage(line string) (*LogMessage, error) {
	sm := logRe.FindStringSubmatch(line)
	if sm == nil {
		return nil, fmt.Errorf("unparsable")
	}

	ts, err := time.Parse("15:04:05", sm[1])
	if err != nil {
		return nil, err
	}

	return &LogMessage{
		Timestamp: ts,
		Source:    sm[2],
		Level:     sm[3],
		Message:   sm[4],
	}, nil
}
