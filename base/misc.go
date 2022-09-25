package base

import (
	"runtime/debug"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/structpb"
)

var ErrConfigIsEmpty = errors.Errorf("config is empty")
var ErrTransitTimeout = errors.Errorf("request transit timeout")
var ErrSessionNotFound = errors.Errorf("session not found")
var ErrUnexpectError = errors.Errorf("unexcept error")

var sdkVersion string
var sdkLang = "go"

func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, module := range info.Deps {
			if module.Path == "github.com/kklab-com/kumoi-agent-golang" {
				sdkVersion = module.Version
			}
		}
	}
}

func SDKVersion() string {
	return sdkVersion
}

func SDKLanguage() string {
	return sdkLang
}

type Metadata = structpb.Struct

func NewMetadata(v map[string]interface{}) *Metadata {
	newStruct, _ := structpb.NewStruct(v)
	return newStruct
}

type transitPoolEntity struct {
	timestamp time.Time
	future    SendFuture
}
