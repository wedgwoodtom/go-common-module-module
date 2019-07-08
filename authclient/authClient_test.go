// +build integration

package authclient

import (
	"log"
	"sync"
	"testing"
	"time"
)

func TestSendAndReceiveMessage(t *testing.T) {

	client, err := New("user@theplatform.com", "GSGSGS!!!", "http://stg-admin.identity.auth.theplatform.com")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(client.GetToken())
	t.Log(client.GetToken())

	err = client.SignOut()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(client.GetToken())
	client.SignOut()
}

func TestResponseDecode(t *testing.T) {
	sampleResponse := `{
"signInResponse": {
"token": "Q4ppbgqEymry1Ml4g4v8sSCeoIC8sICC",
"duration": 315360000000,
"idleTimeout": 14400000,
"userName": "admin@theplatform.com",
"userId": "http://stg-admin.identity.auth.theplatform.com/idm/data/User/mps/1150127438"
}
}`

	signInResp, err := decodeResponse([]byte(sampleResponse))
	if err != nil {
		t.Fatal("response not unmarshaled")
	}

	t.Log("Successfully decoded token:", signInResp.Token)

}

func TestMultThreads(t *testing.T) {

	client, err := New("admin@theplatform.com", "Admin!!!", "http://stg-admin.identity.auth.theplatform.com")
	if err != nil {
		t.Fatal(err)
	}
	//client.SignIn()
	defer client.SignOut()

	var wg sync.WaitGroup
	callGetToken(1000, &wg, &client)
	callGetToken(1000, &wg, &client)
	callGetToken(1000, &wg, &client)

	wg.Wait()
}

func callGetToken(millisecs time.Duration, wg *sync.WaitGroup, ac *AuthClient) {
	wg.Add(1)
	go func() {
		log.Println("Start")
		defer wg.Done()
		now := time.Now()
		later := now.Add(millisecs * time.Millisecond)
		i := 0
		for time.Now().Before(later) {
			ac.GetToken()
			i++
		}
		log.Printf("Done, called GetToken %d times", i)
	}()
}
