package config

type Config struct {
	DefaultBaseURI string `json:"defaultBaseURI" koanf:"defaultBaseURI,required"`
}
