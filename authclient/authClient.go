package authclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

type AuthClient struct {
	mutex sync.Mutex

	token string

	user          string
	password      string
	idmServiceUrl string

	lastSignin     int64
	lastAccess     int64
	signinResponse SignInResponse
}

type SignInResponse struct {
	Token       string `json:"token"`
	Duration    int64  `json:"duration"`
	IdleTimeout int64  `json:"idleTimeout"`
	UserName    string `json:"userName"`
	UserId      string `json:"userId"`
}

type SignInResponseResponse struct {
	SignInResponse SignInResponse `json:"signInResponse"`
}

const oneHourInMillis int64 = 1 * 3600000

func New(user string, password string, idmServiceUrl string) (AuthClient, error) {

	return AuthClient{user: user, password: password, idmServiceUrl: idmServiceUrl}, nil
}

func (a *AuthClient) SignIn() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.signIn()
}

func (a *AuthClient) signIn() error {
	// Upper/Lower case pattern because GO locks are not re-entrant

	url := fmt.Sprintf("%s/idm/web/Authentication/signIn?schema=1.0&form=json",
		a.idmServiceUrl)
	log.Println("--> SignIn", url)

	httpClient := http.Client{
		Timeout: time.Second * 30,
	}

	req, err := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(a.user, a.password)
	response, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	log.Println("<--", string(body))
	signinResponse, err := decodeResponse(body)
	if err != nil {
		return err
	}
	a.token = signinResponse.Token
	a.signinResponse = *signinResponse
	a.lastSignin = currentMillis()
	a.lastAccess = a.lastSignin

	return nil
}

func (a *AuthClient) SignOut() error {
	log.Println("SignOut for ", a.token)

	a.mutex.Lock()
	defer a.mutex.Unlock()

	url := fmt.Sprintf("%s/idm/web/Authentication/signOut?schema=1.0&form=json&token=%s",
		a.idmServiceUrl,
		a.token)

	httpClient := http.Client{
		Timeout: time.Second * 30,
	}

	_, err := httpClient.Get(url)
	if err != nil {
		return err
	}
	a.token = ""

	return nil
}

func (a *AuthClient) GetToken() (string, error) {
	//log.Println("GetToken ")

	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.tokenAccessible() {
		a.lastAccess = currentMillis()
		return a.token, nil
	}

	// refresh when expired or idle timeout exceeded
	err := a.signIn()
	if err != nil {
		log.Println("SignIn failed due to error:", err)
		return "", err
	}

	return a.token, nil
}

func (a *AuthClient) tokenAccessible() bool {
	if a.token == "" {
		log.Println("No Token yet available")
		return false
	}
	currMillis := currentMillis()

	// add a fudge factor of an hour in order to avoid timing issues
	if currMillis > a.lastSignin+a.signinResponse.Duration-oneHourInMillis {
		log.Println("Token duration has been exceeded")
		return false
	}

	if currMillis > a.lastAccess+a.signinResponse.IdleTimeout-oneHourInMillis {
		log.Println("Token idle timeout has been exceeded")
		return false
	}

	return true
}

func currentMillis() int64 {
	return time.Now().UnixNano() / 1000000
}

func decodeResponse(resp []byte) (*SignInResponse, error) {
	var sir SignInResponseResponse
	err := json.Unmarshal([]byte(resp), &sir)
	if err != nil {
		return nil, err
	}

	return &sir.SignInResponse, nil
}
