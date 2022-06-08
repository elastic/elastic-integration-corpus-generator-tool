package fields

import (
	"github.com/elastic/go-ucfg/yaml"
)

type yamlFields []yamlField

type yamlField struct {
	Name       string     `config:"name"`
	Type       string     `config:"type"`
	ObjectType string     `config:"object_type"`
	Value      string     `config:"value"`
	Example    string     `config:"example"`
	Fields     yamlFields `config:"fields"`
}

func loadFieldsFromYaml(f []byte) (yamlFields, error) {
	var keys []yamlField

	cfg, err := yaml.NewConfig(f)
	if err != nil {
		return nil, err
	}
	err = cfg.Unpack(&keys)
	if err != nil {
		return nil, err
	}

	fields := yamlFields{}
	for _, key := range keys {
		fields = append(fields, key.Fields...)
	}
	return fields, nil
}

func collectFields(fieldsFromYaml yamlFields, namePrefix string) Fields {
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
			// There are examples of fields of type "group" with no subfields; ignore these.
			if field.Type != "group" {
				fields = fields.merge(field)
			}
		} else {
			subFields := collectFields(fieldFromYaml.Fields, field.Name)
			fields = fields.merge(subFields...)
		}
	}

	return fields
}
