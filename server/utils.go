package server

// GetGameserversForAgent func
func GetGameserversForAgent(agentHostname string, store GameserverStore) ([]Gameserver, error) {
	gameservers, err := store.ListGameservers()
	if err != nil {
		return nil, err
	}

	agentGameservers := make([]Gameserver, 0)
	for _, gs := range gameservers {
		if gs.Deployment.Agent == agentHostname {
			agentGameservers = append(agentGameservers, gs)
		}
	}
	return agentGameservers, nil
}
