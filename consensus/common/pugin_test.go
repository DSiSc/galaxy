package common

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

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
