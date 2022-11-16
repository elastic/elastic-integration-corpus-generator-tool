package fields

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/url"
	"path"
	"strings"

	"golang.org/x/mod/semver"
)

func MapVersion(ctx context.Context, baseUrl, integration, kibanaVersion string) (string, error) {
	searchUrl, err := makeSearchURL(baseUrl, integration, kibanaVersion)
	if err != nil {
		return "", err
	}

	r, err := getFromURL(ctx, searchUrl.String())
	if err != nil {
		return "", err
	}

	var payload []struct {
		Version string `json:"version"`
	}

	body, err := ioutil.ReadAll(r)
	if err != nil {
		_ = r.Close()
		return "", err
	}

	if err = json.Unmarshal(body, &payload); err != nil {
		return "", err
	}

	if len(payload) == 0 {
		return "", errors.New("empty payload")
	}

	version := payload[0].Version

	// semver is picky, requires the prefix
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	if !semver.IsValid(version) {
		return "", errors.New("invalid version")
	}

	return payload[0].Version, nil
}

func makeSearchURL(baseURL, integration, kibanaVersion string) (*url.URL, error) {

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, searchSlug)

	q := u.Query()
	q.Set(kibanaVersionSlug, kibanaVersion)
	q.Set(packageSlug, integration)
	u.RawQuery = q.Encode()

	return u, nil
}
