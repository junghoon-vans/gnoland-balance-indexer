package repository

import (
	"testing"

	"shared/infra/database"
	"shared/models"

	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type BalanceRepositoryTestSuite struct {
	suite.Suite
	db   *database.Database
	repo BalanceRepository
}

func (suite *BalanceRepositoryTestSuite) SetupSuite() {
	// Use in-memory SQLite database
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	suite.db = &database.Database{DB: gormDB}

	// Create tables
	err = suite.db.DB.AutoMigrate(&models.TokenBalance{})
	suite.Require().NoError(err)

	suite.repo = NewBalanceRepository(suite.db)
}

func (suite *BalanceRepositoryTestSuite) TearDownTest() {
	// Clean up data after each test
	suite.db.Exec("DELETE FROM token_balances")
}

func (suite *BalanceRepositoryTestSuite) TestCreateBalance() {
	balance := &models.TokenBalance{
		Address:   "g1test123",
		TokenPath: "gno.land/r/demo/grc20",
		Amount:    "1000",
	}

	err := suite.repo.CreateBalance(balance)
	suite.Assert().NoError(err)
	suite.Assert().NotZero(balance.ID)
	suite.Assert().NotZero(balance.CreatedAt)
}

func (suite *BalanceRepositoryTestSuite) TestGetBalance() {
	// Create test data
	balance := &models.TokenBalance{
		Address:   "g1test123",
		TokenPath: "gno.land/r/demo/grc20",
		Amount:    "1000",
	}
	err := suite.repo.CreateBalance(balance)
	suite.Require().NoError(err)

	// Test retrieval
	found, err := suite.repo.GetBalance("g1test123", "gno.land/r/demo/grc20")
	suite.Assert().NoError(err)
	suite.Assert().NotNil(found)
	suite.Assert().Equal("g1test123", found.Address)
	suite.Assert().Equal("gno.land/r/demo/grc20", found.TokenPath)
	suite.Assert().Equal("1000", found.Amount)
}

func (suite *BalanceRepositoryTestSuite) TestGetBalanceNotFound() {
	// Query for non-existent balance
	found, err := suite.repo.GetBalance("g1notfound", "gno.land/r/demo/grc20")
	suite.Assert().Error(err)
	suite.Assert().Nil(found)
}

func (suite *BalanceRepositoryTestSuite) TestUpdateBalance() {
	// Create test data
	balance := &models.TokenBalance{
		Address:   "g1test123",
		TokenPath: "gno.land/r/demo/grc20",
		Amount:    "1000",
	}
	err := suite.repo.CreateBalance(balance)
	suite.Require().NoError(err)

	// Update balance
	balance.Amount = "2000"
	err = suite.repo.UpdateBalance(balance)
	suite.Assert().NoError(err)

	// Verify update
	updated, err := suite.repo.GetBalance("g1test123", "gno.land/r/demo/grc20")
	suite.Assert().NoError(err)
	suite.Assert().Equal("2000", updated.Amount)
}

func (suite *BalanceRepositoryTestSuite) TestGetBalanceInTx() {
	// Create test data
	balance := &models.TokenBalance{
		Address:   "g1test123",
		TokenPath: "gno.land/r/demo/grc20",
		Amount:    "1000",
	}
	err := suite.repo.CreateBalance(balance)
	suite.Require().NoError(err)

	// Query within transaction
	tx := suite.db.Begin()
	found, err := suite.repo.GetBalanceInTx(tx, "g1test123", "gno.land/r/demo/grc20")
	tx.Rollback()

	suite.Assert().NoError(err)
	suite.Assert().NotNil(found)
	suite.Assert().Equal("g1test123", found.Address)
	suite.Assert().Equal("1000", found.Amount)
}

func (suite *BalanceRepositoryTestSuite) TestCreateBalanceInTx() {
	tx := suite.db.Begin()
	defer tx.Rollback()

	balance := &models.TokenBalance{
		Address:   "g1test123",
		TokenPath: "gno.land/r/demo/grc20",
		Amount:    "1000",
	}

	err := suite.repo.CreateBalanceInTx(tx, balance)
	suite.Assert().NoError(err)
	suite.Assert().NotZero(balance.ID)

	// Should be queryable within transaction
	found, err := suite.repo.GetBalanceInTx(tx, "g1test123", "gno.land/r/demo/grc20")
	suite.Assert().NoError(err)
	suite.Assert().NotNil(found)

	// Should not be queryable after transaction rollback
	tx.Rollback()
	found, err = suite.repo.GetBalance("g1test123", "gno.land/r/demo/grc20")
	suite.Assert().Error(err)
	suite.Assert().Nil(found)
}

func (suite *BalanceRepositoryTestSuite) TestUpdateBalanceInTx() {
	// Create test data
	balance := &models.TokenBalance{
		Address:   "g1test123",
		TokenPath: "gno.land/r/demo/grc20",
		Amount:    "1000",
	}
	err := suite.repo.CreateBalance(balance)
	suite.Require().NoError(err)

	tx := suite.db.Begin()
	defer tx.Rollback()

	// Update within transaction
	balance.Amount = "2000"
	err = suite.repo.UpdateBalanceInTx(tx, balance)
	suite.Assert().NoError(err)

	// Verify update within transaction
	updated, err := suite.repo.GetBalanceInTx(tx, "g1test123", "gno.land/r/demo/grc20")
	suite.Assert().NoError(err)
	suite.Assert().Equal("2000", updated.Amount)

	// Should maintain original value after transaction rollback
	tx.Rollback()
	original, err := suite.repo.GetBalance("g1test123", "gno.land/r/demo/grc20")
	suite.Assert().NoError(err)
	suite.Assert().Equal("1000", original.Amount)
}

func TestBalanceRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(BalanceRepositoryTestSuite))
}
