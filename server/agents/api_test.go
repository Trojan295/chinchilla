package agents

import (
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
				Cpus:   2,
				Memory: 2024,
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
	assert.Equal(t, "[{\"hostname\":\"localhost\",\"reservedResources\":{\"memory\":1024},\"resources\":{\"cpus\":2,\"memory\":2024},\"usedResources\":{\"memory\":1024}}]", w.Body.String())
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
