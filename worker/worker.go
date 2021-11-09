package worker

type Worker interface {
	Start() (<-chan Result, error)
	Stop()
}
