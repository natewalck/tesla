package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Package rand
const charset = "abcdefghijklmnopqrstuvwxyz" +
  "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
  rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
  b := make([]byte, length)
  for i := range b {
    b[i] = charset[seededRand.Intn(len(charset))]
  }
  return string(b)
}

func String(length int) string {
  return StringWithCharset(length, charset)
}

// End package rand

func main() {
    // Following along with https://tesla-api.timdorr.com/api-basics/authentication#get-https-auth-tesla-com-oauth-2-v-3-authorize
    // Create code_verifier and code_challenger for oAuth request
    codeVerifier := String(86)
    fmt.Println(codeVerifier)
    codeVerifier256 := sha256.Sum256([]byte(codeVerifier))
    fmt.Println(codeVerifier256)
    codeChallenge := base64.URLEncoding.EncodeToString(codeVerifier256[:])
    fmt.Println(codeChallenge)
    stateString := String(10)

    // Step 1
    // GET on https://auth.tesla.com/oauth2/v3/authorize
    authURL := "https://auth.tesla.com/oauth2/v3/authorize"
    req, _ := http.NewRequest("GET", authURL, nil)
    q := req.URL.Query()
    // These are all required
    q.Add("client_id", "ownerapi") // Always ownerapi
    q.Add("code_challenge", codeChallenge) // The codeChallenge from above
    q.Add("code_challenge_method", "S256") // The codeChallenge hash (sha256)
    q.Add("redirect_uri", "https://auth.tesla.com/void/callback") // always this
    q.Add("response_type", "code") // always this
    q.Add("scope", "openid email offline_access") // always this
    q.Add("state", stateString) // Anything. oAuth state value string
    q.Add("login_hint", "") // Email of the account to login with
    req.URL.RawQuery = q.Encode()
    fmt.Println(req.URL.String())
    requestDump, err := httputil.DumpRequest(req, true)
    if err != nil {
      fmt.Println(err)
    }
    fmt.Println("Dumping request")
    fmt.Println(string(requestDump))
    client := &http.Client{}
    resp, err := client.Do(req)

    if err != nil {
        fmt.Println("Errored when sending request to the server")
        return
    }
    defer resp.Body.Close()
    // We need to get the following from the response:
    // Session ID
    // _csrf hidden form input
    // _phase hidden form input
    // _process hidden form input
    // transaction_id hidden form input
    // cancel hidden form input
    // These will be used on the step 2 POST
    var sessionID string
    for _, cookie := range resp.Cookies() {
      if cookie.Name == "tesla-auth.sid" {
        sessionID = cookie.Value
      }
    }
    fmt.Println("Session ID is: ", sessionID)
    document, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
      log.Fatal("Error loading HTTP response body. ", err)
    }
    // Get values from response form that we need for next POST
    csrf, _ := document.Find("input[name='_csrf']").Attr("value")
    phase, _ := document.Find("input[name='_phase']").Attr("value")
    process, _ := document.Find("input[name='_process']").Attr("value")
    transactionID, _ := document.Find("input[name='transaction_id']").Attr("value")
    cancel, _ := document.Find("input[name='cancel']").Attr("value")
    fmt.Println(csrf)
    fmt.Println(phase)
    fmt.Println(process)
    fmt.Println(transactionID)
    fmt.Println(cancel)

    // Step 2
    // POST https://auth.tesla.com/oauth2/v3/authorize

}


