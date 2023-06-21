// Handler function for routes
package controllers

import (
	"encoding/json"
	"kwoc-backend/middleware"
	"kwoc-backend/utils"
	"net/http"

	"kwoc-backend/models"
)

type OAuthReqBodyFields struct {
	// Code generated by Github OAuth
	Code string `json:"code"`
	// `mentor` or `student`
	Type string `json:"type"`
}

type OAuthResBodyFields struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	// `mentor` or `student`
	Type string `json:"type"`
	// Whether the user has newly registered or was registered before
	IsNewUser bool   `json:"isNewUser"`
	Jwt       string `json:"jwt"`
}

func OAuth(w http.ResponseWriter, r *http.Request) {
	app := r.Context().Value(middleware.APP_CTX_KEY).(*middleware.App)
	db := app.Db

	var reqFields = OAuthReqBodyFields{}
	err := json.NewDecoder(r.Body).Decode(&reqFields)

	if err != nil {
		utils.LogErrAndRespond(r, w, err, "Error parsing JSON body parameters.", http.StatusBadRequest)
		return
	}

	if reqFields.Code == "" || reqFields.Type == "" {
		utils.LogWarnAndRespond(r, w, "Empty body parameters.", http.StatusBadRequest)
		return
	}

	// Get a Github OAuth access token
	accessToken, err := utils.GetOauthAccessToken(reqFields.Code)
	if err != nil {
		utils.LogErrAndRespond(r, w, err, "Error getting OAuth access token.", http.StatusInternalServerError)
		return
	}

	// Get the user's information from the Github API
	userInfo, err := utils.GetOauthUserInfo(accessToken)
	if err != nil {
		utils.LogErrAndRespond(r, w, err, "Error getting OAuth user info.", http.StatusInternalServerError)
		return
	}

	if userInfo.Username == "" {
		utils.LogWarnAndRespond(r, w, "Could not get username from the Github API.", http.StatusInternalServerError)
		return
	}

	// Check if the user has already registered
	var isNewUser bool = false

	if reqFields.Type == "student" {
		student := models.Student{}
		db.
			Table("students").
			Where("username = ?", userInfo.Username).
			First(&student)

		isNewUser = student.Username != userInfo.Username
	} else if reqFields.Type == "mentor" {
		mentor := models.Mentor{}
		db.
			Table("mentors").
			Where("username = ?", userInfo.Username).
			First(&mentor)

		isNewUser = mentor.Username != userInfo.Username
	}

	// Generate a JWT string for the user
	jwtString, err := utils.GenerateLoginJwtString(utils.LoginJwtFields{
		Username: userInfo.Username,
	})
	if err != nil {
		utils.LogErrAndRespond(r, w, err, "Error generating a JWT string.", http.StatusInternalServerError)
		return
	}

	resFields := OAuthResBodyFields{
		Username:  userInfo.Username,
		Name:      userInfo.Name,
		Email:     userInfo.Email,
		Type:      reqFields.Type,
		IsNewUser: isNewUser,
		Jwt:       jwtString,
	}

	resJson, err := json.Marshal(resFields)
	if err != nil {
		utils.LogErrAndRespond(r, w, err, "Error generating response JSON.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resJson)

	if err != nil {
		utils.LogErr(r, err, "Error writing the response.")

		return
	}
}
