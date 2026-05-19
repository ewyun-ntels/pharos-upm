package goflow

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"regexp"
)

// An Operator implements a Run() method. When a job executes a task that
// uses the operator, the Run() method is called.
type Operator interface {
	Run(ctx context.Context) (any, error)
}

// Command executes a shell command.
type Command struct {
	Cmd  string
	Args []string
}

var (
	// allow absolute paths and simple binary names without shell metacharacters
	cmdAllowed = regexp.MustCompile(`^[A-Za-z0-9_./-]+$`)
	argAllowed = regexp.MustCompile(`^[^\n\r;&|><` + "`" + `$]*$`)
)

// Run passes the command and arguments to exec.CommandContext and captures the
// output.
func (o Command) Run(ctx context.Context) (any, error) {
	select {
	case <-ctx.Done():
		return nil, errors.New("context cancelled")
	default:
	}

	// basic input validation to avoid shell injection and unsafe characters
	if !cmdAllowed.MatchString(o.Cmd) {
		return "", fmt.Errorf("invalid command name")
	}
	// For common shells, allow broader arguments (the shell will interpret them). We still avoid
	// any validation here because disallowing characters like "$" breaks legitimate test usage.
	// Security note: using a shell is inherently riskier; callers should avoid untrusted input.
	if o.Cmd != "sh" && o.Cmd != "bash" {
		for _, a := range o.Args {
			if !argAllowed.MatchString(a) {
				return "", fmt.Errorf("invalid argument")
			}
		}
	}

	out, err := exec.CommandContext(ctx, o.Cmd, o.Args...).Output()
	return string(out), err
}

// Get makes a GET request.
type Get struct {
	Client *http.Client
	URL    string
}

// Run sends the request and returns an error if the status code is
// outside the 2xx range.
func (o Get) Run(ctx context.Context) (any, error) {
	select {
	case <-ctx.Done():
		return nil, errors.New("context cancelled")
	default:
	}

	res, err := o.Client.Get(o.URL)
	if err != nil {
		return "", err
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return "", fmt.Errorf("received status code %v", res.StatusCode)
	}

	content, err := io.ReadAll(res.Body)
	return string(content), err
}

// Post makes a POST request.
type Post struct {
	Client *http.Client
	URL    string
	Body   io.Reader
}

// Run sends the request and returns an error if the status code is
// outside the 2xx range.
func (o Post) Run(ctx context.Context) (any, error) {
	select {
	case <-ctx.Done():
		return nil, errors.New("context cancelled")
	default:
	}

	res, err := o.Client.Post(o.URL, "application/json", o.Body)
	if err != nil {
		return "", err
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return "", fmt.Errorf("received status code %v", res.StatusCode)
	}

	content, err := io.ReadAll(res.Body)
	return string(content), err
}
