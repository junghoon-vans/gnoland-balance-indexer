package repository

import (
	"testing"

	"shared/pkg/database"
	"shared/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates a test database for testing
func setupTestDB(t *testing.T) (*database.Database, func()) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate the schema
	err = db.AutoMigrate(&models.TokenBalance{})
	require.NoError(t, err)

	testDB := &database.Database{DB: db}

	cleanup := func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}

	return testDB, cleanup
}

// createTestTokenBalance creates a test token balance
func createTestTokenBalance(address, tokenPath, amount string) models.TokenBalance {
	return models.TokenBalance{
		Address:   address,
		TokenPath: tokenPath,
		Amount:    amount,
	}
}

func TestBalanceRepository_GetBalancesByAddress(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBalanceRepository(db)

	// Create test data
	balance1 := createTestTokenBalance("g1address1", "gno.land/r/demo/grc20", "1000")
	balance2 := createTestTokenBalance("g1address1", "gno.land/r/demo/grc21", "2000")
	balance3 := createTestTokenBalance("g1address2", "gno.land/r/demo/grc20", "3000")
	zeroBalance := createTestTokenBalance("g1address1", "gno.land/r/demo/grc22", "0")

	require.NoError(t, db.Create(&balance1).Error)
	require.NoError(t, db.Create(&balance2).Error)
	require.NoError(t, db.Create(&balance3).Error)
	require.NoError(t, db.Create(&zeroBalance).Error)

	// Test getting balances by address
	balances, err := repo.GetBalancesByAddress("g1address1")
	require.NoError(t, err)

	// Should return 2 balances (excluding zero balance)
	assert.Len(t, balances, 2)
	assert.Equal(t, "g1address1", balances[0].Address)
	assert.Equal(t, "g1address1", balances[1].Address)

	// Test with non-existent address
	balances, err = repo.GetBalancesByAddress("nonexistent")
	require.NoError(t, err)
	assert.Len(t, balances, 0)
}

func TestBalanceRepository_GetAllBalances(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBalanceRepository(db)

	// Create test data
	balance1 := createTestTokenBalance("g1address1", "gno.land/r/demo/grc20", "1000")
	balance2 := createTestTokenBalance("g1address2", "gno.land/r/demo/grc21", "2000")
	zeroBalance := createTestTokenBalance("g1address3", "gno.land/r/demo/grc22", "0")

	require.NoError(t, db.Create(&balance1).Error)
	require.NoError(t, db.Create(&balance2).Error)
	require.NoError(t, db.Create(&zeroBalance).Error)

	// Test getting all balances
	balances, err := repo.GetAllBalances()
	require.NoError(t, err)

	// Should return 2 balances (excluding zero balance)
	assert.Len(t, balances, 2)
}

func TestBalanceRepository_GetBalancesByTokenPath(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBalanceRepository(db)

	// Create test data
	balance1 := createTestTokenBalance("g1address1", "gno.land/r/demo/grc20", "1000")
	balance2 := createTestTokenBalance("g1address2", "gno.land/r/demo/grc20", "2000")
	balance3 := createTestTokenBalance("g1address3", "gno.land/r/demo/grc21", "3000")
	zeroBalance := createTestTokenBalance("g1address4", "gno.land/r/demo/grc20", "0")

	require.NoError(t, db.Create(&balance1).Error)
	require.NoError(t, db.Create(&balance2).Error)
	require.NoError(t, db.Create(&balance3).Error)
	require.NoError(t, db.Create(&zeroBalance).Error)

	// Test getting balances by token path
	balances, err := repo.GetBalancesByTokenPath("gno.land/r/demo/grc20")
	require.NoError(t, err)

	// Should return 2 balances for grc20 (excluding zero balance)
	assert.Len(t, balances, 2)
	for _, balance := range balances {
		assert.Equal(t, "gno.land/r/demo/grc20", balance.TokenPath)
	}

	// Test with non-existent token path
	balances, err = repo.GetBalancesByTokenPath("nonexistent")
	require.NoError(t, err)
	assert.Len(t, balances, 0)
}

func TestBalanceRepository_GetBalancesByTokenPathAndAddress(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBalanceRepository(db)

	// Create test data
	balance1 := createTestTokenBalance("g1address1", "gno.land/r/demo/grc20", "1000")
	balance2 := createTestTokenBalance("g1address1", "gno.land/r/demo/grc21", "2000")
	balance3 := createTestTokenBalance("g1address2", "gno.land/r/demo/grc20", "3000")
	zeroBalance := createTestTokenBalance("g1address1", "gno.land/r/demo/grc22", "0")

	require.NoError(t, db.Create(&balance1).Error)
	require.NoError(t, db.Create(&balance2).Error)
	require.NoError(t, db.Create(&balance3).Error)
	require.NoError(t, db.Create(&zeroBalance).Error)

	// Test getting balance by token path and address
	balances, err := repo.GetBalancesByTokenPathAndAddress("gno.land/r/demo/grc20", "g1address1")
	require.NoError(t, err)

	// Should return 1 balance
	assert.Len(t, balances, 1)
	assert.Equal(t, "g1address1", balances[0].Address)
	assert.Equal(t, "gno.land/r/demo/grc20", balances[0].TokenPath)
	assert.Equal(t, "1000", balances[0].Amount)

	// Test with zero balance
	balances, err = repo.GetBalancesByTokenPathAndAddress("gno.land/r/demo/grc22", "g1address1")
	require.NoError(t, err)
	assert.Len(t, balances, 0)

	// Test with non-existent combination
	balances, err = repo.GetBalancesByTokenPathAndAddress("nonexistent", "nonexistent")
	require.NoError(t, err)
	assert.Len(t, balances, 0)
}
