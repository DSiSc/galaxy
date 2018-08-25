package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Role(t *testing.T) {
	asserts := assert.New(t)
	asserts.Equal(0, int(Proposing))
	asserts.Equal(1, int(Propose))
	asserts.Equal(2, int(Approve))
	asserts.Equal(3, int(Reject))
	asserts.Equal(4, int(Committed))
}
