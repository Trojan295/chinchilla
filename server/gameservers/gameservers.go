package gameservers

import (
	"errors"

	"github.com/Trojan295/chinchilla-server/proto"
	"github.com/Trojan295/chinchilla-server/server"
)

type GameserverOptions struct {
	Version    string
	Parameters map[string]string
}

type GameserverMetadata struct {
	Name    string
	Options []GameserverOptions
}

type gameserverManager interface {
	metadata() GameserverMetadata

	endpoint(*server.Gameserver, *proto.Gameserver) (string, error)
	createRunConfiguration(*server.GameserverDefinition) (*proto.GameserverRunConfiguration, error)
}

type GameserverManager struct {
	managers []gameserverManager
}

func NewGameserverManager() GameserverManager {
	return GameserverManager{
		managers: []gameserverManager{minecraftGameserverManager{}, factorioGameserverManager{}},
	}
}

func (handler *GameserverManager) GetSupportedGameservers() []GameserverMetadata {
	metadata := make([]GameserverMetadata, 0)
	for _, manager := range handler.managers {
		metadata = append(metadata, manager.metadata())
	}
	return metadata
}

func (handler *GameserverManager) CreateRunConfiguration(definition *server.GameserverDefinition) (*proto.GameserverRunConfiguration, error) {
	for _, manager := range handler.managers {
		if manager.metadata().Name == definition.Game {
			return manager.createRunConfiguration(definition)
		}
	}

	return nil, errors.New("Not supported game type")
}

func (handler *GameserverManager) Endpoint(gameserver *server.Gameserver, state *proto.Gameserver) (string, error) {
	for _, manager := range handler.managers {
		if manager.metadata().Name == gameserver.Definition.Game {
			return manager.endpoint(gameserver, state)
		}
	}
	return "", errors.New("Not supported game type")
}
