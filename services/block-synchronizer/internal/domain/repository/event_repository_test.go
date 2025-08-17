package repository

import (
	"testing"

	"shared/pkg/database"
	"shared/pkg/models"
	"shared/pkg/testutils"

	"github.com/stretchr/testify/suite"
)

type EventRepositoryTestSuite struct {
	suite.Suite
	db   *database.Database
	repo EventRepository
}

func (suite *EventRepositoryTestSuite) SetupSuite() {
	db, err := testutils.SetupInMemoryDB()
	suite.Require().NoError(err)

	suite.db = db
	suite.repo = NewEventRepository(suite.db)
}

func (suite *EventRepositoryTestSuite) TearDownTest() {
	testutils.CleanupDatabase(suite.db)
}

func (suite *EventRepositoryTestSuite) TestSaveEvent() {
	// First create a transaction to reference
	tx := testutils.CreateTestTransaction(testutils.TestTransactionHash1, 12345)
	err := suite.db.Create(tx).Error
	suite.Require().NoError(err)

	event := testutils.CreateTestTransactionEvent(tx.ID, testutils.TestEventType)

	err = suite.repo.SaveEvent(event)
	suite.Assert().NoError(err)
	suite.Assert().NotZero(event.ID)

	// Verify event was saved
	var savedEvent models.TransactionEvent
	err = suite.db.First(&savedEvent, event.ID).Error
	suite.Assert().NoError(err)
	suite.Assert().Equal(tx.ID, savedEvent.TransactionID)
	suite.Assert().Equal(testutils.TestEventType, savedEvent.Type)
	suite.Assert().Equal("Transfer", savedEvent.Func)
	suite.Assert().Equal(testutils.TestTokenPath, savedEvent.PkgPath)
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

	event := testutils.CreateTestTransactionEvent(tx.ID, testutils.TestEventType)
	err = suite.db.Create(event).Error
	suite.Require().NoError(err)

	attr := testutils.CreateTestTransactionEventAttr(event.ID, "from", testutils.TestAddress1)

	err = suite.repo.SaveEventAttr(attr)
	suite.Assert().NoError(err)
	suite.Assert().NotZero(attr.ID)

	// Verify attribute was saved
	var savedAttr models.TransactionEventAttr
	err = suite.db.First(&savedAttr, attr.ID).Error
	suite.Assert().NoError(err)
	suite.Assert().Equal(event.ID, savedAttr.EventID)
	suite.Assert().Equal("from", savedAttr.Key)
	suite.Assert().Equal(testutils.TestAddress1, savedAttr.Value)
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

	event := testutils.CreateTestTransactionEvent(tx.ID, testutils.TestEventType)
	err = suite.db.Create(event).Error
	suite.Require().NoError(err)

	attrs := []*models.TransactionEventAttr{
		testutils.CreateTestTransactionEventAttr(event.ID, "from", testutils.TestAddress1),
		testutils.CreateTestTransactionEventAttr(event.ID, "to", testutils.TestAddress2),
		testutils.CreateTestTransactionEventAttr(event.ID, "amount", "1000"),
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
	event := testutils.CreateTestTransactionEvent(999999, testutils.TestEventType) // Non-existent transaction ID

	// This should succeed in SQLite as it doesn't enforce foreign key constraints by default
	// In production with PostgreSQL, this would fail
	err := suite.repo.SaveEvent(event)
	suite.Assert().NoError(err) // SQLite allows this
}

func (suite *EventRepositoryTestSuite) TestSaveEventAttrWithInvalidEventID() {
	attr := testutils.CreateTestTransactionEventAttr(999999, "from", testutils.TestAddress1) // Non-existent event ID

	// This should succeed in SQLite as it doesn't enforce foreign key constraints by default
	// In production with PostgreSQL, this would fail
	err := suite.repo.SaveEventAttr(attr)
	suite.Assert().NoError(err) // SQLite allows this
}

func TestEventRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(EventRepositoryTestSuite))
}
