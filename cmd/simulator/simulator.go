package main

import (
	"context"
	"flag"
	"log"
	"math"
	"strings"
	"time"

	"github.com/Trojan295/chinchilla/proto"
	"google.golang.org/grpc"
)

func main() {
	hostname := flag.String("hostname", "", "Agent hostname")
	cpus := flag.Int("cpus", 8, "Agent CPUs count")
	memory := flag.Int("memory", 32, "")
	ipAddressStr := flag.String("ips", "", "IP addresses")
	flag.Parse()

	if *hostname == "" {
		panic("Missing -hostname")
	}

	if *ipAddressStr == "" {
		panic("Missing -ips")
	}

	ipAddresses := strings.Split(*ipAddressStr, ",")

	address := "localhost:10110"

	log.Printf("Running agent simulator")
	log.Printf("Hostname: %s", *hostname)
	log.Printf("CPUs: %d", *cpus)
	log.Printf("Memory: %d GB", *memory)
	log.Printf("Available IPs: %s", ipAddresses)

	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := proto.NewAgentServiceClient(conn)

	ctx := context.Background()

	var runningGameservers []*proto.Gameserver
	var usedMemory = 0

	for {

		agentState := &proto.AgentState{
			Hostname: *hostname,
			Resources: &proto.AgentResources{
				Cpus:        int32(*cpus),
				Memory:      int32(*memory * 1024 * 1024),
				IpAddresses: int32(len(ipAddresses)),
			},
			ResourceUsage: &proto.AgentResourceUsage{
				Memory: int32(usedMemory),
			},
			RunningGameservers: runningGameservers,
		}

		_, err := c.Register(ctx, agentState)
		if err != nil {
			log.Printf("Cannot update agent state: %s", err)
		}

		deployments, err := c.GetGameserverDeployments(ctx, &proto.GetGameserverDeploymentsRequest{
			Hostname: *hostname,
		})

		runningGameservers = make([]*proto.Gameserver, 0)
		usedMemory = 0
		for i, deployment := range deployments.Deployments {
			runningGameservers = append(runningGameservers, &proto.Gameserver{
				UUID:   deployment.UUID,
				Status: proto.GameserverStatus_RUNNING,
				Endpoint: &proto.Endpoint{
					IpAddress: ipAddresses[i%(len(ipAddresses))],
				},
			})
			usedMemory += int(math.Round(float64(deployment.ResourceRequirements.MemoryReservation) * 0.6))
		}

		time.Sleep(5 * time.Second)
	}
}
