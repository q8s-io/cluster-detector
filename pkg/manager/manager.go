package manager

type Manager interface {
	Name() string
	Start()
	Stop()
}
