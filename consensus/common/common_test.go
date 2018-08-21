package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Role(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(0, int(Proposing))
	assert.Equal(1, int(Propose))
	assert.Equal(2, int(Approve))
	assert.Equal(3, int(Reject))
	assert.Equal(4, int(Commited))
}
