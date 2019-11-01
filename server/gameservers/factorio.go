package gameservers

import (
	"fmt"
	"strings"

	"github.com/Trojan295/chinchilla-server/proto"
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

func (manager factorioGameserverManager) createDeployment(definition *server.GameserverDefinition) (*proto.GameserverDeployment, error) {
	envVars := []*proto.EnvironmentVariable{}

	for key, value := range definition.Parameters {
		envVars = append(envVars, &proto.EnvironmentVariable{
			Name:  fmt.Sprintf("CONFIG_%s", strings.ToUpper(key)),
			Value: value,
		})
	}

	return &proto.GameserverDeployment{
		Name:  definition.Name,
		UUID:  definition.UUID,
		Image: fmt.Sprintf("factoriotools/factorio:%s", definition.Version),
		ResourceRequirements: &proto.ResourceRequirements{
			MemoryReservation: 512 * 1024,
		},
		Ports: []*proto.NetworkPort{
			&proto.NetworkPort{
				Protocol:      proto.NetworkProtocol_UDP,
				ContainerPort: 34197,
			},
		},
		Environment: envVars,
	}, nil
}
