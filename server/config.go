package server

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

// Server configuration
type Server struct {
	Port int
}

// Etcd configuration
type Etcd struct {
	Address string
}

// Configuration of agent
type Configuration struct {
	Auth   map[string]interface{}
	Server Server
	Etcd   Etcd
}

// LoadConfig load a Configuration from a toml file
func LoadConfig() (*Configuration, error) {
	dat, err := ioutil.ReadFile("server.toml")
	if err != nil {
		return nil, err
	}

	config := &Configuration{}
	err = toml.Unmarshal(dat, config)
	return config, err
}
