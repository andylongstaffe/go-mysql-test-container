package gorm_test

import (
	"context"
	"github.com/andylongstaffe/go-mysql-test-container/mysqltestcontainer"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	gmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

type IntegrationTestSuite struct {
	suite.Suite

	db             *gorm.DB
	container      *mysqltestcontainer.MySqlTestContainer
	reuseContainer bool
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (suite *IntegrationTestSuite) SetupSuite() {
	t := suite.T()

	// Create container
	container, err := mysqltestcontainer.Create("test")
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
	t.Logf("setupSuite finished")
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	if err := suite.container.GetContainer().Terminate(context.TODO()); err != nil {
		suite.T().Logf("error terminating container - %s", err)
	}
	suite.T().Logf("setupSuite finished")
}

func (suite *IntegrationTestSuite) SetupTest() {
	assert.Nil(suite.T(), suite.db.Exec("delete from test").Error)

}

func (suite *IntegrationTestSuite) TestIntegrationTest() {
	t := suite.T()
	db := suite.db

	assert.Nil(t, db.Exec("insert into test(firstname) values (?)", "andy").Error)

	var name string
	assert.Nil(t, db.Raw("select firstname from test").Scan(&name).Error)
	assert.Equal(t, "andy", name)
}

func (suite *IntegrationTestSuite) TestIntegrationTest2() {
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
