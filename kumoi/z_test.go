package kumoi

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOmegaKeepAlive(t *testing.T) {
	o := NewOmegaBuilder(conf).Connect().Omega()
	<-time.After(50 * time.Second)
	assert.False(t, o.IsClosed())
	assert.False(t, o.IsDisconnected())
	assert.True(t, o.Close().Await().IsSuccess())
}
