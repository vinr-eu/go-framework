package database

import (
	"context"
	"github.com/pkg/errors"
	"github.com/vinr-eu/go-framework/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"os"
	"time"
)

type Repository struct {
	timeout  time.Duration
	client   *mongo.Client
	database *mongo.Database
}

func NewMongoDBRepository(timeout time.Duration, databaseName string, opts ...*options.ClientOptions) (*Repository, error) {
	// Creating logger without any attributes as this is startup.
	logger := log.NewLogger()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if len(opts) == 0 {
		opts = append(opts, options.Client().ApplyURI(os.Getenv("MONGO_DB_URI")))
	}
	// Connect to MongoDB first.
	client, err := mongo.Connect(ctx, opts...)
	if err != nil {
		logger.Error("Connect failed", "err", err)
		return nil, err
	}
	// Ping and check the connection.
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		logger.Error("Ping failed", "err", err)
		return nil, err
	}
	database := client.Database(databaseName)
	logger.Info("MongoDB client connected", "databaseName", databaseName)
	// Return repository for use throughout the application.
	return &Repository{timeout: timeout, client: client, database: database}, nil
}

func (r *Repository) Disconnect() {
	// Creating logger without any attributes as this is shutdown.
	logger := log.NewLogger()
	if r.client == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	// Disconnect from MongoDB.
	err := r.client.Disconnect(ctx)
	if err != nil {
		logger.Error("MongoDB client disconnect failed", "err", err)
	} else {
		logger.Info("MongoDB client disconnected", "dbName", r.database.Name())
	}
}

func (r *Repository) GetName() string {
	return r.database.Name()
}

func (r *Repository) Find(collectionName string, filter interface{}, pageSize int, pageNumber int, results interface{}, sortParams ...string) error {
	findOptions := options.Find()
	findOptions.SetSkip(int64((pageNumber - 1) * pageSize))
	findOptions.SetLimit(int64(pageSize))
	if len(sortParams) != 0 {
		sort := bson.M{}
		for i := 0; i < len(sortParams); i += 2 {
			key := sortParams[i]
			value := sortParams[i+1]
			if value == "asc" {
				sort[key] = 1
			} else {
				sort[key] = -1
			}
		}
		findOptions.SetSort(sort)
	}
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	cur, err := r.database.Collection(collectionName).Find(ctx, filter, findOptions)
	if err != nil {
		return errors.WithStack(err)
	}
	ctx, cancel = context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	err = cur.All(ctx, results)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (r *Repository) FindByID(collectionName string, id string, entity interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	err := r.database.Collection(collectionName).FindOne(ctx, bson.M{"_id": id}).Decode(entity)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (r *Repository) Count(collectionName string, filter interface{}) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	count, err := r.database.Collection(collectionName).CountDocuments(ctx, filter)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	return count, nil
}

func (r *Repository) Create(collectionName string, _ string, entity interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	_, err := r.database.Collection(collectionName).InsertOne(ctx, entity)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (r *Repository) Update(collectionName string, id string, entity interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	_, err := r.database.Collection(collectionName).ReplaceOne(ctx, bson.M{"_id": id}, entity)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (r *Repository) Delete(collectionName string, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	_, err := r.database.Collection(collectionName).DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (r *Repository) Aggregate(collectionName string, pipeline interface{}, results interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	cur, err := r.database.Collection(collectionName).Aggregate(ctx, pipeline)
	if err != nil {
		return errors.WithStack(err)
	}
	ctx, cancel = context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	err = cur.All(ctx, results)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (r *Repository) NewBucket() (*gridfs.Bucket, error) {
	bucket, err := gridfs.NewBucket(r.database)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return bucket, nil
}
