package agents

import (
	"context"
	"log"

	"github.com/Trojan295/chinchilla-server/proto"
	"github.com/Trojan295/chinchilla-server/server"
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

// GetGameServers func
func (server AgentServiceServer) GetGameServers(ctx context.Context, req *proto.GetGameserversRequest) (*proto.GameserverRunConfigurationList, error) {
	gameservers, err := getGameserversForAgent(req.Hostname, server.GameserverStore)
	if err != nil {
		log.Printf("AgentServiceServer GetGameServers error: %v", err)
		return nil, err
	}

	runConfigs := make([]*proto.GameserverRunConfiguration, 0, len(gameservers))
	for _, gs := range gameservers {
		runConfigs = append(runConfigs, gs.RunConfiguration)
	}

	return &proto.GameserverRunConfigurationList{
		Servers: runConfigs,
	}, nil
}
