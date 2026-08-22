package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

func newID(prefix string, now time.Time) string {
	bytes := make([]byte, 5)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%s-%d", prefix, now.UnixNano())
	}
	return fmt.Sprintf("%s-%d-%s", prefix, now.UnixMilli(), hex.EncodeToString(bytes))
}
