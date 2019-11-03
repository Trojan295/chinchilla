package gameservers

import (
	"github.com/Trojan295/chinchilla/proto"
	"github.com/Trojan295/chinchilla/server"
)

type minecraftGameserverManager struct {
}

func (manager minecraftGameserverManager) metadata() GameserverMetadata {
	return GameserverMetadata{
		Name: "Minecraft",
		Options: []GameserverOptions{
			GameserverOptions{
				Version: "1.13.2",
				Parameters: map[string]string{
					"motd": "Motto of the Day",
				},
			},
			GameserverOptions{
				Version: "1.14.1",
				Parameters: map[string]string{
					"motd": "Motto of the Day",
				},
			},
		},
	}
}

func (manager minecraftGameserverManager) createDeployment(definition *server.GameserverDefinition) (*proto.GameserverDeployment, error) {
	envVars := []*proto.EnvironmentVariable{
		&proto.EnvironmentVariable{
			Name:  "EULA",
			Value: "TRUE",
		},
	}

	if motd, ok := definition.Parameters["motd"]; ok {
		envVars = append(envVars, &proto.EnvironmentVariable{
			Name:  "MOTD",
			Value: motd,
		})
	}

	return &proto.GameserverDeployment{
		Name:  definition.Name,
		UUID:  definition.UUID,
		Image: "itzg/minecraft-server",
		ResourceRequirements: &proto.ResourceRequirements{
			MemoryReservation: 1536 * 1024,
		},
		Ports: []*proto.NetworkPort{
			&proto.NetworkPort{
				Protocol:      proto.NetworkProtocol_TCP,
				ContainerPort: 25565,
			},
		},
		Environment: envVars,
	}, nil
}
