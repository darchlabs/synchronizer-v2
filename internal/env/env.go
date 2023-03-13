package env

// define config variables for env
type Env struct {
	IntervalSeconds  int64  `envconfig:"interval_seconds" required:"true"`
	DatabaseFilepath string `envconfig:"database_filepath" required:"true"`
	Port             string `envconfig:"port" required:"true"`
	BaseURL          string `envconfig:"base_url" default:""`
}
