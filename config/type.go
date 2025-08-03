package config

type Config struct {
	App      AppConfig      `yaml:"app" validate:"required"`
	Database DatabaseConfig `yaml:"database" validate:"required"`
	Redis    RedisConfig    `yaml:"redis" validate:"required"`
	Secret   SecretConfig   `yaml:"app" validate:"required"`
	Kafka    KafkaConfig    `yaml:"kafka" validate:"required"`
	Xendit   XenditConfig   `yaml:"xendit" validate:"required"`
	Toggle   ToggleConfig   `yaml:"toggle" validate:"required"`
}

type AppConfig struct {
	Port string `yaml:"port"`
}

type ToggleConfig struct {
	DisableCreateInvoiceDirectly bool `yaml:"disable_create_invoice_directly"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	Port     string `yaml:"port"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
}

type SecretConfig struct {
	JWTSecret string `yaml:"jwt_secret_key"`
}

type KafkaConfig struct {
	Broker string            `yaml:"broker"`
	Topics map[string]string `yaml:"topics"`
}

type XenditConfig struct {
	SecretApiKey string `yaml:"secret_api_key"`
	WebhookToken string `yaml:"webhook_token"`
}
