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
	consensusMap := NewConsensusMap()
	consensusMap.Add(mockHash)
	mapInfo := consensusMap.GetConsensusMap()
	_, ok := mapInfo[mockHash]
	assert.Equal(t, true, ok)
	content := consensusMap.GetConsensusContentByHash(mockHash)
	assert.Equal(t, mockHash, content.digest)
	assert.Equal(t, Initial, content.state)
	content.SetState(ToConsensus)
	assert.Equal(t, Initial, content.state)
	content.SetState(InConsensus)
	assert.Equal(t, InConsensus, content.state)
	ok, _ = content.GetSignByAccount(mockAccounts[0])
	assert.Equal(t, false, ok)
	content.AddSignature(mockAccounts[0], mockSignset[0])
	ok, sign := content.GetSignByAccount(mockAccounts[0])
	assert.Equal(t, true, ok)
	assert.Equal(t, mockSignset[0], sign)
	assert.Equal(t, 1, content.Signatures())
	content.AddSignature(mockAccounts[0], mockSignset[0])
	block := &types.Block{
		Header: &types.Header{
			Height:    uint64(1),
			MixDigest: mockHash,
		},
		HeaderHash: mockHash,
	}
	content.SetContentByHash(errorHash, block)
	content.SetContentByHash(mockHash, block)
	contents, err := content.GetContentByHash(errorHash)
	assert.Equal(t, fmt.Errorf("record not exist with digest %v", errorHash), err)
	assert.Nil(t, contents)

	contents, err = content.GetContentByHash(mockHash)
	assert.Nil(t, err)
	assert.Equal(t, block, contents)

	sigMap, err := content.GetSignMapByHash(mockHash)
	assert.Nil(t, err)
	assert.Equal(t, mockSignset[0], sigMap[mockAccounts[0]])

	sigMap1, err1 := consensusMap.GetConsensusContentSigMapByHash(mockHash)
	assert.Equal(t, err, err1)
	assert.Equal(t, sigMap, sigMap1)
}
