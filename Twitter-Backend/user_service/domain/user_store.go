package domain

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserStore interface {
	Get(ctx context.Context, id primitive.ObjectID) (*User, error)
	GetAll(ctx context.Context) ([]*User, error)
	Post(ctx context.Context, user *User) (*User, error)
	GetOneUser(ctx context.Context, username string) (*User, error)
	DeleteUserByID(ctx context.Context, id primitive.ObjectID) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
}
