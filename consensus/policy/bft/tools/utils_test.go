package tools

import (
	"fmt"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/validator/tools/account"
	"github.com/stretchr/testify/assert"
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

var errorHash = types.Hash{
	0xbe, 0x79, 0x1d, 0x4a, 0xf9, 0x64, 0x8f, 0xc3, 0x7f, 0x94, 0xeb, 0x36, 0x53, 0x19, 0xf6, 0xd0,
	0xa9, 0x78, 0x9f, 0x9c, 0x22, 0x47, 0x2c, 0xa7, 0xa6, 0x12, 0xa9, 0xca, 0x4, 0x13, 0xc1, 0x4,
}

func TestNewConsensusMap(t *testing.T) {
	plugin := NewConsensusPlugin()

	content, err := plugin.GetContentByHash(mockHash)
	assert.Equal(t, fmt.Errorf("content %x not exist, please confirm", mockHash), err)
	assert.Nil(t, content)

	plugin.Add(mockHash, nil)
	content, err = plugin.GetContentByHash(mockHash)
	assert.Nil(t, err)
	assert.NotNil(t, content)

	content, err = plugin.GetContentByHash(mockHash)
	assert.Nil(t, err)
	assert.NotNil(t, content)
	assert.NotNil(t, Initial, content.State())
	err = content.SetState(ToConsensus)
	assert.Equal(t, err, fmt.Errorf("can not move state from %v to %v", Initial, ToConsensus))
	assert.NotNil(t, Initial, content.State())

	err = content.SetState(InConsensus)
	assert.Nil(t, err)
	assert.NotNil(t, InConsensus, content.State())

	signatures := content.Signatures()
	assert.Equal(t, 0, len(signatures))

	ok := content.AddSignature(mockAccounts[0], mockSignset[0])
	assert.Equal(t, true, ok)

	ok = content.AddSignature(mockAccounts[0], mockSignset[0])
	assert.Equal(t, false, ok)

	sign, ok := content.GetSignByAccount(mockAccounts[1])
	assert.Equal(t, false, ok)

	sign, ok = content.GetSignByAccount(mockAccounts[0])
	assert.Equal(t, true, ok)
	assert.Equal(t, mockSignset[0], sign)

	plugin.Remove(mockHash)
	content, err = plugin.GetContentByHash(mockHash)
	assert.Equal(t, fmt.Errorf("content %x not exist, please confirm", mockHash), err)
	assert.Nil(t, content)
}
