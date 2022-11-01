package gorm_test

import (
	"context"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	gmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"

	"github.com/andylongstaffe/go-mysql-test-container/mysqltestcontainer"
	"github.com/stretchr/testify/assert"
)

func TestGorm(t *testing.T) {
	// Create container
	container, err := mysqltestcontainer.Create("test")
	assert.Nil(t, err)
	err = container.GetDb().Ping()
	assert.Nil(t, err)
	defer container.GetContainer().Terminate(context.TODO())

	// Migrate schema
	driver, err := mysql.WithInstance(container.GetDb(), &mysql.Config{})
	assert.Nil(t, err)
	m, err := migrate.NewWithDatabaseInstance(
		"file://./migrations",
		"mysql", driver)
	assert.Nil(t, err)

	err = m.Up()
	assert.Nil(t, err)

	// check schema here via gorm
	db, err := gorm.Open(gmysql.New(gmysql.Config{
		Conn: container.GetDb(),
	}), &gorm.Config{})

	assert.Nil(t, db.Exec("insert into test(firstname) values (?)", "andy").Error)

	var name string
	assert.Nil(t, db.Raw("select firstname from test").Scan(&name).Error)
	assert.Equal(t, "andy", name)
}
