package docker

import (
	"fmt"
	"regexp"
)

var validHex = regexp.MustCompile(`^([a-f0-9]{64})$`)

// ValidateID checks whether an ID string is a valid image ID.
func ValidateID(id string) error {
	if ok := validHex.MatchString(id); !ok {
		return fmt.Errorf("image ID '%s' is invalid ", id)
	}
	return nil
}
