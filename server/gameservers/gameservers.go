package gameservers

import (
	"errors"

	common "github.com/Trojan295/chinchilla-common"
	"github.com/Trojan295/chinchilla-server/server"
)

type GameserverMetadata struct {
	Name       string
	Versions   []string
	Parameters map[string]string
}

type gameserverManager interface {
	metadata() GameserverMetadata

	endpoint(*server.Gameserver, *common.Gameserver) (string, error)
	createRunConfiguration(*server.GameserverDefinition) (*common.GameserverRunConfiguration, error)
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

func (handler *GameserverManager) CreateRunConfiguration(definition *server.GameserverDefinition) (*common.GameserverRunConfiguration, error) {
	for _, manager := range handler.managers {
		if manager.metadata().Name == definition.Game {
			return manager.createRunConfiguration(definition)
		}
	}

	return nil, errors.New("Not supported game type")
}

func (handler *GameserverManager) Endpoint(gameserver *server.Gameserver, state *common.Gameserver) (string, error) {
	for _, manager := range handler.managers {
		if manager.metadata().Name == gameserver.Definition.Game {
			return manager.endpoint(gameserver, state)
		}
	}
	return "", errors.New("Not supported game type")
}
