package hub

import (
	crand "crypto/rand"
	"fmt"
	"net/url"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

const HubClusterName = "hubstream"

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
	// Generate a simple unique ID using timestamp and random bytes
	secret := [32]byte{}
	crand.Read(secret[:])
	id := fmt.Sprintf("hub-%d-%x", time.Now().UnixNano(), secret[:4])

	return &Options{
		Name:       id,
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
		StoreDir:              "./data",
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
