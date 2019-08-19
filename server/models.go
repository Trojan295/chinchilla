package server

import common "github.com/Trojan295/chinchilla-common"

// GameserverDefinition represents the receipe for the game server
type GameserverDefinition struct {
	UUID       string
	Name       string
	Owner      string
	Game       string
	Version    string
	Parameters map[string]string
}

// Gameserver glues GameserverDefinition and GameserverRunConfiguration
type Gameserver struct {
	Definition       GameserverDefinition
	RunConfiguration *common.GameserverRunConfiguration
}

// AgentStore is an interface for an agents storage
type AgentStore interface {
	RegisterAgent(*common.AgentState) error
	ListAgents() ([]common.AgentState, error)
	GetAgentState(UUID string) (*common.AgentState, error)
}

// GameserverStore interface
type GameserverStore interface {
	CreateGameserver(*Gameserver) error
	ListGameservers() ([]Gameserver, error)
	GetGameserver(UUID string) (*Gameserver, error)
	DeleteGameserver(UUID string) error
}
