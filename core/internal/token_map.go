package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"ntels.com/pharos/core/pkg/common"
)

type Token struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

var mutex sync.Mutex
var clientTokens = NewMap[*Token]()

func GetClientToken(address, id, secret string) (*Token, error) {
	key := getKey(address, id, secret)

	if clientTokens.Exist(key) {
		return clientTokens.Get(key), nil
	}

	{
		// Logic to prevent the problem of continuing to get before set when the corresponding function is called simultaneously
		mutex.Lock()
		defer mutex.Unlock()
		if clientTokens.Exist(key) {
			return clientTokens.Get(key), nil
		}
	}

	token, err := getToken(address, id, secret)
	if err != nil {
		return nil, err
	}

	setToken(key, token)

	return clientTokens.Get(key), nil
}

func getKey(address, id, secret string) string {
	return strings.Join([]string{address, id, secret}, ":")
}

func getToken(address, id, secret string) (*Token, error) {
	client := common.GetRestyClient()

	r := client.R()

	r.SetContext(context.Background())

	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("Authorization", "Bearer your_token")

	body := fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s", id, secret)
	r.SetBody([]byte(body))

	response, err := r.Post(address + "/auth/token")
	if err != nil {
		return nil, err
	}

	if response.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("(%s)(%s)", http.StatusText(response.StatusCode()), response.String())
	}

	token := Token{}

	if err := json.Unmarshal(response.Body(), &token); err != nil {
		return &token, err
	}

	return &token, nil
}

func setToken(key string, token *Token) {
	clientTokens.Remove(key, nil)

	go func() {
		time.Sleep(time.Second * time.Duration((float64(token.ExpiresIn))*0.9))
		clientTokens.Remove(key, nil)
	}()

	clientTokens.Set(key, token)
}
