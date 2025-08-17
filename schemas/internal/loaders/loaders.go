package loaders

import (
	"errors"

	pkgschemas "github.com/atombender/go-jsonschema/pkg/schemas"
)

var (
	errCannotLoadSchema = errors.New("cannot load schema")
)

func NewCachedFileLoader() *CachedFileLoader {
	return &CachedFileLoader{
		loader: pkgschemas.NewDefaultMultiLoader([]string{".json"}, []string{".yaml", ".yml"}),
		cache:  make(map[string]*pkgschemas.Schema),
	}
}

type CachedFileLoader struct {
	loader pkgschemas.Loader
	cache  map[string]*pkgschemas.Schema
}

func (l *CachedFileLoader) Load(fileName, parentFileName string) (*pkgschemas.Schema, error) {
	if schema, ok := l.cache[fileName]; ok {
		return schema, nil
	}

	schema, err := l.loader.Load(fileName, parentFileName)
	if err != nil {
		return nil, errors.Join(errCannotLoadSchema, err)
	}

	l.cache[fileName] = schema
	if schema.ID != "" {
		l.cache[schema.ID] = schema
	}

	return schema, nil
}
