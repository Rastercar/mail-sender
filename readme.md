# Maile Sender

Mail sender service

---

## Configuration

configuration is set by a yml config file and enviroment variables, each variable on the yaml file can be overwrittern by a env var,
check `config/config.yml` for details.

when developing its easier to use the yml equivalent of those variables on your `config/config.dev.yml` file and running the service
with `make run_dev` or `go run cmd/main.go --config-file="./config/config.dev.yml"`
