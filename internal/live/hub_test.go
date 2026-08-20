package live

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestUserChannelKey(t *testing.T) {
	if UserChannelKey(0) != "" || UserChannelKey(-1) != "" {
		t.Fatal("invalid ids should be empty")
	}
	if got := UserChannelKey(42); got != "user:42" {
		t.Fatalf("got %q", got)
	}
}

func TestUniquePositive(t *testing.T) {
	got := uniquePositive([]int{0, 2, 2, -1, 3, 2})
	if len(got) != 2 || got[0] != 2 || got[1] != 3 {
		t.Fatalf("got %#v", got)
	}
}

func TestHubSubscribeBroadcastAndUnsubscribe(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	ch := h.Subscribe(ctx, UserChannelKey(1))

	h.Publish(Event{Type: TypeTaskUpdated, TaskID: 9, ActorID: 1}, []int{1})

	select {
	case payload := <-ch:
		var ev Event
		if err := json.Unmarshal(payload, &ev); err != nil {
			t.Fatal(err)
		}
		if ev.Type != TypeTaskUpdated || ev.TaskID != 9 || ev.ActorID != 1 {
			t.Fatalf("unexpected event %#v", ev)
		}
		if ev.Origin != h.Origin() || ev.Timestamp == "" {
			t.Fatalf("missing origin/timestamp %#v", ev)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event")
	}

	cancel()
	time.Sleep(20 * time.Millisecond)
	h.Publish(Event{Type: TypeTaskCreated, TaskID: 1}, []int{1})
	select {
	case <-ch:
		t.Fatal("unsubscribed client should not receive")
	case <-time.After(30 * time.Millisecond):
	}
}

func TestHubDropsWhenBufferFull(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := h.Subscribe(ctx, UserChannelKey(1))
	for i := 0; i < 16; i++ {
		h.Publish(Event{Type: TypeTaskUpdated, TaskID: i + 1}, []int{1})
	}
	n := 0
drain:
	for {
		select {
		case <-ch:
			n++
		default:
			break drain
		}
	}
	if n > 8 {
		t.Fatalf("buffer should drop; got %d", n)
	}
	if n == 0 {
		t.Fatal("expected some delivered events")
	}
}

func TestHubSkipsRemoteEchoFromSameOrigin(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := h.Subscribe(ctx, UserChannelKey(1))

	payload, _ := json.Marshal(Event{Type: TypeTaskUpdated, TaskID: 5, Origin: h.Origin()})
	h.handleRemote(UserChannelKey(1), payload)
	select {
	case <-ch:
		t.Fatal("same-origin redis echo should be skipped")
	case <-time.After(30 * time.Millisecond):
	}

	payload, _ = json.Marshal(Event{Type: TypeTaskUpdated, TaskID: 6, Origin: "other"})
	h.handleRemote(UserChannelKey(1), payload)
	select {
	case raw := <-ch:
		var ev Event
		if err := json.Unmarshal(raw, &ev); err != nil {
			t.Fatal(err)
		}
		if ev.TaskID != 6 {
			t.Fatalf("got %#v", ev)
		}
	case <-time.After(time.Second):
		t.Fatal("expected remote event")
	}
}

func TestAfterTaskChangeNoopsWithoutHub(t *testing.T) {
	hubMu.Lock()
	hub = nil
	hubMu.Unlock()
	AfterTaskChange(1, 1, TypeTaskUpdated)
}

func TestPublishDedupesUserIDs(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := h.Subscribe(ctx, UserChannelKey(7))
	h.Publish(Event{Type: TypeTaskCreated, TaskID: 1}, []int{7, 7, 0, 7})
	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("expected one event")
	}
	select {
	case <-ch:
		t.Fatal("duplicate user id should not deliver twice")
	case <-time.After(30 * time.Millisecond):
	}
}
