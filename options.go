package socketify

import "net/http"

const (
	defaultAddress  = ":8080"
	defaultEndpoint = "/ws"
)

type options struct {
	serveMux    *http.ServeMux
	address     string
	endpoint    string
	checkOrigin func(r *http.Request) bool
}

func defaultOptions() *options {
	return &options{
		serveMux: http.NewServeMux(),
		address:  defaultAddress,
		endpoint: defaultEndpoint,
	}
}

func Options() *options {
	return &options{}
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

func (o *options) SetCheckOriginIgnore() *options {
	o.checkOrigin = func(r *http.Request) bool {
		return true
	}
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
}
