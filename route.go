package socketify

import (
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"github.com/aliforever/encryptionbox"
	"github.com/teris-io/shortid"
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
	var clientID string
	var attributes map[string]interface{}
	var err error

	if s.opts.onConnect != nil {
		clientID, attributes, err = s.opts.onConnect(w, r)
		if err != nil {
			s.opts.logger.Error("Error onConnect", err)
			return
		}
	}

	var ef *encryptionFields

	var headers http.Header

	if s.opts.encryption != nil {
		if s.opts.encryption.Method == EncryptionTypeRsaAes && s.opts.encryption.rsaAes != nil {
			key, err := s.parseRsaPublicKey(r)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				s.opts.logger.Error("Error parsing rsa public key from connection", err)
				return
			}

			ef = &encryptionFields{
				clientPublicKey: key,
			}

			serverPrivateKey, err := s.opts.encryption.rsaAes.privateKey()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				s.opts.logger.Error("Error getting server private key", err)
				return
			}

			ef.serverPrivateKey = serverPrivateKey
		}
	}

	c, err := s.upgrade.Upgrade(w, r, headers)
	if err != nil {
		s.opts.logger.Error("Error upgrading request", err, fmt.Sprintf("Headers: %+v", r.Header))
		return
	}

	if clientID == "" {
		clientID = shortid.MustGenerate()
	}

	client := newConnection(s, c, r, clientID, ef)
	if s.storage != nil {
		s.storage.addClient(client)
	}

	if attributes != nil {
		for key, val := range attributes {
			client.SetAttribute(key, val)
		}
	}

	s.connections <- client
}
