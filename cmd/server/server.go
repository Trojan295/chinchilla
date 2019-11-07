package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/Trojan295/chinchilla/common"
	"github.com/Trojan295/chinchilla/proto"
	"github.com/Trojan295/chinchilla/server"
	"github.com/Trojan295/chinchilla/server/agents"
	"github.com/Trojan295/chinchilla/server/api"
	"github.com/Trojan295/chinchilla/server/auth"
	"github.com/Trojan295/chinchilla/server/stores"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.etcd.io/etcd/client"
	"google.golang.org/grpc"
)

// NewAgentServiceServer constructor
func NewAgentServiceServer(etcdStore *stores.EtcdStore) agents.AgentServiceServer {
	return agents.AgentServiceServer{
		AgentStore:      etcdStore,
		GameserverStore: etcdStore,
	}
}

func runGrpcServer(config *common.Configuration, etcdStore *stores.EtcdStore) {
	port := fmt.Sprintf(":%d", config.Server.Port)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	proto.RegisterAgentServiceServer(s, NewAgentServiceServer(etcdStore))

	log.Printf("Listening for gRPC on %s\n", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func setupRouter(r *gin.Engine, etcd *stores.EtcdStore, elastic *stores.ElasticsearchStore) {
	r.GET("/health/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	agents.MountAgentsAPI(r, etcd, etcd)
	api.MountGameserverAPI(r, etcd, etcd, elastic)
}

var version string

func main() {
	log.Printf("Chinchilla server v%s\n", version)

	config, err := common.LoadConfig()
	if err != nil {
		panic(err)
	}

	cfg := client.Config{
		Endpoints:               []string{config.Etcd.Address},
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}
	etcdStore, err := stores.NewEtcdStore(cfg)
	if err != nil {
		panic(err)
	}

	elasticsearchStore := stores.NewElasticsearchStore(config.Elasticsearch.Address)

	go runGrpcServer(config, etcdStore)
	server.StartMetrics(etcdStore)

	r := gin.Default()
	auth.SetupAuthentication(r, config.Auth)
	setupRouter(r, etcdStore, &elasticsearchStore)
	r.Run(":8081")
}
