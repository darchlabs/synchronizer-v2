package env

type Env struct {
	CronjobIntervalSeconds  int64  `envconfig:"cronjob_interval_seconds" required:"true"`
	DatabaseDSN             string `envconfig:"database_dsn" required:"true"`
	Port                    string `envconfig:"port" required:"true"`
	Debug                   bool   `envconfig:"debug" default:"false"`
	MigrationDir            string `envconfig:"migration_dir" required:"true"`
	NetworksEtherscanURL    string `envconfig:"networks_etherscan_url" required:"true"`
	NetworksEtherscanAPIKey string `envconfig:"networks_etherscan_api_key" required:"true"`
	NetworksNodeURL         string `envconfig:"networks_node_url" required:"true"`
	MaxTransactions         int    `envconfig:"max_transactions" required:"true"`
	WebhooksIntervalSeconds int64  `envconfig:"webhooks_interval_seconds" required:"true"`
}
