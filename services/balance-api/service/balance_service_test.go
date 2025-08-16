package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"balance-api/dto"
	"shared/infra/cache"
	"shared/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockBalanceRepository is a mock implementation of BalanceRepository
type MockBalanceRepository struct {
	mock.Mock
}

func (m *MockBalanceRepository) GetBalancesByAddress(address string) ([]models.TokenBalance, error) {
	args := m.Called(address)
	return args.Get(0).([]models.TokenBalance), args.Error(1)
}

func (m *MockBalanceRepository) GetAllBalances() ([]models.TokenBalance, error) {
	args := m.Called()
	return args.Get(0).([]models.TokenBalance), args.Error(1)
}

func (m *MockBalanceRepository) GetBalancesByTokenPath(tokenPath string) ([]models.TokenBalance, error) {
	args := m.Called(tokenPath)
	return args.Get(0).([]models.TokenBalance), args.Error(1)
}

func (m *MockBalanceRepository) GetBalancesByTokenPathAndAddress(tokenPath, address string) ([]models.TokenBalance, error) {
	args := m.Called(tokenPath, address)
	return args.Get(0).([]models.TokenBalance), args.Error(1)
}

// MockCache is a mock implementation of Cache
type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCache) DeletePattern(ctx context.Context, pattern string) error {
	args := m.Called(ctx, pattern)
	return args.Error(0)
}

func (m *MockCache) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *MockCache) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestBalanceService_GetTokenBalances_WithAddress(t *testing.T) {
	mockRepo := new(MockBalanceRepository)
	mockCache := new(MockCache)
	service := NewBalanceService(mockRepo, mockCache)

	// Test data
	balances := []models.TokenBalance{
		{Address: "g1abcdefghijklmnopqrstuvwxyz1234567890ab", TokenPath: "gno.land/r/demo/grc20", Amount: "1000"},
		{Address: "g1abcdefghijklmnopqrstuvwxyz1234567890ab", TokenPath: "gno.land/r/demo/grc21", Amount: "2000"},
	}

	// Mock cache miss
	mockCache.On("Get", mock.Anything, "balance:address:g1abcdefghijklmnopqrstuvwxyz1234567890ab", mock.Anything).Return(cache.ErrCacheMiss)
	mockCache.On("Set", mock.Anything, "balance:address:g1abcdefghijklmnopqrstuvwxyz1234567890ab", mock.Anything, mock.Anything).Return(nil)

	mockRepo.On("GetBalancesByAddress", "g1abcdefghijklmnopqrstuvwxyz1234567890ab").Return(balances, nil)

	req := dto.BalanceRequest{Address: "g1abcdefghijklmnopqrstuvwxyz1234567890ab"}
	response, err := service.GetTokenBalances(req)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Len(t, response.Balances, 2)
	assert.Equal(t, "gno.land/r/demo/grc20", response.Balances[0].TokenPath)
	assert.Equal(t, "1000", response.Balances[0].Amount)
	assert.Equal(t, "gno.land/r/demo/grc21", response.Balances[1].TokenPath)
	assert.Equal(t, "2000", response.Balances[1].Amount)

	mockRepo.AssertExpectations(t)
}

func TestBalanceService_GetTokenBalances_WithInvalidAddress(t *testing.T) {
	mockRepo := new(MockBalanceRepository)
	mockCache := new(MockCache)
	// Mock cache miss for simplicity in tests
	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(cache.ErrCacheMiss)
	mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(nil)
	service := NewBalanceService(mockRepo, mockCache)

	req := dto.BalanceRequest{Address: "invalid_address"}
	response, err := service.GetTokenBalances(req)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "invalid address format")

	mockRepo.AssertExpectations(t)
}

func TestBalanceService_GetTokenBalances_WithoutAddress(t *testing.T) {
	mockRepo := new(MockBalanceRepository)
	mockCache := new(MockCache)
	// Mock cache miss for simplicity in tests
	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(cache.ErrCacheMiss)
	mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(nil)
	service := NewBalanceService(mockRepo, mockCache)

	// Test data - multiple balances for same token path
	balances := []models.TokenBalance{
		{Address: "g1abcdefghijklmnopqrstuvwxyz1234567890ab", TokenPath: "gno.land/r/demo/grc20", Amount: "1000"},
		{Address: "g1abcdefghijklmnopqrstuvwxyz1234567890cd", TokenPath: "gno.land/r/demo/grc20", Amount: "2000"},
		{Address: "g1abcdefghijklmnopqrstuvwxyz1234567890ef", TokenPath: "gno.land/r/demo/grc21", Amount: "500"},
	}

	mockRepo.On("GetAllBalances").Return(balances, nil)

	req := dto.BalanceRequest{Address: ""}
	response, err := service.GetTokenBalances(req)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Len(t, response.Balances, 2)

	// Check aggregated amounts
	for _, balance := range response.Balances {
		if balance.TokenPath == "gno.land/r/demo/grc20" {
			assert.Equal(t, "3000", balance.Amount) // 1000 + 2000
		} else if balance.TokenPath == "gno.land/r/demo/grc21" {
			assert.Equal(t, "500", balance.Amount)
		}
	}

	mockRepo.AssertExpectations(t)
}

func TestBalanceService_GetTokenBalances_RepositoryError(t *testing.T) {
	mockRepo := new(MockBalanceRepository)
	mockCache := new(MockCache)
	// Mock cache miss for simplicity in tests
	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(cache.ErrCacheMiss)
	mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(nil)
	service := NewBalanceService(mockRepo, mockCache)

	mockRepo.On("GetBalancesByAddress", "g1abcdefghijklmnopqrstuvwxyz1234567890ab").Return([]models.TokenBalance{}, errors.New("database error"))

	req := dto.BalanceRequest{Address: "g1abcdefghijklmnopqrstuvwxyz1234567890ab"}
	response, err := service.GetTokenBalances(req)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to fetch balances")

	mockRepo.AssertExpectations(t)
}

func TestBalanceService_GetTokenBalancesByPath_WithAddress(t *testing.T) {
	mockRepo := new(MockBalanceRepository)
	mockCache := new(MockCache)
	// Mock cache miss for simplicity in tests
	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(cache.ErrCacheMiss)
	mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(nil)
	service := NewBalanceService(mockRepo, mockCache)

	// Test data
	balances := []models.TokenBalance{
		{Address: "g1abcdefghijklmnopqrstuvwxyz1234567890ab", TokenPath: "gno.land/r/demo/grc20", Amount: "1000"},
	}

	mockRepo.On("GetBalancesByTokenPathAndAddress", "gno.land/r/demo/grc20", "g1abcdefghijklmnopqrstuvwxyz1234567890ab").Return(balances, nil)

	response, err := service.GetTokenBalancesByPath("gno.land/r/demo/grc20", "g1abcdefghijklmnopqrstuvwxyz1234567890ab")

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Len(t, response.AccountBalances, 1)
	assert.Equal(t, "g1abcdefghijklmnopqrstuvwxyz1234567890ab", response.AccountBalances[0].Address)
	assert.Equal(t, "gno.land/r/demo/grc20", response.AccountBalances[0].TokenPath)
	assert.Equal(t, "1000", response.AccountBalances[0].Amount)

	mockRepo.AssertExpectations(t)
}

func TestBalanceService_GetTokenBalancesByPath_WithoutAddress(t *testing.T) {
	mockRepo := new(MockBalanceRepository)
	mockCache := new(MockCache)
	// Mock cache miss for simplicity in tests
	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(cache.ErrCacheMiss)
	mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(nil)
	service := NewBalanceService(mockRepo, mockCache)

	// Test data
	balances := []models.TokenBalance{
		{Address: "g1abcdefghijklmnopqrstuvwxyz1234567890ab", TokenPath: "gno.land/r/demo/grc20", Amount: "1000"},
		{Address: "g1abcdefghijklmnopqrstuvwxyz1234567890cd", TokenPath: "gno.land/r/demo/grc20", Amount: "2000"},
	}

	mockRepo.On("GetBalancesByTokenPath", "gno.land/r/demo/grc20").Return(balances, nil)

	response, err := service.GetTokenBalancesByPath("gno.land/r/demo/grc20", "")

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Len(t, response.AccountBalances, 2)
	assert.Equal(t, "g1abcdefghijklmnopqrstuvwxyz1234567890ab", response.AccountBalances[0].Address)
	assert.Equal(t, "g1abcdefghijklmnopqrstuvwxyz1234567890cd", response.AccountBalances[1].Address)

	mockRepo.AssertExpectations(t)
}

func TestBalanceService_GetTokenBalancesByPath_WithURLEncodedPath(t *testing.T) {
	mockRepo := new(MockBalanceRepository)
	mockCache := new(MockCache)
	// Mock cache miss for simplicity in tests
	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(cache.ErrCacheMiss)
	mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(nil)
	service := NewBalanceService(mockRepo, mockCache)

	// Test data
	balances := []models.TokenBalance{
		{Address: "g1abcdefghijklmnopqrstuvwxyz1234567890ab", TokenPath: "gno.land/r/demo/grc20", Amount: "1000"},
	}

	// Mock expects the decoded path
	mockRepo.On("GetBalancesByTokenPath", "gno.land/r/demo/grc20").Return(balances, nil)

	// Pass URL-encoded path
	response, err := service.GetTokenBalancesByPath("gno.land%2Fr%2Fdemo%2Fgrc20", "")

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Len(t, response.AccountBalances, 1)
	assert.Equal(t, "gno.land/r/demo/grc20", response.AccountBalances[0].TokenPath)

	mockRepo.AssertExpectations(t)
}

func TestBalanceService_GetTokenBalancesByPath_WithInvalidAddress(t *testing.T) {
	mockRepo := new(MockBalanceRepository)
	mockCache := new(MockCache)
	// Mock cache miss for simplicity in tests
	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(cache.ErrCacheMiss)
	mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(nil)
	service := NewBalanceService(mockRepo, mockCache)

	response, err := service.GetTokenBalancesByPath("gno.land/r/demo/grc20", "invalid_address")

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "invalid address format")

	mockRepo.AssertExpectations(t)
}

func TestBalanceService_GetTokenBalancesByPath_RepositoryError(t *testing.T) {
	mockRepo := new(MockBalanceRepository)
	mockCache := new(MockCache)
	// Mock cache miss for simplicity in tests
	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(cache.ErrCacheMiss)
	mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(nil)
	service := NewBalanceService(mockRepo, mockCache)

	mockRepo.On("GetBalancesByTokenPath", "gno.land/r/demo/grc20").Return([]models.TokenBalance{}, errors.New("database error"))

	response, err := service.GetTokenBalancesByPath("gno.land/r/demo/grc20", "")

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to fetch balances")

	mockRepo.AssertExpectations(t)
}
