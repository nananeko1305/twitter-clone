package domain

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthStore interface {
	GetAll(ctx context.Context) ([]*Credentials, error)
	Register(ctx context.Context, user *Credentials) error
	GetOneUser(ctx context.Context, username string) (*Credentials, error)
	DeleteUserByID(ctx context.Context, id primitive.ObjectID) error
	GetOneUserByID(ctx context.Context, id primitive.ObjectID) *Credentials
	UpdateUser(ctx context.Context, user *Credentials) error
}
