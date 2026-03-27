package main

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/s-turchinskiy/romanov/internal/4_xml_search_http/handlers"
	"github.com/s-turchinskiy/romanov/internal/4_xml_search_http/models"
	"github.com/s-turchinskiy/romanov/internal/4_xml_search_http/service/xml_repository"
)

type fields struct {
	AccessToken string
	URL         string
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

	hndlrWithWrongFile := handlers.NewHandler(
		xml_repository.NewXMLService("dataset1.xml"),
		10*time.Second)
	srvWithWrongFile := httptest.NewServer(hndlrWithWrongFile)

	srvOnlyImpossibleCases := httptest.NewServer(&TestHandler{})

	for _, tt := range testCases(srv.URL, srvWithWrongFile.URL, srvOnlyImpossibleCases.URL) {
		t.Run(tt.name, func(t *testing.T) {
			clnt := &SearchClient{
				AccessToken: tt.fields.AccessToken,
				URL:         tt.fields.URL,
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

func testCases(url, urlWrong, urlImpossible string) []testCase {
	fieldWithAccessToken := fields{
		AccessToken: "TestToken",
		URL:         url,
	}

	multipleUsersCupidatat := []models.User{
		{
			ID:     0,
			Name:   "Boyd Wolf",
			Age:    22,
			About:  "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n",
			Gender: "male",
		},
		{
			ID:     1,
			Name:   "Hilda Mayer",
			Age:    21,
			About:  "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n",
			Gender: "female",
		},
		{
			ID:     5,
			Name:   "Beulah Stark",
			Age:    30,
			About:  "Enim cillum eu cillum velit labore. In sint esse nulla occaecat voluptate pariatur aliqua aliqua non officia nulla aliqua. Fugiat nostrud irure officia minim cupidatat laborum ad incididunt dolore. Fugiat nostrud eiusmod ex ea nulla commodo. Reprehenderit sint qui anim non ad id adipisicing qui officia Lorem.\n",
			Gender: "female",
		},
		{
			ID:     6,
			Name:   "Jennings Mays",
			Age:    39,
			About:  "Veniam consectetur non non aliquip exercitation quis qui. Aliquip duis ut ad commodo consequat ipsum cupidatat id anim voluptate deserunt enim laboris. Sunt nostrud voluptate do est tempor esse anim pariatur. Ea do amet Lorem in mollit ipsum irure Lorem exercitation. Exercitation deserunt adipisicing nulla aute ex amet sint tempor incididunt magna. Quis et consectetur dolor nulla reprehenderit culpa laboris voluptate ut mollit. Qui ipsum nisi ullamco sit exercitation nisi magna fugiat anim consectetur officia.\n",
			Gender: "male",
		},
	}

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
						ID:     1,
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
		{
			name:   "RequestMultipleUsers",
			fields: fieldWithAccessToken,
			req: models.SearchRequest{
				Query:      "cupidatat",
				Limit:      4,
				OrderBy:    models.OrderByAsIs,
				OrderField: "Id",
			},
			want: &models.SearchResponse{
				Users:    multipleUsersCupidatat,
				NextPage: true,
			},
			wantErr: false,
		},
		{
			name:   "EmptyResponse and 25 maximum",
			fields: fieldWithAccessToken,
			req: models.SearchRequest{
				Query: "some request",
				Limit: 35,
			},
			want: &models.SearchResponse{
				Users:    nil,
				NextPage: false,
			},
			wantErr: false,
		},
		{
			name:   "wrong limit",
			fields: fieldWithAccessToken,
			req: models.SearchRequest{
				Query: "some request",
				Limit: -1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "wrong offset",
			fields: fieldWithAccessToken,
			req: models.SearchRequest{
				Query:  "some request",
				Offset: -1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "StatusUnauthorized",
			fields:  fields{AccessToken: "", URL: url},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "internal error",
			fields:  fields{AccessToken: "aa", URL: urlWrong},
			req:     models.SearchRequest{},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "no url",
			fields:  fields{},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "bad request",
			fields: fieldWithAccessToken,
			req: models.SearchRequest{
				Query:  "some request",
				Offset: 1000,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "bad request ErrorBadOrderField",
			fields: fieldWithAccessToken,
			req: models.SearchRequest{
				Query:      "some request",
				OrderField: "wrong",
				OrderBy:    1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "timeout",
			fields:  fields{AccessToken: "timeout", URL: urlImpossible},
			req:     models.SearchRequest{},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "broken_json_bad_request",
			fields:  fields{AccessToken: "broken_json_bad_request", URL: urlImpossible},
			req:     models.SearchRequest{},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "broken_json_status_ok",
			fields:  fields{AccessToken: "broken_json_status_ok", URL: urlImpossible},
			req:     models.SearchRequest{},
			want:    nil,
			wantErr: true,
		},
	}
}

type TestHandler struct {
}

func (h *TestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	accessToken := r.Header[http.CanonicalHeaderKey("AccessToken")][0]

	switch accessToken {
	case "timeout":
		time.Sleep(1100 * time.Millisecond)
	case "broken_json_bad_request":
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("broken_json"))
	case "broken_json_status_ok":
		_, _ = w.Write([]byte("broken_json"))
	}

}
