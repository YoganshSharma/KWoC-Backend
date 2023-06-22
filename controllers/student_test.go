package controllers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"kwoc-backend/controllers"
	"kwoc-backend/utils"
	"net/http"
	"testing"

	"gorm.io/gorm"
)

func createStudentRegRequest(reqFields *controllers.RegisterStudentReqFields) *http.Request {
	reqBody, _ := json.Marshal(reqFields)

	req, _ := http.NewRequest(
		"POST",
		"/student/form/",
		bytes.NewReader(reqBody),
	)

	return req
}

// Test unauthenticated request to /student/form/
func TestStudentRegNoAuth(t *testing.T) {
	testRequestNoAuth(t, "POST", "/student/form/")
}

// Test request to /student/form/ with invalid jwt
func TestStudentRegInvalidAuth(t *testing.T) {
	testRequestInvalidAuth(t, "POST", "/student/form/")
}

// Test request to /student/form/ with session hijacking attempt
func TestStudentRegSessionHijacking(t *testing.T) {
	// Generate a jwt secret key for testing
	setTestJwtSecretKey()

	testLoginFields := utils.LoginJwtFields{Username: "someuser"}

	someuserJwt, _ := utils.GenerateLoginJwtString(testLoginFields)

	reqFields := controllers.RegisterStudentReqFields{Username: "anotheruser"}

	req := createStudentRegRequest(&reqFields)
	req.Header.Add("Bearer", someuserJwt)

	res := executeRequest(req, nil)

	expectStatusCodeToBe(t, res, http.StatusUnauthorized)
	expectResponseBodyToBe(t, res, "Login username and given username do not match.")
}

// Test a new user registration request to /student/form/ with proper authentication and input
func tStudentRegNewUser(db *gorm.DB, t *testing.T) {
	// Test login fields
	testUsername := getTestUsername()
	testLoginFields := utils.LoginJwtFields{Username: testUsername}

	testJwt, _ := utils.GenerateLoginJwtString(testLoginFields)
	reqFields := controllers.RegisterStudentReqFields{Username: testUsername}

	req := createStudentRegRequest(&reqFields)
	req.Header.Add("Bearer", testJwt)

	res := executeRequest(req, db)

	expectStatusCodeToBe(t, res, http.StatusOK)
	expectResponseBodyToBe(t, res, "Student registration successful.")
}

// Test an existing user registration request to /student/form/ with proper authentication and input
func tStudentRegExistingUser(db *gorm.DB, t *testing.T) {
	// Test login fields
	testUsername := getTestUsername()
	testLoginFields := utils.LoginJwtFields{Username: testUsername}

	testJwt, _ := utils.GenerateLoginJwtString(testLoginFields)
	reqFields := controllers.RegisterStudentReqFields{Username: testUsername}

	req := createStudentRegRequest(&reqFields)
	req.Header.Add("Bearer", testJwt)

	_ = executeRequest(req, db)

	// Execute the same request again
	req = createStudentRegRequest(&reqFields)
	req.Header.Add("Bearer", testJwt)

	res := executeRequest(req, db)

	expectStatusCodeToBe(t, res, http.StatusBadRequest)
	expectResponseBodyToBe(t, res, fmt.Sprintf("Student `%s` already exists.", testUsername))
}

// Test requests to /student/form/ with proper authentication and input
func TestStudentRegOK(t *testing.T) {
	// Set up a local test database path
	db := setTestDB()
	defer unsetTestDB()

	// Generate a jwt secret key for testing
	setTestJwtSecretKey()
	defer unsetTestJwtSecretKey()

	// New student registration test
	t.Run(
		"Test: new student registration.",
		func(t *testing.T) {
			tStudentRegNewUser(db, t)
		},
	)

	// Existing student registration test
	t.Run(
		"Test: existing student registration.",
		func(t *testing.T) {
			tStudentRegExistingUser(db, t)
		},
	)
}