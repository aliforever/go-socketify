package socketify

import "net/http"

type options struct {
	serveMux *http.ServeMux
	address  string
	endpoint string
}

func defaultOptions() *options {
	return &options{
		serveMux: http.NewServeMux(),
		address:  ":8080",
		endpoint: "/ws",
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
