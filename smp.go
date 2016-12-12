package otr4

import (
	sha3 "golang.org/x/crypto/sha3"
)

func generateSMPsecret(intrFingerprint, recvrFingerprint, ssid, secret []byte) []byte {
	h := sha3.New512()
	h.Write(intrFingerprint)
	h.Write(recvrFingerprint)
	h.Write(ssid)
	h.Write(secret)
	return h.Sum(nil)
}
