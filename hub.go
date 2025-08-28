package hub

import (
	crand "crypto/rand"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

const HubClusterName = "hubstream"
const HubIDFile = "hub.id"

// readHubID reads the hub ID from file, returns empty string if file doesn't exist
func readHubID(dataDir string) (string, error) {
	idFile := filepath.Join(dataDir, HubIDFile)
	if _, err := os.Stat(idFile); os.IsNotExist(err) {
		return "", nil // File doesn't exist, will create new ID
	}

	data, err := os.ReadFile(idFile)
	if err != nil {
		return "", fmt.Errorf("failed to read hub ID file: %w", err)
	}

	id := strings.TrimSpace(string(data))
	if id == "" {
		return "", fmt.Errorf("hub ID file is empty")
	}

	return id, nil
}

// writeHubID writes the hub ID to file
func writeHubID(dataDir, id string) error {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	idFile := filepath.Join(dataDir, HubIDFile)
	if err := os.WriteFile(idFile, []byte(id+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to write hub ID file: %w", err)
	}

	return nil
}

// generateHubID generates a new unique hub ID
func generateHubID() (string, error) {
	secret := [32]byte{}
	if _, err := crand.Read(secret[:]); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return fmt.Sprintf("hub-%d-%x", time.Now().UnixNano(), secret[:4]), nil
}

type Options struct {
	Name               string
	Host               string
	Port               int
	AuthorizationToken string
	MaxPayload         Size

	Routes []*url.URL

	ClusterHost         string
	ClusterPort         int
	ClusterUsername     string
	ClusterPassword     string
	ClusterConnPoolSize int
	ClusterPingInterval time.Duration

	JetstreamMaxMemory    Size
	JetstreamMaxStorage   Size
	StreamMaxBufferedMsgs int
	StreamMaxBufferedSize int64
	StoreDir              string
	SyncInterval          time.Duration
	SyncAlways            bool

	LogFile      string `json:"-"`
	LogSizeLimit int64  `json:"-"`
	LogMaxFiles  int64  `json:"-"`
	Syslog       bool   `json:"-"`
	RemoteSyslog string `json:"-"`
}

func DefaultOptions() (*Options, error) {
	// Try to read existing hub ID from file
	dataDir := "./data"
	existingID, err := readHubID(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read hub ID: %w", err)
	}

	var hubID string
	if existingID != "" {
		// Use existing ID from file
		hubID = existingID
	} else {
		// Generate new ID and save to file
		hubID, err = generateHubID()
		if err != nil {
			return nil, fmt.Errorf("failed to generate hub ID: %w", err)
		}

		if err := writeHubID(dataDir, hubID); err != nil {
			return nil, fmt.Errorf("failed to save hub ID: %w", err)
		}
	}

	return &Options{
		Name:       hubID,
		Host:       "0.0.0.0",
		Port:       4222,
		MaxPayload: NewSizeFromMegabytes(8),

		ClusterHost:           "0.0.0.0",
		ClusterPort:           6222,
		ClusterConnPoolSize:   64,
		ClusterPingInterval:   2 * time.Minute,
		JetstreamMaxMemory:    NewSizeFromMegabytes(512),
		JetstreamMaxStorage:   NewSizeFromGigabytes(10),
		StreamMaxBufferedMsgs: 65536,
		StreamMaxBufferedSize: 64 * 1024 * 1024,
		StoreDir:              dataDir,
		SyncInterval:          2 * time.Second,
		SyncAlways:            false,

		LogSizeLimit: 10 * 1024 * 1024,
		LogMaxFiles:  3,
		Syslog:       false,
		RemoteSyslog: "",
	}, nil
}

type Hub struct {
	options       *Options
	server        *server.Server
	inProcessConn *nats.Conn
	jetstreamCtx  nats.JetStreamContext
}

func NewHub(opt *Options) (*Hub, error) {
	if opt == nil {
		return nil, fmt.Errorf("options cannot be nil")
	}

	natsServerOpts := &server.Options{
		ServerName:    opt.Name,
		Host:          opt.Host,
		Port:          opt.Port,
		Authorization: opt.AuthorizationToken,
		MaxPayload:    int32(opt.MaxPayload.Bytes()),

		Routes: opt.Routes,

		NoLog:        opt.LogFile == "",
		LogFile:      opt.LogFile,
		LogSizeLimit: opt.LogSizeLimit,
		LogMaxFiles:  opt.LogMaxFiles,
		Syslog:       opt.Syslog,
		RemoteSyslog: opt.RemoteSyslog,

		Cluster: server.ClusterOpts{
			Name:         HubClusterName,
			Host:         opt.ClusterHost,
			Port:         opt.ClusterPort,
			Username:     opt.ClusterUsername,
			Password:     opt.ClusterPassword,
			PoolSize:     opt.ClusterConnPoolSize,
			PingInterval: opt.ClusterPingInterval,
		},

		JetStream:             true,
		StoreDir:              opt.StoreDir,
		JetStreamMaxMemory:    int64(opt.JetstreamMaxMemory.Bytes()),
		JetStreamMaxStore:     opt.JetstreamMaxStorage.Bytes(),
		StreamMaxBufferedMsgs: opt.StreamMaxBufferedMsgs,
		StreamMaxBufferedSize: opt.StreamMaxBufferedSize,
		SyncInterval:          opt.SyncInterval,
		SyncAlways:            opt.SyncAlways,
	}

	svr, err := server.NewServer(natsServerOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create NATS server: %w", err)
	}

	go svr.Start()

	// Increase timeout for server startup
	if !svr.ReadyForConnections(4 * time.Second) {
		return nil, fmt.Errorf("NATS server failed to start in time")
	}

	conn, err := nats.Connect(svr.ClientURL(), nats.InProcessServer(svr))
	if err != nil {
		return nil, fmt.Errorf("failed to create in-process NATS connection: %w", err)
	}

	js, err := conn.JetStream()
	if err != nil {
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	return &Hub{
		options:       opt,
		server:        svr,
		inProcessConn: conn,
		jetstreamCtx:  js,
	}, nil
}

func (h *Hub) Shutdown() {
	h.inProcessConn.Close()
	h.server.Shutdown()
}
