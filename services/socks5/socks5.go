package socks5

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"sublink/models"
	"sublink/services/mihomo"

	"github.com/metacubex/mihomo/constant"
)

const (
	settingEnabled       = "socks5_enabled"
	settingListenAddress = "socks5_listen_address"
	settingPort          = "socks5_port"
	settingUsername      = "socks5_username"
	settingPassword      = "socks5_password"
	settingNodeID        = "socks5_node_id"
	settingSelection     = "socks5_selection"
	settingRequireAuth   = "socks5_require_auth"

	defaultListenAddress = "127.0.0.1"
	defaultPort          = 1080
	defaultSelection     = "best"
	handshakeTimeout     = 15 * time.Second
	dialTimeout          = 30 * time.Second
)

// Config controls the phase-one SOCKS5 gateway.
type Config struct {
	Enabled       bool
	ListenAddress string
	Port          int
	Username      string
	Password      string
	NodeID        int
	Selection     string
	RequireAuth   bool
	ClearPassword bool
}

// PublicConfig is safe to return from the settings API.
type PublicConfig struct {
	Enabled        bool   `json:"enabled"`
	ListenAddress  string `json:"listenAddress"`
	Port           int    `json:"port"`
	Username       string `json:"username"`
	HasPassword    bool   `json:"hasPassword"`
	MaskedPassword string `json:"maskedPassword,omitempty"`
	NodeID         int    `json:"nodeId"`
	Selection      string `json:"selection"`
	RequireAuth    bool   `json:"requireAuth"`
	Running        bool   `json:"running"`
	BoundAddress   string `json:"boundAddress,omitempty"`
}

func defaultConfig() Config {
	return Config{ListenAddress: defaultListenAddress, Port: defaultPort, Selection: defaultSelection, RequireAuth: true}
}

func LoadConfig() (Config, error) {
	cfg := defaultConfig()
	if value, err := models.GetSetting(settingEnabled); err == nil {
		cfg.Enabled = value == "true"
	}
	if value, err := models.GetSetting(settingListenAddress); err == nil && strings.TrimSpace(value) != "" {
		cfg.ListenAddress = strings.TrimSpace(value)
	}
	if value, err := models.GetSetting(settingPort); err == nil && strings.TrimSpace(value) != "" {
		parsed, parseErr := strconv.Atoi(value)
		if parseErr != nil {
			return cfg, fmt.Errorf("invalid SOCKS5 port: %w", parseErr)
		}
		cfg.Port = parsed
	}
	if value, err := models.GetSetting(settingUsername); err == nil {
		cfg.Username = value
	}
	if value, err := models.GetSetting(settingPassword); err == nil && strings.TrimSpace(value) != "" {
		password, decryptErr := models.DecryptUserAISecret(value)
		if decryptErr != nil {
			return cfg, fmt.Errorf("decrypt SOCKS5 password: %w", decryptErr)
		}
		cfg.Password = password
	}
	if value, err := models.GetSetting(settingNodeID); err == nil && strings.TrimSpace(value) != "" {
		parsed, parseErr := strconv.Atoi(value)
		if parseErr != nil {
			return cfg, fmt.Errorf("invalid SOCKS5 node ID: %w", parseErr)
		}
		cfg.NodeID = parsed
	}
	if value, err := models.GetSetting(settingSelection); err == nil && strings.TrimSpace(value) != "" {
		cfg.Selection = strings.ToLower(strings.TrimSpace(value))
	}
	if value, err := models.GetSetting(settingRequireAuth); err == nil && strings.TrimSpace(value) != "" {
		cfg.RequireAuth = value != "false"
	}
	return NormalizeConfig(cfg)
}

func NormalizeConfig(cfg Config) (Config, error) {
	if strings.TrimSpace(cfg.ListenAddress) == "" {
		cfg.ListenAddress = defaultListenAddress
	}
	cfg.ListenAddress = strings.TrimSpace(cfg.ListenAddress)
	if cfg.Port == 0 {
		cfg.Port = defaultPort
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return cfg, errors.New("SOCKS5 port must be between 1 and 65535")
	}
	switch strings.ToLower(strings.TrimSpace(cfg.Selection)) {
	case "", "best":
		cfg.Selection = defaultSelection
	case "random", "specific":
		cfg.Selection = strings.ToLower(strings.TrimSpace(cfg.Selection))
	default:
		return cfg, errors.New("SOCKS5 selection must be best, random, or specific")
	}
	if cfg.Enabled && cfg.RequireAuth && (strings.TrimSpace(cfg.Username) == "" || cfg.Password == "") {
		return cfg, errors.New("SOCKS5 username and password are required when authentication is enabled")
	}
	if len(cfg.Username) > 255 || len(cfg.Password) > 255 {
		return cfg, errors.New("SOCKS5 username and password must be at most 255 bytes")
	}
	if cfg.Selection == "specific" && cfg.NodeID <= 0 {
		return cfg, errors.New("a node ID is required for specific SOCKS5 selection")
	}
	return cfg, nil
}

func SaveConfig(input Config) (Config, error) {
	cfg, err := NormalizeConfig(input)
	if err != nil {
		return cfg, err
	}
	values := map[string]string{
		settingEnabled:       strconv.FormatBool(cfg.Enabled),
		settingListenAddress: cfg.ListenAddress,
		settingPort:          strconv.Itoa(cfg.Port),
		settingUsername:      strings.TrimSpace(cfg.Username),
		settingNodeID:        strconv.Itoa(cfg.NodeID),
		settingSelection:     cfg.Selection,
		settingRequireAuth:   strconv.FormatBool(cfg.RequireAuth),
	}
	if cfg.ClearPassword {
		values[settingPassword] = ""
	} else if cfg.Password != "" {
		encrypted, encryptErr := models.EncryptUserAISecret(cfg.Password)
		if encryptErr != nil {
			return cfg, fmt.Errorf("encrypt SOCKS5 password: %w", encryptErr)
		}
		values[settingPassword] = encrypted
	} else if existing, getErr := models.GetSetting(settingPassword); getErr != nil || existing == "" {
		values[settingPassword] = ""
	}
	for key, value := range values {
		if err := models.SetSetting(key, value); err != nil {
			return cfg, err
		}
	}
	return LoadConfig()
}

func ToPublicConfig(cfg Config, running bool, boundAddress string) PublicConfig {
	masked := ""
	if cfg.Password != "" {
		masked = "••••••••"
	}
	return PublicConfig{Enabled: cfg.Enabled, ListenAddress: cfg.ListenAddress, Port: cfg.Port, Username: cfg.Username, HasPassword: cfg.Password != "", MaskedPassword: masked, NodeID: cfg.NodeID, Selection: cfg.Selection, RequireAuth: cfg.RequireAuth, Running: running, BoundAddress: boundAddress}
}

// DialFunc allows tests and future routing implementations to replace mihomo dialing.
type DialFunc func(ctx context.Context, node models.Node, host string, port uint16) (net.Conn, error)

func defaultDial(ctx context.Context, node models.Node, host string, port uint16) (net.Conn, error) {
	adapter, err := mihomo.GetMihomoAdapter(node.Link)
	if err != nil {
		return nil, err
	}
	return adapter.DialContext(ctx, &constant.Metadata{NetWork: constant.TCP, Type: constant.SOCKS5, Host: host, DstPort: port})
}

// Server serves SOCKS5 CONNECT requests over a supplied listener.
type Server struct {
	cfg  Config
	dial DialFunc
}

func NewServer(cfg Config, dial DialFunc) (*Server, error) {
	normalized, err := NormalizeConfig(cfg)
	if err != nil {
		return nil, err
	}
	if dial == nil {
		dial = defaultDial
	}
	return &Server{cfg: normalized, dial: dial}, nil
}

func (s *Server) Serve(ctx context.Context, listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return nil
			}
			return err
		}
		go s.serveConn(ctx, conn)
	}
}

func (s *Server) serveConn(ctx context.Context, client net.Conn) {
	defer client.Close()
	_ = client.SetDeadline(time.Now().Add(handshakeTimeout))
	if s.cfg.RequireAuth {
		if err := s.authenticate(client); err != nil {
			return
		}
	} else if err := negotiateNoAuth(client); err != nil {
		return
	}
	host, port, err := readConnectRequest(client)
	if err != nil {
		return
	}
	node, err := selectNodeFunc(s.cfg)
	if err != nil {
		_ = writeReply(client, 0x01)
		return
	}
	dialCtx, cancel := context.WithTimeout(ctx, dialTimeout)
	defer cancel()
	upstream, err := s.dial(dialCtx, node, host, port)
	if err != nil {
		_ = writeReply(client, 0x01)
		return
	}
	defer upstream.Close()
	_ = client.SetDeadline(time.Time{})
	if err := writeReply(client, 0x00); err != nil {
		return
	}
	pump(client, upstream)
}

func (s *Server) authenticate(conn net.Conn) error {
	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil || header[0] != 0x05 {
		return errors.New("invalid SOCKS5 greeting")
	}
	methods := make([]byte, int(header[1]))
	if _, err := io.ReadFull(conn, methods); err != nil {
		return err
	}
	offered := false
	for _, method := range methods {
		if method == 0x02 {
			offered = true
			break
		}
	}
	if !offered {
		_, _ = conn.Write([]byte{0x05, 0xff})
		return errors.New("username/password authentication not offered")
	}
	if _, err := conn.Write([]byte{0x05, 0x02}); err != nil {
		return err
	}
	authHeader := make([]byte, 2)
	if _, err := io.ReadFull(conn, authHeader); err != nil || authHeader[0] != 0x01 {
		return errors.New("invalid SOCKS5 auth version")
	}
	username := make([]byte, int(authHeader[1]))
	if _, err := io.ReadFull(conn, username); err != nil {
		return err
	}
	length := make([]byte, 1)
	if _, err := io.ReadFull(conn, length); err != nil {
		return err
	}
	password := make([]byte, int(length[0]))
	if _, err := io.ReadFull(conn, password); err != nil {
		return err
	}
	if string(username) != s.cfg.Username || string(password) != s.cfg.Password {
		_, _ = conn.Write([]byte{0x01, 0x01})
		return errors.New("invalid SOCKS5 credentials")
	}
	_, err := conn.Write([]byte{0x01, 0x00})
	return err
}

func negotiateNoAuth(conn net.Conn) error {
	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil || header[0] != 0x05 {
		return errors.New("invalid SOCKS5 greeting")
	}
	methods := make([]byte, int(header[1]))
	if _, err := io.ReadFull(conn, methods); err != nil {
		return err
	}
	for _, method := range methods {
		if method == 0x00 {
			_, err := conn.Write([]byte{0x05, 0x00})
			return err
		}
	}
	_, _ = conn.Write([]byte{0x05, 0xff})
	return errors.New("no-auth method not offered")
}

func readConnectRequest(conn net.Conn) (string, uint16, error) {
	header := make([]byte, 4)
	if _, err := io.ReadFull(conn, header); err != nil || header[0] != 0x05 {
		return "", 0, errors.New("invalid SOCKS5 request")
	}
	if header[1] != 0x01 {
		_ = writeReply(conn, 0x07)
		return "", 0, errors.New("SOCKS5 command is not CONNECT")
	}
	var host string
	switch header[3] {
	case 0x01:
		address := make([]byte, net.IPv4len)
		if _, err := io.ReadFull(conn, address); err != nil {
			return "", 0, err
		}
		host = net.IP(address).String()
	case 0x03:
		length := make([]byte, 1)
		if _, err := io.ReadFull(conn, length); err != nil {
			return "", 0, err
		}
		address := make([]byte, int(length[0]))
		if _, err := io.ReadFull(conn, address); err != nil {
			return "", 0, err
		}
		host = string(address)
	case 0x04:
		address := make([]byte, net.IPv6len)
		if _, err := io.ReadFull(conn, address); err != nil {
			return "", 0, err
		}
		host = net.IP(address).String()
	default:
		_ = writeReply(conn, 0x08)
		return "", 0, errors.New("unsupported SOCKS5 address type")
	}
	portBytes := make([]byte, 2)
	if _, err := io.ReadFull(conn, portBytes); err != nil {
		return "", 0, err
	}
	return host, uint16(portBytes[0])<<8 | uint16(portBytes[1]), nil
}

func writeReply(conn net.Conn, code byte) error {
	_, err := conn.Write([]byte{0x05, code, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	return err
}

func pump(client, upstream net.Conn) {
	done := make(chan struct{}, 2)
	go func() { _, _ = io.Copy(upstream, client); done <- struct{}{} }()
	go func() { _, _ = io.Copy(client, upstream); done <- struct{}{} }()
	<-done
}

var selectNodeFunc = selectNode

func selectNode(cfg Config) (models.Node, error) {
	if cfg.Selection == "specific" {
		if node, ok := models.GetNodeByID(cfg.NodeID); ok && strings.TrimSpace(node.Link) != "" {
			return *node, nil
		}
		return models.Node{}, errors.New("configured SOCKS5 node was not found")
	}
	if cfg.Selection == "best" {
		if node, err := models.GetBestProxyNode(); err == nil && node != nil && strings.TrimSpace(node.Link) != "" {
			return *node, nil
		}
	}
	var modelNode models.Node
	nodes, err := modelNode.ListWithFilters(models.NodeFilter{})
	if err != nil {
		return models.Node{}, err
	}
	available := make([]models.Node, 0, len(nodes))
	for _, node := range nodes {
		if strings.TrimSpace(node.Link) != "" {
			available = append(available, node)
		}
	}
	if len(available) == 0 {
		return models.Node{}, errors.New("no proxy nodes are available")
	}
	return available[rand.Intn(len(available))], nil
}

// Manager owns the optional long-lived listener.
type Manager struct {
	mu       sync.Mutex
	listener net.Listener
	cancel   context.CancelFunc
	cfg      Config
	dial     DialFunc
}

var defaultManager = &Manager{}

func DefaultManager() *Manager { return defaultManager }

func (m *Manager) Apply(cfg Config) error {
	normalized, err := NormalizeConfig(cfg)
	if err != nil {
		return err
	}
	m.Stop()
	m.mu.Lock()
	m.cfg = normalized
	dial := m.dial
	m.mu.Unlock()
	if !normalized.Enabled {
		return nil
	}
	listener, err := net.Listen("tcp", net.JoinHostPort(normalized.ListenAddress, strconv.Itoa(normalized.Port)))
	if err != nil {
		return fmt.Errorf("listen SOCKS5: %w", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	m.mu.Lock()
	m.listener = listener
	m.cancel = cancel
	m.mu.Unlock()
	server, err := NewServer(normalized, dial)
	if err != nil {
		_ = listener.Close()
		cancel()
		return err
	}
	go func() {
		if serveErr := server.Serve(ctx, listener); serveErr != nil {
			// Listener failures are reflected by Running=false; avoid crashing the HTTP service.
		}
		m.mu.Lock()
		if m.listener == listener {
			m.listener = nil
			m.cancel = nil
		}
		m.mu.Unlock()
	}()
	return nil
}

func (m *Manager) StartFromSettings() error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	return m.Apply(cfg)
}

func (m *Manager) Stop() {
	m.mu.Lock()
	listener := m.listener
	cancel := m.cancel
	m.listener = nil
	m.cancel = nil
	m.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if listener != nil {
		_ = listener.Close()
	}
}

func (m *Manager) Status() (bool, string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.listener == nil {
		return false, ""
	}
	return true, m.listener.Addr().String()
}

func (m *Manager) SetDialFuncForTest(dial DialFunc) func() {
	m.mu.Lock()
	previous := m.dial
	m.dial = dial
	m.mu.Unlock()
	return func() {
		m.mu.Lock()
		m.dial = previous
		m.mu.Unlock()
	}
}
