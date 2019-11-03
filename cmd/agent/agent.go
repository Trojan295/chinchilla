package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Trojan295/chinchilla/agent"
	"github.com/Trojan295/chinchilla/common"
	"github.com/Trojan295/chinchilla/proto"
	"github.com/docker/docker/client"
	"google.golang.org/grpc"
)

const osMemoryReservation = 512

func getAgentState(hostname string) *proto.AgentState {
	cpuStats, err := agent.GetCPUStats()
	if err != nil {
		log.Fatalf("cannot get cpuStats: %s\n", err)
		return nil
	}

	memoryStats, err := agent.GetMemoryStats()
	if err != nil {
		log.Fatalf("cannot get memoryStats: %s\n", err)
		return nil
	}

	totalAvailableMemory := memoryStats.Total - osMemoryReservation
	if totalAvailableMemory < 0 {
		totalAvailableMemory = 0
	}

	availableMemory := memoryStats.Total - memoryStats.Available - osMemoryReservation
	if availableMemory < 0 {
		availableMemory = 0
	}

	return &proto.AgentState{
		Hostname: hostname,
		Resources: &proto.AgentResources{
			Cpus:   int64(cpuStats.Cores),
			Memory: int64(totalAvailableMemory),
		},
		ResourceUsage: &proto.AgentResourceUsage{
			Memory: int64(availableMemory),
		},
	}
}

var version string

func main() {
	log.Printf("Chinchilla agent v%s\n", version)

	config, err := common.LoadConfig()
	if err != nil {
		panic(err)
	}

	serverAddress := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	ipsFlag := flag.String("ips", "", "")
	flag.Parse()

	if *ipsFlag == "" {
		panic("Set -ips flag")
	}

	ipAddresses := strings.Split(*ipsFlag, ",")

	log.Printf("Connecting to server at %s", serverAddress)
	log.Printf("Using hostname: %s", hostname)

	conn, err := grpc.Dial(serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := proto.NewAgentServiceClient(conn)

	ctx := context.Background()

	docker, _ := client.NewEnvClient()
	manager := agent.NewGameserverManager(docker, docker, ipAddresses)

	for {
		gameservers, err := manager.GetGameservers()
		if err != nil {
			log.Fatalf("Failed to get game servers")
			continue
		}

		agentState := getAgentState(hostname)
		agentState.Resources.IpAddresses = int64(len(ipAddresses))
		agentState.RunningGameservers = gameservers

		c.Register(ctx, agentState)

		targetConfig, err := c.GetGameserverDeployments(ctx, &proto.GetGameserverDeploymentsRequest{
			Hostname: hostname,
		})
		if err != nil {
			log.Fatalf("Failed to get game servers: %s", err)
			continue
		}

		manager.Tick(targetConfig)

		time.Sleep(5 * time.Second)
	}
}
