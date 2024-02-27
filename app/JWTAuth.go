package main

import (
	"encoding/json"
	"github.com/go-chi/jwtauth"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

var UserBD = make(map[string]*User)

// @Summary Register a new user
// @ID registerUser
// @Tags SingUp
// @Accept  json
// @Produce  json
// @Param input body User true "User"
// @Success 201 "User registered successfully"
// @Failure 400 "Invalid request format"
// @Router /register [post]
func SingUpHandler(w http.ResponseWriter, r *http.Request) {
	var singUp User

	if err := json.NewDecoder(r.Body).Decode(&singUp); err != nil {
		http.Error(w, "Invalid SingUp", http.StatusBadRequest)
		return
	}

	if _, exists := UserBD[singUp.Login]; exists {
		http.Error(w, "User already exists", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(singUp.Password), bcrypt.DefaultCost)

	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	UserBD[singUp.Login] = &User{
		Login:    singUp.Login,
		Password: string(hashedPassword),
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User registered successfully"))
}

// @Summary Log in a user
// @ID loginUser
// @Tags SingIn
// @Accept  json
// @Produce  json
// @Param input body User true "User"
// @Success 200 "JWT token"
// @Failure 401 "Invalid credentials"
// @Router /login [post]
func SingInHandler(w http.ResponseWriter, r *http.Request) {
	var singIn User
	var tokenAuth = jwtauth.New("HS256", []byte("mySecret"), nil)
	if err := json.NewDecoder(r.Body).Decode(&singIn); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	user, exists := UserBD[singIn.Login]
	if !exists {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(singIn.Password))
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	_, tokenString, err := tokenAuth.Encode(map[string]interface{}{"username": user.Login})
	if err != nil {
		http.Error(w, "Generate token to failed", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Bearer " + tokenString))
}

func TokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _, err := jwtauth.FromContext(request.Context())
		if err != nil {
			http.Error(writer, "Invalid token", http.StatusForbidden)
			return
		}
		next.ServeHTTP(writer, request)
	})
}
