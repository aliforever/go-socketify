package socketify

import (
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"github.com/aliforever/encryptionbox"
	"net/http"
)

func (s *Server) parseRsaPublicKey(r *http.Request) (*rsa.PublicKey, error) {
	publicKeyPem := r.Header.Get("rsa_public_key_pem_b64")
	if publicKeyPem == "" {
		publicKeyPem = r.URL.Query().Get("rsa_public_key_pem_b64")
	}

	if publicKeyPem == "" {
		return nil, fmt.Errorf("rsa_public_key_pem_base64_not_provided")
	}

	b64, err := base64.StdEncoding.DecodeString(publicKeyPem)
	if err != nil {
		return nil, fmt.Errorf("cant_decode_rsa_public_key_pem_base64_%s", err)
	}

	publicKey, err := encryptionbox.EncryptionBox{}.RSA.PublicKeyFromPKCS1PEMBytes(b64)
	if err != nil {
		return nil, err
	}

	return publicKey, nil
}

func (s *Server) websocketUpgrade(w http.ResponseWriter, r *http.Request) {
	ur := newUpgradeRequest(s, w, r)

	s.upgradeRequests <- ur

	<-ur.done
}
