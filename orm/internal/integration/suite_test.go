package integration

import (
	"fmt"
	"github.com/coderi421/kyuu/orm"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type Suite struct {
	suite.Suite
	driver string
	dsn    string

	db *orm.DB
}

func (s *Suite) SetupSuite() {
	db, err := orm.Open(s.driver, s.dsn)
	require.NoError(s.T(), err)
	err = db.Wait()
	if err != nil {
		panic(fmt.Sprintf("SetupSuite err: %v", err))
	}
	s.db = db
}
