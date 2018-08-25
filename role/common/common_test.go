package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Role(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(0, int(Master))
	assert.Equal(1, int(Slave))
	assert.Equal(2, int(Normal))
	assert.Equal(3, int(UnKnown))
}
