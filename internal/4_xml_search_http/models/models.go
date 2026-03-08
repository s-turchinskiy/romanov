package models

type User struct {
	ID     int
	Name   string
	Age    int
	About  string
	Gender string
}

type SearchResponse struct {
	Users    []User
	NextPage bool
}

type SearchErrorResponse struct {
	Error string
}

const (
	OrderByAsc  = -1
	OrderByAsIs = 0
	OrderByDesc = 1

	ErrorBadOrderField = `OrderField invalid`
)

type SearchRequest struct {
	Limit      int
	Offset     int    // Можно учесть после сортировки
	Query      string // подстрока в 1 из полей
	OrderField string
	OrderBy    int
}
