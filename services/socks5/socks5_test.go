package socks5

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"sublink/models"
)

func TestNormalizeConfigDefaultsAndValidation(t *testing.T) {
	cfg, err := NormalizeConfig(Config{Enabled: true, Username: "user", Password: "pass", RequireAuth: true})
	if err != nil {
		t.Fatalf("normalize config: %v", err)
	}
	if cfg.ListenAddress != defaultListenAddress || cfg.Port != defaultPort || cfg.Selection != defaultSelection || !cfg.RequireAuth {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if _, err := NormalizeConfig(Config{Enabled: true, RequireAuth: true, Username: "user"}); err == nil {
		t.Fatal("expected missing password validation error")
	}
}

func TestServerHandlesAuthenticatedConnectAndPumpsTraffic(t *testing.T) {
	previous := selectNodeFunc
	selectNodeFunc = func(Config) (models.Node, error) {
		return models.Node{ID: 1, Link: "socks5://upstream:1080", Name: "test"}, nil
	}
	t.Cleanup(func() { selectNodeFunc = previous })

	client, serverConn := net.Pipe()
	defer func() { _ = client.Close() }()
	defer func() { _ = serverConn.Close() }()

	upstream, upstreamPeer := net.Pipe()
	defer func() { _ = upstreamPeer.Close() }()
	defer func() { _ = upstream.Close() }()

	server, err := NewServer(Config{Enabled: true, Username: "alice", Password: "secret", RequireAuth: true}, func(context.Context, models.Node, string, uint16) (net.Conn, error) {
		return upstream, nil
	})
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	done := make(chan struct{})
	go func() {
		server.serveConn(context.Background(), serverConn)
		close(done)
	}()

	if _, err := client.Write([]byte{0x05, 0x01, 0x02}); err != nil {
		t.Fatalf("write greeting: %v", err)
	}
	response := make([]byte, 2)
	if _, err := io.ReadFull(client, response); err != nil {
		t.Fatalf("read method response: %v", err)
	}
	if string(response) != string([]byte{0x05, 0x02}) {
		t.Fatalf("unexpected method response: %x", response)
	}
	if _, err := client.Write([]byte{0x01, 0x05, 'a', 'l', 'i', 'c', 'e', 0x06, 's', 'e', 'c', 'r', 'e', 't'}); err != nil {
		t.Fatalf("write auth: %v", err)
	}
	if _, err := io.ReadFull(client, response); err != nil {
		t.Fatalf("read auth response: %v", err)
	}
	if string(response) != string([]byte{0x01, 0x00}) {
		t.Fatalf("unexpected auth response: %x", response)
	}
	request := []byte{0x05, 0x01, 0x00, 0x03, 0x0b}
	request = append(request, []byte("example.com")...)
	request = append(request, 0x01, 0xbb)
	if _, err := client.Write(request); err != nil {
		t.Fatalf("write connect request: %v", err)
	}
	reply := make([]byte, 10)
	if _, err := io.ReadFull(client, reply); err != nil {
		t.Fatalf("read connect reply: %v", err)
	}
	if reply[1] != 0x00 {
		t.Fatalf("connect failed: %x", reply)
	}

	payload := []byte("hello through proxy")
	if _, err := client.Write(payload); err != nil {
		t.Fatalf("write payload: %v", err)
	}
	_ = upstreamPeer.SetReadDeadline(time.Now().Add(time.Second))
	received := make([]byte, len(payload))
	if _, err := io.ReadFull(upstreamPeer, received); err != nil {
		t.Fatalf("read upstream payload: %v", err)
	}
	if string(received) != string(payload) {
		t.Fatalf("unexpected payload: %q", received)
	}
	if _, err := upstreamPeer.Write([]byte("pong")); err != nil {
		t.Fatalf("write upstream response: %v", err)
	}
	_ = client.SetReadDeadline(time.Now().Add(time.Second))
	responsePayload := make([]byte, 4)
	if _, err := io.ReadFull(client, responsePayload); err != nil {
		t.Fatalf("read client response: %v", err)
	}
	if string(responsePayload) != "pong" {
		t.Fatalf("unexpected client response: %q", responsePayload)
	}
	_ = client.Close()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("server connection did not exit")
	}
}

func TestServerRejectsUnsupportedCommand(t *testing.T) {
	client, serverConn := net.Pipe()
	defer func() { _ = client.Close() }()
	defer func() { _ = serverConn.Close() }()
	server, err := NewServer(Config{Enabled: true, Username: "alice", Password: "secret", RequireAuth: true}, func(context.Context, models.Node, string, uint16) (net.Conn, error) {
		t.Fatal("dial should not be called")
		return nil, nil
	})
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	done := make(chan struct{})
	go func() { server.serveConn(context.Background(), serverConn); close(done) }()
	_, _ = client.Write([]byte{0x05, 0x01, 0x02})
	response := make([]byte, 2)
	_, _ = io.ReadFull(client, response)
	_, _ = client.Write([]byte{0x01, 0x05, 'a', 'l', 'i', 'c', 'e', 0x06, 's', 'e', 'c', 'r', 'e', 't'})
	_, _ = io.ReadFull(client, response)
	_, _ = client.Write([]byte{0x05, 0x02, 0x00, 0x01})
	reply := make([]byte, 10)
	if _, err := io.ReadFull(client, reply); err != nil {
		t.Fatalf("read reject reply: %v", err)
	}
	if reply[1] != 0x07 {
		t.Fatalf("expected command-not-supported reply, got %x", reply)
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("server connection did not exit")
	}
}
