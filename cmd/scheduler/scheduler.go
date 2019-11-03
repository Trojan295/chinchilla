package main

import (
	"errors"
	"log"
	"math/rand"
	"time"

	"github.com/Trojan295/chinchilla/common"
	"github.com/Trojan295/chinchilla/server"
	"github.com/Trojan295/chinchilla/server/stores"
	"go.etcd.io/etcd/client"
)

var version string

type agentInfo struct {
	hostname    string
	lastContact time.Time
	ipAddresses int
	totalMemory int
}

type SchedulerService struct {
	config          common.Scheduler
	gameserverStore server.GameserverStore
	agentStore      server.AgentStore
}

func (service *SchedulerService) getAllAgentInfo() ([]agentInfo, error) {
	agentInfos := make([]agentInfo, 0)

	agents, err := service.agentStore.ListAgents()
	if err != nil {
		return agentInfos, err
	}

	for _, agent := range agents {
		agentInfos = append(agentInfos, agentInfo{
			hostname:    agent.State.Hostname,
			lastContact: agent.LastContact,
			ipAddresses: int(agent.State.Resources.IpAddresses),
			totalMemory: int(agent.State.Resources.Memory),
		})
	}

	return agentInfos, nil
}

func (service *SchedulerService) assignAgent(gameserver *server.Gameserver) error {
	agentsInfo, err := service.getAllAgentInfo()
	if err != nil {
		return err
	}

	possibleAgents := make([]agentInfo, 0)

	for _, agent := range agentsInfo {
		agentGss, err := server.GetGameserversForAgent(agent.hostname, service.gameserverStore)
		if err != nil {
			continue
		}

		usedIPs := 0
		memoryReservation := 0

		for _, agentGs := range agentGss {
			memoryReservation += int(agentGs.Deployment.ResourceRequirements.MemoryReservation)

			if agentGs.Definition.Game == gameserver.Definition.Game {
				usedIPs++
			}
		}

		if time.Now().Sub(agent.lastContact).Seconds() > float64(service.config.AgentContactDelay) {
			continue
		}

		if agent.totalMemory-memoryReservation-int(gameserver.Deployment.ResourceRequirements.MemoryReservation) <= 0 {
			continue
		}

		if agent.ipAddresses-usedIPs <= 0 {
			continue
		}

		possibleAgents = append(possibleAgents, agent)
	}

	if len(possibleAgents) == 0 {
		return errors.New("No free agents to assign")
	}

	idx := rand.Intn(len(possibleAgents))
	agent := possibleAgents[idx]
	gameserver.Deployment.Agent = agent.hostname
	return service.gameserverStore.UpdateGameserver(gameserver)
}

func (service *SchedulerService) Tick() error {
	gameservers, err := service.gameserverStore.ListGameservers()
	if err != nil {
		return err
	}

	for _, gameserver := range gameservers {
		if gameserver.Deployment.Agent != "" {
			continue
		}

		log.Printf("Scheduling gameserver %s...", gameserver.Definition.Name)
		if err := service.assignAgent(&gameserver); err != nil {
			log.Printf("ERROR Failed to schedule %s: %s", gameserver.Definition.UUID, err.Error())
		} else {
			log.Printf("Scheduled gameserver %s to %s", gameserver.Definition.Name, gameserver.Deployment.Agent)
		}
	}

	return nil
}

func main() {
	log.Printf("Chinchilla scheduler v%s\n", version)

	config, err := common.LoadConfig()
	if err != nil {
		panic(err)
	}

	log.Printf("Tick interval: %ds", config.Scheduler.Interval)
	log.Printf("Agent contact delay: %ds", config.Scheduler.AgentContactDelay)

	cfg := client.Config{
		Endpoints:               []string{config.Etcd.Address},
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}

	etcdStore, err := stores.NewEtcdStore(cfg)
	if err != nil {
		panic(err)
	}

	service := SchedulerService{
		config:          config.Scheduler,
		gameserverStore: etcdStore,
		agentStore:      etcdStore,
	}

	for {
		err := service.Tick()
		if err != nil {
			log.Printf("Error in tick: %s", err.Error())
		}

		time.Sleep(time.Duration(config.Scheduler.Interval) * time.Second)
	}

}
