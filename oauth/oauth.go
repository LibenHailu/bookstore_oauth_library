package oauth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/LibenHailu/bookstore_oauth_library/oauth/errors"
	"github.com/mercadolibre/golang-restclient/rest"
)

const (
	headerXPublic    = "X-Public"
	headerXClientId  = "X-Client-Id"
	headerXCallerId  = "X-User-Id"
	paramAccessToken = "access_token"
)

var (
	oauthRestClient = rest.RequestBuilder{
		BaseURL: "http://localhost:8081",
		Timeout: 200 * time.Millisecond,
	}
)

// type oauthClient struct {
// }

// type oauthInterface interface {
// }

type accessToken struct {
	Id       string `json:"id"`
	UserId   int64  `json:"user_id"`
	ClientId int64  `json:"client_id"`
}

func IsPublic(request *http.Request) bool {
	if request == nil {
		return true
	}
	return request.Header.Get(headerXPublic) == "true"
}

func GetCallerId(request *http.Request) int64 {
	if request == nil {
		return 0
	}
	callerId, err := strconv.ParseInt(request.Header.Get(headerXCallerId), 10, 64)

	if err != nil {
		return 0
	}
	return callerId
}

func GetClientId(request *http.Request) int64 {
	if request == nil {
		return 0
	}
	clientId, err := strconv.ParseInt(request.Header.Get(headerXClientId), 10, 64)

	if err != nil {
		return 0
	}
	return clientId
}

func AuthenticateRequest(request *http.Request) *errors.RestErr {
	if request == nil {
		return nil
	}
	accessTokenId := strings.TrimSpace(request.URL.Query().Get(paramAccessToken))
	// http://api.bookstore.com/resources?access_token=123
	if accessTokenId == "" {
		return nil
	}

	at, err := getAccessToken(accessTokenId)

	if err != nil {
		if err.Status == http.StatusNotFound {
			return nil
		}
		return err
	}

	request.Header.Add(headerXClientId, fmt.Sprintf("%v", at.ClientId))
	request.Header.Add(headerXCallerId, fmt.Sprintf("%v", at.UserId))

	return nil

}

func getAccessToken(accessTokenId string) (*accessToken, *errors.RestErr) {
	response := oauthRestClient.Get(fmt.Sprint("/oauth/access_token/%s", accessTokenId))
	if response == nil || response.Response == nil {
		return nil, errors.NewInternalServerError("invalid restclient response when trying to get access token")
	}
	if response.StatusCode > 299 {
		var restErr errors.RestErr
		err := json.Unmarshal(response.Bytes(), &restErr)
		if err != nil {
			return nil, errors.NewInternalServerError("invalid error interface when trying to get access token")
		}

		return nil, &restErr
	}

	var at accessToken
	if err := json.Unmarshal(response.Bytes(), &at); err != nil {
		return nil, errors.NewInternalServerError("error when trying to unmarshal access token response")
	}

	return &at, nil
}
