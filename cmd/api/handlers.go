package main

import (
	"errors"
	"log"
	"net/http"
)

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	log.Println("in home")

	var payload = struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Version string `json:"version"`
	}{
		Status:  "active",
		Message: "Go Movies up and running",
		Version: "1.0.0",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *application) AllMovies(w http.ResponseWriter, r *http.Request) {
	log.Println("in all movies")

	movies, err := app.DB.AllMovies()
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, movies)
}

func (app *application) authenticate(w http.ResponseWriter, r *http.Request) {
	log.Println("in authenticate")
	// Read a json payload
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// validate the user against database
	user, err := app.DB.GetUserByEmail(requestPayload.Email)

	if err != nil {
		log.Println("user not found")
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	// check password
	valid, err := user.PasswordMatches(requestPayload.Password)

	if err != nil || !valid {
		log.Println("password not valid", valid)
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	// create a jwt user
	u := jwtUser{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	// generate tokens
	tokens, err := app.auth.GenerateTokenPairs(&u)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	// log.Println(tokens.Token)

	refreshCookie := app.auth.GetRefreshCookie(tokens.RefreshToken)
	// send cookie to browser
	http.SetCookie(w, refreshCookie)

	// w.Write([]byte(tokens.Token))
	app.writeJSON(w, http.StatusAccepted, tokens)
}
