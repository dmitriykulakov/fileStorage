package models

type Clients struct {
	Name         string `gorm:"primaryKey"`
	HashPassword string
}

type PgConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
}
