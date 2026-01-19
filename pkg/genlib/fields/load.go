package fields

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/elastic/go-ucfg/yaml"
)

var ErrNotFound = errors.New("not found")

const (
	fieldsSlug        = "fields"
	packageSlug       = "package"
	dataStreamSlug    = "data_stream"
	searchSlug        = "search"
	kibanaVersionSlug = "kibana.version"
	manifestSlug      = "manifest.yml"
)

type yamlManifest struct {
	Type string `config:"type"`
}

func LoadFields(ctx context.Context, baseURL, integration, dataStream, version string) (Fields, string, error) {
	fieldsContent, dataStreamType, err := getFieldsFilesAndDataStreamType(ctx, baseURL, integration, dataStream, version)
	if err != nil {
		return nil, dataStreamType, err
	}

	if len(fieldsContent) == 0 {
		return nil, dataStreamType, ErrNotFound
	}

	fieldsFromYaml, err := loadFieldsFromYaml(fieldsContent)
	if err != nil {
		return nil, dataStreamType, err
	}

	fields := collectFields(fieldsFromYaml, "")

	fields, err = normaliseFields(fields)
	return fields, dataStreamType, err
}

func LoadFieldsWithTemplateFromString(ctx context.Context, fieldsContent string) (Fields, error) {
	if len(fieldsContent) == 0 {
		return nil, ErrNotFound
	}

	fieldsYaml := []byte("- key: key\n  fields:\n")
	for _, line := range strings.Split(fieldsContent, "\n") {
		fieldsYaml = append(fieldsYaml, []byte(`    `+line+"\n")...)
	}

	fieldsFromYaml, err := loadFieldsFromYaml(fieldsYaml)
	if err != nil {
		return nil, err
	}

	fields := collectFields(fieldsFromYaml, "")

	return normaliseFields(fields)
}

func LoadFieldsWithTemplate(ctx context.Context, fieldYamlPath string) (Fields, error) {
	fieldsFileContent, err := os.ReadFile(fieldYamlPath)
	if err != nil {
		return nil, err
	}

	var fieldsContent string

	key := strings.TrimSuffix(filepath.Base(fieldYamlPath), filepath.Ext(fieldYamlPath))
	keyEntry := fmt.Sprintf("- key: %s\n  fields:\n", key)
	for _, line := range strings.Split(string(fieldsFileContent), "\n") {
		keyEntry += `    ` + line + "\n"
	}

	fieldsContent += keyEntry
	if len(fieldsContent) == 0 {
		return nil, ErrNotFound
	}

	fieldsFromYaml, err := loadFieldsFromYaml([]byte(fieldsContent))
	if err != nil {
		return nil, err
	}

	fields := collectFields(fieldsFromYaml, "")

	return normaliseFields(fields)
}

func makePackageURL(baseURL, integration, version string) (*url.URL, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, packageSlug, integration, version)
	return u, nil
}

func makeDownloadURL(baseURL, donwloadPath string) (*url.URL, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, donwloadPath)
	return u, nil
}

func getFieldsFilesAndDataStreamType(ctx context.Context, baseURL, integration, dataStream, version string) ([]byte, string, error) {
	packageURL, err := makePackageURL(baseURL, integration, version)
	if err != nil {
		return nil, "", err
	}

	r, err := getFromURL(ctx, packageURL.String())
	if err != nil {
		return nil, "", err
	}

	var downloadPayload struct {
		Download string `json:"download"`
	}

	body, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, "", err
	}
	if err = json.Unmarshal(body, &downloadPayload); err != nil {
		return nil, "", err
	}

	downloadURL, err := makeDownloadURL(baseURL, downloadPayload.Download)
	if err != nil {
		return nil, "", err
	}
	r, err = getFromURL(ctx, downloadURL.String())
	defer func(r io.ReadCloser) {
		if r != nil {
			_ = r.Close()
		}
	}(r)

	if err != nil {
		return nil, "", err
	}

	zipContent, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, "", err
	}

	archive, err := zip.NewReader(bytes.NewReader(zipContent), int64(len(zipContent)))
	if err != nil {
		return nil, "", err
	}

	prefixFieldsPath := path.Join(fmt.Sprintf("%s-%s", integration, version), dataStreamSlug, dataStream, fieldsSlug)
	manifestPath := path.Join(fmt.Sprintf("%s-%s", integration, version), dataStreamSlug, dataStream, manifestSlug)

	var dataStreamType string
	var fieldsContent string
	for _, z := range archive.File {
		if z.FileInfo().IsDir() {
			continue
		}

		if !strings.HasPrefix(z.Name, prefixFieldsPath) && !strings.HasPrefix(z.Name, manifestPath) {
			continue
		}

		fieldsFileName := z.Name
		zr, err := z.Open()
		if err != nil {
			if zr != nil {
				_ = zr.Close()
			}
			return nil, "", err
		}

		fieldsFileContent, err := ioutil.ReadAll(zr)
		if err != nil {
			if zr != nil {
				_ = zr.Close()
			}
			return nil, "", err
		}

		_ = zr.Close()

		if strings.HasPrefix(z.Name, prefixFieldsPath) {
			key := strings.TrimSuffix(filepath.Base(fieldsFileName), filepath.Ext(fieldsFileName))
			keyEntry := fmt.Sprintf("- key: %s\n  fields:\n", key)
			for _, line := range strings.Split(string(fieldsFileContent), "\n") {
				keyEntry += `    ` + line + "\n"
			}

			fieldsContent += keyEntry
		}

		if strings.HasPrefix(z.Name, manifestPath) {
			var manifest yamlManifest

			cfg, err := yaml.NewConfig(fieldsFileContent)
			if err != nil {
				return nil, "", err
			}
			err = cfg.Unpack(&manifest)
			if err != nil {
				return nil, "", err
			}

			dataStreamType = manifest.Type
		}
	}

	return []byte(fieldsContent), dataStreamType, nil
}

func getFromURL(ctx context.Context, srcURL string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", srcURL, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
		return nil, ErrNotFound
	}

	return resp.Body, nil
}
