module github.com/mfelipe/go-feijoada/kafka-consumer

go 1.24.3

replace (
	github.com/mfelipe/go-feijoada/schema-validator => ../schema-validator
	github.com/mfelipe/go-feijoada/stream-buffer => ../stream-buffer
)

require (
	github.com/knadh/koanf/parsers/yaml v1.0.0
	github.com/knadh/koanf/providers/env v1.1.0
	github.com/knadh/koanf/providers/file v1.2.0
	github.com/knadh/koanf/v2 v2.2.0
	github.com/mfelipe/go-feijoada/stream-buffer v0.0.0-00010101000000-000000000000
	github.com/twmb/franz-go v1.19.4
	golang.org/x/sync v0.14.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/redis/go-redis/v9 v9.8.0 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.11.2 // indirect
	github.com/valkey-io/valkey-go v1.0.60 // indirect
	golang.org/x/sys v0.32.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
