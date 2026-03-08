package main

import (
	"fmt"
	"net/http/httptest"
	"os"
	"reflect"
	"time"

	"github.com/s-turchinskiy/romanov/internal/4_xml_search_http/handlers"
	"github.com/s-turchinskiy/romanov/internal/4_xml_search_http/models"
	"github.com/s-turchinskiy/romanov/internal/4_xml_search_http/service/xml_repository"
)

type fieldsMain struct {
	AccessToken string
}

type testCaseMain struct {
	name    string
	fields  fieldsMain
	req     models.SearchRequest
	want    *models.SearchResponse
	wantErr bool
}

func main() {
	res, _ := os.Getwd()
	hndlr := handlers.NewHandler(
		xml_repository.NewXMLService(res+"/cmd/4_xml_search_http/dataset.xml"),
		10*time.Second)
	srv := httptest.NewServer(hndlr)

	for _, tt := range testCasesMain() {
		clnt := &SearchClient{
			AccessToken: tt.fields.AccessToken,
			URL:         srv.URL,
		}
		got, err := clnt.FindUsers(tt.req)
		if (err != nil) != tt.wantErr {
			fmt.Printf("FindUsers() error = \"%v\", wantErr %v", err, tt.wantErr)
			return
		}
		if !reflect.DeepEqual(got, tt.want) {
			fmt.Printf("FindUsers() got = %v, want %v", got, tt.want)
		}
	}
}

func testCasesMain() []testCaseMain {
	fieldWithAccessToken := fieldsMain{AccessToken: "TestToken"}

	return []testCaseMain{
		{
			name:   "RequestSingleUser",
			fields: fieldWithAccessToken,
			req: models.SearchRequest{
				Query: "Hilda",
				Limit: 1,
			},
			want: &models.SearchResponse{
				Users: []models.User{
					{
						Id:     1,
						Name:   "Hilda Mayer",
						Age:    21,
						About:  "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n",
						Gender: "female",
					},
				},
				NextPage: false,
			},
			wantErr: false,
		},
	}
}
