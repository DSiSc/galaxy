package participates

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_NewParticipatePolicy(t *testing.T) {
	assert := assert.New(t)
	participate, err := NewParticipatePolicy()
	assert.NotNil(participate)
	assert.Nil(err)
}
