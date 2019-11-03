package gameservers

import (
	"fmt"

	"github.com/Trojan295/chinchilla/proto"
	"github.com/Trojan295/chinchilla/server"
)

type teamspeakGameserverManager struct {
}

func (manager teamspeakGameserverManager) metadata() GameserverMetadata {
	return GameserverMetadata{
		Name: "Teamspeak",
		Options: []GameserverOptions{
			GameserverOptions{
				Version:    "latest",
				Parameters: map[string]string{},
			},
			GameserverOptions{
				Version:    "3.9",
				Parameters: map[string]string{},
			},
		},
	}
}

func (manager teamspeakGameserverManager) createDeployment(definition *server.GameserverDefinition) (*proto.GameserverDeployment, error) {
	envVars := []*proto.EnvironmentVariable{
		&proto.EnvironmentVariable{
			Name:  "TS3SERVER_LICENSE",
			Value: "accept",
		},
	}

	return &proto.GameserverDeployment{
		Name:  definition.Name,
		UUID:  definition.UUID,
		Image: fmt.Sprintf("teamspeak:%s", definition.Version),
		ResourceRequirements: &proto.ResourceRequirements{
			MemoryReservation: 64 * 1024,
		},
		Ports: []*proto.NetworkPort{
			&proto.NetworkPort{
				Protocol:      proto.NetworkProtocol_TCP,
				ContainerPort: 30033,
			},
			&proto.NetworkPort{
				Protocol:      proto.NetworkProtocol_UDP,
				ContainerPort: 9987,
			},
		},
		Environment: envVars,
	}, nil
}
