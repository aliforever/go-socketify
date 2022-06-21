package socketify

type messageType interface {
	Type() int
	Data() ([]byte, error)
	Err() chan error
}
