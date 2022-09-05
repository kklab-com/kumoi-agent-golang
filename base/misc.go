package base

import (
	"time"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/structpb"
)

var ErrConfigIsEmpty = errors.Errorf("config is empty")
var ErrTransitTimeout = errors.Errorf("request transit timeout")
var ErrSessionNotFound = errors.Errorf("session not found")
var ErrUnexpectError = errors.Errorf("unexcept error")

type Metadata = structpb.Struct

func NewMetadata(v map[string]interface{}) *Metadata {
	newStruct, _ := structpb.NewStruct(v)
	return newStruct
}

type transitPoolEntity struct {
	timestamp time.Time
	future    SendFuture
}
