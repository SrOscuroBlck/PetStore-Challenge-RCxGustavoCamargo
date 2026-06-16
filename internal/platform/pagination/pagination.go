package pagination

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Cursor is an opaque keyset position: the sort timestamp of the last row seen
// (created_at for the available list, sold_at for the sold list) plus its id.
type Cursor struct {
	SortKey time.Time
	ID      uuid.UUID
}

func Encode(c Cursor) string {
	raw := fmt.Sprintf("%d:%s", c.SortKey.UnixMicro(), c.ID.String())
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func Decode(encoded string) (Cursor, error) {
	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return Cursor{}, fmt.Errorf("decode cursor: %w", err)
	}
	micros, idStr, found := strings.Cut(string(raw), ":")
	if !found {
		return Cursor{}, errors.New("malformed cursor")
	}
	unixMicros, err := strconv.ParseInt(micros, 10, 64)
	if err != nil {
		return Cursor{}, fmt.Errorf("parse cursor timestamp: %w", err)
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return Cursor{}, fmt.Errorf("parse cursor id: %w", err)
	}
	return Cursor{SortKey: time.UnixMicro(unixMicros).UTC(), ID: id}, nil
}
