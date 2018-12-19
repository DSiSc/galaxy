package common

import (
	"fmt"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/validator/tools/account"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewConsensusMap(t *testing.T) {
	plugin := NewConsensusPlugin()
	latestHeight := plugin.GetLatestBlockHeight()
	assert.Equal(t, uint64(0), latestHeight)
	content, err := plugin.GetContentByHash(mockHash)
	assert.Equal(t, fmt.Errorf("content %x not exist, please confirm", mockHash), err)
	assert.Nil(t, content)
	block := &types.Block{
		Header: &types.Header{
			Height: uint64(1),
		},
	}
	plugin.Add(mockHash, block)
	plugin.Add(mockHash, block)
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

	signMap := content.GetSignMap()
	assert.Equal(t, 1, len(signMap))
	assert.Equal(t, mockSignset[0], signMap[mockAccounts[0]])

	contents := content.GetContentPayload()
	assert.NotNil(t, contents)
	assert.Equal(t, block, contents)

	plugin.SetLatestBlockHeight(uint64(1))
	latestHeight = plugin.GetLatestBlockHeight()
	assert.Equal(t, uint64(1), latestHeight)

	plugin.Remove(mockHash)
	content, err = plugin.GetContentByHash(mockHash)
	assert.Equal(t, fmt.Errorf("content %x not exist, please confirm", mockHash), err)
	assert.Nil(t, content)
}

func TestNewResponseNodes(t *testing.T) {
	walterLevel := 2
	responses := NewResponseNodes(walterLevel, mockAccounts[0])
	state := responses.GetResponseState()
	assert.Equal(t, GoOnline, state)

	accounts, state := responses.AddResponseNodes(mockAccounts[0])
	assert.Equal(t, 1, len(accounts))
	assert.Equal(t, GoOnline, state)

	accounts, state = responses.AddResponseNodes(mockAccounts[1])
	assert.Equal(t, 2, len(accounts))
	assert.Equal(t, Online, state)
}

func TestNewOnlineWizard(t *testing.T) {
	wizard := NewOnlineWizard()
	assert.NotNil(t, wizard)

	walterLevel := 3
	blockHeight := uint64(1)
	viewNum := uint64(1)
	accounts, state := wizard.AddOnlineResponse(blockHeight, mockAccounts[:2], walterLevel, mockAccounts[0], viewNum)
	assert.Equal(t, len(mockAccounts[:2]), len(accounts))
	assert.Equal(t, GoOnline, state)

	state = wizard.GetCurrentState()
	assert.Equal(t, GoOnline, state)

	state = wizard.GetCurrentStateByHeight(blockHeight)
	assert.Equal(t, GoOnline, state)

	currentHeight := wizard.GetCurrentHeight()
	assert.Equal(t, blockHeight, currentHeight)

	master := wizard.GetMasterByBlockHeight(blockHeight)
	assert.Equal(t, mockAccounts[0], master)

	state = wizard.GetResponseNodesStateByBlockHeight(uint64(2))
	assert.Equal(t, GoOnline, state)

	state = wizard.GetResponseNodesStateByBlockHeight(uint64(1))
	assert.Equal(t, GoOnline, state)

	accounts, state = wizard.AddOnlineResponse(blockHeight+1, mockAccounts[2:3], walterLevel, mockAccounts[2], viewNum)
	assert.Equal(t, len(mockAccounts[2:3]), len(accounts))
	assert.Equal(t, GoOnline, state)

	currentHeight = wizard.GetCurrentHeight()
	assert.Equal(t, blockHeight+1, currentHeight)

	master = wizard.GetMasterByBlockHeight(currentHeight)
	assert.Equal(t, mockAccounts[2], master)

	accounts, state = wizard.AddOnlineResponse(blockHeight+1, []account.Account{mockAccounts[3]}, walterLevel, mockAccounts[3], viewNum+1)
	assert.Equal(t, len(mockAccounts[2:]), len(accounts))
	assert.Equal(t, GoOnline, state)
}
