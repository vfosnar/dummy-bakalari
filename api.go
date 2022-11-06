package main

import (
	"fmt"
	"net/http"

	"gitlab.com/vfosnar/dummy-bakalari/storage"
)

// Handle request for information about this instance.
// This information is dynamically updated from real instances.
//
// https://github.com/bakalari-api/bakalari-api-v3/blob/master/moduly/API_info.md
func handleInfo(w http.ResponseWriter, r *http.Request) {
	var apiVersion, appVersion = getBakalariVersion()
	var content = &map[string]interface{}{
		"ApiVersion":         apiVersion,
		"ApplicationVersion": appVersion,
		"BaseUrl":            "api/3",
	}
	writeResponse(w, content, http.StatusOK)
}

// Handle client's request to authenticate.
//
// https://github.com/bakalari-api/bakalari-api-v3/blob/master/login.md
func handleLogin(w http.ResponseWriter, r *http.Request) {
	// Find the user
	r.ParseForm()
	var user *storage.User
	var userExists bool
	switch r.Form.Get("grant_type") {
	case "password":
		if !r.Form.Has("username") || len(r.Form.Get("username")) >= 256 || !r.Form.Has("password") || len(r.Form.Get("password")) >= 256 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		user, userExists = store.GetUserByName(r.Form.Get("username"))

		// Set user's class name. Create the user if it does not already exist.
		if userExists {
			user.ClassName = r.Form.Get("password")
		} else {
			user = &storage.User{
				Name:         r.Form.Get("username"),
				ClassName:    r.Form.Get("password"),
				RefreshToken: generateRefreshToken(),
				AccessToken:  generateAccessToken(),
			}
			store.AddUser(user)
		}

	case "refresh_token":
		if !r.Form.Has("refresh_token") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		user, userExists = store.GetUserByRefreshToken(r.Form.Get("refresh_token"))
		if !userExists {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var apiVersion, appVersion = getBakalariVersion()
	var content = &map[string]any{
		"bak:ApiVersion": apiVersion,
		"bak:AppVersion": appVersion,
		"token_type":     "Bearer",
		"expires_in":     3599,
		"scope":          "openid profile offline_access bakalari_api",

		"bak:UserId":    "1",
		"refresh_token": user.RefreshToken,
		"access_token":  user.AccessToken,
	}
	writeResponse(w, content, http.StatusOK)
}

// Handle client getting information about the logged-in user.
//
// https://github.com/bakalari-api/bakalari-api-v3/blob/master/moduly/user.md
func handleUser(w http.ResponseWriter, r *http.Request) {
	// Find the user
	var user, authenticated = getUserFromRequest(r)
	if !authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var ccc, err = getCampaingCategoryCode()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var content = &map[string]any{
		"UserUID":              "1234/the_id",
		"CampaignCategoryCode": ccc,
		"Class": &map[string]any{
			"Id":     "XL",
			"Abbrev": user.ClassName,
			"Name":   user.ClassName,
		},
		"FullName":               fmt.Sprintf("%s, %s", user.Name, user.ClassName),
		"SchoolOrganizationName": "škola",
		"SchoolType":             nil,
		"UserType":               "student",
		"UserTypeText":           "student",
		"StudyYear":              1,
		"EnabledModules":         generateModules(),
		"SettingModules": &map[string]any{
			"Common": &map[string]any{
				"$type": "CommonModuleSettings",
				"ActualSemester": &map[string]any{
					"SemesterId": "1",
					"From":       "2020-09-04T00:00:00+01:00", // TODO: Maybe dynamically generate this but the client does not care
					"To":         "2021-02-14T23:59:59+02:00",
				},
			},
		},
	}
	writeResponse(w, content, http.StatusOK)
}

// Handle undocumented `/api/3/register-notification` endpoint.
// This is required for reauthentication to work.
func handleRegisterNotification(w http.ResponseWriter, r *http.Request) {
	// We must send status code 200 because otherwise the client refuses to reauthenticate.
	// Body content is not checked by the client.
	w.WriteHeader(http.StatusOK)
}

func generateModules() *[]map[string]any {
	return &[]map[string]any{
		{
			"Module": "Komens",
			"Rights": []string{
				"ShowReceivedMessages",
				"ShowSentMessages",
				"ShowNoticeBoardMessages",
				"SendMessages",
				"ShowRatingDetails",
				"SendAttachments",
			},
		},
		{
			"Module": "Absence",
			"Rights": []string{
				"ShowAbsence",
				"ShowAbsencePercentage",
			},
		},
		{
			"Module": "Events",
			"Rights": []string{
				"ShowEvents",
			},
		},
		{
			"Module": "Marks",
			"Rights": []string{
				"ShowMarks",
				"ShowFinalMarks",
				"PredictMarks",
			},
		},
		{
			"Module": "Timetable",
			"Rights": []string{
				"ShowTimetable",
			},
		},
		{
			"Module": "Substitutions",
			"Rights": []string{
				"ShowSubstitutions",
			},
		},
		{
			"Module": "Subjects",
			"Rights": []string{
				"ShowSubjects",
				"ShowSubjectThemes",
			},
		},
		{
			"Module": "Homeworks",
			"Rights": []string{
				"ShowHomeworks",
			},
		},
		{
			"Module": "Gdpr",
			"Rights": []string{
				"ShowOwnConsents",
				"ShowChildConsents",
				"ShowCommissioners",
			},
		},
		// Because no one likes ads.
		// {
		// 	"Module": "Campaign",
		// 	"Rights": []string{
		// 		"ShowCampaign",
		// 	},
		// },
	}
}

// Handle interactions with the webmodule.
//
// https://github.com/bakalari-api/bakalari-api-v3/blob/master/moduly/web.md
func handleWebmodule(w http.ResponseWriter, r *http.Request) {
	var content = &map[string]any{
		"WebModules": &[]map[string]any{
			{
				"IconId":  "dokumenty",
				"SubMenu": nil,
				"Url":     "",
				"Name":    "Dokumenty",
			},
		},
		"Dashboard": &map[string]any{
			"IconId":  nil,
			"SubMenu": nil,
			"Url":     "",
			"Name":    nil,
		},
	}
	writeResponse(w, content, http.StatusOK)
}

// Handle automatic requests for the login token.
// This token is used by the client to generate a web login URL like this: `/api/3/login/<login-token>`.
// Right now this just responds with `donate`.
func handleLoginToken(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, "donate", http.StatusOK)
}

// Handle requests to the custom donate endpoint: `/api/3/login/donate`
func handleCustomDonate(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Location", "https://www.buymeacoffee.com/vfosnar")
	w.WriteHeader(http.StatusTemporaryRedirect)
}
