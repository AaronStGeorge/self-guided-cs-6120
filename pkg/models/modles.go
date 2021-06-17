package models

import (
	"encoding/json"
	"errors"
)

type Program struct {
	Functions []Function `json:"functions"`
}

type Function struct {
	Args   []Args        `json:"args,omitempty"`
	Instrs []Instruction `json:"instrs"`
	Name   string        `json:"name"`
	Type   *Type         `json:"type,omitempty"`
}

type Args struct {
	Name string `json:"name"`
	Type *Type  `json:"type"`
}

type Instruction struct {
	Args   []string `json:"args,omitempty"`
	Dest   *string  `json:"dest,omitempty"`
	Labels []string `json:"labels,omitempty"`
	Funcs  []string `json:"funcs,omitempty"`
	Op     *string  `json:"op,omitempty"`
	Type   *Type    `json:"type,omitempty"`
	Value  *Value   `json:"value,omitempty"`
	Label  *string  `json:"label,omitempty"`
}

type Value struct {
	// Bril has and int and a float type but we always store a float64 since
	// we're always pulling this out of JSON.
	Float *float64
	Bool  *bool
}

func (v *Value) MarshalJSON() ([]byte, error) {
	switch {
	case v.Float != nil:
		return json.Marshal(*v.Float)
	case v.Bool != nil:
		return json.Marshal(*v.Bool)
	default:
		return nil, errors.New("malformed value")
	}
}

func (v *Value) UnmarshalJSON(data []byte) error {
	var readValue interface{}
	if err := json.Unmarshal(data, &readValue); err != nil {
		return err
	}
	switch t := readValue.(type) {
	case float64:
		v.Float = &t
	case bool:
		v.Bool = &t
	default:
		return errors.New("unknown type")
	}
	return nil
}

type Type struct {
	Primitive     *string
	Parameterized *ParameterizedType
}

func (t *Type) MarshalJSON() ([]byte, error) {
	switch {
	case t.Primitive != nil:
		return json.Marshal(*t.Primitive)
	case t.Parameterized != nil:
		return json.Marshal(t.Parameterized)
	default:
		return nil, errors.New("malformed type")
	}
}

func (t *Type) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch tv := v.(type) {
	case string:
		t.Primitive = &tv
	case map[string]interface{}:
		var pt ParameterizedType
		if err := json.Unmarshal(data, &pt); err != nil {
			return err
		}
		t.Parameterized = &pt
	default:
		return errors.New("unknown type")
	}
	return nil
}

type ParameterizedType struct {
	Parameter string
	Type      Type
}

func (pt *ParameterizedType) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]*Type{pt.Parameter: &pt.Type})
}

func (pt *ParameterizedType) UnmarshalJSON(data []byte) error {
	// Get the data
	var v map[string]interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	// The parameter for the type will be the first (and only) key
	keys := make([]string, 0, 1)
	for k := range v {
		keys = append(keys, k)
	}
	if len(keys) != 1 {
		return errors.New("more than one parameterized type")
	}
	pt.Parameter = keys[0]

	// Send the rest is again a Type
	rest, err := json.Marshal(v[pt.Parameter])
	if err != nil {
		return err
	}
	var t Type
	if err := json.Unmarshal(rest, &t); err != nil {
		return err
	}
	pt.Type = t

	return nil
}
