package mongodb

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"os"
	"time"
	"vinr.eu/go-framework/app"
	"vinr.eu/go-framework/log"
)

type Repository struct {
	timeout  time.Duration
	client   *mongo.Client
	database *mongo.Database
}

type TenantModel struct {
	Id          string     `bson:"_id"`
	TenantId    string     `bson:"tenantId"`
	CreatedTime time.Time  `bson:"createdTime"`
	UpdatedTime *time.Time `bson:"updatedTime"`
}

func NewMongoDBRepository(timeout time.Duration, dbName string, opts ...*options.ClientOptions) (*Repository, error) {
	logger := log.NewLogger()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if len(opts) == 0 {
		opts = append(opts, options.Client().ApplyURI(os.Getenv("MONGO_DB_URI")))
	}
	client, err := mongo.Connect(ctx, opts...)
	if err != nil {
		logger.Error("Connect failed", "err", err)
		return nil, err
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		logger.Error("Ping failed", "err", err)
		return nil, err
	}
	database := client.Database(dbName)
	logger.Info("MongoDB client connected", "dbName", dbName)
	return &Repository{timeout: timeout, client: client, database: database}, nil
}

func (r *Repository) Disconnect() {
	logger := log.NewLogger()
	if r.client == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
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

func (r *Repository) Find(collectionName string, filter interface{}, pageSize int, pageNumber int, results interface{}, sortParams ...string) *app.Error {
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
		return app.NewError(errors.WithStack(err))
	}
	ctx, cancel = context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	err = cur.All(ctx, results)
	if err != nil {
		return app.NewError(errors.WithStack(err))
	}
	return nil
}

func (r *Repository) FindById(collectionName string, id string, entity interface{}) *app.Error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	err := r.database.Collection(collectionName).FindOne(ctx, bson.M{"_id": id}).Decode(entity)
	if err != nil {
		return app.NewError(errors.WithStack(err))
	}
	return nil
}

func (r *Repository) Create(collectionName string, _ string, entity interface{}) *app.Error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	_, err := r.database.Collection(collectionName).InsertOne(ctx, entity)
	if err != nil {
		return app.NewError(errors.WithStack(err))
	}
	return nil
}

func (r *Repository) Update(collectionName string, id string, entity interface{}) *app.Error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	_, err := r.database.Collection(collectionName).ReplaceOne(ctx, bson.M{"_id": id}, entity)
	if err != nil {
		return app.NewError(errors.WithStack(err))
	}
	return nil
}

func (r *Repository) Delete(collectionName string, id string) *app.Error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	_, err := r.database.Collection(collectionName).DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return app.NewError(errors.WithStack(err))
	}
	return nil
}

func (r *Repository) Aggregate(collectionName string, pipeline interface{}, results interface{}) *app.Error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	cur, err := r.database.Collection(collectionName).Aggregate(ctx, pipeline)
	if err != nil {
		return app.NewError(errors.WithStack(err))
	}
	ctx, cancel = context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	err = cur.All(ctx, results)
	if err != nil {
		return app.NewError(errors.WithStack(err))
	}
	return nil
}

func (r *Repository) NewBucket() (*gridfs.Bucket, *app.Error) {
	bucket, err := gridfs.NewBucket(r.database)
	if err != nil {
		return nil, app.NewError(errors.WithStack(err))
	}
	return bucket, nil
}

func (r *Repository) Count(collectionName string, filter interface{}) (int64, *app.Error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	count, err := r.database.Collection(collectionName).CountDocuments(ctx, filter)
	if err != nil {
		return 0, app.NewError(errors.WithStack(err))
	}
	return count, nil
}
