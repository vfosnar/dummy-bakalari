package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"gitlab.com/vfosnar/dummy-bakalari/storage"
)

var store storage.Storage

func main() {
	// Setup random seed
	rand.Seed(time.Now().UnixNano())

	// Setup in-memory user storage
	store = storage.NewMemoryStorage()

	// Handle default fallback route
	http.HandleFunc("/", handleDefault)

	// Handle Bakaláři routes
	http.HandleFunc("/api/3", handleInfo)
	http.HandleFunc("/api/login", handleLogin)
	http.HandleFunc("/api/3/user", handleUser)
	http.HandleFunc("/api/3/register-notification", handleRegisterNotification)
	http.HandleFunc("/api/3/webmodule", handleWebmodule)
	http.HandleFunc("/api/3/logintoken", handleLoginToken)
	http.HandleFunc("/api/3/login/donate", handleCustomDonate)

	// Start Bakaláři version update goroutine
	bakalariVersionCheckLock.Lock()
	bakalariVersionIsBeingUpdated = true
	go updateBakalariVersion()
	bakalariVersionCheckLock.Unlock()

	// Start the server at given address
	var address = os.Getenv("APP_ADDRESS")
	if address == "" {
		address = ":8080" // Default value
	}
	log.Printf("Listening on: %s", address)
	http.ListenAndServe(address, nil)
}

func handleDefault(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		handleHome(w, r)
		return
	}

	log.Printf("Unhandled path: %s\n", r.URL)
	w.WriteHeader(http.StatusBadRequest)
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, HOME_HTML)
}

const HOME_HTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Dummy Bakaláři</title>
    <style>
        :root {
            --c-text: #4c4f69;
            --c-bg: #eff1f5;
            --c-primary: #8839ef;
        }

        @media screen and (prefers-color-scheme: dark) {
            :root {
                --c-text: #cdd6f4;
                --c-bg: #1e1e2e;
                --c-primary: #cba6f7;
            }
        }

        body {
            margin: 2rem;

            color: var(--c-text);
            background-color: var(--c-bg);
            font-size: 20px;
			font-family: -apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,"Helvetica Neue",Arial,sans-serif,"Apple Color Emoji","Segoe UI Emoji","Segoe UI Symbol","Noto Color Emoji";
			text-align: center;
        }

        a {
            color: var(--c-primary);
        }
    </style>
</head>
<body>
    This is a dummy Bakaláři instance. Documentation can be found <a href="https://gitlab.com/vfosnar/dummy-bakalari">here</a> 📃
</body>
</html>`
