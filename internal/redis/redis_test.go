package redisclient

import (
	"fmt"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func TestRedisConnectionCheck(t *testing.T) {
	db, mock := redismock.NewClientMock()

	// Scenario 1: Successful ping
	mock.ExpectPing().SetVal("PONG")
	Rdb = db
	err := RedisConnectionCheck()
	defer db.Close()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Resetting mock for the next scenario
	db, mock = redismock.NewClientMock()

	// Scenario 2: Failed ping
	mock.ExpectPing().SetErr(fmt.Errorf("failed to ping"))
	Rdb = db
	err = RedisConnectionCheck()
	defer db.Close()
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
