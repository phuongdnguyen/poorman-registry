package config

import "time"

type MysqlConfig struct {
	Host     string `mapstructure:"host"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbName"`
}

type Config struct {
	DBConfig            MysqlConfig   `mapstructure:"dbConfig"`
	Images              []Image       `mapstructure:"images"`
	WorkerFetchInterval time.Duration `mapstructure:"workerFetchInterval"`
}

type Image struct {
	Name        string `mapstructure:"name"`
	Constraint  string `mapstructure:"constraint"`
	MainPackage string `mapstructure:"mainPackage"`
}
