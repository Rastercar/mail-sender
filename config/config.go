package config

import (
	"context"
	"flag"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	Debug bool `yaml:"debug" env:"APP_DEBUG"`
}

type MailConfig struct {
	Sender           string `env-required:"true" yaml:"sender" env:"MAIL_SENDER"`
	RetryWaitTime    int    `env-required:"true" yaml:"retry_wait_time" env:"MAIL_RETRY_WAIT_TIME"`
	ReqPerSecLimit   int    `env-required:"true" yaml:"req_per_sec_limit" env:"MAIL_REQ_PER_SEC_LIMIT"`
	MaxRetryAttempts int    `env-required:"true" yaml:"max_retry_attempts" env:"MAIL_MAX_RETRY_ATTEMPTS"`
}

type TracerConfig struct {
	Url         string `env-required:"true" yaml:"url" env:"TRACER_URL"`
	ServiceName string `env-required:"true" yaml:"service_name" env:"TRACER_SERVICE_NAME"`
}

type RmqConfig struct {
	Url               string `env-required:"true" yaml:"url" env:"RMQ_URL"`
	Queue             string `env-required:"true" yaml:"queue" env:"RMQ_QUEUE"`
	ReconnectWaitTime int    `env-required:"true" yaml:"reconnect_wait_time" env:"RMQ_RECONNECT_WAIT_TIME"`
}

type AwsConfig struct {
	Region          string     `env-required:"true" yaml:"region" env:"AWS_REGION"`
	AccessKeyId     string     `env-required:"true" yaml:"access_key_id" env:"AWS_ACCESS_KEY_ID"`
	SecretAccessKey string     `env-required:"true" yaml:"secret_access_key" env:"AWS_SECRET_ACCESS_KEY"`
	Instance        aws.Config `env-required:"false"`
}

type Config struct {
	App    AppConfig    `yaml:"app"`
	Rmq    RmqConfig    `yaml:"rmq"`
	Aws    AwsConfig    `yaml:"aws"`
	Mail   MailConfig   `yaml:"mail"`
	Tracer TracerConfig `yaml:"tracer"`
}

func setAwsEnvVars(awsCfg AwsConfig) {
	// Important: LoadDefaultConfig uses enviroment variables to
	// determine the aws SDK session, so if the aws config was set
	// with the .yml file it would wrongfully use the env vars instead
	// see: https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/

	os.Setenv("AWS_REGION", awsCfg.Region)
	os.Setenv("AWS_ACCESS_KEY_ID", awsCfg.AccessKeyId)
	os.Setenv("AWS_SECRET_ACCESS_KEY", awsCfg.SecretAccessKey)
}

func Parse() (*Config, error) {
	var cfgFilePath = flag.String("config-file", "/etc/config.yml", "A filepath to the yml file containing the microservice configuration")
	flag.Parse()

	cfg := &Config{}

	if err := cleanenv.ReadConfig(*cfgFilePath, cfg); err != nil {
		return nil, err
	}

	setAwsEnvVars(cfg.Aws)

	awsCfgInstance, err := awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithRegion(cfg.Aws.Region))
	if err != nil {
		return nil, err
	}

	cfg.Aws.Instance = awsCfgInstance

	return cfg, nil
}
