package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

func StdErr(msg string) {
	if os.Stderr != nil {
		lines := strings.Split(msg, "\n")
		if len(lines) > 1 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
		for i, line := range lines {
			now := time.Now().Format("2006-01-02 15:04:05 -0700 MST")
			lines[i] = now + "  " + line
		}
		os.Stderr.Write([]byte(strings.Join(lines, "\n") + "\n"))
	}
}

func StdErrF(format string, a ...any) {
	StdErr(fmt.Sprintf(format, a...))
}

func StdErrCtx(ctx map[string]interface{}) {
	b, err := json.Marshal(ctx)
	if err != nil {
		return
	}

	StdErr(string(b))
}
