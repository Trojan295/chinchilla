package stores

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Trojan295/chinchilla/server"
	"go.etcd.io/etcd/client"
)

// EtcdStore is a etcd implementation of the AgentStore
type EtcdStore struct {
	keysAPI client.KeysAPI
}

// NewEtcdStore creates an EtcdAgentStore from a etcd.Config
func NewEtcdStore(config client.Config) (*EtcdStore, error) {
	etcdClient, err := client.New(config)
	if err != nil {
		return nil, err
	}

	kapi := client.NewKeysAPI(etcdClient)

	kapi.Set(context.Background(), "/gameservers", "", &client.SetOptions{
		Dir:       true,
		PrevExist: "false",
	})
	kapi.Set(context.Background(), "/agents", "", &client.SetOptions{
		Dir:       true,
		PrevExist: "false",
	})

	return &EtcdStore{
		keysAPI: kapi,
	}, nil
}

// RegisterAgent registers a new Agent
func (store *EtcdStore) RegisterAgent(agent *server.Agent) error {
	value, _ := json.Marshal(agent)

	_, err := store.keysAPI.Set(
		context.Background(),
		fmt.Sprintf("/agents/%s/state", agent.State.Hostname),
		string(value), nil,
	)

	return err
}

// ListAgents returns a AgentDetails list
func (store *EtcdStore) ListAgents() ([]server.Agent, error) {
	agents := make([]server.Agent, 0)

	agentsRes, err := store.keysAPI.Get(context.Background(), "/agents", nil)
	if err != nil {
		return agents, err
	}

	for _, agentNode := range agentsRes.Node.Nodes {
		detailsRes, _ := store.keysAPI.Get(context.Background(), fmt.Sprintf("%s/state", agentNode.Key), nil)
		var agentDetails server.Agent
		json.Unmarshal([]byte(detailsRes.Node.Value), &agentDetails)
		agents = append(agents, agentDetails)
	}
	return agents, err
}

// GetAgentState func
func (store *EtcdStore) GetAgent(UUID string) (*server.Agent, error) {
	agentStateRes, err := store.keysAPI.Get(context.Background(), fmt.Sprintf("/agents/%s/state", UUID), nil)
	if err != nil {
		return nil, err
	}

	var agentState server.Agent
	json.Unmarshal([]byte(agentStateRes.Node.Value), &agentState)
	return &agentState, nil
}

// ListGameservers returns a Gameserver list
func (store *EtcdStore) ListGameservers() ([]server.Gameserver, error) {
	gameservers := make([]server.Gameserver, 0)

	gsRes, err := store.keysAPI.Get(context.Background(), "/gameservers", nil)
	if err != nil {
		log.Println(err.Error())
		return gameservers, err
	}

	for _, gsNode := range gsRes.Node.Nodes {
		definitionRes, err := store.keysAPI.Get(context.Background(), gsNode.Key, nil)
		if err != nil {
			continue
		}

		gs := server.Gameserver{}
		json.Unmarshal([]byte(definitionRes.Node.Value), &gs)
		gameservers = append(gameservers, gs)
	}

	return gameservers, nil
}

// GetGameserver func
func (store *EtcdStore) GetGameserver(UUID string) (*server.Gameserver, error) {
	gsRes, err := store.keysAPI.Get(context.Background(), fmt.Sprintf("/gameservers/%s", UUID), nil)
	if err != nil {
		return nil, err
	}

	gs := &server.Gameserver{}
	json.Unmarshal([]byte(gsRes.Node.Value), gs)
	return gs, nil
}

// DeleteGameserver func
func (store *EtcdStore) DeleteGameserver(UUID string) error {
	_, err := store.keysAPI.Delete(context.Background(), fmt.Sprintf("/gameservers/%s", UUID), &client.DeleteOptions{
		Dir:       true,
		Recursive: true,
	})
	return err
}

// CreateGameserver func
func (store *EtcdStore) CreateGameserver(gs *server.Gameserver) error {
	gsData, _ := json.Marshal(*gs)
	_, err := store.keysAPI.Create(context.Background(), fmt.Sprintf("/gameservers/%s", gs.Definition.UUID), string(gsData))
	return err
}

// UpdateGameserver func
func (store *EtcdStore) UpdateGameserver(gs *server.Gameserver) error {
	gsData, _ := json.Marshal(*gs)
	_, err := store.keysAPI.Update(context.Background(), fmt.Sprintf("/gameservers/%s", gs.Definition.UUID), string(gsData))
	return err
}
