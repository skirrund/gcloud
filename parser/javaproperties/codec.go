package javaproperties

import (
	"bytes"
	"sort"
	"strings"

	"github.com/skirrund/gcloud/parser/properties"
	"github.com/spf13/cast"
)

// Codec implements the encoding.Encoder and encoding.Decoder interfaces for Java properties encoding.
type Codec struct {
	KeyDelimiter string

	Properties *properties.Properties
}

func (c *Codec) Encode(v map[string]any) ([]byte, error) {
	if c.Properties == nil {
		c.Properties = properties.NewProperties()
	}

	flattened := map[string]any{}

	flattened = flattenAndMergeMap(flattened, v, "", c.keyDelimiter())

	keys := make([]string, 0, len(flattened))

	for key := range flattened {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		_, _, err := c.Properties.Set(key, cast.ToString(flattened[key]))
		if err != nil {
			return nil, err
		}
	}

	var buf bytes.Buffer

	_, err := c.Properties.WriteComment(&buf, "#", properties.UTF8)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (c *Codec) Decode(b []byte, v map[string]any) error {
	var err error
	c.Properties, err = properties.Load(b, properties.UTF8)
	if err != nil {
		return err
	}

	for _, key := range c.Properties.Keys() {
		// ignore existence check: we know it's there
		value, _ := c.Properties.Get(key)

		// recursively build nested maps
		path := strings.Split(key, c.keyDelimiter())
		lastKey := strings.ToLower(path[len(path)-1])
		deepestMap := deepSearch(v, path[0:len(path)-1])

		// set innermost value
		deepestMap[lastKey] = value
	}

	return nil
}

func (c Codec) keyDelimiter() string {
	if c.KeyDelimiter == "" {
		return "."
	}

	return c.KeyDelimiter
}

// THIS CODE IS COPIED HERE: IT SHOULD NOT BE MODIFIED
// AT SOME POINT IT WILL BE MOVED TO A COMMON PLACE
// deepSearch scans deep maps, following the key indexes listed in the
// sequence "path".
// The last value is expected to be another map, and is returned.
//
// In case intermediate keys do not exist, or map to a non-map value,
// a new map is created and inserted, and the search continues from there:
// the initial map "m" may be modified!
func deepSearch(m map[string]any, path []string) map[string]any {
	for _, k := range path {
		m2, ok := m[k]
		if !ok {
			// intermediate key does not exist
			// => create it and continue from there
			m3 := make(map[string]any)
			m[k] = m3
			m = m3
			continue
		}
		m3, ok := m2.(map[string]any)
		if !ok {
			// intermediate key is a value
			// => replace with a new map
			m3 = make(map[string]any)
			m[k] = m3
		}
		// continue search from here
		m = m3
	}
	return m
}

// flattenAndMergeMap recursively flattens the given map into a new map
// Code is based on the function with the same name in the main package.
// TODO: move it to a common place.
func flattenAndMergeMap(shadow, m map[string]any, prefix, delimiter string) map[string]any {
	if shadow != nil && prefix != "" && shadow[prefix] != nil {
		// prefix is shadowed => nothing more to flatten
		return shadow
	}
	if shadow == nil {
		shadow = make(map[string]any)
	}

	var m2 map[string]any
	if prefix != "" {
		prefix += delimiter
	}
	for k, val := range m {
		fullKey := prefix + k
		switch val := val.(type) {
		case map[string]any:
			m2 = val
		case map[any]any:
			m2 = cast.ToStringMap(val)
		default:
			// immediate value
			shadow[strings.ToLower(fullKey)] = val
			continue
		}
		// recursively merge to shadow map
		shadow = flattenAndMergeMap(shadow, m2, fullKey, delimiter)
	}
	return shadow
}
