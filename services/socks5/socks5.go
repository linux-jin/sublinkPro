package socks5

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"sublink/models"
	"sublink/utils"

	"github.com/metacubex/mihomo/constant"
)

const (
	settingEnabled                      = "socks5_enabled"
	settingListenAddress                = "socks5_listen_address"
	settingPort                         = "socks5_port"
	settingUsername                     = "socks5_username"
	settingPassword                     = "socks5_password"
	settingNodeID                       = "socks5_node_id"
	settingSelection                    = "socks5_selection"
	settingRequireAuth                  = "socks5_require_auth"
	settingMaxAttempts                  = "socks5_max_attempts"
	settingDialTimeoutSeconds           = "socks5_dial_timeout_seconds"
	settingFailureCooldownSeconds       = "socks5_failure_cooldown_seconds"
	settingSpecificFallback             = "socks5_specific_fallback"
	settingMaxConnections               = "socks5_max_connections"
	settingMaxConnectionsPerClient      = "socks5_max_connections_per_client"
	settingIdleTimeoutSeconds           = "socks5_idle_timeout_seconds"
	settingMaxConnectionDurationSeconds = "socks5_max_connection_duration_seconds"
	settingHealthCheckEnabled           = "socks5_health_check_enabled"
	settingHealthCheckIntervalSeconds   = "socks5_health_check_interval_seconds"
	settingHealthCheckTimeoutSeconds    = "socks5_health_check_timeout_seconds"

	defaultListenAddress                = "127.0.0.1"
	defaultPort                         = 1080
	defaultSelection                    = "best"
	defaultMaxAttempts                  = 1
	defaultDialTimeoutSeconds           = 30
	defaultFailureCooldownSeconds       = 0
	defaultMaxConnections               = 256
	defaultMaxConnectionsPerClient      = 32
	defaultIdleTimeoutSeconds           = 600
	defaultMaxConnectionDurationSeconds = 0
	defaultHealthCheckIntervalSeconds   = 60
	defaultHealthCheckTimeoutSeconds    = 5
	handshakeTimeout                    = 15 * time.Second
)

// Config controls the SOCKS5 gateway.
type Config struct {
	Enabled                      bool
	ListenAddress                string
	Port                         int
	Username                     string
	Password                     string
	NodeID                       int
	Selection                    string
	RequireAuth                  bool
	MaxAttempts                  int
	DialTimeoutSeconds           int
	FailureCooldownSeconds       int
	SpecificFallback             bool
	MaxConnections               int
	MaxConnectionsPerClient      int
	IdleTimeoutSeconds           int
	MaxConnectionDurationSeconds int
	HealthCheckEnabled           bool
	HealthCheckIntervalSeconds   int
	HealthCheckTimeoutSeconds    int
	ClearPassword                bool
}

// PublicConfig is safe to return from the settings API.
type PublicConfig struct {
	Enabled                      bool   `json:"enabled"`
	ListenAddress                string `json:"listenAddress"`
	Port                         int    `json:"port"`
	Username                     string `json:"username"`
	HasPassword                  bool   `json:"hasPassword"`
	MaskedPassword               string `json:"maskedPassword,omitempty"`
	NodeID                       int    `json:"nodeId"`
	Selection                    string `json:"selection"`
	RequireAuth                  bool   `json:"requireAuth"`
	MaxAttempts                  int    `json:"maxAttempts"`
	DialTimeoutSeconds           int    `json:"dialTimeoutSeconds"`
	FailureCooldownSeconds       int    `json:"failureCooldownSeconds"`
	SpecificFallback             bool   `json:"specificFallback"`
	MaxConnections               int    `json:"maxConnections"`
	MaxConnectionsPerClient      int    `json:"maxConnectionsPerClient"`
	IdleTimeoutSeconds           int    `json:"idleTimeoutSeconds"`
	MaxConnectionDurationSeconds int    `json:"maxConnectionDurationSeconds"`
	HealthCheckEnabled           bool   `json:"healthCheckEnabled"`
	HealthCheckIntervalSeconds   int    `json:"healthCheckIntervalSeconds"`
	HealthCheckTimeoutSeconds    int    `json:"healthCheckTimeoutSeconds"`
	Running                      bool   `json:"running"`
	BoundAddress                 string `json:"boundAddress,omitempty"`
}

func defaultConfig() Config {
	return Config{ListenAddress: defaultListenAddress, Port: defaultPort, Selection: defaultSelection, RequireAuth: true, MaxAttempts: defaultMaxAttempts, DialTimeoutSeconds: defaultDialTimeoutSeconds, FailureCooldownSeconds: defaultFailureCooldownSeconds, MaxConnections: defaultMaxConnections, MaxConnectionsPerClient: defaultMaxConnectionsPerClient, IdleTimeoutSeconds: defaultIdleTimeoutSeconds, MaxConnectionDurationSeconds: defaultMaxConnectionDurationSeconds, HealthCheckIntervalSeconds: defaultHealthCheckIntervalSeconds, HealthCheckTimeoutSeconds: defaultHealthCheckTimeoutSeconds}
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
	if value, err := models.GetSetting(settingMaxAttempts); err == nil && strings.TrimSpace(value) != "" {
		parsed, parseErr := strconv.Atoi(value)
		if parseErr != nil {
			return cfg, fmt.Errorf("invalid SOCKS5 max attempts: %w", parseErr)
		}
		cfg.MaxAttempts = parsed
	}
	if value, err := models.GetSetting(settingDialTimeoutSeconds); err == nil && strings.TrimSpace(value) != "" {
		parsed, parseErr := strconv.Atoi(value)
		if parseErr != nil {
			return cfg, fmt.Errorf("invalid SOCKS5 dial timeout: %w", parseErr)
		}
		cfg.DialTimeoutSeconds = parsed
	}
	if value, err := models.GetSetting(settingFailureCooldownSeconds); err == nil && strings.TrimSpace(value) != "" {
		parsed, parseErr := strconv.Atoi(value)
		if parseErr != nil {
			return cfg, fmt.Errorf("invalid SOCKS5 failure cooldown: %w", parseErr)
		}
		cfg.FailureCooldownSeconds = parsed
	}
	if value, err := models.GetSetting(settingSpecificFallback); err == nil && strings.TrimSpace(value) != "" {
		cfg.SpecificFallback = value == "true"
	}
	if value, err := models.GetSetting(settingHealthCheckEnabled); err == nil && strings.TrimSpace(value) != "" {
		cfg.HealthCheckEnabled = value == "true"
	}
	for key, target := range map[string]*int{settingMaxConnections: &cfg.MaxConnections, settingMaxConnectionsPerClient: &cfg.MaxConnectionsPerClient, settingIdleTimeoutSeconds: &cfg.IdleTimeoutSeconds, settingMaxConnectionDurationSeconds: &cfg.MaxConnectionDurationSeconds, settingHealthCheckIntervalSeconds: &cfg.HealthCheckIntervalSeconds, settingHealthCheckTimeoutSeconds: &cfg.HealthCheckTimeoutSeconds} {
		if value, err := models.GetSetting(key); err == nil && strings.TrimSpace(value) != "" {
			parsed, parseErr := strconv.Atoi(value)
			if parseErr != nil {
				return cfg, fmt.Errorf("invalid SOCKS5 setting %s: %w", key, parseErr)
			}
			*target = parsed
		}
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
	case "random", "round_robin", "specific":
		cfg.Selection = strings.ToLower(strings.TrimSpace(cfg.Selection))
	default:
		return cfg, errors.New("SOCKS5 selection must be best, random, round_robin, or specific")
	}
	if cfg.MaxAttempts == 0 {
		cfg.MaxAttempts = defaultMaxAttempts
	}
	if cfg.MaxAttempts < 1 || cfg.MaxAttempts > 5 {
		return cfg, errors.New("SOCKS5 max attempts must be between 1 and 5")
	}
	if cfg.DialTimeoutSeconds == 0 {
		cfg.DialTimeoutSeconds = defaultDialTimeoutSeconds
	}
	if cfg.DialTimeoutSeconds < 1 || cfg.DialTimeoutSeconds > 120 {
		return cfg, errors.New("SOCKS5 dial timeout must be between 1 and 120 seconds")
	}
	if cfg.FailureCooldownSeconds < 0 || cfg.FailureCooldownSeconds > 3600 {
		return cfg, errors.New("SOCKS5 failure cooldown must be between 0 and 3600 seconds")
	}
	if cfg.Enabled && !cfg.RequireAuth && !isLoopbackListenAddress(cfg.ListenAddress) {
		return cfg, errors.New("SOCKS5 authentication can only be disabled on a loopback listen address")
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
	if cfg.Selection == "specific" && cfg.SpecificFallback && cfg.MaxAttempts < 2 {
		return cfg, errors.New("SOCKS5 max attempts must be at least 2 when specific-node fallback is enabled")
	}
	if cfg.MaxConnections == 0 {
		cfg.MaxConnections = defaultMaxConnections
	}
	if cfg.MaxConnections < 1 || cfg.MaxConnections > 10000 {
		return cfg, errors.New("SOCKS5 max connections must be between 1 and 10000")
	}
	if cfg.MaxConnectionsPerClient == 0 {
		cfg.MaxConnectionsPerClient = defaultMaxConnectionsPerClient
	}
	if cfg.MaxConnectionsPerClient < 1 || cfg.MaxConnectionsPerClient > cfg.MaxConnections {
		return cfg, errors.New("SOCKS5 per-client connection limit is invalid")
	}
	if cfg.IdleTimeoutSeconds < 0 || cfg.IdleTimeoutSeconds > 86400 {
		return cfg, errors.New("SOCKS5 idle timeout must be between 0 and 86400 seconds")
	}
	if cfg.MaxConnectionDurationSeconds < 0 || cfg.MaxConnectionDurationSeconds > 604800 {
		return cfg, errors.New("SOCKS5 max connection duration must be between 0 and 604800 seconds")
	}
	if cfg.HealthCheckIntervalSeconds == 0 {
		cfg.HealthCheckIntervalSeconds = defaultHealthCheckIntervalSeconds
	}
	if cfg.HealthCheckIntervalSeconds < 10 || cfg.HealthCheckIntervalSeconds > 3600 {
		return cfg, errors.New("SOCKS5 health check interval must be between 10 and 3600 seconds")
	}
	if cfg.HealthCheckTimeoutSeconds == 0 {
		cfg.HealthCheckTimeoutSeconds = defaultHealthCheckTimeoutSeconds
	}
	if cfg.HealthCheckTimeoutSeconds < 1 || cfg.HealthCheckTimeoutSeconds > 30 {
		return cfg, errors.New("SOCKS5 health check timeout must be between 1 and 30 seconds")
	}
	return cfg, nil
}

func SaveConfig(input Config) (Config, error) {
	cfg, err := NormalizeConfig(input)
	if err != nil {
		return cfg, err
	}
	values := map[string]string{
		settingEnabled:                      strconv.FormatBool(cfg.Enabled),
		settingListenAddress:                cfg.ListenAddress,
		settingPort:                         strconv.Itoa(cfg.Port),
		settingUsername:                     strings.TrimSpace(cfg.Username),
		settingNodeID:                       strconv.Itoa(cfg.NodeID),
		settingSelection:                    cfg.Selection,
		settingRequireAuth:                  strconv.FormatBool(cfg.RequireAuth),
		settingMaxAttempts:                  strconv.Itoa(cfg.MaxAttempts),
		settingDialTimeoutSeconds:           strconv.Itoa(cfg.DialTimeoutSeconds),
		settingFailureCooldownSeconds:       strconv.Itoa(cfg.FailureCooldownSeconds),
		settingSpecificFallback:             strconv.FormatBool(cfg.SpecificFallback),
		settingMaxConnections:               strconv.Itoa(cfg.MaxConnections),
		settingMaxConnectionsPerClient:      strconv.Itoa(cfg.MaxConnectionsPerClient),
		settingIdleTimeoutSeconds:           strconv.Itoa(cfg.IdleTimeoutSeconds),
		settingMaxConnectionDurationSeconds: strconv.Itoa(cfg.MaxConnectionDurationSeconds),
		settingHealthCheckEnabled:           strconv.FormatBool(cfg.HealthCheckEnabled),
		settingHealthCheckIntervalSeconds:   strconv.Itoa(cfg.HealthCheckIntervalSeconds),
		settingHealthCheckTimeoutSeconds:    strconv.Itoa(cfg.HealthCheckTimeoutSeconds),
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
	return PublicConfig{Enabled: cfg.Enabled, ListenAddress: cfg.ListenAddress, Port: cfg.Port, Username: cfg.Username, HasPassword: cfg.Password != "", MaskedPassword: masked, NodeID: cfg.NodeID, Selection: cfg.Selection, RequireAuth: cfg.RequireAuth, MaxAttempts: cfg.MaxAttempts, DialTimeoutSeconds: cfg.DialTimeoutSeconds, FailureCooldownSeconds: cfg.FailureCooldownSeconds, SpecificFallback: cfg.SpecificFallback, MaxConnections: cfg.MaxConnections, MaxConnectionsPerClient: cfg.MaxConnectionsPerClient, IdleTimeoutSeconds: cfg.IdleTimeoutSeconds, MaxConnectionDurationSeconds: cfg.MaxConnectionDurationSeconds, HealthCheckEnabled: cfg.HealthCheckEnabled, HealthCheckIntervalSeconds: cfg.HealthCheckIntervalSeconds, HealthCheckTimeoutSeconds: cfg.HealthCheckTimeoutSeconds, Running: running, BoundAddress: boundAddress}
}

// DialFunc allows tests and future routing implementations to replace mihomo dialing.
type DialFunc func(ctx context.Context, node models.Node, host string, port uint16) (net.Conn, error)

// Server serves SOCKS5 CONNECT requests over a supplied listener.
type Server struct {
	cfg         Config
	dial        DialFunc
	pool        *adapterPool
	router      *nodeRouter
	registry    *sessionRegistry
	healthProbe HealthProbeFunc
	healthSweep healthSweepState
	connMu      sync.Mutex
	activeConns int
	connZero    chan struct{}
	closing     bool
	closeOnce   sync.Once
	closed      chan struct{}
}

func NewServer(cfg Config, dial DialFunc) (*Server, error) {
	normalized, err := NormalizeConfig(cfg)
	if err != nil {
		return nil, err
	}
	server := &Server{cfg: normalized, dial: dial, router: newNodeRouter(), registry: newSessionRegistry(normalized.MaxConnections, normalized.MaxConnectionsPerClient), closed: make(chan struct{})}
	if server.dial == nil {
		server.pool = newAdapterPool(nil)
		server.dial = server.dialWithPool
	}
	server.healthProbe = server.defaultHealthProbe
	return server, nil
}

func (s *Server) Serve(ctx context.Context, listener net.Listener) error {
	s.startHealthLoop(ctx)
	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return nil
			}
			return err
		}
		s.connMu.Lock()
		if s.closing {
			s.connMu.Unlock()
			_ = conn.Close()
			continue
		}
		if s.activeConns == 0 {
			s.connZero = make(chan struct{})
		}
		s.activeConns++
		s.connMu.Unlock()
		go func() {
			defer s.finishConnection()
			s.serveConn(ctx, conn)
		}()
	}
}

func (s *Server) finishConnection() {
	s.connMu.Lock()
	if s.activeConns > 0 {
		s.activeConns--
		if s.activeConns == 0 && s.connZero != nil {
			close(s.connZero)
			s.connZero = nil
		}
	}
	s.connMu.Unlock()
}

func (s *Server) serveConn(ctx context.Context, client net.Conn) {
	session, connCtx, ok := s.registry.register(ctx, client)
	if !ok {
		_ = client.Close()
		return
	}
	ctx = connCtx
	watchDone := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			session.close()
		case <-watchDone:
		}
	}()
	established := false
	defer func() {
		close(watchDone)
		if !established {
			s.registry.recordFailure()
		}
		s.registry.remove(session.ID)
		_ = client.Close()
	}()
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
	session.mu.Lock()
	session.Target = net.JoinHostPort(host, strconv.Itoa(int(port)))
	session.mu.Unlock()
	nodes, err := s.router.candidates(s.cfg)
	if err != nil {
		_ = writeReply(client, 0x01)
		return
	}
	var upstream net.Conn
	for _, node := range nodes {
		dialCtx, cancel := context.WithTimeout(ctx, time.Duration(s.cfg.DialTimeoutSeconds)*time.Second)
		upstream, err = s.dial(dialCtx, node, host, port)
		cancel()
		if err == nil {
			s.router.health.recordSuccess(node)
			session.setUpstream(upstream, node)
			break
		}
		s.router.health.recordFailureReason(node, time.Duration(s.cfg.FailureCooldownSeconds)*time.Second, s.sanitizeProbeError(node, err))
	}
	if upstream == nil {
		_ = writeReply(client, 0x01)
		return
	}
	defer func() { _ = upstream.Close() }()
	_ = client.SetDeadline(time.Time{})
	if err := writeReply(client, 0x00); err != nil {
		return
	}
	established = true
	s.registry.recordSuccess()
	pumpWithSession(ctx, session, client, upstream, s.cfg.IdleTimeoutSeconds, s.cfg.MaxConnectionDurationSeconds)
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
	usernameOK := subtle.ConstantTimeCompare(username, []byte(s.cfg.Username)) == 1
	passwordOK := subtle.ConstantTimeCompare(password, []byte(s.cfg.Password)) == 1
	if !usernameOK || !passwordOK {
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

func (s *Server) dialWithPool(ctx context.Context, node models.Node, host string, port uint16) (net.Conn, error) {
	adapter, lease, err := s.pool.acquire(node)
	if err != nil {
		return nil, err
	}
	conn, err := adapter.DialContext(ctx, &constant.Metadata{NetWork: constant.TCP, Type: constant.SOCKS5, Host: host, DstPort: port})
	if err != nil {
		lease.release(true)
		return nil, err
	}
	return &pooledConn{Conn: conn, lease: lease}, nil
}

func (s *Server) Close() {
	s.closeOnce.Do(func() {
		s.connMu.Lock()
		s.closing = true
		s.connMu.Unlock()
		s.stopHealthChecks()
		if s.registry != nil {
			s.registry.closeAll()
		}
		s.connMu.Lock()
		activeConnections := s.activeConns
		done := s.connZero
		s.connMu.Unlock()
		if activeConnections > 0 && done != nil {
			timer := time.NewTimer(5 * time.Second)
			select {
			case <-done:
				if !timer.Stop() {
					<-timer.C
				}
			case <-timer.C:
			}
		}
		if s.pool != nil {
			s.pool.close()
		}
		close(s.closed)
	})
	<-s.closed
}

func (s *Server) Connections() []ConnectionSnapshot { return s.registry.snapshot() }
func (s *Server) Stats() GatewaySnapshot            { return s.registry.snapshotStats() }

func isLoopbackListenAddress(address string) bool {
	host := strings.TrimSpace(strings.Trim(address, "[]"))
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// Manager owns the optional long-lived listener.
type Manager struct {
	mu       sync.Mutex
	listener net.Listener
	cancel   context.CancelFunc
	cfg      Config
	dial     DialFunc
	server   *Server
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
	var listenConfig net.ListenConfig
	listener, err := listenConfig.Listen(context.Background(), "tcp", net.JoinHostPort(normalized.ListenAddress, strconv.Itoa(normalized.Port)))
	if err != nil {
		return fmt.Errorf("listen SOCKS5: %w", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	server, err := NewServer(normalized, dial)
	if err != nil {
		_ = listener.Close()
		cancel()
		return err
	}
	m.mu.Lock()
	m.listener = listener
	m.cancel = cancel
	m.server = server
	m.mu.Unlock()
	go func() {
		if serveErr := server.Serve(ctx, listener); serveErr != nil {
			utils.Warn("SOCKS5 listener stopped: %v", serveErr)
		}
		server.Close()
		m.mu.Lock()
		if m.listener == listener {
			m.listener = nil
			m.cancel = nil
			m.server = nil
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
	server := m.server
	m.listener = nil
	m.cancel = nil
	m.server = nil
	m.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if listener != nil {
		_ = listener.Close()
	}
	if server != nil {
		server.Close()
	}
}

func (m *Manager) GatewaySnapshot() GatewaySnapshot {
	m.mu.Lock()
	server := m.server
	m.mu.Unlock()
	if server == nil {
		return GatewaySnapshot{}
	}
	return server.Stats()
}

func (m *Manager) Connections() []ConnectionSnapshot {
	m.mu.Lock()
	server := m.server
	m.mu.Unlock()
	if server == nil {
		return []ConnectionSnapshot{}
	}
	return server.Connections()
}

func (m *Manager) HealthSnapshot() HealthSnapshot {
	m.mu.Lock()
	server := m.server
	m.mu.Unlock()
	if server == nil {
		return HealthSnapshot{Nodes: []NodeHealthSnapshot{}}
	}
	return server.HealthSnapshot()
}

func (m *Manager) TriggerHealthProbe() (started bool, available bool) {
	m.mu.Lock()
	server := m.server
	m.mu.Unlock()
	if server == nil {
		return false, false
	}
	return server.triggerHealthProbe()
}

func (m *Manager) CloseConnection(id string) bool {
	m.mu.Lock()
	server := m.server
	m.mu.Unlock()
	return server != nil && server.registry.close(id)
}

func (m *Manager) CloseAllConnections() {
	m.mu.Lock()
	server := m.server
	m.mu.Unlock()
	if server != nil {
		server.registry.closeAll()
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
