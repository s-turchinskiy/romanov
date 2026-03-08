package service

import (
	"errors"
	"sort"

	"github.com/s-turchinskiy/romanov/internal/4_xml_search_http/models"
)

type Servicer interface {
	Users(query, orderField string, orderBy, offset, limit int) ([]models.User, *ServiceError)
}

func SortClients(sortType string, sortOrder int, clients []models.User) ([]models.User, error) {
	switch sortOrder {
	case models.OrderByAsIs:
		return clients, nil
	case models.OrderByAsc:
		return sortAsc(sortType, clients)
	case models.OrderByDesc:
		return sortDesc(sortType, clients)
	default:
		return nil, errors.New("wrong sort order")
	}
}

func sortAsc(sortType string, clients []models.User) ([]models.User, error) {
	switch sortType {
	case "Id":
		sort.Slice(clients, func(i, j int) bool {
			return clients[i].Id < clients[j].Id
		})
	case "Age":
		sort.Slice(clients, func(i, j int) bool {
			return clients[i].Age < clients[j].Age
		})
	case "":
		fallthrough
	case "Name":
		sort.Slice(clients, func(i, j int) bool {
			return clients[i].Name < clients[j].Name
		})
	default:
		return nil, errors.New(models.ErrorBadOrderField)
	}

	return clients, nil
}

func sortDesc(sortType string, clients []models.User) ([]models.User, error) {
	switch sortType {
	case "Id":
		sort.Slice(clients, func(i, j int) bool {
			return clients[i].Id > clients[j].Id
		})
	case "Age":
		sort.Slice(clients, func(i, j int) bool {
			return clients[i].Age > clients[j].Age
		})
	case "":
		fallthrough
	case "Name":
		sort.Slice(clients, func(i, j int) bool {
			return clients[i].Name > clients[j].Name
		})
	default:
		return nil, errors.New(models.ErrorBadOrderField)
	}

	return clients, nil
}
