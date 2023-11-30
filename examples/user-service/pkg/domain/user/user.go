package user

const CollectionName = "users"

type Entity struct {
	Id           string `bson:"_id"`
	FirstName    string `bson:"firstName"`
	LastName     string `bson:"lastName"`
	EmailAddress string `bson:"emailAddress"`
}
