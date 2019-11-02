package gameservers

import (
	"log"
	"net/http"

	"github.com/Trojan295/chinchilla/server"
	"github.com/Trojan295/chinchilla/server/auth"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

type supportedGameserversResponse []supportedGameserverResponse

type supportedGameserverOptions struct {
	Version    string            `json:"version"`
	Parameters map[string]string `json:"parameters"`
}

type supportedGameserverResponse struct {
	Name    string                       `json:"name"`
	Options []supportedGameserverOptions `json:"options"`
}

type getGameserverResponse struct {
	UUID    string  `json:"uuid"`
	Name    string  `json:"name"`
	Game    string  `json:"game"`
	Version string  `json:"version"`
	Address *string `json:"address"`
	Status  string  `json:"status"`
}

type listGameserversResponse []getGameserverResponse

type createGameserverRequest struct {
	Name       string            `json:"name" binding:"required"`
	Game       string            `json:"game" binding:"required"`
	Version    string            `json:"version" binding:"required"`
	Parameters map[string]string `json:"parameters" binding:"required"`
}

type createGameserverResponse getGameserverResponse

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

func (api *gameserversAPI) getSupportedGameservers(c *gin.Context) {
	metadata := api.gameserverManager.GetSupportedGameservers()

	res := supportedGameserversResponse{}
	for _, gs := range metadata {
		optionsRes := make([]supportedGameserverOptions, 0)
		for _, option := range gs.Options {
			optionsRes = append(optionsRes, supportedGameserverOptions{
				Version:    option.Version,
				Parameters: option.Parameters,
			})
		}

		res = append(res, supportedGameserverResponse{
			Name:    gs.Name,
			Options: optionsRes,
		})
	}

	c.JSON(http.StatusOK, res)
}

func (api *gameserversAPI) createGameserver(c *gin.Context) {
	var body createGameserverRequest
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

	deployment, err := api.gameserverManager.CreateGameserverDeployment(&gs.Definition)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err})
		return
	}

	gs.Deployment = deployment
	api.gameserverStore.CreateGameserver(&gs)

	response := createGameserverResponse{
		UUID:    gs.Definition.UUID,
		Name:    gs.Definition.Name,
		Game:    gs.Definition.Game,
		Status:  "UNKNOWN",
		Version: gs.Definition.Version,
	}

	c.JSON(http.StatusAccepted, response)
}

func (api *gameserversAPI) listGameservers(c *gin.Context) {
	userID := c.GetString("userID")

	gameservers, err := api.gameserverStore.ListGameservers()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
	}

	resp := listGameserversResponse{}
	for _, gameserver := range gameservers {
		if gameserver.Definition.Owner != userID {
			continue
		}

		agentState, err := api.agentsStore.GetAgentState(gameserver.Deployment.Agent)

		var address string
		status := "UNKNOWN"
		if err != nil {
			log.Printf("gameserversAPI GetAgentState error: %v", err)
		} else {
			for _, agentServer := range agentState.RunningGameservers {
				if agentServer.UUID == gameserver.Definition.UUID {
					status = string(agentServer.Status.String())
					address, _ = api.gameserverManager.Endpoint(&gameserver, agentServer)
				}
			}
		}

		resp = append(resp, getGameserverResponse{
			UUID:    gameserver.Definition.UUID,
			Name:    gameserver.Definition.Name,
			Game:    gameserver.Definition.Game,
			Version: gameserver.Definition.Version,
			Address: &address,
			Status:  status,
		})

	}

	c.JSON(http.StatusOK, resp)
}

func (api *gameserversAPI) deleteGameserver(c *gin.Context) {
	UUID := c.Param("uuid")
	userID := c.GetString("userID")

	gameserver, err := api.gameserverStore.GetGameserver(UUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	if gameserver.Definition.Owner != userID {
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	err = api.gameserverStore.DeleteGameserver(UUID)
	if err != nil {
		log.Printf("gameserversAPI deleteGameserver error: %v", err)
		c.JSON(http.StatusServiceUnavailable, "")
		return
	}
	c.JSON(http.StatusAccepted, gin.H{})
}
