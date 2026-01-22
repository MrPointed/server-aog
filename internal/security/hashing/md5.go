package hashing

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

func MD5(text string) string {
	hash := md5.Sum([]byte(text))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}
