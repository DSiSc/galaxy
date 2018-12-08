package tools

import (
	"fmt"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus/common"
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

func TestAbs(t *testing.T) {
	var x = uint64(2)
	var y = uint64(4)
	z := int64(x - y)
	assert.Equal(t, int64(-2), z)
	z = Abs(x, y)
	assert.Equal(t, int64(2), z)
}

func TestNewViewChange(t *testing.T) {
	viewChange := NewViewChange()
	assert.NotNil(t, viewChange)
	assert.Equal(t, common.DefaultViewNum, viewChange.currentView)
	assert.Equal(t, common.DefaultWalterLevel, viewChange.GetWalterLevel())

	viewNum := viewChange.GetCurrentViewNum()
	assert.Equal(t, common.DefaultViewNum, viewNum)

	mockCurrentViewNum := uint64(10)
	viewChange.SetCurrentViewNum(mockCurrentViewNum)
	assert.Equal(t, mockCurrentViewNum, viewChange.GetCurrentViewNum())
}

func TestNewRequests(t *testing.T) {
	mockToChange := uint8(2)
	request := NewRequests(mockToChange)
	assert.NotNil(t, request)
	assert.NotNil(t, common.Viewing, request.GetViewRequestState())
	assert.NotNil(t, mockToChange, request.toChange)
	assert.NotNil(t, 0, len(request.nodes))
}

func TestViewChange_AddViewRequest(t *testing.T) {
	viewChange := NewViewChange()
	mockViewNum := uint64(1)
	mockToChange := uint8(2)
	change, err := viewChange.AddViewRequest(mockViewNum, mockToChange)
	assert.Nil(t, err)
	assert.NotNil(t, change)

	change, err = viewChange.AddViewRequest(mockViewNum, mockToChange)
	assert.Nil(t, err)
	assert.NotNil(t, change)

	state := change.ReceiveViewRequestByAccount(mockAccounts[0])
	assert.Equal(t, common.Viewing, state)

	state = change.ReceiveViewRequestByAccount(mockAccounts[1])
	assert.Equal(t, common.ViewEnd, state)

	mockViewNum = uint64(2)
	change, err = viewChange.AddViewRequest(mockViewNum, mockToChange)
	assert.Nil(t, change)
	assert.NotNil(t, err)
	expect := fmt.Errorf("diff of current view %d and request view %d beyond walter level %d",
		common.DefaultViewNum, mockViewNum, common.DefaultWalterLevel)
	assert.Equal(t, expect, err)
}
