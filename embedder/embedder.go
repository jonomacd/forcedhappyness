package embedder

import (
	"io"
	"strings"

	"github.com/dyatlov/go-oembed/oembed"
)

var (
	Oembed     *oembed.Oembed
	ba         = baseAugmenter{}
	augmenters = map[string]Augmenter{
		"Gfycat":  gyfcatAugmenter{},
		"Twitter": twitterAugmenter{},
	}
)

func Init(providers io.Reader) error {
	Oembed = oembed.NewOembed()

	return Oembed.ParseProviders(providers)
}

type Augmenter interface {
	Augment(item *oembed.Item, url string) (*oembed.Info, error)
}

func AugmentItem(item *oembed.Item, url string) (*oembed.Info, error) {
	if a, ok := augmenters[item.ProviderName]; ok {
		return a.Augment(item, url)
	}

	return ba.Augment(item, url)
}

type baseAugmenter struct{}

func (baseAugmenter) Augment(item *oembed.Item, url string) (*oembed.Info, error) {
	i, err := item.FetchOembed(oembed.Options{
		URL: url,
	})

	return i, err
}

type gyfcatAugmenter struct {
	baseAugmenter
}

func (a gyfcatAugmenter) Augment(item *oembed.Item, url string) (*oembed.Info, error) {
	info, err := a.baseAugmenter.Augment(item, url)
	if err != nil {
		return nil, err
	}

	info.HTML = strings.Replace(info.HTML, "position:relative;padding-bottom:calc(100% / 0.0)", "position:relative;padding-bottom:calc(100% / 1.0)", -1)
	return info, err
}

type twitterAugmenter struct {
	baseAugmenter
}

func (a twitterAugmenter) Augment(item *oembed.Item, url string) (*oembed.Info, error) {
	if !strings.Contains(item.EndpointURL, "omit_script") {
		item.EndpointURL = item.EndpointURL[:len(item.EndpointURL)-5] + "&hide_thread=true&dnt=true&omit_script=true&url="
	}
	return a.baseAugmenter.Augment(item, url)
}
