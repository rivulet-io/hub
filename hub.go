package hub

import (
	crand "crypto/rand"
	"fmt"
	"math"
	"math/rand/v2"
	"net/url"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"gosuda.org/randflake"
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
	secret := [32]byte{}
	crand.Read(secret[:])
	generator, err := randflake.NewGenerator(rand.Int64(), 0, math.MaxInt64, secret[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create ID generator: %w", err)
	}
	id, err := generator.GenerateString()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ID: %w", err)
	}
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
		StoreDir:              "data",
		SyncInterval:          2 * time.Second,
		SyncAlways:            false,

		LogSizeLimit: 10 * 1024 * 1024,
		LogMaxFiles:  3,
		Syslog:       false,
		RemoteSyslog: "",
	}, nil
}

type Hub struct {
	server        *server.Server
	inProcessConn *nats.Conn
}

func NewHub(opt *Options) (*Hub, error) {
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

	conn, err := nats.Connect(svr.ClientURL(), nats.InProcessServer(svr))
	if err != nil {
		return nil, fmt.Errorf("failed to create in-process NATS connection: %w", err)
	}

	return &Hub{
		server:        svr,
		inProcessConn: conn,
	}, nil
}
