module email

go 1.24.3

require (
	github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger v0.0.0
	github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/metrics v0.0.0
	github.com/google/uuid v1.6.0
	github.com/joho/godotenv v1.5.1
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.20.5
	github.com/rabbitmq/amqp091-go v1.10.0
	github.com/redis/go-redis/v9 v9.11.0
	github.com/sendgrid/sendgrid-go v3.16.1+incompatible
	github.com/stretchr/testify v1.10.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.55.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/sendgrid/rest v2.6.9+incompatible // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/sys v0.22.0 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger => ../pkg/logger

replace github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/metrics => ../pkg/metrics
