package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"gitlab.com/vfosnar/dummy-bakalari/storage"
)

var API_ACCESS_HEADER_PATTERN = regexp.MustCompile("^Bearer (.+)$")

// Authenticate the user.
func getUserFromRequest(r *http.Request) (user *storage.User, authenticated bool) {
	// Extract access token from the header
	var header = r.Header.Get("authorization")
	var match = API_ACCESS_HEADER_PATTERN.FindStringSubmatch(header)
	if match == nil {
		return nil, false
	}
	var accessToken = match[1]

	// Find the user using given access token in the storage
	var exists bool
	user, exists = store.GetUserByAccessToken(accessToken)
	if !exists {
		return nil, false
	}
	return user, true
}

// Write JSON serialized response.
func writeResponse(w http.ResponseWriter, content any, statusCode int) {
	// Serialize response body
	var s, err = json.Marshal(content)
	if err != nil {
		return
	}

	// Write response
	w.Header().Add("content-type", "application/json; charset=utf-8")
	w.Header().Add("content-length", fmt.Sprint(len(s)))
	w.WriteHeader(statusCode)
	w.Write(s)
}

// Internal function for doing a GET request to an offical Bakaláři server API.
// The `response` parameter must be a pointer.
func apiGetJson(instance string, endpoint string, response any) (err error) {
	// Make a request
	data, err := apiGetRequest(instance, endpoint)
	if err != nil {
		return err
	}

	// Unmarshal the data
	err = json.Unmarshal(*data, response)
	if err != nil {
		return err
	}
	return nil
}

func apiGetRequest(instance string, endpoint string) (data *[]byte, err error) {
	// Build a request
	request, _ := http.NewRequest("GET", instance+endpoint, nil)
	request.Header.Set("content-type", "application/json; charset=utf-8")
	request.Header.Set("accept", "application/json; charset=utf-8")

	// Fetch header
	httpResponse, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	// Fetch and return the data
	body, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}
	return &body, nil
}

type apiResponseMunicipalityCity struct {
	Name        string `json:"name"`
	SchoolCount int    `json:"schoolCount"`
}

type apiResponseMunicipalityCityDetails struct {
	Name    string                          `json:"name"`
	Schools []apiResponseMunicipalitySchool `json:"schools"`
}

type apiResponseMunicipalitySchool struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	SchoolUrl string `json:"schoolUrl"`
}

type apiResponseInfo struct {
	ApiVersion         string
	ApplicationVersion string
	BaseUrl            string
}

var bakalariVersionCheckLock sync.Mutex
var bakalariVersionApiValue = "3.23.0"
var bakalariVersionAppValue = "1.52.1102.1"
var bakalariVersionFetchTime time.Time
var bakalariVersionIsBeingUpdated = false

// Get the most used Bakaláři version.
func getBakalariVersion() (api string, app string) {
	// Check if version data should be updated
	bakalariVersionCheckLock.Lock()
	if bakalariVersionFetchTime.Add(time.Hour).Before(time.Now()) && !bakalariVersionIsBeingUpdated {
		bakalariVersionIsBeingUpdated = true
		go updateBakalariVersion()
	}
	bakalariVersionCheckLock.Unlock()

	// Return current stored value
	return bakalariVersionApiValue, bakalariVersionAppValue
}

func updateBakalariVersion() {
	log.Println("Running Bakaláři version update.")

	// Request list of all available cities
	var cities []apiResponseMunicipalityCity
	if err := apiGetJson("https://sluzby.bakalari.cz", "/api/v1/municipality", &cities); err != nil {
		log.Println("Updating Bakaláři version failed:", err.Error())
		return
	}

	// Find the most used Bakaláři version
	apiVersionFrequency := make(map[string]int)
	appVersionFrequency := make(map[string]int)
	for x := 0; x < 10; x++ {
		// Pick a random city with at least 1 school
		cityIndex := rand.Intn(len(cities))
		cityName := cities[cityIndex].Name
		encodedCityName := url.PathEscape(cityName)
		var cityDetails apiResponseMunicipalityCityDetails
		if err := apiGetJson("https://sluzby.bakalari.cz", "/api/v1/municipality/"+encodedCityName, &cityDetails); err != nil {
			log.Println("Updating Bakaláři version failed:", err.Error())
			x--
			continue
		}
		if len(cityDetails.Schools) == 0 {
			x--
			continue
		}

		// Pick a random school
		schoolIndex := rand.Intn(len(cityDetails.Schools))
		schoolUrl := strings.TrimRight(cityDetails.Schools[schoolIndex].SchoolUrl, "/")

		// Request school's information
		var schoolInfo apiResponseInfo
		if err := apiGetJson(schoolUrl, "/api/3", &schoolInfo); err != nil {
			log.Printf("Failed to fetch API version from \"%s\".\n", schoolUrl)
			x--
			continue
		}

		// Add API version to the frequency map
		if _, exists := apiVersionFrequency[schoolInfo.ApiVersion]; !exists {
			apiVersionFrequency[schoolInfo.ApiVersion] = 0
		}
		apiVersionFrequency[schoolInfo.ApiVersion]++
		// Add Application version to the frequency map
		if _, exists := appVersionFrequency[schoolInfo.ApplicationVersion]; !exists {
			appVersionFrequency[schoolInfo.ApplicationVersion] = 0
		}
		appVersionFrequency[schoolInfo.ApplicationVersion]++
	}
	// Find the most frequent occurrences
	var bestApiVersion string
	var bestApiVersionFrequency int
	for apiVersion, frequency := range apiVersionFrequency {
		if frequency > bestApiVersionFrequency {
			bestApiVersion = apiVersion
			bestApiVersionFrequency = frequency
		}
	}
	var bestAppVersion string
	var bestAppVersionFrequency int
	for appVersion, frequency := range appVersionFrequency {
		if frequency > bestAppVersionFrequency {
			bestAppVersion = appVersion
			bestAppVersionFrequency = frequency
		}
	}

	bakalariVersionApiValue = bestApiVersion
	bakalariVersionAppValue = bestAppVersion
	bakalariVersionFetchTime = time.Now()
	bakalariVersionIsBeingUpdated = false
	log.Println("Finished Bakaláři version update.")
}

// Generate a campaign category code. We don't actually care about the value but it should be a valid token.
//
// https://github.com/bakalari-api/bakalari-api-v3/blob/master/moduly/user.md#v%C3%BDznam-campaigncategorycode
func getCampaingCategoryCode() (string, error) {
	var content = map[string]any{
		"sid": "1234",
		"ut":  69,
		"sy":  1,
	}

	// Serialize the content
	var s, err = json.Marshal(content)
	if err != nil {
		return "", err
	}

	// Encode using base64 and return
	return base64.URLEncoding.EncodeToString([]byte(s)), nil
}

func generateAccessToken() string {
	return base64.URLEncoding.EncodeToString(generateRandomBytes(1500))
}

func generateRefreshToken() string {
	return base64.URLEncoding.EncodeToString(generateRandomBytes(1500))
}

func generateRandomBytes(n int) []byte {
	array := make([]byte, n)
	for i := range array {
		array[i] = byte(rand.Intn(256))
	}
	return array
}
