package thing

type Repository interface {
	Exists() (bool, error)
	Pull(path string) error
}
