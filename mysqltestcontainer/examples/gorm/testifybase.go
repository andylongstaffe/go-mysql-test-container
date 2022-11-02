package gorm

import "github.com/stretchr/testify/suite"

// ** Base

type IntegrationBase struct {
	suite.Suite
	Str string
}

func (suite *IntegrationBase) SetupSuite() {
	suite.T().Logf("running (base) SetupSuite")
	suite.Str = "initialised"
}
