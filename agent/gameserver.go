package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
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
	args.Add("label", "chinchilla.agent.gameserver")

	containers, err := manager.containers.ContainerList(ctx, types.ContainerListOptions{
		Filters: args,
	})
	if err != nil {
		return make([]*proto.Gameserver, 0), err
	}

	gameservers := make([]*proto.Gameserver, 0, len(containers))

	for _, cont := range containers {

		gameservers = append(gameservers, &proto.Gameserver{
			UUID:   cont.Labels["chinchilla.agent.gameserver"],
			Status: proto.GameserverStatus_RUNNING,
			Endpoint: &proto.Endpoint{
				IpAddress: cont.Ports[0].IP,
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

func createGameserverContainerConfig(gameserverConfig *proto.GameserverDeployment) *container.Config {
	envs := make([]string, 0)
	for _, variable := range gameserverConfig.Environment {
		env := fmt.Sprintf("%s=%s", variable.Name, variable.Value)
		envs = append(envs, env)
	}

	return &container.Config{
		Image: gameserverConfig.Image,
		Env:   envs,
		Labels: map[string]string{
			"chinchilla.agent.gameserver": gameserverConfig.UUID,
		},
	}
}

func createGameserverHostConfig(deployment *proto.GameserverDeployment, allIPs []string) (*container.HostConfig, error) {
	ipAddress, err := findFreeIPAddress(deployment.Ports, allIPs)
	if err != nil {
		return nil, err
	}

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
				HostIP:   *ipAddress,
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
	}, nil
}

func findFreeIPAddress(ports []*proto.NetworkPort, allIPs []string) (*string, error) {
	for _, ip := range allIPs {
		isFree := true
		for _, networkPort := range ports {
			if networkPort.Protocol == proto.NetworkProtocol_TCP {
				addr := &net.TCPAddr{
					IP:   net.ParseIP(ip),
					Port: int(networkPort.ContainerPort),
				}
				ln, err := net.ListenTCP("tcp", addr)

				if err != nil {
					isFree = false
					break
				} else {
					ln.Close()
				}
			} else {
				addr := &net.UDPAddr{
					IP:   net.ParseIP(ip),
					Port: int(networkPort.ContainerPort),
				}
				ln, err := net.ListenUDP("udp", addr)

				if err != nil {
					isFree = false
					break
				} else {
					ln.Close()
				}
			}
		}

		if isFree {
			return &ip, nil
		}
	}

	return nil, errors.New("Cannot find free IP")
}

func (manager *GameserverManager) createGameserverContainer(gameServer *proto.GameserverDeployment) error {
	ctx := context.Background()

	reader, err := manager.image.ImagePull(ctx, gameServer.Image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, reader)

	hostConfig, err := createGameserverHostConfig(gameServer, manager.ipAddresses)
	if err != nil {
		return err
	}

	container, err := manager.containers.ContainerCreate(ctx,
		createGameserverContainerConfig(gameServer),
		hostConfig,
		nil,
		gameServer.UUID,
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

func isForRemoval(server *proto.Gameserver, servers []*proto.GameserverDeployment) bool {
	for _, server := range servers {
		if server.UUID == server.UUID {
			return false
		}
	}
	return true
}
