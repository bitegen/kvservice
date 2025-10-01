package utils

import (
	"cloud/config"
	"fmt"
)

func MakeDSN(cfg config.PostgresConfig) string {
	if cfg.Pool.MaxConns <= 0 {
		return fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s",
			cfg.User,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.DbName,
		)
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?pool_max_conns=%d",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DbName,
		cfg.Pool.MaxConns,
	)
}
