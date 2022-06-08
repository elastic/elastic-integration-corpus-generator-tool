package fields

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
)

var ErrNotFound = errors.New("Not found")

const (
	fieldsSlug        = "fields"
	packageSlug       = "package"
	dataStreamSlug    = "data_stream"
	searchSlug        = "search"
	kibanaVersionSlug = "kibana.version"
	manifestSlug      = "manifest.yml"
)

func LoadFields(ctx context.Context, baseURL, integration, dataStream, version string) (Fields, error) {

	packageURL, err := makePackageURL(baseURL, integration, version)
	if err != nil {
		return nil, err
	}

	fieldsContent, err := getFieldsFiles(ctx, packageURL, dataStream)
	if err != nil {
		return nil, err
	}

	if len(fieldsContent) == 0 {
		return nil, ErrNotFound
	}

	fieldsFromYaml, err := loadFieldsFromYaml(fieldsContent)
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

func getFieldsFiles(ctx context.Context, packageURL *url.URL, dataStream string) ([]byte, error) {
	body, err := getFromURL(ctx, packageURL.String())
	if err != nil {
		return nil, err
	}

	var assetsPayload struct {
		Assets []string `json:"assets"`
	}

	if err = json.Unmarshal(body, &assetsPayload); err != nil {
		return nil, err
	}

	fieldsFilesURL := make([]string, 0)
	prefixFieldsPath := path.Join(packageURL.Path, dataStreamSlug, dataStream, fieldsSlug)
	for _, assetPath := range assetsPayload.Assets {
		if !strings.HasPrefix(assetPath, prefixFieldsPath) {
			continue
		}

		fieldsFilesURL = append(fieldsFilesURL, assetPath)
	}

	var fieldsContent string
	for _, fieldsFileURL := range fieldsFilesURL {
		packageURL.Path = fieldsFileURL

		body, err := getFromURL(ctx, packageURL.String())
		if err != nil {
			return nil, err
		}

		key := strings.TrimSuffix(filepath.Base(fieldsFileURL), filepath.Ext(fieldsFileURL))
		keyEntry := fmt.Sprintf("- key: %s\n  fields:\n", key)
		for _, line := range strings.Split(string(body), "\n") {
			keyEntry += `    ` + line + "\n"
		}

		fieldsContent += keyEntry
	}

	return []byte(fieldsContent), nil
}

func getFromURL(ctx context.Context, srcURL string) ([]byte, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", srcURL, nil)

	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrNotFound
	}

	return ioutil.ReadAll(resp.Body)
}

func makeManifestURL(baseURL, integration, stream, version string) (*url.URL, error) {

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	// https://epr.elastic.co/package/endpoint/8.2.0/data_stream/process/manifest.yml
	u.Path = path.Join(u.Path, packageSlug, integration, version, dataStreamSlug, stream, manifestSlug)

	return u, nil
}
