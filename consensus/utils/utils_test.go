package utils

import (
	"fmt"
	"github.com/DSiSc/validator/tools/account"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAccountFilter(t *testing.T) {
	var exist = false
	for _, account := range mockAccounts {
		if account == mockAccounts[1] {
			exist = true
			break
		}
	}
	assert.Equal(t, true, exist)

	var blackList = make([]account.Account, 0)
	blackList = append(blackList, mockAccounts[1])
	peers := AccountFilter([]account.Account{mockAccounts[1]}, mockAccounts)
	assert.Equal(t, len(mockAccounts)-1, len(peers))
	exist = false
	for _, account := range peers {
		if account == mockAccounts[1] {
			exist = true
			break
		}
	}
	assert.Equal(t, false, exist)
}

func TestGetAccountById(t *testing.T) {
	account := GetAccountById(mockAccounts, mockAccounts[1].Extension.Id)
	assert.Equal(t, account, mockAccounts[1])
}

func TestAbs(t *testing.T) {
	var x = uint64(2)
	var y = uint64(4)
	z := int64(x - y)
	assert.Equal(t, int64(-2), z)
	z = Abs(x, y)
	assert.Equal(t, int64(2), z)
}

func TestGetAccountWithMinId(t *testing.T) {
	var nodes []account.Account
	nodes = mockAccounts
	minNode := GetAccountWithMinId(nodes)
	assert.Equal(t, mockAccounts[0], minNode)
	timer := time.NewTimer(10 * time.Second)
	_, _, timeSecond := time.Now().Clock()
	fmt.Printf("receive timer and time %v.\n", timeSecond)
	timer.Reset(5 * time.Second)
	index := 0
	for {
		select {
		case <-timer.C:
			_, _, timeSecond := time.Now().Clock()
			fmt.Printf("receive timer and time %v.\n", timeSecond)
			timer.Reset(3 * time.Second)
			index = index + 1
			if index == 3 {
				timer.Stop()
				return
			}
		default:
		}
	}
}
