package store

import (
	"auth_service/domain"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	DATABASE   = "user_credentials"
	COLLECTION = "credentials"
)

type AuthMongoDBStore struct {
	credentials *mongo.Collection
}

func NewAuthMongoDBStore(client *mongo.Client) domain.AuthStore {
	auths := client.Database(DATABASE).Collection(COLLECTION)
	return &AuthMongoDBStore{
		credentials: auths,
	}
}

func (store *AuthMongoDBStore) GetAll(ctx context.Context) ([]*domain.Credentials, error) {

	filter := bson.D{{}}
	return store.filter(filter)
}

func (store *AuthMongoDBStore) Register(ctx context.Context, user *domain.Credentials) error {
	//vratiti u jednom trenutku
	user.Verified = true

	result, err := store.credentials.InsertOne(context.TODO(), user)
	if err != nil {
		return err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)

	return nil
}

func (store *AuthMongoDBStore) UpdateUser(ctx context.Context, user *domain.Credentials) error {

	_, err := store.credentials.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set": user})
	if err != nil {
		return err
	}

	return nil
}

func (store *AuthMongoDBStore) GetOneUser(ctx context.Context, username string) (*domain.Credentials, error) {

	filter := bson.M{"username": username}

	user, err := store.filterOne(filter)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (store *AuthMongoDBStore) GetOneUserByID(ctx context.Context, id primitive.ObjectID) *domain.Credentials {

	filter := bson.M{"_id": id}

	var user domain.Credentials
	err := store.credentials.FindOne(context.TODO(), filter, nil).Decode(&user)
	if err != nil {
		return nil
	}

	return &user
}

func (store *AuthMongoDBStore) DeleteUserByID(ctx context.Context, id primitive.ObjectID) error {

	_, err := store.credentials.DeleteMany(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	return nil
}

func (store *AuthMongoDBStore) filter(filter interface{}) ([]*domain.Credentials, error) {
	cursor, err := store.credentials.Find(context.TODO(), filter)
	defer cursor.Close(context.TODO())

	if err != nil {
		return nil, err
	}
	return decode(cursor)
}

func (store *AuthMongoDBStore) filterOne(filter interface{}) (user *domain.Credentials, err error) {
	result := store.credentials.FindOne(context.TODO(), filter)
	err = result.Decode(&user)
	return
}

func decode(cursor *mongo.Cursor) (users []*domain.Credentials, err error) {
	cursor.Next(context.TODO())
	var user domain.Credentials
	err = cursor.Decode(&user)
	if err != nil {
		return
	}
	users = append(users, &user)

	err = cursor.Err()
	return
}
