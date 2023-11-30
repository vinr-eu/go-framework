package managing

import (
	"vinr.eu/go-framework/app"
	"vinr.eu/go-framework/database"
	"vinr.eu/go-framework/user-service/internal/code"
	"vinr.eu/go-framework/user-service/pkg/api/managing"
	"vinr.eu/go-framework/user-service/pkg/domain/user"
)

func ReadUser(repository *database.Repository, id string, _ map[string]string) (*managing.ReadUserResponse, *app.Error) {
	var entity user.Entity
	err := repository.FindById(user.CollectionName, id, &entity)
	if err != nil {
		return nil, app.NewErrorWithCode(err, code.ErrCode101DataFetchFailed)
	}
	response := managing.ReadUserResponse{
		Id:           entity.Id,
		FirstName:    entity.FirstName,
		LastName:     entity.LastName,
		EmailAddress: entity.EmailAddress,
	}
	return &response, nil
}
