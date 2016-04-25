package format

import (
	"bufio"

	"github.com/42wim/syslogparser"
	"github.com/42wim/syslogparser/rfc3164cisco"
)

type RFC3164cisco struct{}

func (f *RFC3164cisco) GetParser(line []byte) syslogparser.LogParser {
	return rfc3164cisco.NewParser(line)
}

func (f *RFC3164cisco) GetSplitFunc() bufio.SplitFunc {
	return nil
}
