package fields

import (
	"regexp"
	"sort"
	"strings"
)

type Fields []Field

func (f Fields) Len() int           { return len(f) }
func (f Fields) Less(i, j int) bool { return f[i].Name < f[j].Name }
func (f Fields) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

type Field struct {
	Name       string
	Type       string
	ObjectType string
	Example    string
	Value      any
}

func (fields Fields) merge(fieldsToMerge ...Field) Fields {
	merged := false
	for _, field := range fieldsToMerge {
		for _, currentField := range fields {
			if currentField.Name != field.Name {
				continue
			}

			if currentField.Example > field.Example {
				field.Example = currentField.Example
			}

			if currentField.Value != nil && field.Value == nil {
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
