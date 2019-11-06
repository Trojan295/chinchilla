package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/Trojan295/chinchilla/proto"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// GameserverState holds the state of a game server
type GameserverState struct {
	Running bool
}

// GameserverManager struct
type GameserverManager struct {
	containers  client.ContainerAPIClient
	image       client.ImageAPIClient
	ipAddresses []string
}

// NewGameserverManager creates a GameserverManager instance
func NewGameserverManager(containersAPI client.ContainerAPIClient, imageAPI client.ImageAPIClient, ipAddresses []string) *GameserverManager {
	return &GameserverManager{
		containers:  containersAPI,
		image:       imageAPI,
		ipAddresses: ipAddresses,
	}
}

// Tick func
func (manager *GameserverManager) Tick(deploymentConfig *proto.GetGameserverDeploymentsResponse) error {
	deployments := deploymentConfig.Deployments

	runningServers, _ := manager.GetGameservers()

	for _, runningServer := range runningServers {
		if isForRemoval(runningServer, deployments) {
			serverUUID := runningServer.UUID
			log.Printf("Gameserver %s marked for removal", serverUUID)
			manager.RemoveGameserver(serverUUID)
			log.Printf("Gameserver %s removed", serverUUID)
		}
	}

	for _, server := range deployments {
		runningServer, err := manager.GetGameserver(server.UUID)
		if err != nil {
			log.Printf("Error in GameserverManager::Tick: %s", err)
			continue
		}

		if runningServer == nil {
			log.Printf("Creating gameserver %s...", server.UUID)
			err := manager.CreateGameserver(server)
			if err != nil {
				log.Printf("Error whie creating %s: %s", server.UUID, err)
			}
			log.Printf("Created gameserver %s", server.UUID)
		}
	}
	return nil
}

// GetGameservers returns all gamservers
func (manager *GameserverManager) GetGameservers() ([]*proto.Gameserver, error) {
	ctx := context.Background()

	args := filters.NewArgs()
	args.Add("label", "chinchilla.gameserver.uuid")

	containers, err := manager.containers.ContainerList(ctx, types.ContainerListOptions{
		Filters: args,
	})
	if err != nil {
		return make([]*proto.Gameserver, 0), err
	}

	gameservers := make([]*proto.Gameserver, 0, len(containers))

	for _, cont := range containers {
		gameservers = append(gameservers, &proto.Gameserver{
			UUID:   cont.Labels["chinchilla.gameserver.uuid"],
			Status: proto.GameserverStatus_RUNNING,
			Endpoint: &proto.Endpoint{
				IpAddress: cont.Labels["chinchilla.gameserver.ip_address"],
			},
		})
	}

	return gameservers, nil
}

// GetGameserver returns gameserver
func (manager *GameserverManager) GetGameserver(UUID string) (*proto.Gameserver, error) {
	servers, err := manager.GetGameservers()
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		if server.UUID == UUID {
			return server, nil
		}
	}
	return nil, nil
}

// CreateGameserver creates a complete server
func (manager *GameserverManager) CreateGameserver(gameServer *proto.GameserverDeployment) error {
	err := manager.createGameserverContainer(gameServer)
	return err
}

// RemoveGameserver removes a gameserver
func (manager *GameserverManager) RemoveGameserver(uuid string) error {
	return manager.removeGameServerContainer(uuid)
}

func createGameserverContainerConfig(gameserverConfig *proto.GameserverDeployment, ipAddress string) *container.Config {
	envs := make([]string, 0)
	for _, variable := range gameserverConfig.Environment {
		env := fmt.Sprintf("%s=%s", variable.Name, variable.Value)
		envs = append(envs, env)
	}

	return &container.Config{
		Image: gameserverConfig.Image,
		Env:   envs,
		Labels: map[string]string{
			"chinchilla.gameserver.uuid":       gameserverConfig.UUID,
			"chinchilla.gameserver.ip_address": ipAddress,
		},
	}
}

func createGameserverHostConfig(deployment *proto.GameserverDeployment, ipAddress string) *container.HostConfig {

	portBindings := nat.PortMap{}
	for _, port := range deployment.Ports {
		var protocolName string
		if port.Protocol == proto.NetworkProtocol_TCP {
			protocolName = "tcp"
		} else {
			protocolName = "udp"
		}
		key, _ := nat.NewPort(protocolName, fmt.Sprintf("%d", port.ContainerPort))
		value := []nat.PortBinding{
			{
				HostIP:   ipAddress,
				HostPort: fmt.Sprintf("%d", port.ContainerPort),
			},
		}
		portBindings[key] = value
	}

	return &container.HostConfig{
		PortBindings: portBindings,
		Resources: container.Resources{
			Memory:            deployment.ResourceRequirements.MemoryLimit * 1024,
			MemoryReservation: deployment.ResourceRequirements.MemoryReservation * 1024,
		},
	}
}

func (manager *GameserverManager) findFreeIPAddress(ports []*proto.NetworkPort, allIPs []string) (*string, error) {
	testPort := ports[0]

	var testPortFilter string
	if testPort.Protocol == proto.NetworkProtocol_TCP {
		testPortFilter = fmt.Sprintf("%d/tcp", testPort.ContainerPort)
	} else {
		testPortFilter = fmt.Sprintf("%d/udp", testPort.ContainerPort)
	}

	for _, ip := range allIPs {
		args := filters.NewArgs()
		args.Add("label", fmt.Sprintf("chinchilla.gameserver.ip_address=%s", ip))
		args.Add("publish", testPortFilter)

		containers, err := manager.containers.ContainerList(context.Background(), types.ContainerListOptions{
			Filters: args,
		})
		if err != nil {
			continue
		}

		if len(containers) == 0 {
			return &ip, nil
		}
	}

	return nil, errors.New("Cannot find free IP")
}

func (manager *GameserverManager) createGameserverContainer(deployment *proto.GameserverDeployment) error {
	ctx := context.Background()

	reader, err := manager.image.ImagePull(ctx, deployment.Image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, reader)

	ipAddress, err := manager.findFreeIPAddress(deployment.Ports, manager.ipAddresses)
	if err != nil {
		return err
	}

	container, err := manager.containers.ContainerCreate(ctx,
		createGameserverContainerConfig(deployment, *ipAddress),
		createGameserverHostConfig(deployment, *ipAddress),
		nil,
		deployment.UUID,
	)

	if err != nil {
		return err
	}

	if err := manager.containers.ContainerStart(ctx, container.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	return nil
}

func (manager *GameserverManager) removeGameServerContainer(containerID string) error {
	ctx := context.Background()
	return manager.containers.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{Force: true})
}

func isForRemoval(server *proto.Gameserver, deployments []*proto.GameserverDeployment) bool {
	for _, deployment := range deployments {
		if deployment.UUID == server.UUID {
			return false
		}
	}
	return true
}
