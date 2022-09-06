package base

import (
	"fmt"
	"time"
)

const (
	DefaultServiceScheme         = "https"
	DefaultServiceDomain         = "omega.kklab.com"
	DefaultConnectTimeout        = 10 * time.Second
	DefaultTransitTimeout        = 10 * time.Second
	DefaultKeepAlivePingInterval = time.Minute
)

type Config struct {
	Scheme         string
	Domain         string
	ConnectTimeout time.Duration
	TransitTimeout time.Duration
	AppId          string
	Token          string
}

func NewConfig(appId string, token string) *Config {
	return &Config{
		Scheme:         DefaultServiceScheme,
		Domain:         DefaultServiceDomain,
		ConnectTimeout: DefaultConnectTimeout,
		TransitTimeout: DefaultTransitTimeout,
		AppId:          appId,
		Token:          token,
	}
}

func (c *Config) discoveryUri() string {
	return fmt.Sprintf("%s://gd.%s/discovery/%s", c.Scheme, c.Domain, c.AppId)
}
