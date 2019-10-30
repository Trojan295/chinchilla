package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/Trojan295/chinchilla-server/proto"
	"github.com/Trojan295/chinchilla-server/server"
	"github.com/Trojan295/chinchilla-server/server/agents"
	"github.com/Trojan295/chinchilla-server/server/auth"
	"github.com/Trojan295/chinchilla-server/server/gameservers"
	"github.com/Trojan295/chinchilla-server/server/stores"
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

func runGrpcServer(config *server.Configuration, etcdStore *stores.EtcdStore) {
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

func setupRouter(etcd *stores.EtcdStore) *gin.Engine {
	secret, ok := os.LookupEnv("AUTH0_KEY")
	if !ok {
		panic("Missing AUTH0_KEY variable")
	}

	r := gin.Default()
	r.Use(auth.JWTToken(secret))

	r.GET("/health/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	agents.MountAgentsAPI(r, etcd, etcd)
	gameservers.MountGameserverAPI(r, etcd, etcd)

	return r
}

var version string

func main() {
	log.Printf("Chinchilla server v%s\n", version)

	config, err := server.LoadConfig()
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
	go runGrpcServer(config, etcdStore)
	server.StartMetrics(etcdStore)

	r := setupRouter(etcdStore)

	r.Run(":8080")
}
