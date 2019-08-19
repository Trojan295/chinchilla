package agents

import (
	"log"
	"net/http"

	"github.com/Trojan295/chinchilla-server/server"
	"github.com/Trojan295/chinchilla-server/server/auth"
	"github.com/gin-gonic/gin"
)

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

	response := make([]gin.H, 0)

	for _, agent := range agents {
		var reservedMemory *int
		gameservers, err := getGameserversForAgent(agent.Hostname, api.gameserverStore)
		if err == nil {
			acc := 0
			for _, gs := range gameservers {
				acc += int(gs.RunConfiguration.ResourceRequirements.MemoryReservation)
			}
			reservedMemory = &acc
		}

		response = append(response, gin.H{
			"hostname": agent.Hostname,
			"resources": gin.H{
				"cpus":   agent.Resources.Cpus,
				"memory": agent.Resources.Memory,
			},
			"usedResources": gin.H{
				"memory": agent.ResourceUsage.Memory,
			},
			"reservedResources": gin.H{
				"memory": reservedMemory,
			},
		})
	}

	c.JSON(http.StatusOK, response)
}
