package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/atombender/go-jsonschema/pkg/generator"
	"github.com/mfelipe/go-feijoada/schemas/internal/loaders"
)

//go:embed schemas/*.json
var schemasFS embed.FS

// This is an exercise on running the generator programmatically, as the original module doesn't offer much configuration on input and output.
// It generates Go models from any JSON Schema files located in the `schemas` directory, using values dynamically extracted from the JSON schemas
// for naming the output files, packages and structs. But is not worth it, would be better to rewrite the generator to support this directly.
func main() {
	loader := loaders.NewCachedFileLoader()

	var files []string
	var schemaMappings = make([]generator.SchemaMapping, 0)
	if err := fs.WalkDir(schemasFS, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		loadedSchema, err := loader.Load(absPath, "")
		if err != nil {
			return err
		}
		schemaIdURL, err := url.Parse(loadedSchema.ID)
		if err != nil {
			return err
		}

		sPath := schemaIdURL.EscapedPath()
		preffix := "/schemas/"
		if strings.HasPrefix(sPath, preffix) {
			splitParams := strings.Split(strings.TrimPrefix(sPath, preffix), "/")
			if len(splitParams) < 2 {
				return fmt.Errorf("invalid schema path %s, expected at least 2 segments after %s", sPath, preffix)
			}
			name := splitParams[0]
			version := strings.ReplaceAll(splitParams[1], ".", "_")

			schemaMappings = append(schemaMappings, generator.SchemaMapping{
				SchemaID:    loadedSchema.ID,
				PackageName: fmt.Sprintf("github.com/mfelipe/go-feijoada/schemas/models/%s/v%s", name, version),
				OutputName:  fmt.Sprintf("models/%s/%s/%s.go", name, version, name),
			})

		}

		files = append(files, absPath)
		return nil
	}); err != nil {
		panic(err)
	}

	cfg := generator.Config{
		Warner: func(message string) {
			fmt.Printf("Warning: %s\n", message)
		},
		ExtraImports:              true,
		SchemaMappings:            schemaMappings,
		StructNameFromTitle:       true,
		Tags:                      []string{"json"},
		OnlyModels:                true,
		MinSizedInts:              true,
		MinimalNames:              false,
		DisableReadOnlyValidation: false,
		DisableCustomTypesForMaps: true,
		Loader:                    loader,
		DefaultPackageName:        "models",
		ResolveExtensions:         []string{"json"},
	}

	gen, err := generator.New(cfg)
	if err != nil {
		panic(err)
	}

	for _, fileName := range files {
		if err = gen.DoFile(fileName); err != nil {
			panic(err)
		}
	}

	sources, err := gen.Sources()
	if err != nil {
		panic(err)
	}

	for fileName, source := range sources {
		if fileName == "-" {
			if _, err = os.Stdout.Write(source); err != nil {
				panic(err)
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(fileName), 0755); err != nil {
				panic(err)
			}
			w, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				panic(err)
			}
			if _, err = w.Write(source); err != nil {
				panic(err)
			}
			_ = w.Close()
		}
	}
}
