package handler

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type Pipe struct {
	client  *ssh.Client
	session *ssh.Session

	stdinWriter  io.Writer
	stdoutReader io.Reader
}

func (pipe *Pipe) connect(address, user, password string) error {
	// TODO 차후 SSH 추가할때 config에서 읽어오도록 수정
	// Determine known_hosts path: env override or default to ~/.ssh/known_hosts
	knownHostsPath := os.Getenv("TARZAN_SSH_KNOWN_HOSTS")
	if knownHostsPath == "" {
		home, _ := os.UserHomeDir()
		knownHostsPath = filepath.Join(home, ".ssh", "known_hosts")
	}

	hostKeyCallback, err := knownhosts.New(knownHostsPath)
	if err != nil {
		return err
	}

	clientConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: hostKeyCallback,
	}

	if client, err := ssh.Dial("tcp", address, clientConfig); err != nil {
		return err
	} else {
		pipe.client = client
	}

	if session, err := pipe.client.NewSession(); err != nil {
		return err
	} else {
		pipe.session = session
	}

	if stdinWriter, err := pipe.session.StdinPipe(); err != nil {
		return err
	} else {
		pipe.stdinWriter = stdinWriter
	}

	if stdoutReader, err := pipe.session.StdoutPipe(); err != nil {
		return err
	} else {
		pipe.stdoutReader = stdoutReader
	}

	if err := pipe.session.RequestPty("xterm", 80, 40, ssh.TerminalModes{}); err != nil {
		return err
	}

	if err := pipe.session.Shell(); err != nil {
		return err
	}

	return nil
}

func (pipe *Pipe) close() {
	if pipe.session != nil {
		if err := pipe.session.Close(); err != nil {
			slog.Error(err.Error())
		}

		pipe.session = nil
	}

	if pipe.client != nil {
		if err := pipe.client.Close(); err != nil {
			slog.Error(err.Error())
		}

		pipe.client = nil
	}
}

func (pipe *Pipe) read() (string, error) {
	buffer := make([]byte, 1024)

	if n, err := pipe.stdoutReader.Read(buffer); err != nil {
		return "", err
	} else {
		return string(buffer[:n]), nil
	}
}

func (pipe *Pipe) write(command string) (int, error) {
	command += "\n"

	return pipe.stdinWriter.Write([]byte(command))
}
