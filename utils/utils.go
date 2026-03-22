package utils

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"
)

func GetOrCreateRequestID(r *http.Request) string {
	if id := r.Header.Get("X-Request-ID"); id != "" {
		return id
	}

	if r.Header == nil {
		r.Header = make(http.Header)
	}

	var b [16]byte
	id := time.Now().UTC().Format("20060102150405.000000000")
	if _, err := rand.Read(b[:]); err == nil {
		id = hex.EncodeToString(b[:])
	}
	r.Header.Set("X-Request-ID", id)

	return id
}
