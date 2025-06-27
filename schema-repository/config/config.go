package config

import (
	"path/filepath"

	utilscfg "github.com/mfelipe/go-feijoada/utils/config"
	utilslog "github.com/mfelipe/go-feijoada/utils/log"
)

const (
	Prefix = "SR"
)

func Load() *Server {
	path, err := filepath.Abs("../config/base.yaml")
	if err != nil {
		panic(err)
	}

	var cfg Server
	utilscfg.Load[Server](Prefix, path, &cfg)

	return &cfg
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
