package socketify

import (
	"fmt"
	"github.com/teris-io/shortid"
	"net/http"
)

type UpgradeRequest struct {
	server *Server
	wr     http.ResponseWriter
	r      *http.Request

	done       chan bool
	clientID   string
	attributes map[string]interface{}
}

func newUpgradeRequest(server *Server, wr http.ResponseWriter, r *http.Request) *UpgradeRequest {
	return &UpgradeRequest{
		server:     server,
		wr:         wr,
		r:          r,
		attributes: map[string]interface{}{},
		done:       make(chan bool),
	}
}

func (u *UpgradeRequest) SetClientID(clientID string) *UpgradeRequest {
	u.clientID = clientID

	return u
}

func (u *UpgradeRequest) SetAttribute(key string, val interface{}) *UpgradeRequest {
	u.attributes[key] = val

	return u
}

func (u *UpgradeRequest) WriteResponse(statusCode int, header http.Header, response []byte) (int, error) {
	defer func() {
		u.done <- true
	}()

	u.wr.WriteHeader(statusCode)

	for key, values := range header {
		for _, value := range values {
			u.wr.Header().Add(key, value)
		}
	}

	return u.wr.Write(response)
}

func (u *UpgradeRequest) Request() *http.Request {
	return u.r
}

func (u *UpgradeRequest) Upgrade() (*Connection, error) {
	defer func() {
		u.done <- true
	}()

	var ef *encryptionFields

	if u.server.opts.encryption != nil {
		// TODO: Implement Encryption
		if u.server.opts.encryption.Method == EncryptionTypeRsaAes && u.server.opts.encryption.rsaAes != nil {
			key, err := u.server.parseRsaPublicKey(u.r)
			if err != nil {
				u.wr.WriteHeader(http.StatusBadRequest)
				u.wr.Write([]byte(err.Error()))
				u.server.opts.logger.Error("Error parsing rsa public key from connection", err)
				return nil, err
			}

			ef = &encryptionFields{
				clientPublicKey: key,
			}

			serverPrivateKey, err := u.server.opts.encryption.rsaAes.privateKey()
			if err != nil {
				u.wr.WriteHeader(http.StatusInternalServerError)
				u.wr.Write([]byte(err.Error()))
				u.server.opts.logger.Error("Error getting server private key", err)
				return nil, err
			}

			ef.serverPrivateKey = serverPrivateKey
		}
	}

	var headers http.Header

	c, err := u.server.upgrade.Upgrade(u.wr, u.r, headers)
	if err != nil {
		u.server.opts.logger.Error("Error upgrading request", err, fmt.Sprintf("Headers: %+v", u.r.Header))
		return nil, err
	}

	if u.clientID == "" {
		u.clientID = shortid.MustGenerate()
	}

	connection := newConnection(u.server, c, u.clientID, ef)
	if u.server.storage != nil {
		u.server.storage.addClient(connection)
	}

	return connection, nil
}
