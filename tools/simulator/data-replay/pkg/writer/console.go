package writer

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

type ConsoleWriter struct{}

func (w *ConsoleWriter) Type() string {
	return "console"
}

func (w *ConsoleWriter) Write(ctx map[string]interface{}) error {
	b, err := json.Marshal(ctx)
	if err != nil {
		return errors.Wrap(err, "json marshal error")
	}

	fmt.Println(string(b))
	return nil
}

func NewConsoleWriter() (*ConsoleWriter, error) {
	return &ConsoleWriter{}, nil
}
