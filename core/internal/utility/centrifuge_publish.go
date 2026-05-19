package utility

import (
	"context"

	"github.com/centrifugal/centrifuge-go"
	"ntels.com/pharos/core/pkg/common"
)

func CentrifugePublish(endpoint, channel, data string) error {
	client := centrifuge.NewJsonClient(
		endpoint,
		centrifuge.Config{
			TLSConfig: common.BuildDefaultClientTLS(),
		})

	if err := client.Connect(); err != nil {
		return err
	}
	defer client.Close()

	subscription, err := client.NewSubscription(channel, centrifuge.SubscriptionConfig{
		Recoverable: true,
		JoinLeave:   true,
	})
	if err != nil {
		return err
	}

	if err := subscription.Subscribe(); err != nil {
		return err
	}

	if _, err := subscription.Publish(context.Background(), []byte(data)); err != nil {
		return err
	}

	return nil
}
