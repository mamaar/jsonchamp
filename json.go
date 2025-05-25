package jsonchamp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func marshalValue(m any) ([]byte, error) {
	m = toLargestType(m)
	buf := &bytes.Buffer{}

	switch v := m.(type) {
	case string:
		js, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("error marshalling string: %w", err)
		}

		buf.Write(js)
	case bool:
		j, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("error marshalling bool: %w", err)
		}

		buf.Write(j)
	case int64:
		j, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("error marshalling int: %w", err)
		}

		buf.Write(j)
	case float64:
		buf.WriteString(strconv.FormatFloat(v, 'f', -1, 64))

		if !strings.Contains(buf.String(), ".") {
			buf.WriteString(".0")
		}
	case *Map:
		buf.WriteString("{")

		keys := v.Keys()
		for i, k := range keys {
			buf.Write(marshalKey(k))
			buf.WriteString(":")

			v, _ := v.Get(k)

			d, err := marshalValue(v)
			if err != nil {
				return nil, err
			}

			buf.Write(d)

			if i < len(keys)-1 {
				buf.WriteString(",")
			}
		}

		buf.WriteString("}")
	case nil:
		buf.WriteString("null")
	default:
		unknown := reflect.ValueOf(v)

		switch {
		case unknown.Kind() == reflect.Slice:
			buf.WriteString("[")

			for i := range unknown.Len() {
				valMarshal, err := marshalValue(unknown.Index(i).Interface())
				if err != nil {
					return nil, fmt.Errorf("could not marshal slice item: %w", err)
				}

				buf.Write(valMarshal)

				if i < unknown.Len()-1 {
					buf.WriteString(",")
				}
			}

			buf.WriteString("]")
		default:
			return nil, fmt.Errorf("could not marshal type: %T", v)
		}
	}

	return buf.Bytes(), nil
}

func marshalKey(m string) []byte {
	key, err := marshalValue(m)
	if err != nil {
		return nil
	}

	return key
}

func marshal(m *Map) ([]byte, error) {
	return marshalValue(m)
}

// MarshalJSON marshals a Map into a JSON object.
func (m *Map) MarshalJSON() ([]byte, error) {
	b, err := marshal(m)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func unmarshalArray(dec *json.Decoder) ([]any, error) {
	var arr []any

	for {
		token, err := dec.Token()
		if err != nil {
			return nil, fmt.Errorf("could not get token: %w", err)
		}

		if token == json.Delim(']') {
			return arr, nil
		}

		switch v := token.(type) {
		case json.Delim:
			switch v {
			case '{':
				newMap := New()

				newMap, err := unmarshalMap(dec, newMap)
				if err != nil {
					return nil, fmt.Errorf("error unmarshalling map: %w", err)
				}

				arr = append(arr, newMap)
			case '[':
				newArr, err := unmarshalArray(dec)
				if err != nil {
					return nil, fmt.Errorf("could not unmarshal array: %w", err)
				}

				arr = append(arr, newArr)
			default:
				return nil, fmt.Errorf("unexpected delimiter %c", v)
			}
		case string:
			arr = append(arr, v)
		case json.Number:
			if strings.Contains(string(v), ".") {
				f, err := v.Float64()
				if err != nil {
					return nil, fmt.Errorf("could not convert number to float: %w", err)
				}

				arr = append(arr, f)
			} else {
				i, err := v.Int64()
				if err != nil {
					return nil, fmt.Errorf("could not convert number to int: %w", err)
				}

				arr = append(arr, i)
			}
		case bool:
			arr = append(arr, v)
		default:
			return nil, fmt.Errorf("unexpected type %T", v)
		}
	}
}

func unmarshalMap(dec *json.Decoder, m *Map) (*Map, error) {
	for {
		keyToken, err := dec.Token()
		if err != nil {
			return nil, fmt.Errorf("could not get key: %w", err)
		}

		if keyToken == json.Delim('}') {
			return m, nil
		}

		keyString, isString := keyToken.(string)
		if !isString {
			return nil, fmt.Errorf("expected string key, got %T", keyToken)
		}

		valueToken, err := dec.Token()
		if err != nil {
			return nil, fmt.Errorf("could not get value: %w", err)
		}

		switch v := valueToken.(type) {
		case string:
			m = m.Set(keyString, v)
		case json.Number:
			if strings.Contains(string(v), ".") {
				f, err := v.Float64()
				if err != nil {
					return nil, fmt.Errorf("could not convert number to float: %w", err)
				}

				m = m.Set(keyString, f)
			} else {
				i, err := v.Int64()
				if err != nil {
					return nil, fmt.Errorf("could not convert number to int: %w", err)
				}

				m = m.Set(keyString, i)
			}
		case bool:
			m = m.Set(keyString, v)
		case json.Delim:
			switch v {
			case '{':
				newMap := New()

				newMap, err := unmarshalMap(dec, newMap)
				if err != nil {
					return nil, err
				}

				m = m.Set(keyString, newMap)
			case '[':
				arr, err := unmarshalArray(dec)
				if err != nil {
					return nil, fmt.Errorf("could not unmarshal array: %w", err)
				}

				m = m.Set(keyString, arr)
			default:
				return nil, fmt.Errorf("unexpected delimiter %c", v)
			}
		default:
			return nil, fmt.Errorf("unexpected type %T", v)
		}
	}
}

// UnmarshalJSON unmarshals a JSON object into a Map.
func (m *Map) UnmarshalJSON(d []byte) error {
	if m.hasher == nil {
		m.hasher = defaultMapOptions.hasher()
	}

	if m.root == nil {
		m.root = &bitmasked{
			level:      0,
			valueMap:   0,
			subMapsMap: 0,
			values:     make([]node, 0, 5),
		}
	}

	dec := json.NewDecoder(bytes.NewReader(d))
	dec.UseNumber()

	token, err := dec.Token()
	if err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	if token != json.Delim('{') {
		return fmt.Errorf("expected object, got %T", token)
	}

	newMap, err := unmarshalMap(dec, m)
	if err != nil {
		return err
	}

	*m = *newMap

	return nil
}
