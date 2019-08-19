package gameservers

import (
	"log"
	"math/rand"
	"net/http"

	common "github.com/Trojan295/chinchilla-common"
	"github.com/Trojan295/chinchilla-server/server"
	"github.com/Trojan295/chinchilla-server/server/auth"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

type gameserversAPI struct {
	agentsStore       server.AgentStore
	gameserverStore   server.GameserverStore
	gameserverManager GameserverManager
}

// MountGameserverAPI func
func MountGameserverAPI(r *gin.Engine, agStore server.AgentStore, gsStore server.GameserverStore) {
	api := gameserversAPI{agStore, gsStore, NewGameserverManager()}

	group := r.Group("/gameservers/")
	group.OPTIONS("/", api.getSupportedGameservers)
	group.GET("/", auth.LoginRequired(), api.listGameservers)
	group.POST("/", auth.LoginRequired(), api.createGameserver)
	group.DELETE("/:uuid/", auth.LoginRequired(), api.deleteGameserver)
}

type gameserverCreate struct {
	Name       string            `json:"name" binding:"required"`
	Game       string            `json:"game" binding:"required"`
	Version    string            `json:"version" binding:"required"`
	Parameters map[string]string `json:"parameters" binding:"required"`
}

type gameserverStatusResponse struct {
	UUID    string `json:"uuid" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Game    string `json:"game" binding:"required"`
	Version string `json:"version" binding:"required"`
	Address string `json:"address"`
}

func (api *gameserversAPI) getSupportedGameservers(c *gin.Context) {
	metadata := api.gameserverManager.GetSupportedGameservers()
	c.JSON(http.StatusOK, metadata)
}

func (api *gameserversAPI) createGameserver(c *gin.Context) {
	var body gameserverCreate
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	owner := c.GetString("userID")
	uuid := uuid.NewV4().String()

	gs := server.Gameserver{
		Definition: server.GameserverDefinition{
			UUID:       uuid,
			Name:       body.Name,
			Owner:      owner,
			Game:       body.Game,
			Version:    body.Version,
			Parameters: body.Parameters,
		},
	}

	runConfig, err := api.gameserverManager.CreateRunConfiguration(&gs.Definition)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err})
		return
	}
	gs.RunConfiguration = runConfig

	agents, _ := api.agentsStore.ListAgents()
	idx := rand.Intn(len(agents))
	gs.RunConfiguration.Agent = agents[idx].Hostname

	api.gameserverStore.CreateGameserver(&gs)
	c.JSON(http.StatusAccepted, gin.H{"status": "Order accepted"})
}

func (api *gameserversAPI) listGameservers(c *gin.Context) {
	userID := c.GetString("userID")

	gameservers, err := api.gameserverStore.ListGameservers()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
	}

	resp := make([]*gameserverStatusResponse, 0)
	for _, gameserver := range gameservers {
		if gameserver.Definition.Owner != userID {
			continue
		}

		var gameserverState *common.GameserverState
		agentState, err := api.agentsStore.GetAgentState(gameserver.RunConfiguration.Agent)
		if err != nil {
			log.Printf("gameserversAPI GetAgentState error: %v", err)
		} else {
			for i := range agentState.GameServersState {
				if agentState.GameServersState[i].Uuid == gameserver.Definition.UUID {
					gameserverState = agentState.GameServersState[i]
				}
			}
		}

		address, _ := api.gameserverManager.Endpoint(&gameserver, gameserverState)
		resp = append(resp, &gameserverStatusResponse{
			Address: address,
			Game:    gameserver.Definition.Game,
			Name:    gameserver.Definition.Name,
			UUID:    gameserver.Definition.UUID,
			Version: gameserver.Definition.Version,
		})
	}

	c.JSON(http.StatusOK, resp)
}

func (api *gameserversAPI) deleteGameserver(c *gin.Context) {
	UUID := c.Param("uuid")
	err := api.gameserverStore.DeleteGameserver(UUID)
	if err != nil {
		log.Printf("gameserversAPI deleteGameserver error: %v", err)
		c.JSON(http.StatusServiceUnavailable, "")
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"status": "Deleting"})
}
