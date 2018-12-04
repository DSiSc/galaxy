package tools

import (
	"github.com/DSiSc/validator/tools/account"
	"github.com/magiconair/properties/assert"
	"testing"
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
