package gameservers

import (
	"errors"
	"fmt"

	"github.com/Trojan295/chinchilla-server/proto"
	"github.com/Trojan295/chinchilla-server/server"
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

func (manager minecraftGameserverManager) createRunConfiguration(definition *server.GameserverDefinition) (*proto.GameserverRunConfiguration, error) {
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

	return &proto.GameserverRunConfiguration{
		Name:  definition.Name,
		UUID:  definition.UUID,
		Image: "itzg/minecraft-server",
		ResourceRequirements: &proto.ResourceRequirements{
			MemoryReservation: 1532,
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

func (manager minecraftGameserverManager) endpoint(gameserver *server.Gameserver, state *proto.Gameserver) (string, error) {
	if state != nil {
		for _, port := range state.PortMappings {
			if port.ContainerPort == 25565 && port.Protocol == proto.NetworkProtocol_TCP {
				return fmt.Sprintf("%s:%d", gameserver.RunConfiguration.Agent, port.HostPort), nil
			}
		}
	}
	return "", errors.New("Endpoint not ready")
}
