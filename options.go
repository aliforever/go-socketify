package socketify

import (
	"github.com/teris-io/shortid"
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
	idFunc        func(r *http.Request) string
}

func defaultOptions() *options {
	return &options{
		serveMux: http.NewServeMux(),
		address:  defaultAddress,
		endpoint: defaultEndpoint,
		logger:   logger{},
	}
}

func Options() *options {
	return &options{}
}

func (o *options) SetIdSetterFunction(fn func(r *http.Request) string) *options {
	o.idFunc = fn
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
	if o.idFunc == nil {
		o.idFunc = func(_ *http.Request) string {
			return shortid.MustGenerate()
		}
	}
}
