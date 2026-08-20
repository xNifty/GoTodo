package live

import (
	"context"
	"strings"

	"github.com/redis/go-redis/v9"
)

const redisChannelPrefix = "gotodo:sse:"

type redisBridge struct {
	publisher func(channel string, payload []byte)
}

func newRedisBridge(client *redis.Client, onMessage func(channel string, payload []byte)) *redisBridge {
	if client == nil {
		return nil
	}
	ctx := context.Background()
	go subscribeLoop(ctx, client, redisChannelPrefix+"*", func(channel string, payload []byte) {
		key := strings.TrimPrefix(channel, redisChannelPrefix)
		onMessage(key, payload)
	})
	return &redisBridge{
		publisher: func(channel string, payload []byte) {
			_ = client.Publish(ctx, redisChannelPrefix+channel, payload).Err()
		},
	}
}

func (b *redisBridge) publish(channel string, payload []byte) {
	if b != nil && b.publisher != nil {
		b.publisher(channel, payload)
	}
}

func subscribeLoop(ctx context.Context, client *redis.Client, pattern string, handler func(channel string, payload []byte)) {
	pubsub := client.PSubscribe(ctx, pattern)
	ch := pubsub.Channel()
	for msg := range ch {
		if msg == nil {
			continue
		}
		handler(msg.Channel, []byte(msg.Payload))
	}
}
