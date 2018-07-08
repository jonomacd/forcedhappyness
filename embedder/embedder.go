package embedder

import (
	"io"

	"github.com/dyatlov/go-oembed/oembed"
)

var (
	Oembed *oembed.Oembed
)

func Init(providers io.Reader) error {
	Oembed = oembed.NewOembed()

	return Oembed.ParseProviders(providers)
}
