package utils

import (
	"cloud/internal/config"
	"fmt"
)

func MakeDSN(cfg config.PostgresConfig) string {
	base := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DbName,
	)

	sslParam := "sslmode=disable"

	return fmt.Sprintf("%s?%s", base, sslParam)
}
