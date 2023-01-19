package EventStore

type Logger interface {
	Errorf(template string, args ...interface{})
	Debugf(template string, args ...interface{})
	Error(err error)
}
