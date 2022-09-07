package base

import (
	"os"
	"sync/atomic"
	"testing"
	"time"

	concurrent "github.com/kklab-com/goth-concurrent"
	value2 "github.com/kklab-com/goth-kkutil/value"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
	"github.com/stretchr/testify/assert"
)

var appId = ""
var token = ""
var domain = ""
var engine *Engine
var agentBuilder *AgentBuilder

const Timeout = time.Second

func TestMain(m *testing.M) {
	appId = os.Getenv("TEST_APP_ID")
	token = os.Getenv("TEST_TOKEN")
	domain = os.Getenv("TEST_DOMAIN")
	conf := NewConfig(appId, token)
	conf.Domain = domain
	engine = NewEngine(conf)
	agentBuilder = NewAgentBuilder(engine)
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

func TestSessionOperation(t *testing.T) {
	sf := engine.connect()
	assert.NotEmpty(t, sf.Session())
	ss := sf.Session()
	ssp := ss.Ping()
	ssp.Await()
	assert.NotEmpty(t, ss.Ping().TransitFrame())
	assert.NotEmpty(t, ss.Hello().TransitFrame().GetHello())
	assert.NotEmpty(t, ss.ServerTime().TransitFrame().GetServerTime())
	assert.NotEmpty(t, ss.Ch())
	assert.NotEmpty(t, ss.GetEngine())
	assert.NotEmpty(t, ss.GetId())
	assert.NotEmpty(t, ss.GetName())
	assert.NotEmpty(t, ss.GetSubject())
	assert.NotEmpty(t, ss.SetMetadata(NewMetadata(map[string]interface{}{"N": "NV"})).AwaitTimeout(time.Second).IsSuccess())
	assert.NotEmpty(t, ss.GetMetadata().GetFields()["N"].GetStringValue() == "NV")
	assert.True(t, ss.Close().AwaitTimeout(time.Second).IsSuccess())
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
	count := 0
	s.(*session).transitPool.Range(func(key, value any) bool {
		count++
		println(value2.JsonMarshal(value))
		return true
	})

	rs.(*session).transitPool.Range(func(key, value any) bool {
		count++
		println(value2.JsonMarshal(value))
		return true
	})

	assert.True(t, count == 0)
	assert.True(t, s.Close().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, rs.Close().AwaitTimeout(Timeout).IsSuccess())
}

func TestSessionsMessage(t *testing.T) {
	rs1 := engine.connect().Session()
	rs2 := engine.connect().Session()
	s := engine.connect().Session()
	wg := concurrent.WaitGroup{}
	f := concurrent.NewFuture()
	wg.Add(300)
	count := int64(0)
	rs1.OnMessage(func(msg *omega.TransitFrame) {
		wg.Done()
		atomic.AddInt64(&count, 1)
	})

	rs2.OnMessage(func(msg *omega.TransitFrame) {
		wg.Done()
		atomic.AddInt64(&count, 1)
	})

	s.OnMessage(func(msg *omega.TransitFrame) {
		wg.Done()
		atomic.AddInt64(&count, 1)
	})

	go func() {
		wg.Wait()
		f.Completable().Complete(nil)
	}()

	for i := 0; i < 100; i++ {
		s.SendRequest(&omega.TransitFrame_SessionsMessage{SessionsMessage: &omega.SessionsMessage{ToSessions: []string{rs1.GetId(), rs2.GetId(), s.GetId()}, Message: "!!"}})
	}

	assert.True(t, f.AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, count == 300)
	assert.True(t, rs1.Close().AwaitTimeout(time.Second).IsSuccess())
	assert.True(t, s.Close().AwaitTimeout(time.Second).IsSuccess())
}
