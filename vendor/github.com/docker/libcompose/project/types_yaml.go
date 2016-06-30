package project

import (
	"fmt"
	"strings"

	"github.com/flynn/go-shlex"
)

// Stringorslice represents a string or an array of strings.
// TODO use docker/docker/pkg/stringutils.StrSlice once 1.9.x is released.
type Stringorslice struct {
	parts []string
}

// MarshalYAML implements the Marshaller interface.
func (s Stringorslice) MarshalYAML() (tag string, value interface{}, err error) {
	return "", s.parts, nil
}

func toStrings(s []interface{}) ([]string, error) {
	if len(s) == 0 {
		return nil, nil
	}
	r := make([]string, len(s))
	for k, v := range s {
		if sv, ok := v.(string); ok {
			r[k] = sv
		} else {
			return nil, fmt.Errorf("Cannot unmarshal '%v' of type %T into a string value", v, v)
		}
	}
	return r, nil
}

// UnmarshalYAML implements the Unmarshaller interface.
func (s *Stringorslice) UnmarshalYAML(tag string, value interface{}) error {
	switch value := value.(type) {
	case []interface{}:
		parts, err := toStrings(value)
		if err != nil {
			return err
		}
		s.parts = parts
	case string:
		s.parts = []string{value}
	default:
		return fmt.Errorf("Failed to unmarshal Stringorslice: %#v", value)
	}
	return nil
}

// Len returns the number of parts of the Stringorslice.
func (s *Stringorslice) Len() int {
	if s == nil {
		return 0
	}
	return len(s.parts)
}

// Slice gets the parts of the StrSlice as a Slice of string.
func (s *Stringorslice) Slice() []string {
	if s == nil {
		return nil
	}
	return s.parts
}

// NewStringorslice creates an Stringorslice based on the specified parts (as strings).
func NewStringorslice(parts ...string) Stringorslice {
	return Stringorslice{parts}
}

// Command represents a docker command, can be a string or an array of strings.
// FIXME why not use Stringorslice (type Command struct { Stringorslice }
type Command struct {
	parts []string
}

// MarshalYAML implements the Marshaller interface.
func (s Command) MarshalYAML() (tag string, value interface{}, err error) {
	return "", s.parts, nil
}

// UnmarshalYAML implements the Unmarshaller interface.
func (s *Command) UnmarshalYAML(tag string, value interface{}) error {
	switch value := value.(type) {
	case []interface{}:
		parts, err := toStrings(value)
		if err != nil {
			return err
		}
		s.parts = parts
	case string:
		parts, err := shlex.Split(value)
		if err != nil {
			return err
		}
		s.parts = parts
	default:
		return fmt.Errorf("Failed to unmarshal Command: %#v", value)
	}
	return nil
}

// ToString returns the parts of the command as a string (joined by spaces).
func (s *Command) ToString() string {
	return strings.Join(s.parts, " ")
}

// Slice gets the parts of the Command as a Slice of string.
func (s *Command) Slice() []string {
	return s.parts
}

// NewCommand create a Command based on the specified parts (as strings).
func NewCommand(parts ...string) Command {
	return Command{parts}
}

// SliceorMap represents a slice or a map of strings.
type SliceorMap struct {
	parts map[string]string
}

// MarshalYAML implements the Marshaller interface.
func (s SliceorMap) MarshalYAML() (tag string, value interface{}, err error) {
	return "", s.parts, nil
}

// UnmarshalYAML implements the Unmarshaller interface.
func (s *SliceorMap) UnmarshalYAML(tag string, value interface{}) error {
	switch value := value.(type) {
	case map[interface{}]interface{}:
		parts := map[string]string{}
		for k, v := range value {
			if sk, ok := k.(string); ok {
				if sv, ok := v.(string); ok {
					parts[sk] = sv
				} else {
					return fmt.Errorf("Cannot unmarshal '%v' of type %T into a string value", v, v)
				}
			} else {
				return fmt.Errorf("Cannot unmarshal '%v' of type %T into a string value", k, k)
			}
		}
		s.parts = parts
	case []interface{}:
		parts := map[string]string{}
		for _, s := range value {
			if str, ok := s.(string); ok {
				str := strings.TrimSpace(str)
				keyValueSlice := strings.SplitN(str, "=", 2)

				key := keyValueSlice[0]
				val := ""
				if len(keyValueSlice) == 2 {
					val = keyValueSlice[1]
				}
				parts[key] = val
			} else {
				return fmt.Errorf("Cannot unmarshal '%v' of type %T into a string value", s, s)
			}
		}
		s.parts = parts
	default:
		return fmt.Errorf("Failed to unmarshal SliceorMap: %#v", value)
	}
	return nil
}

// MapParts get the parts of the SliceorMap as a Map of string.
func (s *SliceorMap) MapParts() map[string]string {
	if s == nil {
		return nil
	}
	return s.parts
}

// NewSliceorMap creates a new SliceorMap based on the specified parts (as map of string).
func NewSliceorMap(parts map[string]string) SliceorMap {
	return SliceorMap{parts}
}

// MaporEqualSlice represents a slice of strings that gets unmarshal from a
// YAML map into 'key=value' string.
type MaporEqualSlice struct {
	parts []string
}

// MarshalYAML implements the Marshaller interface.
func (s MaporEqualSlice) MarshalYAML() (tag string, value interface{}, err error) {
	return "", s.parts, nil
}

func toSepMapParts(value map[interface{}]interface{}, sep string) ([]string, error) {
	if len(value) == 0 {
		return nil, nil
	}
	parts := make([]string, 0, len(value))
	for k, v := range value {
		if sk, ok := k.(string); ok {
			if sv, ok := v.(string); ok {
				parts = append(parts, sk+sep+sv)
			} else {
				return nil, fmt.Errorf("Cannot unmarshal '%v' of type %T into a string value", v, v)
			}
		} else {
			return nil, fmt.Errorf("Cannot unmarshal '%v' of type %T into a string value", k, k)
		}
	}
	return parts, nil
}

// UnmarshalYAML implements the Unmarshaller interface.
func (s *MaporEqualSlice) UnmarshalYAML(tag string, value interface{}) error {
	switch value := value.(type) {
	case []interface{}:
		parts, err := toStrings(value)
		if err != nil {
			return err
		}
		s.parts = parts
	case map[interface{}]interface{}:
		parts, err := toSepMapParts(value, "=")
		if err != nil {
			return err
		}
		s.parts = parts
	default:
		return fmt.Errorf("Failed to unmarshal MaporEqualSlice: %#v", value)
	}
	return nil
}

// Slice gets the parts of the MaporEqualSlice as a Slice of string.
func (s *MaporEqualSlice) Slice() []string {
	return s.parts
}

// NewMaporEqualSlice creates a new MaporEqualSlice based on the specified parts.
func NewMaporEqualSlice(parts []string) MaporEqualSlice {
	return MaporEqualSlice{parts}
}

// MaporColonSlice represents a slice of strings that gets unmarshal from a
// YAML map into 'key:value' string.
type MaporColonSlice struct {
	parts []string
}

// MarshalYAML implements the Marshaller interface.
func (s MaporColonSlice) MarshalYAML() (tag string, value interface{}, err error) {
	return "", s.parts, nil
}

// UnmarshalYAML implements the Unmarshaller interface.
func (s *MaporColonSlice) UnmarshalYAML(tag string, value interface{}) error {
	switch value := value.(type) {
	case []interface{}:
		parts, err := toStrings(value)
		if err != nil {
			return err
		}
		s.parts = parts
	case map[interface{}]interface{}:
		parts, err := toSepMapParts(value, ":")
		if err != nil {
			return err
		}
		s.parts = parts
	default:
		return fmt.Errorf("Failed to unmarshal MaporColonSlice: %#v", value)
	}
	return nil
}

// Slice gets the parts of the MaporColonSlice as a Slice of string.
func (s *MaporColonSlice) Slice() []string {
	return s.parts
}

// NewMaporColonSlice creates a new MaporColonSlice based on the specified parts.
func NewMaporColonSlice(parts []string) MaporColonSlice {
	return MaporColonSlice{parts}
}

// MaporSpaceSlice represents a slice of strings that gets unmarshal from a
// YAML map into 'key value' string.
type MaporSpaceSlice struct {
	parts []string
}

// MarshalYAML implements the Marshaller interface.
func (s MaporSpaceSlice) MarshalYAML() (tag string, value interface{}, err error) {
	return "", s.parts, nil
}

// UnmarshalYAML implements the Unmarshaller interface.
func (s *MaporSpaceSlice) UnmarshalYAML(tag string, value interface{}) error {
	switch value := value.(type) {
	case []interface{}:
		parts, err := toStrings(value)
		if err != nil {
			return err
		}
		s.parts = parts
	case map[interface{}]interface{}:
		parts, err := toSepMapParts(value, " ")
		if err != nil {
			return err
		}
		s.parts = parts
	default:
		return fmt.Errorf("Failed to unmarshal MaporSpaceSlice: %#v", value)
	}
	return nil
}

// Slice gets the parts of the MaporSpaceSlice as a Slice of string.
func (s *MaporSpaceSlice) Slice() []string {
	return s.parts
}

// NewMaporSpaceSlice creates a new MaporSpaceSlice based on the specified parts.
func NewMaporSpaceSlice(parts []string) MaporSpaceSlice {
	return MaporSpaceSlice{parts}
}
