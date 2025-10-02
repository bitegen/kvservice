package migrator

import (
	"cloud/config"
	"cloud/utils"
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func RunMigrations(cfg config.PostgresConfig, dir string) error {
	dsn := utils.MakeDSN(cfg)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.Up(db, dir); err != nil {
		return err
	}
	return nil
}
