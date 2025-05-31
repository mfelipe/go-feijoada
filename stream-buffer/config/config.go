package config

import "time"

type Config struct {
	Redis  *Server `json:"redis" koanf:"redis,required_without=Valkey"`
	Valkey *Server `json:"valkey" koanf:"valkey,required_without=Redis"`
	Stream Stream  `json:"stream" koanf:"stream,required"`
}

type Server struct {
	IsCluster  bool   `json:"isCluster" koanf:"isCluster"`
	Address    string `json:"address" koanf:"address"`
	Username   string `json:"username" koanf:"username"`
	Password   string `json:"password" koanf:"password"`
	ClientName string `json:"clientName" koanf:"clientName"`
}

type Stream struct {
	Name      string        `json:"name" koanf:"name,required"`
	Group     string        `json:"group" koanf:"group,required"`
	ReadCount int64         `json:"readCount" koanf:"readCount,required,gt=10"`
	Block     time.Duration `json:"block" koanf:"block,required,gte=10000000"`
}
