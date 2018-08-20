package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_GetConfigItem(t *testing.T) {
	assert := assert.New(t)
	config := New(ConfigName)
	assert.NotNil(&config)

	participatesStruct := config.GetConfigItem("participates")
	assert.NotNil(participatesStruct)

	pasedItem := config.GetConfigItem("participates.policy")
	assert.NotNil(pasedItem)
	assert.Equal("solo", pasedItem.(string), "they should be equal")
}
