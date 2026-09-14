package socks5

import (
	"context"
	"errors"
	"io"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"

	"sublink/models"
)

func withCandidateNodes(t *testing.T, nodes []models.Node) {
	t.Helper()
	previous := listCandidateNodesFunc
	listCandidateNodesFunc = func(Config) ([]models.Node, error) {
		copyNodes := append([]models.Node(nil), nodes...)
		return copyNodes, nil
	}
	t.Cleanup(func() { listCandidateNodesFunc = previous })
}

func TestNormalizeConfigPhaseTwoValidation(t *testing.T) {
	cfg, err := NormalizeConfig(Config{ListenAddress: "127.0.0.1", Selection: "round_robin"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MaxAttempts != 1 || cfg.DialTimeoutSeconds != 30 || cfg.FailureCooldownSeconds != 0 {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if _, err := NormalizeConfig(Config{MaxAttempts: 6}); err == nil {
		t.Fatal("expected max attempts validation error")
	}
	if _, err := NormalizeConfig(Config{DialTimeoutSeconds: 121}); err == nil {
		t.Fatal("expected dial timeout validation error")
	}
	if _, err := NormalizeConfig(Config{FailureCooldownSeconds: -1}); err == nil {
		t.Fatal("expected cooldown validation error")
	}
	if _, err := NormalizeConfig(Config{Enabled: true, ListenAddress: "0.0.0.0", RequireAuth: false}); err == nil {
		t.Fatal("expected open proxy protection error")
	}
	if _, err := NormalizeConfig(Config{Selection: "specific", NodeID: 1, SpecificFallback: true, MaxAttempts: 1}); err == nil {
		t.Fatal("expected specific fallback attempt validation error")
	}
}

func TestNodeRouterRoundRobin(t *testing.T) {
	nodes := []models.Node{{ID: 1, Link: "a"}, {ID: 2, Link: "b"}, {ID: 3, Link: "c"}}
	withCandidateNodes(t, nodes)
	router := newNodeRouter()
	cfg := Config{Selection: "round_robin", MaxAttempts: 3, DialTimeoutSeconds: 30}
	first, err := router.candidates(cfg)
	if err != nil {
		t.Fatal(err)
	}
	second, err := router.candidates(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual([]int{first[0].ID, second[0].ID}, []int{1, 2}) {
		t.Fatalf("unexpected rotation: first=%v second=%v", first, second)
	}
}

func TestNodeRouterSpecificFallback(t *testing.T) {
	previous := listCandidateNodesFunc
	listCandidateNodesFunc = func(cfg Config) ([]models.Node, error) {
		if cfg.SpecificFallback {
			return []models.Node{{ID: 1, Link: "a"}, {ID: 2, Link: "b"}}, nil
		}
		return []models.Node{{ID: 1, Link: "a"}}, nil
	}
	t.Cleanup(func() { listCandidateNodesFunc = previous })
	router := newNodeRouter()
	strict, err := router.candidates(Config{Selection: "specific", NodeID: 1, MaxAttempts: 5, DialTimeoutSeconds: 30})
	if err != nil || len(strict) != 1 {
		t.Fatalf("strict candidates=%v err=%v", strict, err)
	}
	fallback, err := router.candidates(Config{Selection: "specific", NodeID: 1, MaxAttempts: 5, DialTimeoutSeconds: 30, SpecificFallback: true})
	if err != nil || len(fallback) != 2 {
		t.Fatalf("fallback candidates=%v err=%v", fallback, err)
	}
}

func TestNodeHealthCooldownAndProbe(t *testing.T) {
	health := newNodeHealth()
	now := time.Unix(100, 0)
	health.now = func() time.Time { return now }
	nodes := []models.Node{{ID: 1, Link: "a"}, {ID: 2, Link: "b"}}
	health.recordFailure(nodes[0], 10*time.Second)
	filtered := health.filter(nodes)
	if len(filtered) != 1 || filtered[0].ID != 2 {
		t.Fatalf("unexpected filtered nodes: %+v", filtered)
	}
	health.recordFailure(nodes[1], 20*time.Second)
	filtered = health.filter(nodes)
	if len(filtered) != 1 || filtered[0].ID != 1 {
		t.Fatalf("expected earliest cooldown probe, got %+v", filtered)
	}
	now = now.Add(11 * time.Second)
	filtered = health.filter(nodes)
	if len(filtered) != 1 || filtered[0].ID != 1 {
		t.Fatalf("expected expired node, got %+v", filtered)
	}
}

func TestServerRetriesBeforeSuccessReply(t *testing.T) {
	withCandidateNodes(t, []models.Node{{ID: 1, Link: "a"}, {ID: 2, Link: "b"}})
	var mu sync.Mutex
	attempts := make([]int, 0, 2)
	upstream, upstreamPeer := net.Pipe()
	defer func() { _ = upstreamPeer.Close() }()
	server, err := NewServer(Config{Selection: "best", RequireAuth: false, MaxAttempts: 2, DialTimeoutSeconds: 1, FailureCooldownSeconds: 30}, func(_ context.Context, node models.Node, _ string, _ uint16) (net.Conn, error) {
		mu.Lock()
		attempts = append(attempts, node.ID)
		mu.Unlock()
		if node.ID == 1 {
			return nil, errors.New("first failed")
		}
		return upstream, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	client, serverConn := net.Pipe()
	defer func() { _ = client.Close() }()
	go server.serveConn(context.Background(), serverConn)
	_, _ = client.Write([]byte{0x05, 0x01, 0x00})
	method := make([]byte, 2)
	_, _ = io.ReadFull(client, method)
	request := []byte{0x05, 0x01, 0x00, 0x03, 0x0b}
	request = append(request, []byte("example.com")...)
	request = append(request, 0x00, 0x50)
	_, _ = client.Write(request)
	reply := make([]byte, 10)
	if _, err := io.ReadFull(client, reply); err != nil {
		t.Fatal(err)
	}
	if reply[1] != 0x00 {
		t.Fatalf("unexpected reply: %x", reply)
	}
	mu.Lock()
	defer mu.Unlock()
	if !reflect.DeepEqual(attempts, []int{1, 2}) {
		t.Fatalf("attempts=%v", attempts)
	}
}
