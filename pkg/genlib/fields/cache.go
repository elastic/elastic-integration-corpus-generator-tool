package fields

import (
	"context"
	"golang.org/x/sync/semaphore"
	"sync"

	"github.com/elastic/go-ucfg/yaml"
)

const (
	ProductionBaseURL = "https://epr.elastic.co/"
	maxParallel       = 4
)

type tuple struct {
	integration string
	stream      string
	version     string
}

type Manifest struct {
	Title   string `config:"title"`
	Type    string `config:"type"`
	DataSet string `config:"dataset"`
}

type CacheOption func(*Cache)

func WithBaseUrl(url string) CacheOption {
	return func(c *Cache) {
		c.baseUrl = url
	}
}

type Cache struct {
	mut      sync.RWMutex
	sema     *semaphore.Weighted
	baseUrl  string
	fields   map[tuple]Fields
	manifest map[tuple]Manifest
}

func NewCache(opts ...CacheOption) *Cache {
	c := &Cache{
		baseUrl:  ProductionBaseURL,
		sema:     semaphore.NewWeighted(maxParallel),
		fields:   make(map[tuple]Fields),
		manifest: make(map[tuple]Manifest),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (f *Cache) LoadFields(ctx context.Context, integration, stream, version string) (Fields, error) {
	var err error

	t := tuple{
		integration: integration,
		stream:      stream,
		version:     version,
	}

	f.mut.RLock()
	flds, ok := f.fields[t]
	f.mut.RUnlock()

	if ok {
		return flds, nil
	}

	// Limit the number of parallel outbound transactions
	if err = f.sema.Acquire(ctx, 1); err != nil {
		return nil, err
	}

	defer f.sema.Release(1)

	// Check again after aquiring semaphore; fields may have been retrieved by another thread
	f.mut.RLock()
	flds, ok = f.fields[t]
	f.mut.RUnlock()

	if !ok {

		if flds, err = LoadFields(ctx, f.baseUrl, integration, stream, version); err != nil {
			return nil, err
		} else {
			f.mut.Lock()
			f.fields[t] = flds
			f.mut.Unlock()
		}
	}

	return flds, nil
}

func (f *Cache) LoadManifest(ctx context.Context, integration, stream, version string) (*Manifest, error) {
	var err error

	t := tuple{
		integration: integration,
		stream:      stream,
		version:     version,
	}

	f.mut.RLock()
	manifest, ok := f.manifest[t]
	f.mut.RUnlock()

	if ok {
		return &manifest, nil
	}

	// Pull the manifest file
	url, err := makeManifestURL(f.baseUrl, integration, stream, version)
	if err != nil {
		return nil, err
	}

	data, err := getFromURL(ctx, url.String())
	if err != nil {
		return nil, err
	}

	// deserialize
	cfg, err := yaml.NewConfig(data)
	if err != nil {
		return nil, err
	}
	err = cfg.Unpack(&manifest)
	if err != nil {
		return nil, err
	}

	f.mut.Lock()
	f.manifest[t] = manifest
	f.mut.Unlock()

	return &manifest, nil
}
