package main

import (
	"encoding/json"
	"errors"
	"fmt"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"log"
	"sync"
)

func main() {
	s := newStorage()
	n := maelstrom.NewNode()

	n.Handle("txn", func(msg maelstrom.Message) error {
		// Unmarshal the message body as a loosely-typed map.
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return fmt.Errorf("unmarshaling message body: %w", err)
		}

		txnBytes, err := json.Marshal(body["txn"])
		if err != nil {
			return fmt.Errorf("marshaling bory txn: %w", err)
		}

		var txn [][]any
		if err := json.Unmarshal(txnBytes, &txn); err != nil {
			return fmt.Errorf("unmarshaling message body: %w", err)
		}

		txnResp := make([][]any, 0, len(txn))

		for i, t := range txn {
			if len(t) != 3 {
				return fmt.Errorf("malformed transaction %d: must have 3 parameters, but has %d", i, len(t))
			}

			operation, ok := t[0].(string)
			if !ok {
				return fmt.Errorf("parsing operation value %q in transaction %d", t[0], i)
			}

			switch operation {
			case "r":
				key, ok := t[1].(float64)
				if !ok {
					return fmt.Errorf("parsing key value %q in transaction %d", t[1], i)
				}

				value, err := s.read(int(key))
				switch {
				case err == nil:
					txnResp = append(txnResp, []any{"r", key, value})
				case errors.Is(err, errKeyNotFound):
					txnResp = append(txnResp, []any{"r", key, nil})
				default:
					return fmt.Errorf("reading key %d in transaction %d: %w", int(key), i, err)
				}

			case "w":
				key, ok := t[1].(float64)
				if !ok {
					return fmt.Errorf("parsing key value %q in transaction %d", t[1], i)
				}

				value, ok := t[2].(float64)
				if !ok {
					return fmt.Errorf("parsing value value %q in transaction %d", t[2], i)
				}

				respValue, err := s.write(int(key), int(value))
				if err != nil {
					return fmt.Errorf("writing key %d in transaction %d: %w", int(key), i, err)
				}

				txnResp = append(txnResp, []any{"w", key, respValue})

			default:
				return fmt.Errorf("unknown operation %q in transaction %d", operation, i)
			}
		}

		resp := map[string]any{
			"type":        "txn_ok",
			"in_reply_to": body["msg_id"],
			"txn":         txnResp,
		}

		return n.Reply(msg, resp)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

var errKeyNotFound = errors.New("key not found")

// storage is a structure capable of storing values in key-value format.
type storage struct {
	data map[int]int
	m    *sync.Mutex
}

func newStorage() storage {
	return storage{
		data: make(map[int]int),
		m:    &sync.Mutex{},
	}
}

// read reads the value of a key from storage.
//
// Returns error if the key does not exist.
func (s storage) read(key int) (int, error) {
	v, found := s.data[key]
	if !found {
		return 0, errKeyNotFound
	}

	return v, nil
}

// write writes a value to a key in storage.
func (s storage) write(key int, value int) (int, error) {
	s.m.Lock()
	defer s.m.Unlock()

	s.data[key] = value

	v, found := s.data[key]
	if !found {
		return 0, errKeyNotFound
	}

	return v, nil
}
