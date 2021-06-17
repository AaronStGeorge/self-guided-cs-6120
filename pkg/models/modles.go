package models

import (
	"encoding/json"
	"errors"
)

type Program struct {
	Functions []Function `json:"functions"`
}

type Function struct {
	Instrs []Instruction `json:"instrs"`
	Name   string        `json:"name"`
}

type Instruction struct {
	Dest   *string  `json:"dest,omitempty"`
	Op     *string  `json:"op,omitempty"`
	Type   *Type    `json:"type,omitempty"`
	Value  *Value   `json:"value,omitempty"`
	Labels []string `json:"labels,omitempty"`
	Label  *string  `json:"label,omitempty"`
	Args   []string `json:"args,omitempty"`
}

type Value struct {
	Int  *int
	Bool *bool
}

func (b *Value) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch t := v.(type) {
	case float64:
		i := int(t)
		b.Int = &i
	case bool:
		b.Bool = &t
	default:
		return errors.New("unknown type")
	}
	return nil
}

type Type struct {
	Primitive     *string
	Parameterized *ParameterizedType
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

func (pt *ParameterizedType) UnmarshalJSON(data []byte) error {
	//// Remove whitespace
	//buffer := new(bytes.Buffer)
	//if err := json.Compact(buffer, data); err != nil {
	//	return err
	//}
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
