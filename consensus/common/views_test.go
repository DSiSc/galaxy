package common

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewViewChange(t *testing.T) {
	viewChange := NewViewChange()
	assert.NotNil(t, viewChange)
	assert.Equal(t, DefaultViewNum, viewChange.currentView)
	assert.Equal(t, DefaultWalterLevel, viewChange.GetWalterLevel())

	viewNum := viewChange.GetCurrentViewNum()
	assert.Equal(t, DefaultViewNum, viewNum)

	mockCurrentViewNum := uint64(10)
	viewChange.SetCurrentViewNum(mockCurrentViewNum)
	assert.Equal(t, mockCurrentViewNum, viewChange.GetCurrentViewNum())
}

func TestNewRequests(t *testing.T) {
	mockToChange := uint8(2)
	request := NewRequests(mockToChange)
	assert.NotNil(t, request)
	assert.NotNil(t, Viewing, request.GetViewRequestState())
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
	assert.Equal(t, Viewing, state)

	state = change.ReceiveViewRequestByAccount(mockAccounts[1])
	assert.Equal(t, ViewEnd, state)

	mockViewNum = uint64(2)
	change, err = viewChange.AddViewRequest(mockViewNum, mockToChange)
	assert.Nil(t, change)
	assert.NotNil(t, err)
	expect := fmt.Errorf("diff of current view %d and request view %d beyond walter level %d",
		DefaultViewNum, mockViewNum, DefaultWalterLevel)
	assert.Equal(t, expect, err)

	mockViewNum = uint64(1)
	viewRequest := viewChange.GetRequestByViewNum(mockViewNum)
	assert.NotNil(t, viewRequest)
	assert.Equal(t, ViewEnd, viewRequest.state)
	assert.Equal(t, 2, len(viewRequest.GetReceivedAccounts()))
	viewChange.RemoveRequest(mockViewNum)
	viewRequest = viewChange.GetRequestByViewNum(mockViewNum)
	assert.Nil(t, viewRequest)
}
