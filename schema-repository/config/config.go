package config

import (
	_ "embed"
	utilscfg "github.com/mfelipe/go-feijoada/utils/config"
	utilslog "github.com/mfelipe/go-feijoada/utils/log"
)

const (
	prefix = "SR"
)

//go:embed base.yaml
var baseCfg []byte

func Load() *Server {
	return utilscfg.Load[Server](prefix, baseCfg)
}

type Server struct {
	Port       int             `json:"port" koanf:"port,required"`
	Log        utilslog.Config `json:"log" koanf:"log"`
	Repository Repository      `json:"repository" koanf:"repository,required"`
}

type Repository struct {
	Redis  *RepoServer `json:"redis" koanf:"redis,required_without=Valkey"`
	Valkey *RepoServer `json:"valkey" koanf:"valkey,required_without=Redis"`
	Data   RepoData    `json:"data" koanf:"data,required"`
}

type RepoData struct {
	KeyPrefix    string `json:"keyPrefix" koanf:"keyPrefix,required"`
	KeySeparator string `json:"keySeparator" koanf:"keySeparator,required"`
}

type RepoServer struct {
	IsCluster  bool   `json:"isCluster" koanf:"isCluster"`
	Address    string `json:"address" koanf:"address"`
	Username   string `json:"username" koanf:"username"`
	Password   string `json:"password" koanf:"password"`
	ClientName string `json:"clientName" koanf:"clientName"`
}
