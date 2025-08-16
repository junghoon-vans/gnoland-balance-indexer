package repository

import (
	"testing"

	"shared/infra/database"
	"shared/models"

	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type EventRepositoryTestSuite struct {
	suite.Suite
	db   *database.Database
	repo EventRepository
}

func (suite *EventRepositoryTestSuite) SetupSuite() {
	// Use in-memory SQLite database
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	suite.db = &database.Database{DB: gormDB}

	// Create tables
	err = suite.db.DB.AutoMigrate(&models.Transaction{}, &models.TransactionEvent{}, &models.TransactionEventAttr{})
	suite.Require().NoError(err)

	suite.repo = NewEventRepository(suite.db)
}

func (suite *EventRepositoryTestSuite) TearDownTest() {
	// Clean up data after each test
	suite.db.Exec("DELETE FROM transaction_event_attrs")
	suite.db.Exec("DELETE FROM transaction_events")
	suite.db.Exec("DELETE FROM transactions")
}

func (suite *EventRepositoryTestSuite) TestSaveEvent() {
	// First create a transaction to reference
	tx := &models.Transaction{
		Hash:        "0x123abc",
		Index:       0,
		BlockHeight: 12345,
		Success:     true,
		GasWanted:   100000,
		GasUsed:     50000,
		Memo:        "test transaction",
	}
	err := suite.db.Create(tx).Error
	suite.Require().NoError(err)

	event := &models.TransactionEvent{
		TransactionID: tx.ID,
		Type:          "gno.land/r/demo/grc20",
		Func:          "Transfer",
		PkgPath:       "gno.land/r/demo/grc20",
	}

	err = suite.repo.SaveEvent(event)
	suite.Assert().NoError(err)
	suite.Assert().NotZero(event.ID)

	// Verify event was saved
	var savedEvent models.TransactionEvent
	err = suite.db.First(&savedEvent, event.ID).Error
	suite.Assert().NoError(err)
	suite.Assert().Equal(tx.ID, savedEvent.TransactionID)
	suite.Assert().Equal("gno.land/r/demo/grc20", savedEvent.Type)
	suite.Assert().Equal("Transfer", savedEvent.Func)
	suite.Assert().Equal("gno.land/r/demo/grc20", savedEvent.PkgPath)
}

func (suite *EventRepositoryTestSuite) TestSaveEventAttr() {
	// First create a transaction and event to reference
	tx := &models.Transaction{
		Hash:        "0x123abc",
		Index:       0,
		BlockHeight: 12345,
		Success:     true,
		GasWanted:   100000,
		GasUsed:     50000,
		Memo:        "test transaction",
	}
	err := suite.db.Create(tx).Error
	suite.Require().NoError(err)

	event := &models.TransactionEvent{
		TransactionID: tx.ID,
		Type:          "gno.land/r/demo/grc20",
		Func:          "Transfer",
		PkgPath:       "gno.land/r/demo/grc20",
	}
	err = suite.db.Create(event).Error
	suite.Require().NoError(err)

	attr := &models.TransactionEventAttr{
		EventID: event.ID,
		Key:     "from",
		Value:   "g1from123",
	}

	err = suite.repo.SaveEventAttr(attr)
	suite.Assert().NoError(err)
	suite.Assert().NotZero(attr.ID)

	// Verify attribute was saved
	var savedAttr models.TransactionEventAttr
	err = suite.db.First(&savedAttr, attr.ID).Error
	suite.Assert().NoError(err)
	suite.Assert().Equal(event.ID, savedAttr.EventID)
	suite.Assert().Equal("from", savedAttr.Key)
	suite.Assert().Equal("g1from123", savedAttr.Value)
}

func (suite *EventRepositoryTestSuite) TestSaveMultipleEventAttrs() {
	// First create a transaction and event to reference
	tx := &models.Transaction{
		Hash:        "0x123abc",
		Index:       0,
		BlockHeight: 12345,
		Success:     true,
		GasWanted:   100000,
		GasUsed:     50000,
		Memo:        "test transaction",
	}
	err := suite.db.Create(tx).Error
	suite.Require().NoError(err)

	event := &models.TransactionEvent{
		TransactionID: tx.ID,
		Type:          "gno.land/r/demo/grc20",
		Func:          "Transfer",
		PkgPath:       "gno.land/r/demo/grc20",
	}
	err = suite.db.Create(event).Error
	suite.Require().NoError(err)

	attrs := []*models.TransactionEventAttr{
		{
			EventID: event.ID,
			Key:     "from",
			Value:   "g1from123",
		},
		{
			EventID: event.ID,
			Key:     "to",
			Value:   "g1to456",
		},
		{
			EventID: event.ID,
			Key:     "amount",
			Value:   "1000",
		},
	}

	// Save all attributes
	for _, attr := range attrs {
		err := suite.repo.SaveEventAttr(attr)
		suite.Assert().NoError(err)
		suite.Assert().NotZero(attr.ID)
	}

	// Verify all attributes were saved
	var count int64
	suite.db.Model(&models.TransactionEventAttr{}).Where("event_id = ?", event.ID).Count(&count)
	suite.Assert().Equal(int64(3), count)
}

func (suite *EventRepositoryTestSuite) TestSaveEventWithInvalidTransactionID() {
	event := &models.TransactionEvent{
		TransactionID: 999999, // Non-existent transaction ID
		Type:          "gno.land/r/demo/grc20",
		Func:          "Transfer",
		PkgPath:       "gno.land/r/demo/grc20",
	}

	// This should succeed in SQLite as it doesn't enforce foreign key constraints by default
	// In production with PostgreSQL, this would fail
	err := suite.repo.SaveEvent(event)
	suite.Assert().NoError(err) // SQLite allows this
}

func (suite *EventRepositoryTestSuite) TestSaveEventAttrWithInvalidEventID() {
	attr := &models.TransactionEventAttr{
		EventID: 999999, // Non-existent event ID
		Key:     "from",
		Value:   "g1from123",
	}

	// This should succeed in SQLite as it doesn't enforce foreign key constraints by default
	// In production with PostgreSQL, this would fail
	err := suite.repo.SaveEventAttr(attr)
	suite.Assert().NoError(err) // SQLite allows this
}

func TestEventRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(EventRepositoryTestSuite))
}
