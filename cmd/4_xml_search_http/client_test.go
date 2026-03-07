package main

import (
	"github.com/s-turchinskiy/romanov/internal/4_xml_search_http/handlers"
	"github.com/s-turchinskiy/romanov/internal/4_xml_search_http/models"
	"github.com/s-turchinskiy/romanov/internal/4_xml_search_http/service/xml_repository"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

type fields struct {
	AccessToken string
}

type testCase struct {
	name    string
	fields  fields
	req     models.SearchRequest
	want    *models.SearchResponse
	wantErr bool
}

func TestSearchClient_FindUsers(t *testing.T) {
	hndlr := handlers.NewHandler(
		xml_repository.NewXMLService("dataset.xml"),
		10*time.Second)
	srv := httptest.NewServer(hndlr)

	for _, tt := range testCases() {
		t.Run(tt.name, func(t *testing.T) {
			clnt := &SearchClient{
				AccessToken: tt.fields.AccessToken,
				URL:         srv.URL,
			}
			got, err := clnt.FindUsers(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindUsers() error = \"%v\", wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindUsers() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func testCases() []testCase {

	fieldWithAccessToken := fields{AccessToken: "TestToken"}

	return []testCase{
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
