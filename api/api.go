package api

import (
	"crypto/md5"
	"encoding/hex"
)

type LiveApi interface {
	GetRealUrl(string) (*Room, error)
}

func MD5(str []byte) string {
	h := md5.New()
	h.Write(str)
	return hex.EncodeToString(h.Sum(nil))
}
