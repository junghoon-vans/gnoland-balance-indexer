package testhelpers

import (
	"context"
	"time"

	"block-synchronizer/dto"
	"shared/infra/database"
	"shared/models"

	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupInMemoryDB creates an in-memory SQLite database for testing
func SetupInMemoryDB() (*database.Database, error) {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	db := &database.Database{DB: gormDB}

	// Auto-migrate all models
	err = db.DB.AutoMigrate(
		&models.Block{},
		&models.Transaction{},
		&models.TransactionMsg{},
		&models.TransactionEvent{},
		&models.TransactionEventAttr{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// CleanupDatabase cleans all tables in the test database
func CleanupDatabase(db *database.Database) {
	db.Exec("DELETE FROM transaction_event_attrs")
	db.Exec("DELETE FROM transaction_events")
	db.Exec("DELETE FROM transaction_msgs")
	db.Exec("DELETE FROM transactions")
	db.Exec("DELETE FROM blocks")
}

// CreateTestBlock creates a test block with default values
func CreateTestBlock(height int64, hash string) *models.Block {
	return &models.Block{
		Hash:     hash,
		Height:   height,
		Time:     time.Now(),
		NumTxs:   5,
		TotalTxs: 100,
	}
}

// CreateTestTransaction creates a test transaction with default values
func CreateTestTransaction(hash string, blockHeight int64) *models.Transaction {
	return &models.Transaction{
		Hash:        hash,
		Index:       0,
		BlockHeight: blockHeight,
		Success:     true,
		GasWanted:   100000,
		GasUsed:     50000,
		Memo:        "test transaction",
	}
}

// CreateTestTransactionEvent creates a test transaction event with default values
func CreateTestTransactionEvent(transactionID uint, eventType string) *models.TransactionEvent {
	return &models.TransactionEvent{
		TransactionID: transactionID,
		Type:          eventType,
		Func:          "Transfer",
		PkgPath:       "gno.land/r/demo/grc20",
	}
}

// CreateTestTransactionEventAttr creates a test transaction event attribute
func CreateTestTransactionEventAttr(eventID uint, key, value string) *models.TransactionEventAttr {
	return &models.TransactionEventAttr{
		EventID: eventID,
		Key:     key,
		Value:   value,
	}
}

// CreateTestGraphQLBlock creates a test GraphQL block with default values
func CreateTestGraphQLBlock(height int64, hash string) *dto.GraphQLBlock {
	return &dto.GraphQLBlock{
		Hash:     hash,
		Height:   height,
		Time:     time.Now().Format(time.RFC3339),
		NumTxs:   5,
		TotalTxs: 10,
	}
}

// CreateTestGraphQLTransaction creates a test GraphQL transaction with default values
func CreateTestGraphQLTransaction(hash string) *dto.GraphQLTransaction {
	return &dto.GraphQLTransaction{
		Hash: hash,
	}
}

// CreateTestGraphQLEvent creates a test GraphQL event with default values
func CreateTestGraphQLEvent(eventType string, attrs []dto.GraphQLEventAttr) *dto.GraphQLEvent {
	return &dto.GraphQLEvent{
		Type:  eventType,
		Attrs: attrs,
	}
}

// CreateTestGraphQLEventAttrs creates test GraphQL event attributes
func CreateTestGraphQLEventAttrs(count int) []dto.GraphQLEventAttr {
	attrs := make([]dto.GraphQLEventAttr, count)
	for i := 0; i < count; i++ {
		attrs[i] = dto.GraphQLEventAttr{
			Key:   "key" + string(rune('0'+i)),
			Value: "value" + string(rune('0'+i)),
		}
	}
	return attrs
}

// CreateTestContext creates a test context with timeout
func CreateTestContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	return ctx
}

// MockExpectation represents a mock expectation helper
type MockExpectation struct {
	Mock   *mock.Mock
	Method string
	Args   []interface{}
	Return []interface{}
	Times  int
}

// SetupMockExpectations sets up multiple mock expectations at once
func SetupMockExpectations(expectations []MockExpectation) {
	for _, exp := range expectations {
		call := exp.Mock.On(exp.Method, exp.Args...)
		if len(exp.Return) > 0 {
			call.Return(exp.Return...)
		}
		if exp.Times > 0 {
			if exp.Times == 1 {
				call.Once()
			} else {
				call.Times(exp.Times)
			}
		}
	}
}

// AssertMockExpectations asserts all mock expectations
func AssertMockExpectations(t mock.TestingT, mocks ...*mock.Mock) {
	for _, m := range mocks {
		m.AssertExpectations(t)
	}
}

// Common test data constants
const (
	TestBlockHash1       = "0x123abc"
	TestBlockHash2       = "0x456def"
	TestTransactionHash1 = "0x789ghi"
	TestTransactionHash2 = "0xabcdef"
	TestAddress1         = "g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d"
	TestAddress2         = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
	TestTokenPath        = "gno.land/r/demo/grc20"
	TestEventType        = "Transfer"
)
