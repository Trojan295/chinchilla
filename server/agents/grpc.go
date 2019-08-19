package agents

import (
	"context"
	"log"

	common "github.com/Trojan295/chinchilla-common"
	"github.com/Trojan295/chinchilla-server/server"
)

// AgentServiceServer implements the gRPC AgentService server
type AgentServiceServer struct {
	AgentStore      server.AgentStore
	GameserverStore server.GameserverStore
}

// Register handles registration of a new agent
func (server AgentServiceServer) Register(ctx context.Context, agentState *common.AgentState) (*common.Empty, error) {
	log.Printf("agentServiceServer Register: register agent %s",
		agentState.Hostname)

	err := server.AgentStore.RegisterAgent(agentState)

	return &common.Empty{}, err
}

// GetGameServers func
func (server AgentServiceServer) GetGameServers(ctx context.Context, req *common.GetGameserversRequest) (*common.GameserverList, error) {
	gameservers, err := getGameserversForAgent(req.Hostname, server.GameserverStore)
	if err != nil {
		log.Printf("AgentServiceServer GetGameServers error: %v", err)
		return nil, err
	}

	runConfigs := make([]*common.GameserverRunConfiguration, 0, len(gameservers))
	for _, gs := range gameservers {
		runConfigs = append(runConfigs, gs.RunConfiguration)
	}

	return &common.GameserverList{
		Servers: runConfigs,
	}, nil
}
