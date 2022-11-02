package gorm_test

import (
	"context"
	"github.com/andylongstaffe/go-mysql-test-container/mysqltestcontainer"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/google/martian/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	gmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

type IntegrationWithReuseSuite struct {
	suite.Suite

	db                   *gorm.DB
	container            *mysqltestcontainer.MySqlTestContainer
	useExistingContainer bool
}

func TestIntegrationWithReuseSuite(t *testing.T) {
	suite.Run(t, new(IntegrationWithReuseSuite))
}

func (suite *IntegrationWithReuseSuite) SetupSuite() {
	t := suite.T()
	//suite.useExistingContainer = true

	name := "integration_test_testify_test"

	if suite.useExistingContainer && isContainerRunning(name) {
		log.Debugf("using existing container %s", name)
		// get db connection (and add to suite)
		dsn := "root:password@tcp(127.0.0.1:3306)/integration_test_testify_test?charset=utf8&parseTime=true"
		db, err := gorm.Open(gmysql.New(gmysql.Config{
			DSN: dsn,
		}), &gorm.Config{})
		if err != nil {
			t.Fatalf("error getting db connection - %s", err)
		}
		suite.db = db

		// suite.container will be left as null (no need to manage container)
	} else {
		// Create container
		container, err := mysqltestcontainer.CreateWithConfig(mysqltestcontainer.Config{
			DB: mysqltestcontainer.DbConfig{
				RootPassword: "password",
				ExposedPorts: []string{"3306/tcp", "33060/tcp"},
				Name:         name,
				Image:        "mariadb:10.5",
			},
		})
		assert.Nil(t, err)
		err = container.GetDb().Ping()
		assert.Nil(t, err)

		// Migrate schema
		driver, err := mysql.WithInstance(container.GetDb(), &mysql.Config{})
		assert.Nil(t, err)
		m, err := migrate.NewWithDatabaseInstance(
			"file://./migrations",
			"mysql", driver)
		assert.Nil(t, err)

		err = m.Up()
		assert.Nil(t, err)

		db, err := gorm.Open(gmysql.New(gmysql.Config{
			Conn: container.GetDb(),
		}), &gorm.Config{})
		if err != nil {
			t.Fatalf("error getting db connection - %s", err)
		}

		suite.container = container
		suite.db = db
	}

	t.Logf("setupSuite finished")
}

func (suite *IntegrationWithReuseSuite) TearDownSuite() {
	if !suite.useExistingContainer {
		if err := suite.container.GetContainer().Terminate(context.TODO()); err != nil {
			suite.T().Logf("error terminating container - %s", err)
		} else {
			suite.T().Logf("container terminated")
		}
	}
	suite.T().Logf("tearDownSuite finished")
}

func (suite *IntegrationWithReuseSuite) SetupTest() {
	assert.Nil(suite.T(), suite.db.Exec("delete from test").Error)
}

func (suite *IntegrationWithReuseSuite) TestIntegrationTest() {
	t := suite.T()
	db := suite.db

	assert.Nil(t, db.Exec("insert into test(firstname) values (?)", "andy").Error)

	var name string
	assert.Nil(t, db.Raw("select firstname from test").Scan(&name).Error)
	assert.Equal(t, "andy", name)
}

func (suite *IntegrationWithReuseSuite) TestIntegrationTest2() {
	t := suite.T()

	db, err := gorm.Open(gmysql.New(gmysql.Config{
		Conn: suite.container.GetDb(),
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("error getting db connection - %s", err)
	}

	assert.Nil(t, db.Exec("insert into test(firstname) values (?)", "marc").Error)

	var name string
	assert.Nil(t, db.Raw("select firstname from test").Scan(&name).Error)
	assert.Equal(t, "marc", name)
}

func isContainerRunning(name string) bool {
	containers := listContainersByName()
	for _, c := range containers {
		for _, n := range c {
			if n == "/"+name {
				return true
			}
		}
	}
	return false
}

func listContainersByName() [][]string {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	var names [][]string
	for _, container := range containers {
		names = append(names, container.Names)
	}
	return names
}
