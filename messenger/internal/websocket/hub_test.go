package websocket

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func newTestClient(id uuid.UUID) *Client {
	return &Client{
		Send:     make(chan []byte, 16),
		Done:     make(chan struct{}),
		UserUUID: id,
	}
}

func broadcastTo(userIDs ...uuid.UUID) *BroadcastMessage {
	recipients := make(map[uuid.UUID]struct{}, len(userIDs))
	for _, id := range userIDs {
		recipients[id] = struct{}{}
	}
	return &BroadcastMessage{Recipients: recipients, Message: []byte("hello")}
}

func recv(t *testing.T, ch chan []byte) []byte {
	t.Helper()
	select {
	case msg := <-ch:
		return msg
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for message")
		return nil
	}
}

func recvNothing(t *testing.T, ch chan []byte) {
	t.Helper()
	select {
	case msg := <-ch:
		t.Fatalf("unexpected message: %s", msg)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestDeliverOnlyToRecipients(t *testing.T) {
	hub := NewHub(nil, nil, nil)
	alice := newTestClient(uuid.New())
	bob := newTestClient(uuid.New())
	hub.register(alice)
	hub.register(bob)

	hub.deliver(broadcastTo(alice.UserUUID))

	recv(t, alice.Send)
	recvNothing(t, bob.Send)
}

func TestDeliverExcludesSender(t *testing.T) {
	hub := NewHub(nil, nil, nil)
	sender := newTestClient(uuid.New())
	receiver := newTestClient(uuid.New())
	hub.register(sender)
	hub.register(receiver)

	msg := broadcastTo(sender.UserUUID, receiver.UserUUID)
	msg.ExcludeUID = sender.UserUUID
	hub.deliver(msg)

	recv(t, receiver.Send)
	recvNothing(t, sender.Send)
}

func TestUnregisterKeepsOtherConnectionsOfSameUser(t *testing.T) {
	hub := NewHub(nil, nil, nil)
	id := uuid.New()
	c1 := newTestClient(id)
	c2 := newTestClient(id)
	hub.register(c1)
	hub.register(c2)

	hub.unregister(c1)
	select {
	case <-c1.Done:
	case <-time.After(time.Second):
		t.Fatal("Done not closed after unregister")
	}

	hub.deliver(broadcastTo(id))
	recv(t, c2.Send)
	recvNothing(t, c1.Send)
}

func TestUnregisterUnknownClientIsNoop(t *testing.T) {
	hub := NewHub(nil, nil, nil)
	c := newTestClient(uuid.New())

	// Клиент не регистрировался — повторный unregister не должен паниковать
	hub.unregister(c)
	hub.unregister(c)

	select {
	case <-c.Done:
		t.Fatal("Done must not be closed for unknown client")
	case <-time.After(50 * time.Millisecond):
	}
}

func TestRunLoopRegisterUnregister(t *testing.T) {
	hub := NewHub(nil, nil, nil)
	go hub.Run()

	c := newTestClient(uuid.New())
	hub.Register <- c
	time.Sleep(100 * time.Millisecond)
	hub.Unregister <- c

	select {
	case <-c.Done:
	case <-time.After(time.Second):
		t.Fatal("Done not closed after unregister via Run loop")
	}
}
