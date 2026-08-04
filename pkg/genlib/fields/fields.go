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
	Value      string
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
	// Trim excess capacity so that callers sharing this slice across goroutines
	// cannot inadvertently share the backing array when they append to it.
	// normaliseFields allocates cap=len(fields) but may keep fewer entries when
	// wildcard fields are filtered; without this trim the surplus capacity is
	// a latent data-race vector (see fbxj-vtfy-yrbl-mlyw).
	normalisedFields = normalisedFields[:len(normalisedFields):len(normalisedFields)]
	return normalisedFields, nil
}
