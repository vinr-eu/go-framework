package managing

import (
	"github.com/vinr-eu/go-framework/app"
	"github.com/vinr-eu/go-framework/database"
	"github.com/vinr-eu/go-framework/user-service/internal/code"
	"github.com/vinr-eu/go-framework/user-service/pkg/api/managing"
	"github.com/vinr-eu/go-framework/user-service/pkg/domain/user"
)

func ViewUser(repository *database.Repository, id string, _ map[string]string) (*managing.ViewUserResponse, *app.Error) {
	var entity user.Entity
	if err := repository.FindByID(user.CollectionName, id, &entity); err != nil { // Simplify as there are no returns.
		return nil, app.NewErrorWithCode(err, code.ErrCode101DataFetchFailed)
	}
	return &managing.ViewUserResponse{
		Id:           entity.Id,
		FirstName:    entity.FirstName,
		LastName:     entity.LastName,
		EmailAddress: entity.EmailAddress,
	}, nil
}
