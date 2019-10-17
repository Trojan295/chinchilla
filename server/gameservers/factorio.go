package gameservers

import (
	"errors"
	"fmt"
	"strings"

	common "github.com/Trojan295/chinchilla-common"
	"github.com/Trojan295/chinchilla-server/server"
)

type factorioGameserverManager struct {
}

func (manager factorioGameserverManager) metadata() GameserverMetadata {
	return GameserverMetadata{
		Name: "Factorio",
		Options: []GameserverOptions{
			GameserverOptions{
				Version: "0.16.51",
				Parameters: map[string]string{
					"description": "Description of the server",
				},
			},
			GameserverOptions{
				Version: "0.17.63",
				Parameters: map[string]string{
					"description": "Description of the server",
				},
			},
		},
	}
}

func (manager factorioGameserverManager) createRunConfiguration(definition *server.GameserverDefinition) (*common.GameserverRunConfiguration, error) {
	envVars := []*common.EnvironmentVariable{}

	for key, value := range definition.Parameters {
		envVars = append(envVars, &common.EnvironmentVariable{
			Name:  fmt.Sprintf("CONFIG_%s", strings.ToUpper(key)),
			Value: value,
		})
	}

	return &common.GameserverRunConfiguration{
		Name:  definition.Name,
		UUID:  definition.UUID,
		Image: fmt.Sprintf("factoriotools/factorio:%s", definition.Version),
		ResourceRequirements: &common.ResourceRequirements{
			MemoryReservation: 512,
		},
		Ports: []*common.NetworkPort{
			&common.NetworkPort{
				Protocol:      common.NetworkProtocol_UDP,
				ContainerPort: 34197,
			},
		},
		Environment: envVars,
	}, nil
}

func (manager factorioGameserverManager) endpoint(server *server.Gameserver, runningServer *common.Gameserver) (string, error) {
	if runningServer != nil {
		for _, port := range runningServer.PortMappings {
			if port.ContainerPort == 34197 && port.Protocol == common.NetworkProtocol_UDP {
				return fmt.Sprintf("%s:%d", server.RunConfiguration.Agent, port.HostPort), nil
			}
		}
	}
	return "", errors.New("Endpoint not ready")

}
