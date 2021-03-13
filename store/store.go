package store

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Storage interface that describes application store
type Storage interface {
	Database() *mongo.Database
	Connect() error
	Disconnect() error
}

// Store is a struct that controlls database operations
type Store struct {
	URL    string
	DBName string
	client *mongo.Client
}

// New creates a new storage
func New(URL, DBName string) (*Store, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(URL))

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	s := &Store{
		URL:    URL,
		DBName: DBName,
		client: client,
	}

	return s, nil
}

// Database returns current database
func (s *Store) Database() *mongo.Database {
	return s.client.Database(s.DBName)
}

// Connect to database
func (s *Store) Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err := s.client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
		return err
	}
	err = s.client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

// Disconnect store
func (s *Store) Disconnect() error {
	return s.client.Disconnect(context.TODO())
}
