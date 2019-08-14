package main

import (
	"context"
	"fmt"

	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/qlik-oss/enigma-go"
)

func main() {

	// Set up the dialer and Header
	dialer := enigma.Dialer{}

	//Create cookiejar
	jar, err := dummyCookieJar()
	if err != nil {
		fmt.Print(err)
	}

	dialer.Jar = jar

	// Create a session with header
	ctx := context.Background()
	global, _ := enigma.Dialer{}.Dial(ctx, "ws://localhost:9076/app/engineData", nil)

	if err != nil {
		fmt.Println("Could not connect", err)
		panic(err)
	}
	version, _ := global.EngineVersion(ctx)
	global.DisconnectFromServer()
	fmt.Println(fmt.Sprintf("Connected to engine version %s.", version.ComponentVersion))

}

// Creates a cookiejar fills it with cookies
func dummyCookieJar() (jar http.CookieJar, err error) {

	// Fill Jar with cookies
	jar, err = cookiejar.New(nil)
	header := http.Header{}
	exp := fmt.Sprintf("%v", time.Now().Local().Add(time.Hour*time.Duration(48)).UTC())
	header.Add("Set-Cookie", "_session=a518840f-893b-4baf-bdf8-10d78ec14bf5; path=/; expires="+exp+"; secure; httponly")
	header.Add("Set-Cookie", "_grant=1d3cdfb9-25d0-42b2-8274-d4b11b97a475; path=/interaction/1d3cdfb9-25d0-42b2-8274-d4b11b97a475; expires="+exp+"; secure; httponly")
	header.Add("Set-Cookie", "_grant=1d3cdfb9-25d0-42b2-8274-d4b11b97a475; path=/auth/1d3cdfb9-25d0-42b2-8274-d4b11b97a475; expires="+exp+"; secure; httponly")
	response := http.Response{Header: header}
	cookies := response.Cookies()

	// Set the cookies
	url, err := url.Parse("ws://localhost:9076/app/engineData")
	jar.SetCookies(url, cookies)

	return
}
