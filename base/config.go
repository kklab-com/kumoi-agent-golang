package base

import "fmt"

const (
	DefaultServiceScheme = "https"
	DefaultServiceDomain = "omega.kklab.com"
)

type Config struct {
	Scheme string
	Domain string
	AppId  string
	Token  string
}

func NewConfig(appId string, token string) *Config {
	return &Config{
		Scheme: DefaultServiceScheme,
		Domain: DefaultServiceDomain,
		AppId:  appId,
		Token:  token,
	}
}

func (c *Config) discoveryUri() string {
	return fmt.Sprintf("%s://gd.%s/discovery/%s", c.Scheme, c.Domain, c.AppId)
}
