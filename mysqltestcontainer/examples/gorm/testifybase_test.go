package gorm_test

import (
	"github.com/andylongstaffe/go-mysql-test-container/mysqltestcontainer/examples/gorm"
	"github.com/stretchr/testify/suite"
	"testing"
)

// ** Instance

func TestIntegrationTestInstanceSuite(t *testing.T) {
	suite.Run(t, new(IntegrationInstance))
}

type IntegrationInstance struct {
	gorm.IntegrationBase
}

func (suite *IntegrationInstance) SetupSuite() {
	suite.IntegrationBase.SetupSuite()
	suite.T().Logf("running (instance) SetupSuite")
}

func (suite *IntegrationInstance) Test1() {
	suite.T().Logf("running test 1")
	suite.T().Logf("str: %s", suite.Str)
}

func (suite *IntegrationInstance) Test2() {
	suite.T().Logf("running test 2")
}
