package agents

import (
	"context"
	"log"

	"github.com/Trojan295/chinchilla/proto"
	"github.com/Trojan295/chinchilla/server"
	root "github.com/Trojan295/chinchilla/server"
)

// AgentServiceServer implements the gRPC AgentService server
type AgentServiceServer struct {
	AgentStore      server.AgentStore
	GameserverStore server.GameserverStore
}

// Register handles registration of a new agent
func (server AgentServiceServer) Register(ctx context.Context, agentState *proto.AgentState) (*proto.Empty, error) {
	log.Printf("agentServiceServer Register: register agent %s",
		agentState.Hostname)

	err := server.AgentStore.RegisterAgent(agentState)

	return &proto.Empty{}, err
}

// GetGameserverDeployments func
func (server AgentServiceServer) GetGameserverDeployments(ctx context.Context, req *proto.GetGameserverDeploymentsRequest) (*proto.GetGameserverDeploymentsResponse, error) {
	gameservers, err := root.GetGameserversForAgent(req.Hostname, server.GameserverStore)
	if err != nil {
		log.Printf("AgentServiceServer GetGameServers error: %v", err)
		return nil, err
	}

	runConfigs := make([]*proto.GameserverDeployment, 0, len(gameservers))
	for _, gs := range gameservers {
		runConfigs = append(runConfigs, gs.Deployment)
	}

	return &proto.GetGameserverDeploymentsResponse{
		Deployments: runConfigs,
	}, nil
}
