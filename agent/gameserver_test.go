package agent

import (
	"reflect"
	"testing"

	"github.com/Trojan295/chinchilla/proto"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
)

func TestCreateGameserverContainerConfig(t *testing.T) {
	runConfig := &proto.GameserverDeployment{
		Agent: "",
		UUID:  "787a1b9d-6371-44d4-bd0b-d3c94077ad6b",
		Environment: []*proto.EnvironmentVariable{
			&proto.EnvironmentVariable{
				Name:  "name",
				Value: "testserver",
			},
		},
		Image: "minecraft:1.3.8",
	}

	containerConfig := createGameserverContainerConfig(runConfig)

	if containerConfig.Image != runConfig.Image {
		t.Errorf("Wrong image")
	}

	if !reflect.DeepEqual(containerConfig.Env, []string{"name=testserver"}) {
		t.Errorf("Wrong env vars: %v", containerConfig.Env)
	}

	if !reflect.DeepEqual(containerConfig.Labels, map[string]string{
		"chinchilla.agent.gameserver": runConfig.UUID,
	}) {
		t.Errorf("Wrong labels: %v", containerConfig.Labels)
	}
}

func TestCreateGameserverHostConfig(t *testing.T) {
	deployment := &proto.GameserverDeployment{
		ResourceRequirements: &proto.ResourceRequirements{
			MemoryLimit:       1234567,
			MemoryReservation: 123456,
		},
		Ports: []*proto.NetworkPort{
			&proto.NetworkPort{
				Protocol:      proto.NetworkProtocol_TCP,
				ContainerPort: 1024,
			},
		},
	}

	hostConfig, err := createGameserverHostConfig(deployment, []string{"127.0.0.1"})
	if err != nil {
		t.Fatalf(err.Error())
	}

	if !reflect.DeepEqual(hostConfig.PortBindings, nat.PortMap{
		"1024/tcp": []nat.PortBinding{
			{
				HostIP:   "127.0.0.1",
				HostPort: "1024",
			},
		},
	}) {
		t.Errorf("Wrong port mappings: %v", hostConfig.PortBindings)
	}

	assert.Equal(t, int64(1234567*1024), hostConfig.Resources.Memory)
	assert.Equal(t, int64(123456*1024), hostConfig.Resources.MemoryReservation)
}
