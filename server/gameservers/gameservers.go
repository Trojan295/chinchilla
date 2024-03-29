package gameservers

import (
	"errors"

	"github.com/Trojan295/chinchilla/proto"
	"github.com/Trojan295/chinchilla/server"
)

// GameserverOptions describes a deployable gameserver option
type GameserverOptions struct {
	Version    string
	Parameters map[string]string
}

// GameserverMetadata describes a gameserver
type GameserverMetadata struct {
	Name    string
	Options []GameserverOptions
}

type gameserverManager interface {
	metadata() GameserverMetadata
	createDeployment(*server.GameserverDefinition) (*proto.GameserverDeployment, error)
}

// GameserverManager struct
type GameserverManager struct {
	managers []gameserverManager
}

// NewGameserverManager creates a new GameserverManager
func NewGameserverManager() GameserverManager {
	return GameserverManager{
		managers: []gameserverManager{
			minecraftGameserverManager{},
			factorioGameserverManager{},
			teamspeakGameserverManager{},
		},
	}
}

// GetSupportedGameservers returns GameserverMetadata for all registered gameservers
func (handler *GameserverManager) GetSupportedGameservers() []GameserverMetadata {
	metadata := make([]GameserverMetadata, 0)
	for _, manager := range handler.managers {
		metadata = append(metadata, manager.metadata())
	}
	return metadata
}

// CreateGameserverDeployment func
func (handler *GameserverManager) CreateGameserverDeployment(definition *server.GameserverDefinition) (*proto.GameserverDeployment, error) {
	for _, manager := range handler.managers {
		if manager.metadata().Name == definition.Game {
			return manager.createDeployment(definition)
		}
	}

	return nil, errors.New("Not supported game type")
}

// Endpoint returns the endpoint for the gameserver
func (handler *GameserverManager) Endpoint(gameserver *server.Gameserver, state *proto.Gameserver) (string, error) {
	if state.Endpoint == nil {
		return "", errors.New("Endpoint not ready")
	}

	return state.Endpoint.IpAddress, nil
}
