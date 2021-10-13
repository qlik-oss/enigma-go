package main

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/qlik-oss/enigma-go/v3"
)

func main() {

	// Create dialer
	dialer := enigma.Dialer{}

	//Add a cookiejar
	jar, err := dummyCookieJar()
	if err != nil {
		fmt.Print(err)
	}

	//Set dialer jar
	dialer.Jar = jar

	//Print all session cookies with the URL "https://www.qlik.com/"
	urlTest, err := url.Parse("https://www.qlik.com/")

	cookies := dialer.Jar.Cookies(urlTest)
	if len(cookies) == 0 {
		fmt.Println("No cookies with this url")
	} else {
		for _, cookie := range cookies {
			if cookie.Name == "_session" {
				fmt.Printf("Session cookie found, id: %v \n", cookie)
			} else {
				fmt.Println("No session cookies found")
			}
		}

	}
}

func dummyCookieJar() (jar http.CookieJar, err error) {

	// Create an empty jar
	jar, err = cookiejar.New(nil)

	//Fill the jar with cookies
	header := http.Header{}
	exp := fmt.Sprintf("%v", time.Now().Local().Add(time.Hour*time.Duration(48)).UTC())
	header.Add("Set-Cookie", "_session=a518840f-893b-4baf-bdf8-10d78ec14bf5; path=/; expires="+exp+"; secure; httponly")
	header.Add("Set-Cookie", "_grant=1d3cdfb9-25d0-42b2-8274-d4b11b97a475; path=/interaction/1d3cdfb9-25d0-42b2-8274-d4b11b97a475; expires="+exp+"; secure; httponly")
	header.Add("Set-Cookie", "_grant=1d3cdfb9-25d0-42b2-8274-d4b11b97a475; path=/auth/1d3cdfb9-25d0-42b2-8274-d4b11b97a475; expires="+exp+"; secure; httponly")

	response := http.Response{Header: header}
	cookies := response.Cookies()

	// Set the cookies
	url, err := url.Parse("https://www.qlik.com")
	jar.SetCookies(url, cookies)

	return
}
