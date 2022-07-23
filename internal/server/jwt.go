// jwt authentication inspired by a blog post: https://www.sohamkamani.com/golang/jwt-authentication/

package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"go.opencensus.io/trace"

	"github.com/Neurostep/todo/pkg/tools/logging"
)

var jwtKey = []byte("secret")

var users = map[string]string{
	"user": "password",
}

type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

type AuthResponse struct {
	Token   string `json:"token"`
	Expires int64  `json:"expires"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func (r *api) signin(c *gin.Context) {
	ctx, span := trace.StartSpan(c.Request.Context(), "signin")
	defer span.End()
	logger := logging.FromContext(ctx, r.logger)

	var creds Credentials
	if err := c.ShouldBindJSON(&creds); err != nil {
		errs := extractBindErrors(err)
		respondErrors(c, logger, http.StatusBadRequest, errs...)
		return
	}

	expectedPassword, ok := users[creds.Username]

	if !ok || expectedPassword != creds.Password {
		respondErrors(c, logger, http.StatusUnauthorized, newError("signin", "incorrect password"))
		return
	}

	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &Claims{
		Username: creds.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{Time: expirationTime},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		respondErrors(c, logger, http.StatusInternalServerError, newError("signin", err.Error()))
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token:   tokenString,
		Expires: expirationTime.Unix(),
	})
}

func (r *api) refresh(c *gin.Context) {
	ctx, span := trace.StartSpan(c.Request.Context(), "refresh")
	defer span.End()
	logger := logging.FromContext(ctx, r.logger)

	cookie, err := c.Request.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			respondErrors(c, logger, http.StatusUnauthorized, newError("refresh", err.Error()))
			return
		}
		respondErrors(c, logger, http.StatusBadRequest, newError("refresh", err.Error()))
		return
	}
	tknStr := cookie.Value
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if !tkn.Valid {
		respondErrors(c, logger, http.StatusUnauthorized, newError("refresh", "not valid token"))
		return
	}
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			respondErrors(c, logger, http.StatusUnauthorized, newError("refresh", err.Error()))
			return
		}
		respondErrors(c, logger, http.StatusBadRequest, newError("refresh", err.Error()))
		return
	}

	if time.Unix(claims.ExpiresAt.Unix(), 0).Sub(time.Now()) > 30*time.Second {
		respondErrors(c, logger, http.StatusBadRequest, newError("refresh", "too early refresh"))
		return
	}

	expirationTime := time.Now().Add(5 * time.Minute)
	claims.ExpiresAt = &jwt.NumericDate{Time: expirationTime}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		respondErrors(c, logger, http.StatusInternalServerError, newError("refresh", err.Error()))
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token:   tokenString,
		Expires: expirationTime.Unix(),
	})
}

func authMiddleware(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	claims := &Claims{}

	tkn, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if !tkn.Valid {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.Next()
}
