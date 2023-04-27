package env

type Env struct {
	IntervalSeconds         string `envconfig:"interval_seconds" required:"true"`
	DatabaseDSN             string `envconfig:"database_dsn" required:"true"`
	Port                    string `envconfig:"port" required:"true"`
	Debug                   bool   `envconfig:"debug" default:"false"`
	MigrationDir            string `envconfig:"migration_dir" required:"true"`
	NetworksEtherscanURL    string `envconfig:"networks_etherscan_url" required:"true"`
	NetworksEtherscanAPIKey string `envconfig:"networks_etherscan_api_key" required:"true"`
	NetworksNodeURL         string `envconfig:"networks_node_url" required:"true"`
}
