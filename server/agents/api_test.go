package agents

import (
	"encoding/json"
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

func TestListAgents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	agentStore := mocks.NewMockAgentStore(ctrl)
	agentStore.EXPECT().ListAgents().Return([]proto.AgentState{
		proto.AgentState{
			Hostname: "localhost",
			Resources: &proto.AgentResources{
				Cpus:        2,
				Memory:      2048,
				IpAddresses: 2,
			},
			ResourceUsage: &proto.AgentResourceUsage{
				Memory: 1024,
			},
		},
	}, nil).AnyTimes()

	gameserverStore := mocks.NewMockGameserverStore(ctrl)
	gameserverStore.EXPECT().ListGameservers().Return([]server.Gameserver{
		server.Gameserver{
			Deployment: &proto.GameserverDeployment{
				Agent: "localhost",
				ResourceRequirements: &proto.ResourceRequirements{
					MemoryReservation: 1024,
				},
			},
		},
	}, nil).AnyTimes()

	router := utils.SetupRouter()
	MountAgentsAPI(router, agentStore, gameserverStore)

	claims := map[string]interface{}{
		"permissions": []string{"read:agents"},
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/agents/", nil)
	req.Header.Add("authorization", "Bearer "+utils.BuildToken(claims))

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var res listAgentsResponse
	json.Unmarshal(w.Body.Bytes(), &res)
	assert.Equal(t, "localhost", res[0].Hostname)
	assert.Equal(t, 2, res[0].Resources.Cpus)
	assert.Equal(t, 2048, res[0].Resources.Memory)
	assert.Equal(t, 1024, res[0].UsedResources.Memory)
	assert.Equal(t, 1024, res[0].ReservedResources.Memory)
	assert.Equal(t, 2, res[0].Resources.IPAddresses)
}

func TestListAgentsWhenNotAuthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	agentStore := mocks.NewMockAgentStore(ctrl)
	gameserverStore := mocks.NewMockGameserverStore(ctrl)

	router := utils.SetupRouter()
	MountAgentsAPI(router, agentStore, gameserverStore)

	claims := map[string]interface{}{
		"permissions": []string{},
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/agents/", nil)
	req.Header.Add("authorization", "Bearer "+utils.BuildToken(claims))

	router.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)
}
