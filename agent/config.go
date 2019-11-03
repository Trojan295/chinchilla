package agent

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

// Server configuration
type Server struct {
	Address string
}

// Agent configuration
type Agent struct {
	Hostname string
}

// Configuration of agent
type Configuration struct {
	Server Server
	Agent  Agent
}

// LoadConfig load a Configuration from a toml file
func LoadConfig() (*Configuration, error) {
	dat, err := ioutil.ReadFile("agent.toml")
	if err != nil {
		return nil, err
	}

	config := &Configuration{}
	err = toml.Unmarshal(dat, config)
	return config, err
}
