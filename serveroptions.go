package socketify

import (
	"net/http"
)

const (
	defaultAddress  = ":8080"
	defaultEndpoint = "/ws"
)

type options struct {
	serveMux      *http.ServeMux
	address       string
	endpoint      string
	checkOrigin   func(r *http.Request) bool
	logger        Logger
	enableStorage bool
	onConnect     func(wr http.ResponseWriter, r *http.Request) (connectionID string, attributes map[string]interface{}, err error)
}

func defaultOptions() *options {
	return &options{
		serveMux: http.NewServeMux(),
		address:  defaultAddress,
		endpoint: defaultEndpoint,
		logger:   logger{},
	}
}

func ServerOptions() *options {
	return &options{}
}

func (o *options) SetOnConnect(fn func(wr http.ResponseWriter, r *http.Request) (string, map[string]interface{}, error)) *options {
	o.onConnect = fn
	return o
}

func (o *options) SetEndpoint(endpoint string) *options {
	o.endpoint = endpoint
	return o
}

func (o *options) SetServeMux(mux *http.ServeMux) *options {
	o.serveMux = mux
	return o
}

func (o *options) SetAddress(address string) *options {
	o.address = address
	return o
}

func (o *options) SetCheckOrigin(checkOriginFn func(r *http.Request) bool) *options {
	o.checkOrigin = checkOriginFn
	return o
}

func (o *options) SetLogger(l Logger) *options {
	o.logger = l
	return o
}

func (o *options) IgnoreCheckOrigin() *options {
	o.checkOrigin = func(r *http.Request) bool {
		return true
	}
	return o
}

func (o *options) EnableStorage() *options {
	o.enableStorage = true
	return o
}

func (o *options) fillDefaults() {
	if o.address == "" {
		o.address = defaultAddress
	}
	if o.endpoint == "" {
		o.endpoint = defaultEndpoint
	}
	if o.serveMux == nil {
		o.serveMux = http.NewServeMux()
	}
	if o.logger == nil {
		o.logger = logger{}
	}
}
