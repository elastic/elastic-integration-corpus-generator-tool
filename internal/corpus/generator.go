// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package corpus

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-ucfg/yaml"
	"github.com/spf13/afero"
)

const (
	dataStreamSlug = "data_stream"
	fieldsSlug     = "fields"
	packageSlug    = "package"
)

// timestamp represent a function providing a timestamp.
// It's used to allow replacing the value with a known one during testing.
type timestamp func() int64

type Fields []Field

func (f Fields) Len() int           { return len(f) }
func (f Fields) Less(i, j int) bool { return f[i].Name < f[j].Name }
func (f Fields) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

type Field struct {
	Name       string
	Type       string
	ObjectType string
	Example    string
	Value      string
}

type YamlFields []YamlField

type YamlField struct {
	Name       string     `config:"name"`
	Type       string     `config:"type"`
	ObjectType string     `config:"object_type"`
	Value      string     `config:"value"`
	Example    string     `config:"example"`
	Fields     YamlFields `config:"fields"`
}

type Config []ConfigField

type ConfigField struct {
	Name        string      `config:"name"`
	Fuzziness   int         `config:"fuzziness"`
	Range       int         `config:"range"`
	Cardinality int         `config:"cardinality"`
	ObjectKeys  []string    `config:"object_keys"`
	Value       interface{} `config:"value"`
}

func (c Config) getField(fieldName string) (ConfigField, bool) {
	for _, field := range c {
		if fieldName == field.Name {
			return field, true
		}
	}

	return ConfigField{}, false
}
func LoadConfig(configFile string) (Config, error) {
	if len(configFile) == 0 {
		return nil, nil
	}

	configFile = os.ExpandEnv(configFile)
	if _, err := os.Stat(configFile); err != nil {
		return nil, err
	}

	config, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	return loadConfigFromYaml(config)
}

func loadConfigFromYaml(c []byte) (Config, error) {
	var config Config

	cfg, err := yaml.NewConfig(c)
	if err != nil {
		return nil, err
	}
	err = cfg.Unpack(&config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func NewGenerator(config Config, fs afero.Fs, location string) (GeneratorCorpus, error) {
	keywordRegex, err := regexp.Compile("(\\.|-|_|\\s){1,1}")

	if err != nil {
		return GeneratorCorpus{}, err
	}

	return GeneratorCorpus{
		config:                       config,
		fs:                           fs,
		location:                     location,
		timestamp:                    time.Now().Unix,
		constantKeyword:              make(map[string]string, 0),
		keywordFieldValueRegex:       keywordRegex,
		objectFieldNameReplacer:      strings.NewReplacer(".*", ""),
		objectFieldValueNameReplacer: strings.NewReplacer(".*", "", ".", "-"),
	}, nil
}

// TestNewGenerator sets up a GeneratorCorpus configured to be used in testing.
func TestNewGenerator() GeneratorCorpus {
	f, _ := NewGenerator(nil, afero.NewMemMapFs(), "testdata")
	f.timestamp = func() int64 { return 1647345675 }
	return f
}

type GeneratorCorpus struct {
	config   Config
	fs       afero.Fs
	location string
	// timestamp allow overriding value in tests
	timestamp timestamp

	// cache for constantKeyword
	constantKeyword map[string]string

	// regexp for keyword processing
	keywordFieldValueRegex *regexp.Regexp

	// replacer for object type processing
	objectFieldNameReplacer *strings.Replacer

	// replacer for object type processin
	objectFieldValueNameReplacer *strings.Replacer
}

func (gc GeneratorCorpus) Location() string {
	return gc.location
}

// bulkPayloadFilename computes the bulkPayloadFilename for the corpus to be generated.
// To provide unique names the provided slug is prepended with current timestamp.
func (gc GeneratorCorpus) bulkPayloadFilename(integrationPackage, dataStream, packageVersion string) string {
	slug := integrationPackage + "-" + dataStream + "-" + packageVersion
	filename := fmt.Sprintf("%d-%s.ndjson", gc.timestamp(), sanitizeFilename(slug))
	return filename
}

var corpusLocPerm = os.FileMode(0770)
var corpusPerm = os.FileMode(0660)

func loadFieldsFromYaml(f []byte) (YamlFields, error) {
	var keys []YamlField

	cfg, err := yaml.NewConfig(f)
	if err != nil {
		return nil, err
	}
	err = cfg.Unpack(&keys)
	if err != nil {
		return nil, err
	}

	fields := YamlFields{}
	for _, key := range keys {
		fields = append(fields, key.Fields...)
	}
	return fields, nil
}

func mergeFields(fields Fields, fieldsToMerge ...Field) Fields {
	merged := false
	for _, field := range fieldsToMerge {
		for _, currentField := range fields {
			if currentField.Name != field.Name {
				continue
			}

			if currentField.Example > field.Example {
				field.Example = currentField.Example
			}

			if currentField.Value > field.Value {
				field.Value = currentField.Value
			}

			merged = true
			break
		}

		if !merged {
			fields = append(fields, field)
		}
	}

	return fields
}

func (gc GeneratorCorpus) eventFieldFaker(field Field, i int, totEvents int, event map[string]interface{}, events []map[string]interface{}) (map[string]interface{}, error) {
	if len(field.Value) > 0 {
		event[field.Name] = field.Value
		return event, nil
	}

	configField, fieldHasConfig := gc.config.getField(field.Name)
	if fieldHasConfig && configField.Value != nil {
		event[field.Name] = configField.Value
		return event, nil
	}

	if i > 0 && fieldHasConfig && configField.Cardinality > 0 {
		cardinalityIndex := int(math.Ceil(float64(totEvents) * (float64(configField.Cardinality) / 100.)))
		originalFieldName := field.Name
		if strings.HasSuffix(field.Name, ".*") {
			originalFieldName = gc.objectFieldNameReplacer.Replace(field.Name)
		}

		if i%cardinalityIndex > 0 {
			event[originalFieldName] = events[i-1][originalFieldName]
			return event, nil
		}
	}

	switch field.Type {
	case "date":
		event[field.Name] = gofakeit.DateRange(time.Now().Add(-1*time.Duration(int64(rand.Intn(3600))*time.Second.Nanoseconds())), time.Now().Add(time.Duration(int64(rand.Intn(3600))*time.Second.Nanoseconds())))
		break
	case "ip":
		event[field.Name] = gofakeit.IPv4Address()
		break
	case "double":
		fallthrough
	case "long":
		var dummy string
		var totDigit int
		var fuzziness int
		if fieldHasConfig {
			totDigit = configField.Range
			fuzziness = configField.Fuzziness
		}

		if totDigit > 0 {
			dummy = strconv.Itoa(rand.Intn(totDigit))
		} else if len(field.Example) > 0 {
			totDigit := len(field.Example)
			dummy = gofakeit.DigitN(uint(totDigit))
		} else {
			dummy = gofakeit.Digit()
		}

		dummyInt, err := strconv.Atoi(dummy)
		if err != nil {
			return event, err
		}

		if field.Type == "long" {
			if fuzziness > 0 && i > 0 {
				previousDummyInt := events[i-1][field.Name].(int)
				adjustedRatio := 1. - float64(rand.Intn(fuzziness))/100.
				if rand.Int()%2 == 0 {
					adjustedRatio = 1. + float64(rand.Intn(fuzziness))/100.
				}
				dummyInt = int(math.Ceil(float64(previousDummyInt) * adjustedRatio))
			}
			event[field.Name] = dummyInt
		} else {
			dummyFloat := float64(dummyInt) / rand.Float64()
			if fuzziness > 0 && i > 0 {
				previousDummyFloat := events[i-1][field.Name].(float64)
				adjustedRatio := 1. + float64(rand.Intn(fuzziness))/100.
				if rand.Int()%2 == 0 {
					adjustedRatio = 1. + float64(rand.Intn(fuzziness))/100.
				}
				dummyFloat = previousDummyFloat * adjustedRatio
			}
			event[field.Name] = dummyFloat
		}

		break
	case "constant_keyword":
		fallthrough
	case "keyword":
		var dummy string
		if len(field.Example) > 0 {
			totWords := len(gc.keywordFieldValueRegex.Split(field.Example, -1))
			dummy = gofakeit.Sentence(totWords)
			var joiner string
			if strings.Contains(field.Example, "\\.") {
				joiner = "\\."
			} else if strings.Contains(field.Example, "-") {
				joiner = "-"
			} else if strings.Contains(field.Example, "_") {
				joiner = "_"
			} else if strings.Contains(field.Example, " ") {
				joiner = " "
			}

			dummy = gc.keywordFieldValueRegex.ReplaceAllString(strings.ToLower(strings.TrimSuffix(dummy, ".")), joiner)
		} else {
			dummy = strings.ToLower(gofakeit.Word())
		}

		if field.Type == "constant_keyword" {
			if value, ok := gc.constantKeyword[field.Name]; !ok {
				gc.constantKeyword[field.Name] = dummy
			} else {
				dummy = value
			}
		}

		event[field.Name] = dummy
		break
	case "boolean":
		event[field.Name] = rand.Int()%2 == 0
		break
	case "object":
		if len(field.ObjectType) > 0 {
			field.Type = field.ObjectType
		} else {
			field.Type = "keyword"
		}

		originalFieldName := gc.objectFieldNameReplacer.Replace(field.Name)
		totObjectsKeys := rand.Intn(10)
		objectsKeys := make([]string, 0, totObjectsKeys)
		if len(configField.ObjectKeys) > 0 {
			for _, objectsKey := range configField.ObjectKeys {
				objectsKeys = append(objectsKeys, objectsKey)
			}
		} else {
			for i := 0; i < totObjectsKeys; i++ {
				objectsKeys = append(objectsKeys, fmt.Sprintf("%s-%d", gc.objectFieldValueNameReplacer.Replace(field.Name), i))
			}
		}
		objectEvent := make(map[string]interface{}, totObjectsKeys)
		for _, objectsKey := range objectsKeys {
			field.Name = objectsKey
			var err error
			currentEvent := make(map[string]interface{})
			currentEvent, err = gc.eventFieldFaker(field, i, totEvents, currentEvent, events)
			if err != nil {
				return event, err
			}

			objectEvent[field.Name] = currentEvent[field.Name]
		}

		event[originalFieldName] = objectEvent
		break
	default:
		event[field.Name] = gofakeit.Sentence(rand.Intn(25))
	}

	return event, nil
}
func (gc GeneratorCorpus) eventsFromFields(fields Fields, totEvents int) ([]map[string]interface{}, error) {
	events := make([]map[string]interface{}, totEvents)

	var err error
	for i := 0; i < totEvents; i++ {
		event := make(map[string]interface{})
		for _, field := range fields {
			event, err = gc.eventFieldFaker(field, i, totEvents, event, events)
			if err != nil {
				return nil, err
			}
		}

		events[i] = event
	}

	return events, nil
}

func normaliseFields(fields Fields) (Fields, error) {
	sort.Sort(fields)
	normalisedFields := make(Fields, 0, len(fields))
	for _, field := range fields {
		if !strings.Contains(field.Name, "*") {
			normalisedFields = append(normalisedFields, field)
			continue
		}

		normalizationPattern := strings.NewReplacer(".", "\\.", "*", ".+").Replace(field.Name)
		re, err := regexp.Compile(normalizationPattern)
		if err != nil {
			return nil, err
		}

		hasMatch := false
		for _, otherField := range fields {
			if otherField.Name == field.Name {
				continue
			}

			if re.MatchString(otherField.Name) {
				hasMatch = true
				break
			}
		}

		if !hasMatch {
			normalisedFields = append(normalisedFields, field)
		}
	}

	sort.Sort(normalisedFields)
	return normalisedFields, nil
}

func collectFields(fieldsFromYaml YamlFields, namePrefix string) Fields {
	fields := make(Fields, 0, len(fieldsFromYaml))
	for _, fieldFromYaml := range fieldsFromYaml {
		field := Field{
			Type:       fieldFromYaml.Type,
			ObjectType: fieldFromYaml.ObjectType,
			Example:    fieldFromYaml.Example,
			Value:      fieldFromYaml.Value,
		}

		if len(namePrefix) == 0 {
			field.Name = fieldFromYaml.Name
		} else {
			field.Name = namePrefix + "." + fieldFromYaml.Name
		}

		if len(fieldFromYaml.Fields) == 0 {
			fields = mergeFields(fields, field)
		} else {
			subFields := collectFields(fieldFromYaml.Fields, field.Name)
			fields = mergeFields(fields, subFields...)
		}
	}

	return fields
}

// Generate generates a bulk request corpus and persist it to file.
func (gc GeneratorCorpus) Generate(packageRegistryBaseURL, integrationPackage, dataStream, packageVersion string, totEvents int) (string, error) {
	if err := gc.fs.MkdirAll(gc.location, corpusLocPerm); err != nil {
		return "", fmt.Errorf("cannot generate corpus location folder: %v", err)
	}

	packageURL, err := url.Parse(packageRegistryBaseURL)
	if err != nil {
		return "", err
	}
	packageURL.Path = path.Join(packageSlug, integrationPackage, packageVersion)

	fieldsContent, err := gc.getFieldsFiles(packageURL, dataStream)
	if err != nil {
		return "", err
	}

	fieldsFromYaml, err := loadFieldsFromYaml(fieldsContent)

	fields := collectFields(fieldsFromYaml, "")
	fields, err = normaliseFields(fields)
	if err != nil {
		return "", err
	}

	events, err := gc.eventsFromFields(fields, totEvents)
	if err != nil {
		return "", err
	}

	payload := []byte("")
	createPayload := []byte(`{ "create" : { "_index": "metrics-` + integrationPackage + `.` + dataStream + `-default" } }` + "\n")
	for _, event := range events {
		payload = append(payload, createPayload...)
		eventPayload, err := json.Marshal(event)
		if err != nil {
			return "", err
		}

		payload = append(payload, eventPayload...)
		payload = append(payload, []byte("\n")...)
	}

	payloadFilename := path.Join(gc.location, gc.bulkPayloadFilename(integrationPackage, dataStream, packageVersion))
	return payloadFilename, afero.WriteFile(gc.fs, payloadFilename, payload, corpusPerm)
}

func getFromURL(getURL *url.URL) ([]byte, error) {
	resp, err := http.Get(getURL.String())
	defer func(r *http.Response) {
		if r != nil && r.Body != nil {
			_ = resp.Body.Close()
		}
	}(resp)

	if err != nil {
		return nil, err
	}

	body, err := func(r *http.Response) ([]byte, error) {
		if r != nil && r.Body != nil {
			return ioutil.ReadAll(r.Body)
		}

		return nil, errors.New("not valid response")
	}(resp)

	if err != nil {
		return nil, err
	}

	return body, nil
}

func (gc GeneratorCorpus) getFieldsFiles(packageIntegrationURL *url.URL, dataStream string) ([]byte, error) {
	body, err := getFromURL(packageIntegrationURL)

	var assetsPayload struct {
		Assets []string `json:"assets"`
	}

	err = json.Unmarshal(body, &assetsPayload)

	if err != nil {
		return nil, err
	}

	fieldsFilesURL := make([]string, 0)
	prefixFieldsPath := path.Join(packageIntegrationURL.Path, dataStreamSlug, dataStream, fieldsSlug)
	for _, assetPath := range assetsPayload.Assets {
		if !strings.HasPrefix(assetPath[1:], prefixFieldsPath) {
			continue
		}

		fieldsFilesURL = append(fieldsFilesURL, assetPath)
	}

	fieldsContent := ""
	for _, fieldsFileURL := range fieldsFilesURL {
		packageIntegrationURL.Path = fieldsFileURL
		body, err := getFromURL(packageIntegrationURL)
		if err != nil {
			return nil, err
		}

		keyEntry := fmt.Sprintf("- key: %s\n  fields:\n", strings.TrimSuffix(filepath.Base(fieldsFileURL), filepath.Ext(fieldsFileURL)))
		for _, line := range strings.Split(string(body), "\n") {
			keyEntry += `    ` + line + "\n"
		}

		fieldsContent += keyEntry
	}

	return []byte(fieldsContent), nil
}

// sanitizeFilename takes care of removing dangerous elements from a string so it can be safely
// used as a bulkPayloadFilename.
// NOTE: does not prevent command injection or ensure complete escaping of input
func sanitizeFilename(s string) string {
	s = strings.Replace(s, " ", "-", -1)
	s = strings.Replace(s, ":", "-", -1)
	s = strings.Replace(s, "/", "-", -1)
	s = strings.Replace(s, "\\", "-", -1)
	return s
}
