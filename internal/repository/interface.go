package repository

type Repository interface {
	Save(id, original string) error
	Get(id string) (string, error)
}
