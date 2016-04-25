package format

import (
	"bufio"

	"github.com/42wim/syslogparser"
)

type Format interface {
	GetParser([]byte) syslogparser.LogParser
	GetSplitFunc() bufio.SplitFunc
}
