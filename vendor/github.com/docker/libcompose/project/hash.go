package project

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"reflect"
	"sort"
)

// GetServiceHash computes and returns a hash that will identify a service.
// This hash will be then used to detect if the service definition/configuration
// have changed and needs to be recreated.
func GetServiceHash(name string, config *ServiceConfig) string {
	hash := sha1.New()

	io.WriteString(hash, name)

	//Get values of Service through reflection
	val := reflect.ValueOf(config).Elem()

	//Create slice to sort the keys in Service Config, which allow constant hash ordering
	serviceKeys := []string{}

	//Create a data structure of map of values keyed by a string
	unsortedKeyValue := make(map[string]interface{})

	//Get all keys and values in Service Configuration
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		keyField := val.Type().Field(i)

		serviceKeys = append(serviceKeys, keyField.Name)
		unsortedKeyValue[keyField.Name] = valueField.Interface()
	}

	//Sort serviceKeys alphabetically
	sort.Strings(serviceKeys)

	//Go through keys and write hash
	for _, serviceKey := range serviceKeys {
		serviceValue := unsortedKeyValue[serviceKey]

		io.WriteString(hash, fmt.Sprintf("\n  %v: ", serviceKey))

		switch s := serviceValue.(type) {
		case SliceorMap:
			sliceKeys := []string{}
			for lkey := range s.MapParts() {
				sliceKeys = append(sliceKeys, lkey)
			}
			sort.Strings(sliceKeys)

			for _, sliceKey := range sliceKeys {
				io.WriteString(hash, fmt.Sprintf("%s=%v, ", sliceKey, s.MapParts()[sliceKey]))
			}
		case MaporEqualSlice:
			sliceKeys := s.Slice()
			// do not sort keys as the order matters

			for _, sliceKey := range sliceKeys {
				io.WriteString(hash, fmt.Sprintf("%s, ", sliceKey))
			}
		case MaporColonSlice:
			sliceKeys := s.Slice()
			// do not sort keys as the order matters

			for _, sliceKey := range sliceKeys {
				io.WriteString(hash, fmt.Sprintf("%s, ", sliceKey))
			}
		case MaporSpaceSlice:
			sliceKeys := s.Slice()
			// do not sort keys as the order matters

			for _, sliceKey := range sliceKeys {
				io.WriteString(hash, fmt.Sprintf("%s, ", sliceKey))
			}
		case Command:
			sliceKeys := s.Slice()
			// do not sort keys as the order matters

			for _, sliceKey := range sliceKeys {
				io.WriteString(hash, fmt.Sprintf("%s, ", sliceKey))
			}
		case Stringorslice:
			sliceKeys := s.Slice()
			sort.Strings(sliceKeys)

			for _, sliceKey := range sliceKeys {
				io.WriteString(hash, fmt.Sprintf("%s, ", sliceKey))
			}
		case []string:
			sliceKeys := s
			sort.Strings(sliceKeys)

			for _, sliceKey := range sliceKeys {
				io.WriteString(hash, fmt.Sprintf("%s, ", sliceKey))
			}
		default:
			io.WriteString(hash, fmt.Sprintf("%v", serviceValue))
		}
	}

	return hex.EncodeToString(hash.Sum(nil))
}
