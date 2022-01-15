package repository

type UserItem struct {
	UserId         string `dynamo:"UserId,hash"`
	ModelTypeAndId string `dynamo:",range"`
	FirstName      string `dynamo:"First Name"`
	LastName       string `dynamo:"Last Name"`
}
