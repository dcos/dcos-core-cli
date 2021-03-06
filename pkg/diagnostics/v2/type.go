package v2

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Type represents a bundle type
type Type int

// Possible types of diagnostics bundles
const (
	Local Type = iota
	Cluster
)

func (s Type) String() string {
	return typeToString[s]
}

var typeToString = map[Type]string{
	Local:   "Local",
	Cluster: "Cluster",
}

var stringToType = map[string]Type{
	"Local":   Local,
	"Cluster": Cluster,
}

// MarshalJSON marshals the enum as a quoted json string
func (s Type) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(typeToString[s])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (s *Type) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	t, ok := stringToType[j]
	if !ok {
		return fmt.Errorf("%s is not valid type", string(b))
	}
	*s = t
	return nil
}
