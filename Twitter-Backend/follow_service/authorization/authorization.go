package authorization

import (
	"github.com/casbin/casbin"
	"github.com/cristalhq/jwt/v4"
	"log"
	"net/http"
	"os"
	"strings"
)

var jwtKey = []byte(os.Getenv("SECRET_KEY"))

var verifier, _ = jwt.NewVerifierHS(jwt.HS256, jwtKey)

func Authorizer(e *casbin.Enforcer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		fn := func(w http.ResponseWriter, r *http.Request) {

			if r.Header.Get("Authorization") == "" {
				res, err := e.EnforceSafe("NotLoggedIn", r.URL.Path, r.Method)
				if err != nil {
					log.Println("enforce error")
					http.Error(w, "unauthorized user", http.StatusUnauthorized)
					return
				}
				log.Println(res)
				if res {
					log.Println("redirect")
					next.ServeHTTP(w, r)
				} else {
					http.Error(w, "forbidden", http.StatusForbidden)
					return
				}

			} else {

				token, err := GetToken(r)
				if err != nil {
					log.Println(err)
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}

				claims := GetMapClaims(token.Bytes())

				res, err := e.EnforceSafe(claims["userType"], r.URL.Path, r.Method)
				if err != nil {
					log.Println("enforce error")
					http.Error(w, "unauthorized user", http.StatusUnauthorized)
					return
				}
				log.Println(res)

				if res {
					log.Println("redirect")
					next.ServeHTTP(w, r)
				} else {
					http.Error(w, "forbidden", http.StatusForbidden)
					return
				}
			}

		}

		return http.HandlerFunc(fn)
	}
}

func GetToken(r *http.Request) (*jwt.Token, error) {
	bearer := r.Header.Get("Authorization")
	bearerToken := strings.Split(bearer, "Bearer ")
	tokenString := bearerToken[1]
	token, err := jwt.Parse([]byte(tokenString), verifier)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return token, nil
}

func GetMapClaims(tokenBytes []byte) map[string]string {
	var claims map[string]string

	err := jwt.ParseClaims(tokenBytes, verifier, &claims)
	if err != nil {
		log.Println(err)
	}

	return claims
}
