package kumoi

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOmegaKeepAlive(t *testing.T) {
	o := NewOmegaBuilder(conf).Connect().Get()
	<-time.After(10 * time.Second)
	assert.False(t, o.IsClosed())
	assert.True(t, o.Close().Await().IsSuccess())
}
