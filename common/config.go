package common

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

// Server configuration
type Server struct {
	Host string
	Port int
}

// Etcd configuration
type Etcd struct {
	Address string
}

// Agent configuration
type Agent struct {
	IPAddresses string
}

type Scheduler struct {
	Interval          int
	AgentContactDelay int
}

// Configuration of agent
type Configuration struct {
	Agent     Agent
	Auth      map[string]interface{}
	Server    Server
	Scheduler Scheduler
	Etcd      Etcd
}

// LoadConfig load a Configuration from a toml file
func LoadConfig() (*Configuration, error) {
	dat, err := ioutil.ReadFile("chinchilla.toml")
	if err != nil {
		return nil, err
	}

	config := &Configuration{}
	err = toml.Unmarshal(dat, config)
	return config, err
}
