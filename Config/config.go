package config

import (
	"context"
	"fmt"
	"io/ioutil"
	"sync"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Db   Database            `yaml:"database"`
	Lnds map[string]*LndInfo `yaml:"lnds"`
	Btc  BitcoinInfo         `yaml:"btc"`
}

type Database struct {
	Type     string
	Address  string
	User     string
	Password string
}

type LndInfo struct {
	Address  string `yaml:"address"`
	Cert     string `yaml:"cert"`
	Macaroon string `yaml:"macaroon"`
}

type BitcoinInfo struct {
	Address  string `yaml:"address"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Mainnet  bool   `yaml:"mainnet"`
}

var Cfg *Config

var cso sync.Once

func GetCfg(ctx context.Context) *Config {
	cso.Do(func() {
		Cfg = newCfg(ctx)
	})

	return Cfg
}

func newCfg(ctx context.Context) *Config {
	cfgpath := ctx.Value("configfile")

	yamlFile, err := ioutil.ReadFile(cfgpath.(string))

	if err != nil {
		fmt.Println("Read Yaml File with ERROR :", err)
	}

	err = yaml.Unmarshal(yamlFile, &Cfg)

	if err != nil {
		fmt.Println("Decode Yaml File with ERROR :", err)
	}

	return Cfg
}
