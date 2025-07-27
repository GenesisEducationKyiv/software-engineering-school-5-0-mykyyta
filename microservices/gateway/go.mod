module gateway

go 1.24.3

require (
	github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger v0.0.0
	github.com/google/uuid v1.6.0
	go.uber.org/zap v1.27.0
	gopkg.in/yaml.v3 v3.0.1
)

require go.uber.org/multierr v1.10.0 // indirect

replace github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger => ../pkg/logger
