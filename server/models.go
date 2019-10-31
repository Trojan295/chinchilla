package server

import "github.com/Trojan295/chinchilla-server/proto"

// GameserverDefinition represents the receipe for the game server
type GameserverDefinition struct {
	UUID       string
	Name       string
	Owner      string
	Game       string
	Version    string
	Parameters map[string]string
}

// Gameserver glues GameserverDefinition and GameserverDeployment
type Gameserver struct {
	Definition GameserverDefinition
	Deployment *proto.GameserverDeployment
}

// AgentStore is an interface for an agents storage
type AgentStore interface {
	RegisterAgent(*proto.AgentState) error
	ListAgents() ([]proto.AgentState, error)
	GetAgentState(UUID string) (*proto.AgentState, error)
}

// GameserverStore interface
type GameserverStore interface {
	CreateGameserver(*Gameserver) error
	ListGameservers() ([]Gameserver, error)
	GetGameserver(UUID string) (*Gameserver, error)
	DeleteGameserver(UUID string) error
}
