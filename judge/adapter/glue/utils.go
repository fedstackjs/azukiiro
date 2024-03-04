package glue

import (
	"bytes"
	"fmt"
)

func parseKVLine(line []byte) (string, string, error) {
	parts := bytes.SplitN(line, []byte("="), 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid line: does not contain key and value separated by '='")
	}
	key := string(parts[0])
	value := string(parts[1])
	return key, value, nil
}
