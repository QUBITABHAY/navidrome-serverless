package postgres

import (
	"fmt"
	"strconv"
)

func parseID(s string, out *int) (bool, error) {
	val, err := strconv.Atoi(s)
	if err != nil {
		return false, fmt.Errorf("parsing integer id: %w", err)
	}
	*out = val
	return true, nil
}

func idToString(id int) string {
	return strconv.Itoa(id)
}
