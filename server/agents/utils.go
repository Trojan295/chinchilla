package agents

import (
	"github.com/Trojan295/chinchilla-server/server"
)

func getGameserversForAgent(agentHostname string, store server.GameserverStore) ([]server.Gameserver, error) {
	gameservers, err := store.ListGameservers()
	if err != nil {
		return nil, err
	}

	agentGameservers := make([]server.Gameserver, 0)
	for _, gs := range gameservers {
		if gs.RunConfiguration.Agent == agentHostname {
			agentGameservers = append(agentGameservers, gs)
		}
	}
	return agentGameservers, nil
}
