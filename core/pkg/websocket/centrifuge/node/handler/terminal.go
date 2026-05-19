package handler

import (
	"encoding/base64"
	"errors"
	"io"
	"log/slog"

	"github.com/centrifugal/centrifuge"
	"github.com/thedevsaddam/gojsonq/v2"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/internal"
)

type Terminal struct {
	Client *centrifuge.Client

	Pipes *internal.Map[*Pipe]
}

func (terminal *Terminal) OnSubscribe(e centrifuge.SubscribeEvent, cb centrifuge.SubscribeCallback) {
	address := gojsonq.New().FromString(string(e.Data)).Find("address")
	username := gojsonq.New().FromString(string(e.Data)).Find("username")
	password := gojsonq.New().FromString(string(e.Data)).Find("password")
	if address == nil || username == nil || password == nil {
		cb(centrifuge.SubscribeReply{}, external.ErrorInvalidDataFormat)
		return
	}

	pipe := Pipe{}
	if err := pipe.connect(address.(string), username.(string), password.(string)); err != nil {
		cb(centrifuge.SubscribeReply{}, err)
		return
	}

	terminal.Pipes.Set(e.Channel, &pipe)

	go func() {
		for {
			select {
			case <-terminal.Client.Context().Done():
				return
			default:
				if result, err := pipe.read(); errors.Is(err, io.EOF) {
					return
				} else if err != nil {
					slog.Error(err.Error())
				} else if err := terminal.Client.Send([]byte(`{"message":"` + base64.StdEncoding.EncodeToString([]byte(result)) + `"}`)); errors.Is(err, io.EOF) {
					return
				} else if err != nil {
					slog.Error(err.Error())
				}
			}
		}
	}()

	cb(centrifuge.SubscribeReply{}, nil)
}

func (terminal *Terminal) OnUnsubscribe(e centrifuge.UnsubscribeEvent) {
	terminal.Pipes.Remove(e.Channel, func(pipe *Pipe) { pipe.close() })
}

func (terminal *Terminal) OnPublish(e centrifuge.PublishEvent, cb centrifuge.PublishCallback) {
	command := gojsonq.New().FromString(string(e.Data)).Find("command")
	if command == nil {
		cb(centrifuge.PublishReply{}, external.ErrorInvalidDataFormat)
		return
	}

	if pipe := terminal.Pipes.Get(e.Channel); pipe == nil {
		cb(centrifuge.PublishReply{}, external.ErrorInvalidType)
		return
	} else if _, err := pipe.write(command.(string)); err != nil {
		cb(centrifuge.PublishReply{}, err)
		return
	}
}
