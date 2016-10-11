package dot

import (
	"fmt"
	"strings"
)

// Quote takes an arbitrary DOT ID and escapes any quotes that is contains.
// The resulting string is quoted again to guarantee that it is a valid ID.
// DOT graph IDs can be any double-quoted string
// See http://www.graphviz.org/doc/info/lang.html
func Quote(id string) string {
	return fmt.Sprintf(`"%s"`, strings.Replace(id, `"`, `\"`, -1))
}
