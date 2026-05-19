package writer

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"text/template"
	"time"

	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type OpenSearchWriter struct {
	Address  string `yaml:"address,omitempty"`
	Index    string `yaml:"index,omitempty"`
	Template string `yaml:"template,omitempty"`
	Pipeline string `yaml:"pipeline,omitempty"`

	client *opensearch.Client
	tmpl   *template.Template
}

func (w *OpenSearchWriter) init(config yaml.Node) error {
	if w == nil {
		return errors.New("nil pointer")
	}

	var err error

	err = config.Decode(w)
	if err != nil {
		return errors.Wrap(err, "decode yaml error")
	}

	if w.Index != "" && w.Template != "" {
		return errors.New("use only one of index and template")
	}

	if w.Template != "" {
		tmpl := template.New("index template")
		tmpl, err := tmpl.Parse(w.Template)
		if err != nil {
			return errors.Wrap(err, "create template error")
		}
		w.tmpl = tmpl
	}

	w.client, err = opensearch.NewClient(opensearch.Config{
		Addresses: []string{w.Address},
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	})
	if err != nil {
		return errors.Wrap(err, "create opensearch client error")
	}
	return nil
}

func (w *OpenSearchWriter) Type() string {
	return "open-search"
}

func (w *OpenSearchWriter) Write(ctx map[string]interface{}) error {
	if w == nil || w.client == nil {
		return errors.New("nil pointer")
	}

	var t time.Time
	if ts, ok := ctx["timestamp"]; !ok {
		t = time.Now()
	} else if t, ok = ts.(time.Time); !ok {
		t = time.Now()
	} else {
		delete(ctx, "timestamp")
	}
	ctx["@timestamp"] = t

	index := w.Index
	if w.tmpl != nil {
		buffer := bytes.NewBuffer(nil)
		err := w.tmpl.Execute(buffer, ctx)
		if err != nil {
			return errors.Wrap(err, "generate index from template error")
		}
		index = buffer.String()
	}

	b, err := json.Marshal(ctx)
	if err != nil {
		return errors.Wrap(err, "marshal json error")
	}

	req := opensearchapi.IndexRequest{
		Index: index,
		Body:  bytes.NewReader(b),
	}
	if w.Pipeline != "" {
		req.Pipeline = w.Pipeline
	}
	response, err := req.Do(context.Background(), w.client)
	if err != nil {
		return errors.Wrap(err, "opensearch insert error")
	}
	_ = response.Body.Close()

	return nil
}

func NewOpenSearchWriter(config yaml.Node) (*OpenSearchWriter, error) {
	writer := &OpenSearchWriter{}

	err := writer.init(config)
	if err != nil {
		return nil, errors.Wrap(err, "opensearch writer init error")
	}

	return writer, nil
}
