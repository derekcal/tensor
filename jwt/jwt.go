package jwt

import (
	"github.com/pearsonappeng/tensor/db"
	"github.com/pearsonappeng/tensor/models/common"

	"errors"
	"time"

	"github.com/Sirupsen/logrus"
	"gopkg.in/dgrijalva/jwt-go.v3"
	"gopkg.in/mgo.v2/bson"
)

type LocalToken struct {
	Token  string
	Expire string
}

func NewAuthToken(t *LocalToken) error {
	// Initial middleware default setting.
	HeaderAuthMiddleware.MiddlewareInit()

	// Create the token
	token := jwt.New(jwt.GetSigningMethod(HeaderAuthMiddleware.SigningAlgorithm))
	claims := token.Claims.(jwt.MapClaims)

	var admin common.User

	if err := db.Users().Find(bson.M{"username": "admin"}).One(&admin); err != nil {
		logrus.Errorln("User not found, Create JWT Token faild")
		return errors.New("User not found, Create JWT Token faild")
	}

	expire := time.Now().Add(HeaderAuthMiddleware.Timeout)
	claims["id"] = admin.ID
	claims["exp"] = expire.Unix()
	claims["orig_iat"] = time.Now().Unix()

	tokenString, err := token.SignedString(HeaderAuthMiddleware.Key)

	if err != nil {
		logrus.Errorln("Create JWT Token faild")
		return errors.New("Create JWT Token faild")
	}

	t.Token = tokenString
	t.Expire = expire.Format(time.RFC3339)
	return nil
}
