package socketify

import "crypto/rsa"

type encryptionMethod string

const (
	EncryptionTypeRsaAes encryptionMethod = "RSA/AES"
)

type encryption struct {
	Method encryptionMethod

	rsaAes *encryptionRsaAes
}

type encryptionRsaAes struct {
	privateKey func() (*rsa.PrivateKey, error)
}

type encryptionFields struct {
	clientPublicKey *rsa.PublicKey

	serverPrivateKey *rsa.PrivateKey
}
