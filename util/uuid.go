package util

import (
	"crypto/rand"
	"encoding/hex"
	"log"
)

func GenUUID() string {
	uuid := make([]byte, 16)
	n, err := rand.Read(uuid)
	if n != len(uuid) || err != nil {
		log.Printf("UUID generation failed. Save/Load functions won't work. (%v)", err)
		return ""
	}
	// TODO: verify the two lines implement RFC 4122 correctly
	uuid[8] = 0x80 // variant bits see page 5
	uuid[4] = 0x40 // version 4 Pseudo Random, see page 7

	return hex.EncodeToString(uuid)
}
