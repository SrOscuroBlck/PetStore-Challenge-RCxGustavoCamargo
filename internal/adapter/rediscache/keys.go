package rediscache

import (
	"fmt"

	"github.com/google/uuid"
)

// emptyCursorToken stands in for the first-page cursor so a key never ends bare.
const emptyCursorToken = "_"

func genKey(storeID uuid.UUID) string {
	return fmt.Sprintf("catalog:%s:gen", storeID)
}

func pageKey(storeID uuid.UUID, generation int64, limit int, cursor string) string {
	if cursor == "" {
		cursor = emptyCursorToken
	}
	return fmt.Sprintf("catalog:%s:g%d:l%d:c%s", storeID, generation, limit, cursor)
}
