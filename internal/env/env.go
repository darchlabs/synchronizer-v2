package env

type Env struct {
	IntervalSeconds string `envconfig:"interval_seconds" required:"true"`
	DatabaseDSN     string `envconfig:"database_dsn" required:"true"`
	Port            string `envconfig:"port" required:"true"`
	Debug           bool   `envconfig:"debug" default:"false"`
	MigrationDir    string `envconfig:"migration_dir" required:"true"`
	EtherscanApiURL string `envconfig:"etherscan_api_url" required:"true"`
	EtherscanApiKey string `envconfig:"etherscan_api_key" required:"true"`
}
