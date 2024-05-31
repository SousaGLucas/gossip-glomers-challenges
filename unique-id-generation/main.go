package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ardanlabs/conf/v3"
	"github.com/go-redis/redis/v8"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	generator, err := newIDGenetator()
	if err != nil {
		log.Fatal(fmt.Errorf("creating id generator: %w", err))
	}

	n := maelstrom.NewNode()

	n.Handle("generate", func(msg maelstrom.Message) error {
		// Unmarshal the message body as a loosely-typed map.
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return fmt.Errorf("unmarshaling message body: %w", err)
		}

		// Generate new unique id.
		id, err := generator.generate(context.Background())
		if err != nil {
			return fmt.Errorf("generating new id: %w", err)
		}

		resp := map[string]any{
			"msg_id":      id,
			"type":        "generate_ok",
			"in_reply_to": body["msg_id"],
			"id":          id,
		}

		return n.Reply(msg, resp)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

type idGenerator struct {
	cli *redis.Client
}

func newIDGenetator() (idGenerator, error) {
	type config struct {
		Host string `conf:"env:REDIS_ADDR,default:localhost"`
		Port string `conf:"env:REDIS_PORT,default:6379"`
	}

	var cfg config

	_, err := conf.Parse("", &cfg)
	if err != nil {
		return idGenerator{}, fmt.Errorf("parsing config: %w", err)
	}

	cli := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
	})

	err = cli.Ping(context.Background()).Err()
	if err != nil {
		return idGenerator{}, fmt.Errorf("pinging redis: %w", err)
	}

	return idGenerator{
		cli: cli,
	}, nil
}

func (g idGenerator) generate(ctx context.Context) (int64, error) {
	return g.cli.Incr(ctx, "unique_identifier").Result()
}
