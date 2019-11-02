package server

import (
	"time"

	"github.com/Trojan295/chinchilla/proto"
)

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

type Agent struct {
	LastContact time.Time
	State       proto.AgentState
}

// AgentStore is an interface for an agents storage
type AgentStore interface {
	RegisterAgent(*Agent) error
	ListAgents() ([]Agent, error)
	GetAgent(UUID string) (*Agent, error)
}

// GameserverStore interface
type GameserverStore interface {
	CreateGameserver(*Gameserver) error
	UpdateGameserver(*Gameserver) error
	ListGameservers() ([]Gameserver, error)
	GetGameserver(UUID string) (*Gameserver, error)
	DeleteGameserver(UUID string) error
}
