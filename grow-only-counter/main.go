package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()
	kv := maelstrom.NewSeqKV(n)

	n.Handle("add", func(msg maelstrom.Message) error {
		// Unmarshal the message body as a loosely-typed map.
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return fmt.Errorf("unmarshaling message body: %w", err)
		}

		delta, ok := body["delta"].(float64)
		if !ok {
			return fmt.Errorf("parsing delta %v in 'add' handler: is not a float64", body["delta"])
		}

		// Trying to update the counter consistently, avoiding race conditions.
		//
		// Writing is only done if the counter value at the time of writing is equal to
		// the read value.
		//
		// It is expected here that there will always be a moment when the value can be
		// written.

		for {
			counter, err := kv.ReadInt(context.Background(), "counter")
			if err != nil {
				var mErr *maelstrom.RPCError
				if !errors.As(err, &mErr) {
					if mErr.Code != maelstrom.KeyDoesNotExist {
						return fmt.Errorf("reading counter from kv in 'add' handler: %w", err)
					}
				}
			}

			if err = kv.CompareAndSwap(context.Background(), "counter", counter, counter+int(delta), true); err == nil {
				break
			}
		}

		resp := map[string]any{
			"type":        "add_ok",
			"in_reply_to": body["msg_id"],
		}

		return n.Reply(msg, resp)
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		// Unmarshal the message body as a loosely-typed map.
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return fmt.Errorf("unmarshaling message body: %w", err)
		}

		delta, err := kv.ReadInt(context.Background(), "counter")
		if err != nil {
			var mErr *maelstrom.RPCError
			if !errors.As(err, &mErr) {
				if mErr.Code != maelstrom.KeyDoesNotExist {
					return fmt.Errorf("reading counter from kv in 'read' handler: %w", err)
				}
			}
		}

		resp := map[string]any{
			"type":        "read_ok",
			"in_reply_to": body["msg_id"],
			"value":       delta,
		}

		return n.Reply(msg, resp)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
