package format

import (
	"bufio"

	"github.com/42wim/syslogparser"
	"github.com/42wim/syslogparser/rfc3164"
)

type RFC3164 struct{}

func (f *RFC3164) GetParser(line []byte) syslogparser.LogParser {
	return rfc3164.NewParser(line)
}

func (f *RFC3164) GetSplitFunc() bufio.SplitFunc {
	return nil
}
