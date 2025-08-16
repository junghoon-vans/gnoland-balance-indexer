package repository

import (
	"testing"

	"shared/infra/database"
	"shared/testutils"

	"github.com/stretchr/testify/suite"
)

type BalanceRepositoryTestSuite struct {
	suite.Suite
	db   *database.Database
	repo BalanceRepository
}

func (suite *BalanceRepositoryTestSuite) SetupSuite() {
	db, err := testutils.SetupInMemoryDB()
	suite.Require().NoError(err)
	
	suite.db = db
	suite.repo = NewBalanceRepository(suite.db)
}

func (suite *BalanceRepositoryTestSuite) TearDownTest() {
	testutils.CleanupDatabase(suite.db)
}

func (suite *BalanceRepositoryTestSuite) TestCreateBalance() {
	balance := testutils.CreateTestBalance(testutils.TestAddress1, testutils.TestTokenPath, "1000")

	err := suite.repo.CreateBalance(balance)
	suite.Assert().NoError(err)
	suite.Assert().NotZero(balance.ID)
	suite.Assert().NotZero(balance.CreatedAt)
}

func (suite *BalanceRepositoryTestSuite) TestGetBalance() {
	// Create test data
	balance := testutils.CreateTestBalance(testutils.TestAddress1, testutils.TestTokenPath, "1000")
	err := suite.repo.CreateBalance(balance)
	suite.Require().NoError(err)

	// Test retrieval
	found, err := suite.repo.GetBalance(testutils.TestAddress1, testutils.TestTokenPath)
	suite.Assert().NoError(err)
	suite.Assert().NotNil(found)
	suite.Assert().Equal(testutils.TestAddress1, found.Address)
	suite.Assert().Equal(testutils.TestTokenPath, found.TokenPath)
	suite.Assert().Equal("1000", found.Amount)
}

func (suite *BalanceRepositoryTestSuite) TestGetBalanceNotFound() {
	// Query for non-existent balance
	found, err := suite.repo.GetBalance("g1notfound", testutils.TestTokenPath)
	suite.Assert().Error(err)
	suite.Assert().Nil(found)
}

func (suite *BalanceRepositoryTestSuite) TestUpdateBalance() {
	// Create test data
	balance := testutils.CreateTestBalance(testutils.TestAddress1, testutils.TestTokenPath, "1000")
	err := suite.repo.CreateBalance(balance)
	suite.Require().NoError(err)

	// Update balance
	balance.Amount = "2000"
	err = suite.repo.UpdateBalance(balance)
	suite.Assert().NoError(err)

	// Verify update
	updated, err := suite.repo.GetBalance(testutils.TestAddress1, testutils.TestTokenPath)
	suite.Assert().NoError(err)
	suite.Assert().Equal("2000", updated.Amount)
}

func (suite *BalanceRepositoryTestSuite) TestGetBalanceInTx() {
	// Create test data
	balance := testutils.CreateTestBalance(testutils.TestAddress1, testutils.TestTokenPath, "1000")
	err := suite.repo.CreateBalance(balance)
	suite.Require().NoError(err)

	// Query within transaction
	tx := suite.db.Begin()
	found, err := suite.repo.GetBalanceInTx(tx, testutils.TestAddress1, testutils.TestTokenPath)
	tx.Rollback()

	suite.Assert().NoError(err)
	suite.Assert().NotNil(found)
	suite.Assert().Equal(testutils.TestAddress1, found.Address)
	suite.Assert().Equal("1000", found.Amount)
}

func (suite *BalanceRepositoryTestSuite) TestCreateBalanceInTx() {
	tx := suite.db.Begin()
	defer tx.Rollback()

	balance := testutils.CreateTestBalance(testutils.TestAddress1, testutils.TestTokenPath, "1000")

	err := suite.repo.CreateBalanceInTx(tx, balance)
	suite.Assert().NoError(err)
	suite.Assert().NotZero(balance.ID)

	// Should be queryable within transaction
	found, err := suite.repo.GetBalanceInTx(tx, testutils.TestAddress1, testutils.TestTokenPath)
	suite.Assert().NoError(err)
	suite.Assert().NotNil(found)

	// Should not be queryable after transaction rollback
	tx.Rollback()
	found, err = suite.repo.GetBalance(testutils.TestAddress1, testutils.TestTokenPath)
	suite.Assert().Error(err)
	suite.Assert().Nil(found)
}

func (suite *BalanceRepositoryTestSuite) TestUpdateBalanceInTx() {
	// Create test data
	balance := testutils.CreateTestBalance(testutils.TestAddress1, testutils.TestTokenPath, "1000")
	err := suite.repo.CreateBalance(balance)
	suite.Require().NoError(err)

	tx := suite.db.Begin()
	defer tx.Rollback()

	// Update within transaction
	balance.Amount = "2000"
	err = suite.repo.UpdateBalanceInTx(tx, balance)
	suite.Assert().NoError(err)

	// Verify update within transaction
	updated, err := suite.repo.GetBalanceInTx(tx, testutils.TestAddress1, testutils.TestTokenPath)
	suite.Assert().NoError(err)
	suite.Assert().Equal("2000", updated.Amount)

	// Should maintain original value after transaction rollback
	tx.Rollback()
	original, err := suite.repo.GetBalance(testutils.TestAddress1, testutils.TestTokenPath)
	suite.Assert().NoError(err)
	suite.Assert().Equal("1000", original.Amount)
}

func TestBalanceRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(BalanceRepositoryTestSuite))
}
