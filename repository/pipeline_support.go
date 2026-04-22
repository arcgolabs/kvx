package repository

import (
	"errors"

	"github.com/arcgolabs/kvx"
)

type pipelineProvider interface {
	Pipeline() kvx.Pipeline
}

// ErrExpiration reports that an expiration value must be greater than zero.
var ErrExpiration = errors.New("expiration <= 0")
