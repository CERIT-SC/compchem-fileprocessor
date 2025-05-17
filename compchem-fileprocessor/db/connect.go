package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"fi.muni.cz/invenio-file-processor/v2/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func CreatePgPool(
	ctx context.Context,
	logger *zap.Logger,
	pgConfig *config.Postgres,
	migrations string,
) (*pgxpool.Pool, error) {
	logger.Info("Creating pg pool for config",
		zap.String("host", pgConfig.Host),
		zap.String("port", pgConfig.Port),
		zap.String("database", pgConfig.Database),
	)
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := doMigration(logger, pgConfig, migrations); err != nil {
		return nil, err
	}

	return doConnect(ctx, logger, pgConfig)
}

func doMigration(
	logger *zap.Logger,
	pgConfig *config.Postgres,
	migrations string,
) error {
	logger.Info("Running migrations", zap.String("migration-source", migrations))
	db, err := sql.Open("postgres", createPgUrl(pgConfig)+"?sslmode=disable")
	if err != nil {
		logger.Error(
			"Error when migrating pg",
			zap.String("host", pgConfig.Host),
			zap.String("port", pgConfig.Port),
			zap.String("database", pgConfig.Database),
		)
		return err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.Error(
			"Error getting driver instance during migration",
			zap.String("host", pgConfig.Host),
			zap.String("port", pgConfig.Port),
			zap.String("database", pgConfig.Database),
		)
		return err
	}

	migration, err := migrate.NewWithDatabaseInstance(
		migrations,
		"postgres",
		driver,
	)
	if err != nil {
		logger.Error(
			"Error obtaining migration instance",
			zap.String("host", pgConfig.Host),
			zap.String("port", pgConfig.Port),
			zap.String("database", pgConfig.Database),
		)
		return err
	}

	err = migration.Up()
	if err != nil {
		logger.Error(
			"Error migrating schema",
			zap.String("host", pgConfig.Host),
			zap.String("port", pgConfig.Port),
			zap.String("database", pgConfig.Database),
		)
		return err
	}

	return nil
}

func doConnect(
	ctx context.Context,
	logger *zap.Logger,
	pgConfig *config.Postgres,
) (*pgxpool.Pool, error) {
	dbpool, err := pgxpool.New(ctx, createPgUrl(pgConfig))
	if err != nil {
		logger.Error(
			"Error when connecting to pg",
			zap.String("host", pgConfig.Host),
			zap.String("port", pgConfig.Port),
			zap.String("database", pgConfig.Database),
		)
		return nil, err
	}
	if err := dbpool.Ping(ctx); err != nil {
		logger.Error(
			"Failed to ping database",
			zap.String("host", pgConfig.Host),
			zap.String("port", pgConfig.Port),
			zap.String("database", pgConfig.Database),
			zap.Error(err),
		)
		return nil, err
	}

	logger.Info(
		"Successfully connected to database",
		zap.String("host", pgConfig.Host),
		zap.String("database", pgConfig.Database),
	)

	return dbpool, nil
}

func createPgUrl(pgConfig *config.Postgres) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		pgConfig.Auth.Username,
		pgConfig.Auth.Password,
		pgConfig.Host,
		pgConfig.Port,
		pgConfig.Database,
	)
}
