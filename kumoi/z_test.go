package kumoi

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOmegaKeepAlive(t *testing.T) {
	o := NewOmegaBuilder(engine).Connect().Omega()
	<-time.After(10 * time.Second)
	assert.False(t, o.IsClosed())
	assert.False(t, o.IsDisconnected())
	assert.True(t, o.Close().Await().IsSuccess())
}
