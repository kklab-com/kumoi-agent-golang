package base

import (
	"os"
	"testing"
	"time"

	concurrent "github.com/kklab-com/goth-concurrent"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
	"github.com/stretchr/testify/assert"
)

var appId = ""
var token = ""
var domain = ""
var engine *Engine

const Timeout = time.Second

func TestMain(m *testing.M) {
	appId = os.Getenv("TEST_APP_ID")
	token = os.Getenv("TEST_TOKEN")
	domain = os.Getenv("TEST_DOMAIN")
	conf := NewConfig(appId, token)
	conf.Domain = domain
	engine = NewEngine(conf)
	m.Run()
}

func TestSessionEstablish(t *testing.T) {
	cf := engine.connect()
	assert.NotNil(t, cf.GetTimeout(Timeout))
	assert.NotNil(t, cf.Session())
	s := cf.Session()
	assert.NotEmpty(t, s.GetId())
	assert.NotEmpty(t, s.GetName())
	assert.NotEmpty(t, s.Ping().Get())
	assert.True(t, s.Close().AwaitTimeout(Timeout).IsSuccess())
}

func TestSessionMessage(t *testing.T) {
	rmsg := "test message"
	rs := engine.connect().Session()
	s := engine.connect().Session()
	assert.NotEmpty(t, s)
	assert.False(t, s.IsClosed())
	assert.NotEmpty(t, s.GetId())
	assert.NotEmpty(t, s.Ping().Get())
	vf := concurrent.NewFuture()
	rs.OnMessage(func(msg *omega.TransitFrame) {
		assert.Equal(t, rmsg, msg.GetSessionMessage().Message)
		vf.Completable().Complete(nil)
	})

	srs := s.GetRemoteSession(rs.GetId()).Session()
	srs.SendMessage(rmsg)
	srs.Close().AwaitTimeout(Timeout)
	if _, f := s.(*session).remoteSessions.Load(srs.GetId()); f {
		assert.Fail(t, "found in remoteSessions")
	}

	assert.True(t, vf.AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, s.Close().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, rs.Close().AwaitTimeout(Timeout).IsSuccess())
}
