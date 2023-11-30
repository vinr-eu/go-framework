package main

import (
	"net/http"
	"os"
	"time"
	"vinr.eu/go-framework/app"
	"vinr.eu/go-framework/code"
	"vinr.eu/go-framework/database"
	"vinr.eu/go-framework/handler"
	"vinr.eu/go-framework/server"
)

type ReadUserResponse struct {
	Id           string `json:"id"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	EmailAddress string `json:"emailAddress"`
}

const CollectionName = "users"

type User struct {
	Id           string `bson:"_id"`
	FirstName    string `bson:"firstName"`
	LastName     string `bson:"lastName"`
	EmailAddress string `bson:"emailAddress"`
}

var (
	ErrCode101DataFetchFailed = code.NewCode("us101e", "Data fetch failed")
)

func main() {
	repository, err := database.NewMongoDBRepository(10*time.Second, "userService")
	if err != nil {
		os.Exit(1)
	}
	defer repository.Disconnect()

	idleConnectionsClosed := make(chan struct{})

	mux := http.NewServeMux()

	mux.Handle("/managing/queries/read-user",
		handler.NewQueryByIDHandler[ReadUserResponse](repository, ReadUser, handleAppError, mapAuthHeaders))

	server.StartHttpServer(mux, idleConnectionsClosed)

	<-idleConnectionsClosed
}

func ReadUser(repository *database.Repository, id string, _ map[string]string) (*ReadUserResponse, *app.Error) {
	var entity User
	appErr := repository.FindById(CollectionName, id, &entity)
	if appErr != nil {
		appErr.SetCode(ErrCode101DataFetchFailed)
		return nil, appErr
	}
	response := ReadUserResponse{
		Id:           entity.Id,
		FirstName:    entity.FirstName,
		LastName:     entity.LastName,
		EmailAddress: entity.EmailAddress,
	}
	return &response, nil
}

func handleAppError(appErr app.Error, w http.ResponseWriter, _ *http.Request) {
	switch appErr.GetCode() {
	case ErrCode101DataFetchFailed:
		w.WriteHeader(http.StatusNotFound)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func mapAuthHeaders(_ *http.Request) map[string]string {
	return map[string]string{}
}
