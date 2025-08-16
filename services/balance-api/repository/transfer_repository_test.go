package repository

import (
	"testing"
	"time"

	"shared/infra/database"
	"shared/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDBForTransfer creates a test database for transfer testing
func setupTestDBForTransfer(t *testing.T) (*database.Database, func()) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate the schema
	err = db.AutoMigrate(&models.TokenTransfer{})
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

// createTestTokenTransfer creates a test token transfer
func createTestTokenTransfer(fromAddress, toAddress, tokenPath, amount string, createdAt time.Time) models.TokenTransfer {
	return models.TokenTransfer{
		FromAddress: fromAddress,
		ToAddress:   toAddress,
		TokenPath:   tokenPath,
		Amount:      amount,
		CreatedAt:   createdAt,
	}
}

func TestTransferRepository_GetTransfersByAddress(t *testing.T) {
	db, cleanup := setupTestDBForTransfer(t)
	defer cleanup()

	repo := NewTransferRepository(db)

	// Create test data with different timestamps
	now := time.Now()
	transfer1 := createTestTokenTransfer("g1address1", "g1address2", "gno.land/r/demo/grc20", "1000", now.Add(-3*time.Hour))
	transfer2 := createTestTokenTransfer("g1address2", "g1address1", "gno.land/r/demo/grc20", "500", now.Add(-2*time.Hour))
	transfer3 := createTestTokenTransfer("g1address3", "g1address4", "gno.land/r/demo/grc20", "2000", now.Add(-1*time.Hour))
	transfer4 := createTestTokenTransfer("g1address1", "g1address3", "gno.land/r/demo/grc21", "300", now)

	require.NoError(t, db.Create(&transfer1).Error)
	require.NoError(t, db.Create(&transfer2).Error)
	require.NoError(t, db.Create(&transfer3).Error)
	require.NoError(t, db.Create(&transfer4).Error)

	// Test getting transfers by address (g1address1 should have 3 transfers)
	transfers, err := repo.GetTransfersByAddress("g1address1", 0)
	require.NoError(t, err)
	assert.Len(t, transfers, 3)

	// Check if transfers are ordered by created_at DESC
	assert.True(t, transfers[0].CreatedAt.After(transfers[1].CreatedAt))
	assert.True(t, transfers[1].CreatedAt.After(transfers[2].CreatedAt))

	// Test with limit
	transfers, err = repo.GetTransfersByAddress("g1address1", 2)
	require.NoError(t, err)
	assert.Len(t, transfers, 2)

	// Test with non-existent address
	transfers, err = repo.GetTransfersByAddress("nonexistent", 0)
	require.NoError(t, err)
	assert.Len(t, transfers, 0)
}

func TestTransferRepository_GetAllTransfers(t *testing.T) {
	db, cleanup := setupTestDBForTransfer(t)
	defer cleanup()

	repo := NewTransferRepository(db)

	// Create test data with different timestamps
	now := time.Now()
	transfer1 := createTestTokenTransfer("g1address1", "g1address2", "gno.land/r/demo/grc20", "1000", now.Add(-3*time.Hour))
	transfer2 := createTestTokenTransfer("g1address2", "g1address3", "gno.land/r/demo/grc20", "500", now.Add(-2*time.Hour))
	transfer3 := createTestTokenTransfer("g1address3", "g1address4", "gno.land/r/demo/grc21", "2000", now.Add(-1*time.Hour))
	transfer4 := createTestTokenTransfer("g1address4", "g1address1", "gno.land/r/demo/grc21", "300", now)

	require.NoError(t, db.Create(&transfer1).Error)
	require.NoError(t, db.Create(&transfer2).Error)
	require.NoError(t, db.Create(&transfer3).Error)
	require.NoError(t, db.Create(&transfer4).Error)

	// Test getting all transfers without limit
	transfers, err := repo.GetAllTransfers(0)
	require.NoError(t, err)
	assert.Len(t, transfers, 4)

	// Check if transfers are ordered by created_at DESC
	for i := 0; i < len(transfers)-1; i++ {
		assert.True(t, transfers[i].CreatedAt.After(transfers[i+1].CreatedAt) || transfers[i].CreatedAt.Equal(transfers[i+1].CreatedAt))
	}

	// Test with limit
	transfers, err = repo.GetAllTransfers(2)
	require.NoError(t, err)
	assert.Len(t, transfers, 2)

	// The first transfer should be the most recent one
	assert.Equal(t, "g1address4", transfers[0].FromAddress)
	assert.Equal(t, "g1address1", transfers[0].ToAddress)
}

func TestTransferRepository_GetTransfersByAddress_EdgeCases(t *testing.T) {
	db, cleanup := setupTestDBForTransfer(t)
	defer cleanup()

	repo := NewTransferRepository(db)

	// Test with empty database
	transfers, err := repo.GetTransfersByAddress("g1address1", 0)
	require.NoError(t, err)
	assert.Len(t, transfers, 0)

	// Test with negative limit (should be treated as no limit)
	now := time.Now()
	transfer1 := createTestTokenTransfer("g1address1", "g1address2", "gno.land/r/demo/grc20", "1000", now)
	require.NoError(t, db.Create(&transfer1).Error)

	transfers, err = repo.GetTransfersByAddress("g1address1", -1)
	require.NoError(t, err)
	assert.Len(t, transfers, 1)
}

func TestTransferRepository_GetAllTransfers_EdgeCases(t *testing.T) {
	db, cleanup := setupTestDBForTransfer(t)
	defer cleanup()

	repo := NewTransferRepository(db)

	// Test with empty database
	transfers, err := repo.GetAllTransfers(0)
	require.NoError(t, err)
	assert.Len(t, transfers, 0)

	// Test with negative limit (should be treated as no limit)
	now := time.Now()
	transfer1 := createTestTokenTransfer("g1address1", "g1address2", "gno.land/r/demo/grc20", "1000", now)
	require.NoError(t, db.Create(&transfer1).Error)

	transfers, err = repo.GetAllTransfers(-1)
	require.NoError(t, err)
	assert.Len(t, transfers, 1)
}
