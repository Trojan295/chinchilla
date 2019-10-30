package gameservers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Trojan295/chinchilla-server/mocks"
	"github.com/Trojan295/chinchilla-server/proto"
	"github.com/Trojan295/chinchilla-server/server"
	"github.com/Trojan295/chinchilla-server/server/utils"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestListAvailableGameservers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	agentStore := mocks.NewMockAgentStore(ctrl)
	gameserverStore := mocks.NewMockGameserverStore(ctrl)

	router := utils.SetupRouter()
	MountGameserverAPI(router, agentStore, gameserverStore)

	claims := map[string]interface{}{}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/gameservers/", nil)
	req.Header.Add("authorization", "Bearer "+utils.BuildToken(claims))

	router.ServeHTTP(w, req)

	res := supportedGameserversResponse{}
	json.Unmarshal(w.Body.Bytes(), &res)
	assert.Len(t, res, 2)
	assert.Equal(t, "Minecraft", res[0].Name)
	assert.Equal(t, "Factorio", res[1].Name)

	fmt.Println(w.Body.String())

	assert.Equal(t, 200, w.Code)
}

func TestListUserGameservers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gameserver := server.Gameserver{
		RunConfiguration: &proto.GameserverRunConfiguration{
			Agent: "localhost",
		},
		Definition: server.GameserverDefinition{
			UUID:    "someuuid",
			Name:    "my server",
			Game:    "Minecraft",
			Version: "1.12",
			Owner:   "user1",
		},
	}
	otherGameserver := server.Gameserver{
		RunConfiguration: &proto.GameserverRunConfiguration{
			Agent: "localhost",
		},
		Definition: server.GameserverDefinition{
			UUID:    "otheruuid",
			Name:    "my server",
			Game:    "Minecraft",
			Version: "1.12",
			Owner:   "user2",
		},
	}

	gameserverInstance := &proto.Gameserver{
		UUID:   gameserver.Definition.UUID,
		Status: proto.GameserverStatus_RUNNING,
		PortMappings: []*proto.NetworkPortMapping{
			&proto.NetworkPortMapping{
				Protocol:      proto.NetworkProtocol_TCP,
				ContainerPort: 25565,
				HostPort:      25565,
			},
		},
	}

	agentStore := mocks.NewMockAgentStore(ctrl)
	agentStore.EXPECT().
		GetAgentState("localhost").
		Return(&proto.AgentState{
			RunningGameservers: []*proto.Gameserver{gameserverInstance},
		}, nil).
		AnyTimes()

	gameserverStore := mocks.NewMockGameserverStore(ctrl)
	gameserverStore.EXPECT().
		ListGameservers().
		Return([]server.Gameserver{gameserver, otherGameserver}, nil).
		AnyTimes()

	router := utils.SetupRouter()
	MountGameserverAPI(router, agentStore, gameserverStore)

	claims := map[string]interface{}{
		"sub": "user1",
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/gameservers/", nil)
	req.Header.Add("authorization", "Bearer "+utils.BuildToken(claims))

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	res := listGameserversResponse{}
	json.Unmarshal(w.Body.Bytes(), &res)
	assert.Len(t, res, 1)
	assert.Equal(t, "someuuid", res[0].UUID)
	assert.Equal(t, "my server", res[0].Name)
	assert.Equal(t, "Minecraft", res[0].Game)
	assert.Equal(t, "1.12", res[0].Version)
	assert.Equal(t, "localhost:25565", *res[0].Address)
	assert.Equal(t, "RUNNING", res[0].Status)
}

func TestCreateNewServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	agentStore := mocks.NewMockAgentStore(ctrl)
	agentStore.EXPECT().
		ListAgents().
		Return([]proto.AgentState{
			proto.AgentState{
				Hostname: "localhost",
			},
		}, nil).
		Times(1)

	gameserverStore := mocks.NewMockGameserverStore(ctrl)
	gameserverStore.EXPECT().
		CreateGameserver(gomock.Any()).
		Return(nil).
		Times(1)

	router := utils.SetupRouter()
	MountGameserverAPI(router, agentStore, gameserverStore)

	claims := map[string]interface{}{
		"sub": "user1",
	}
	payload := createGameserverRequest{
		Name:    "My server",
		Game:    "Minecraft",
		Version: "1.12",
		Parameters: map[string]string{
			"motd": "hello all!",
		},
	}
	payloadBytes, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/gameservers/", bytes.NewReader(payloadBytes))

	req.Header.Add("authorization", "Bearer "+utils.BuildToken(claims))

	router.ServeHTTP(w, req)

	assert.Equal(t, 202, w.Code)

	res := createGameserverResponse{}
	json.Unmarshal(w.Body.Bytes(), &res)

	assert.Equal(t, "My server", res.Name)
	assert.Equal(t, "Minecraft", res.Game)
	assert.Nil(t, res.Address)
	assert.Equal(t, "UNKNOWN", res.Status)
	assert.Equal(t, "1.12", res.Version)
}

func TestCannotDeleteOtherUserServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	agentStore := mocks.NewMockAgentStore(ctrl)

	gameserverStore := mocks.NewMockGameserverStore(ctrl)
	gameserver := server.Gameserver{
		Definition: server.GameserverDefinition{
			Owner: "otheruser",
		},
	}
	gameserverStore.EXPECT().
		GetGameserver("serverUUID").
		Return(&gameserver, nil).
		Times(1)

	router := utils.SetupRouter()
	MountGameserverAPI(router, agentStore, gameserverStore)

	claims := map[string]interface{}{
		"sub": "user1",
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/gameservers/serverUUID/", nil)

	req.Header.Add("authorization", "Bearer "+utils.BuildToken(claims))
	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
}

func TestDeleteServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	agentStore := mocks.NewMockAgentStore(ctrl)

	gameserverStore := mocks.NewMockGameserverStore(ctrl)
	gameserverStore.EXPECT().
		DeleteGameserver("serverUUID").
		Return(nil).
		Times(1)

	gameserver := server.Gameserver{
		Definition: server.GameserverDefinition{
			Owner: "user1",
		},
	}
	gameserverStore.EXPECT().
		GetGameserver("serverUUID").
		Return(&gameserver, nil).
		Times(1)

	router := utils.SetupRouter()
	MountGameserverAPI(router, agentStore, gameserverStore)

	claims := map[string]interface{}{
		"sub": "user1",
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/gameservers/serverUUID/", nil)

	req.Header.Add("authorization", "Bearer "+utils.BuildToken(claims))
	router.ServeHTTP(w, req)

	assert.Equal(t, 202, w.Code)
}
