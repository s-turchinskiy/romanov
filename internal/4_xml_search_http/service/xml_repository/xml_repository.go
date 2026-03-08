package xml_repository

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/s-turchinskiy/romanov/internal/4_xml_search_http/models"
	"github.com/s-turchinskiy/romanov/internal/4_xml_search_http/service"
)

type XMLService struct {
	filename string
}

type XMLClient struct {
	ID            int    `xml:"id"`
	GUID          string `xml:"guid"`
	IsActive      bool   `xml:"isActive"`
	Balance       string `xml:"balance"`
	Picture       string `xml:"picture"`
	Age           int    `xml:"age"`
	EyeColor      string `xml:"eyeColor"`
	FirstName     string `xml:"first_name"`
	LastName      string `xml:"last_name"`
	Gender        string `xml:"gender"`
	Company       string `xml:"company"`
	Email         string `xml:"email"`
	Phone         string `xml:"phone"`
	Address       string `xml:"address"`
	About         string `xml:"about"`
	Registered    string `xml:"registered"`
	FavoriteFruit string `xml:"favoriteFruit"`
}

type Clients struct {
	Clients []XMLClient `xml:"row"`
}

var (
	errCantOpenFile = errors.New("can't open file")
	errWrongOffset  = errors.New("wrong offset")
	errWrongLimit   = errors.New("wrong limit")
)

func NewXMLService(filename string) *XMLService {
	return &XMLService{
		filename: filename,
	}
}

func (s *XMLService) Users(query, orderField string, orderBy, offset, limit int) ([]models.User, *service.ServiceError) {
	if _, err := os.Stat(s.filename); err != nil && errors.Is(err, os.ErrNotExist) {
		return nil, service.NewServiceError(service.InternalError, fmt.Errorf("file %v not exist", s.filename))
	}

	file, err := os.Open(s.filename)
	if err != nil {
		return nil, service.NewServiceError(service.InternalError, errCantOpenFile)
	}
	defer func() {
		_ = file.Close()
	}()

	var clients Clients
	err = xml.NewDecoder(file).Decode(&clients)
	if err != nil {
		return nil, service.NewServiceError(service.InternalError, err)
	}

	var result []models.User
	for _, client := range clients.Clients {
		name := client.FirstName + " " + client.LastName
		if strings.Contains(name, query) || strings.Contains(client.About, query) {
			var user models.User
			user.About = client.About
			user.Age = client.Age
			user.Gender = client.Gender
			user.ID = client.ID
			user.Name = name
			result = append(result, user)
		}
	}

	result, err = service.SortClients(orderField, orderBy, result)
	if err != nil {
		return nil, service.NewServiceError(service.BadRequest, err)
	}

	if offset < len(result)+1 && offset >= 0 {
		result = result[offset:]
	} else {
		return nil, service.NewServiceError(service.BadRequest, errWrongOffset)
	}

	if limit < len(result) {
		result = result[:limit]
	} else if limit < 0 {
		return nil, service.NewServiceError(service.BadRequest, errWrongLimit)
	}

	return result, nil
}
