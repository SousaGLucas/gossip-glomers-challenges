package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()

	n.Handle("send", func(msg maelstrom.Message) error {
		// Unmarshal the message body as a loosely-typed map.
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return fmt.Errorf("unmarshaling message body: %w", err)
		}

		keyRaw, found := body["key"]
		if !found {
			return errors.New("missing key in message body in 'send' handler")
		}

		key, ok := keyRaw.(string)
		if !ok {
			return fmt.Errorf("parsing key value %q in 'send' handler", keyRaw)
		}

		messageRaw, found := body["msg"]
		if !found {
			return errors.New("missing message in message body in 'send' handler")
		}

		message, ok := messageRaw.(float64)
		if !ok {
			return fmt.Errorf("parsing message value %q in 'send' handler", messageRaw)
		}

		offset := enqueueRecord(key, int(message))

		resp := map[string]any{
			"type":        "send_ok",
			"in_reply_to": body["msg_id"],
			"offset":      offset,
		}

		return n.Reply(msg, resp)
	})

	n.Handle("poll", func(msg maelstrom.Message) error {
		// Unmarshal the message body as a loosely-typed map.
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return fmt.Errorf("unmarshaling message body: %w", err)
		}

		offsetsRaw, found := body["offsets"]
		if !found {
			return errors.New("missing offsets in message body in 'poll' handler")
		}

		offsetsBytes, err := json.Marshal(offsetsRaw)
		if err != nil {
			return fmt.Errorf("marshaling offsets in 'poll' handler: %w", err)
		}

		var offsets map[string]int
		if err := json.Unmarshal(offsetsBytes, &offsets); err != nil {
			return fmt.Errorf("unmarshaling offsets message body in 'poll' handler: %w", err)
		}

		msgs := make(map[string][][]int, len(offsets))

		for key, offset := range offsets {
			records, err := listRecords(key, offset)
			if err != nil {
				return fmt.Errorf("listing records for queue %q in 'poll' handler: %w", key, err)
			}

			logs := make([][]int, 0, len(records))

			for _, r := range records {
				logs = append(logs, []int{r.offset, r.message})
			}

			msgs[key] = logs
		}

		resp := map[string]any{
			"type":        "poll_ok",
			"in_reply_to": body["msg_id"],
			"msgs":        msgs,
		}

		return n.Reply(msg, resp)
	})

	n.Handle("commit_offsets", func(msg maelstrom.Message) error {
		// Unmarshal the message body as a loosely-typed map.
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return fmt.Errorf("unmarshaling message body: %w", err)
		}

		offsetsRaw, found := body["offsets"]
		if !found {
			return errors.New("missing offsets in message body in 'commit_offsets' handler")
		}

		offsetsBytes, err := json.Marshal(offsetsRaw)
		if err != nil {
			return fmt.Errorf("marshaling offsets in 'commit_offsets' handler: %w", err)
		}

		var offsets map[string]int
		if err := json.Unmarshal(offsetsBytes, &offsets); err != nil {
			return fmt.Errorf("unmarshaling offsets message body in 'commit_offsets' handler: %w", err)
		}

		for key, offset := range offsets {
			if err := commitOffsets(key, offset); err != nil {
				return fmt.Errorf("committing offsets for queue %q in 'commit_offsets' handler: %w", key, err)
			}
		}

		resp := map[string]any{
			"type":        "commit_offsets_ok",
			"in_reply_to": body["msg_id"],
		}

		return n.Reply(msg, resp)
	})

	n.Handle("list_committed_offsets", func(msg maelstrom.Message) error {
		// Unmarshal the message body as a loosely-typed map.
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return fmt.Errorf("unmarshaling message body: %w", err)
		}

		keysRaw, found := body["keys"]
		if !found {
			return errors.New("missing keys in message body in 'list_committed_offsets' handler")
		}

		keysBytes, err := json.Marshal(keysRaw)
		if err != nil {
			return fmt.Errorf("marshaling offsets in 'list_committed_offsets' handler: %w", err)
		}

		var keys []string
		if err := json.Unmarshal(keysBytes, &keys); err != nil {
			return fmt.Errorf("unmarshaling keys message body in 'list_committed_offsets' handler: %w", err)
		}

		offsets := make(map[string]int, len(keys))

		for _, key := range keys {
			offset, err := getLastCommitedOffset(key)
			if err != nil {
				return fmt.Errorf("getting the last commited offset for queue %q in 'list_committed_offsets' handler: %w", key, err)
			}

			offsets[key] = offset
		}

		resp := map[string]any{
			"type":        "list_committed_offsets_ok",
			"in_reply_to": body["msg_id"],
			"offsets":     offsets,
		}

		return n.Reply(msg, resp)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

// queues stores the created queues in memory.
var queues = map[string]*queue{}

// enqueueRecord enqueues a record into a queue.
//
// Returns the offset of the queued record.
//
// If the queue does not exist, an error is returned.
// If the queue does not exist, create a new queue and enqueue.
func enqueueRecord(key string, message int) int {
	q, found := queues[key]
	if !found || q == nil {
		q = &queue{
			key:     key,
			records: []record{},
			mu:      &sync.Mutex{},
		}

		queues[key] = q
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	lastOffset := 0

	if len(q.records) > 0 {
		lastOffset = q.records[len(q.records)-1].offset
	}

	newOffset := lastOffset + 1

	q.records = append(q.records, record{
		offset:   newOffset,
		message:  message,
		commited: false,
	})

	queues[key] = q

	return newOffset
}

// listRecords returns the list of records from a queue, filtering records with
// offset greater than the specified offset.
func listRecords(key string, startOffset int) ([]record, error) {
	q, found := queues[key]
	if !found || q == nil {
		return nil, fmt.Errorf("queue %q not found", key)
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	if q.records == nil {
		return make([]record, 0), nil
	}

	records := make([]record, 0, len(q.records))

	for _, r := range q.records {
		if r.offset >= startOffset {
			records = append(records, r)
		}
	}

	return records, nil
}

// commitOffsets commits records up to the specified offset.
//
// If the queue does not exist, an error is returned.
func commitOffsets(key string, untilOffset int) error {
	q, found := queues[key]
	if !found || q == nil {
		return fmt.Errorf("queue %q not found", key)
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	for i, r := range q.records {
		if r.offset <= untilOffset && !r.commited {
			q.records[i].commited = true
		}
	}

	queues[key] = q

	return nil
}

// getLastCommitedOffset returns the last committed offset of a queue.
//
// If the queue does not exist, an error is returned.
func getLastCommitedOffset(key string) (int, error) {
	q, found := queues[key]
	if !found || q == nil {
		return 0, fmt.Errorf("queue %q not found", key)
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	var lastCommitedOffset int

	for _, r := range q.records {
		if r.commited {
			lastCommitedOffset = r.offset
		}
	}

	return lastCommitedOffset, nil
}

// queue represents a queue and its records.
type queue struct {
	key     string
	records []record
	mu      *sync.Mutex
}

// record represents a record in a queue.
type record struct {
	offset   int
	message  int
	commited bool
}
