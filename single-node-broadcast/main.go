package main

import (
	"encoding/json"
	"fmt"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"log"
)

func main() {
	messages := make([]int, 0)

	n := maelstrom.NewNode()

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		// Unmarshal the message body as a loosely-typed map.
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return fmt.Errorf("unmarshaling message body: %w", err)
		}

		message, ok := body["message"].(float64)
		if !ok {
			return fmt.Errorf("parsing message %v: is not a float64", body["message"])
		}

		// Stores message ID.
		messages = append(messages, int(message))

		resp := map[string]any{
			"type":        "broadcast_ok",
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

		resp := map[string]any{
			"type":        "read_ok",
			"in_reply_to": body["msg_id"],
			"messages":    messages,
		}

		return n.Reply(msg, resp)
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		// Unmarshal the message body as a loosely-typed map.
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return fmt.Errorf("unmarshaling message body: %w", err)
		}

		resp := map[string]any{
			"type":        "topology_ok",
			"in_reply_to": body["msg_id"],
		}

		return n.Reply(msg, resp)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
