package server

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/snehamirajkar9/TCP-leastconnectionLB/balancer"
)

type Server interface {
	Start()
	Stop()
	GetAuthScheme() *AuthScheme
}

type serve struct {
	wg         sync.WaitGroup
	listener   net.Listener
	shutdown   chan struct{}
	connection chan *tls.Conn
	balancer   balancer.Balancer
	authScheme *AuthScheme
}

type Config struct {
	Address        string
	Upstreams      []string
	CaCertFile     string
	ServerCertFile string
	ServerKeyFile  string
}

func NewServer(config Config) (Server, error) {
	tlsConfig, err := getTlsConfig(config)
	if err != nil {
		return nil, err
	}
	log.Info().Msg("Starting server ...")
	listener, err := tls.Listen("tcp", config.Address, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on address %s: %w", config.Address, err)
	}
	// TODO hard coded burst and rate would need to go to a config
	balance := balancer.NewBalancer(balancer.Config{Burst: 20, Rate: 20, Upstreams: config.Upstreams})
	auth := NewAuthScheme()
	return &serve{
		listener:   listener,
		shutdown:   make(chan struct{}),
		connection: make(chan *tls.Conn, 100),
		balancer:   balance,
		authScheme: auth,
	}, nil
}

func getTlsConfig(c Config) (*tls.Config, error) {
	caCert, err := os.ReadFile(c.CaCertFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load ca certificate: %w", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	cer, err := tls.LoadX509KeyPair(c.ServerCertFile, c.ServerKeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %w", err)
	}

	cfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
		Certificates: []tls.Certificate{cer},
		ClientAuth:   tls.RequireAndVerifyClientCert, // set mutual tls
		ClientCAs:    caCertPool,
		RootCAs:      caCertPool,
	}
	cfg.Rand = rand.Reader

	return cfg, nil
}

func (s *serve) acceptConnections() {
	defer s.wg.Done()

	for {
		select {
		case <-s.shutdown:
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				continue
			}
			s.connection <- conn.(*tls.Conn)
		}
	}
}

func (s *serve) handleConnections() {
	defer s.wg.Done()

	for {
		select {
		case <-s.shutdown:
			return
		case conn := <-s.connection:
			go s.handleConnection(conn)
		}
	}
}

func (s *serve) handleConnection(conn *tls.Conn) {
	defer conn.Close()

	// need to do handshake first to get the certificates
	// that part is done when the first bytes are exchanged otherwise
	// stated in the tls package documentation
	// "Most uses of this package need not call HandshakeContext explicitly: the
	// first Read or Write will call it automatically."
	err := conn.Handshake()
	if err != nil {
		log.Error().Err(err).Msg("Failed to handshake")
		return
	}
	peersCerts := conn.ConnectionState().PeerCertificates
	if len(peersCerts) == 0 {
		log.Warn().Msg("No certificate found")
		return
	}
	clientId := conn.ConnectionState().PeerCertificates[0].Subject.CommonName
	allowed := s.authScheme.GetAllowedUpstreams(clientId)
	if allowed == nil {
		log.Warn().Str("client", clientId).Msg("Client not allowed")
		return
	}
	err = s.balancer.Balance(conn, clientId, allowed)
	if err != nil {
		if errors.Is(err, &balancer.RateLimiterError{}) {
			log.Warn().Err(err).Msg("Connection rate limited")
		}
		log.Error().Err(err).Msg("Error forwarding")
	}
}

// Start start accepting and handling connections.
func (s *serve) Start() {
	s.wg.Add(2)
	go s.acceptConnections()
	go s.handleConnections()
}

// Stop stop accepting and handling connections.
func (s *serve) Stop() {
	close(s.shutdown)
	s.balancer.Stop()
	s.listener.Close()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return
	case <-time.After(time.Second):
		log.Warn().Msg("Timed out waiting for connections to finish.")
		return
	}
}

// GetAuthScheme returns the auth scheme.
func (s *serve) GetAuthScheme() *AuthScheme {
	return s.authScheme
}
