package mysqltestcontainer

import (
	"context"
	"fmt"
	"github.com/hooligram/kifu"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	rootUsername = "root"
	rootPassword = "password"
	defaultImage = "mariadb:10.5"
)

// Create creates a container containing MySQL in Docker & returns the connection info along with
// the created database.
func Create(name string) (*MySqlTestContainer, error) {
	return CreateWithConfig(Config{
		DB: DbConfig{
			RootPassword: rootPassword,
			ExposedPorts: []string{"3306/tcp", "33060/tcp"},
			Name:         name,
			Image:        defaultImage,
		},
	})
}

type DbConfig struct {
	RootPassword string
	ExposedPorts []string
	Name         string
	Image        string
}

type Config struct {
	DB DbConfig
}

// CreateWithConfig creates a container containing MySQL in Docker & returns the connection info along with
// the created database.
func CreateWithConfig(cfg Config) (*MySqlTestContainer, error) {
	kifu.Info("Starting MySQL test container...")
	req := testcontainers.ContainerRequest{
		Image:        cfg.DB.Image,
		ExposedPorts: cfg.DB.ExposedPorts,
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": cfg.DB.RootPassword,
			"MYSQL_DATABASE":      cfg.DB.Name,
		},
		WaitingFor: wait.ForLog("3306"),
		Name:       cfg.DB.Name,
		SkipReaper: true,
	}
	ctx := context.Background()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}
	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "3306/tcp")
	p := fmt.Sprint(port.Int())
	kifu.Info("Connecting to MySQL inside test container: host=%v, port=%v, database_name=%v", host, p, cfg.DB.Name)
	db, err := open(host, p, cfg.DB.RootPassword, cfg.DB.Name)
	if err != nil {
		return nil, err
	}
	kifu.Info("MySQL test container started successfully!")
	mySql := &MySqlTestContainer{
		db: db,
		dbInfo: &DbInfo{
			Username: rootUsername,
			Password: cfg.DB.RootPassword,
			Ip:       host,
			Port:     p,
			DbName:   cfg.DB.Name,
		},
		container: container,
	}
	return mySql, nil
}
