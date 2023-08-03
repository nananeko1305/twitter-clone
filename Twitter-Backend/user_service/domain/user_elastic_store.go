package domain

type UserElasticStore interface {
	Get(username string) (*User, error)
	Post(user User) error
	CheckIndex()
	Search(search Search) ([]*User, error)
}
