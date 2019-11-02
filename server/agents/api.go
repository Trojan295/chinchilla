package agents

import (
	"log"
	"net/http"

	"github.com/Trojan295/chinchilla/server"
	"github.com/Trojan295/chinchilla/server/auth"
	"github.com/gin-gonic/gin"
)

type agentResources struct {
	Cpus        int `json:"cpus"`
	Memory      int `json:"memory"`
	IPAddresses int `json:"ipAddresses"`
}

type agentUsedResources struct {
	Memory int `json:"memory"`
}

type agentReservedResources struct {
	Memory int `json:"memory"`
}

type getAgentReponse struct {
	Hostname          string                  `json:"hostname"`
	Resources         *agentResources         `json:"resources"`
	UsedResources     *agentUsedResources     `json:"usedResources"`
	ReservedResources *agentReservedResources `json:"reservedResources"`
}

type listAgentsResponse []getAgentReponse

type agentsAPI struct {
	agentsStore     server.AgentStore
	gameserverStore server.GameserverStore
}

// MountAgentsAPI mounts the Agents API
func MountAgentsAPI(r *gin.Engine, agentsStore server.AgentStore, gameserverStore server.GameserverStore) {
	api := agentsAPI{
		agentsStore,
		gameserverStore,
	}

	group := r.Group("/agents/")
	group.GET("/", auth.Auth0Permission("read:agents"), api.getAgents)
}

func (api *agentsAPI) getAgents(c *gin.Context) {
	agents, err := api.agentsStore.ListAgents()
	if err != nil {
		log.Printf("AgentAPI getAgents error: %s", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Cannot list agents"})
	}

	var response listAgentsResponse

	for _, agent := range agents {
		var reservedMemory *int
		gameservers, err := server.GetGameserversForAgent(agent.State.Hostname, api.gameserverStore)
		if err == nil {
			acc := 0
			for _, gs := range gameservers {
				acc += int(gs.Deployment.ResourceRequirements.MemoryReservation)
			}
			reservedMemory = &acc
		}

		response = append(response, getAgentReponse{
			Hostname: agent.State.Hostname,
			Resources: &agentResources{
				Cpus:        int(agent.State.Resources.Cpus),
				Memory:      int(agent.State.Resources.Memory),
				IPAddresses: int(agent.State.Resources.IpAddresses),
			},
			UsedResources: &agentUsedResources{
				Memory: int(agent.State.ResourceUsage.Memory),
			},
			ReservedResources: &agentReservedResources{
				Memory: *reservedMemory,
			},
		})
	}

	c.JSON(http.StatusOK, response)
}
