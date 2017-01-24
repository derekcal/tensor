package jwt

import (
	"github.com/pearsonappeng/tensor/db"
	"github.com/pearsonappeng/tensor/models"
	"github.com/pearsonappeng/tensor/util"
	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/appleboy/gin-jwt.v2"
	"gopkg.in/mgo.v2/bson"
	"net/mail"
	"strings"
	"time"
)

var HeaderAuthMiddleware *jwt.GinJWTMiddleware

func init() {
	HeaderAuthMiddleware = &jwt.GinJWTMiddleware{
		Realm:      "api",
		Key:        []byte(util.Config.Salt),
		Timeout:    time.Hour,
		MaxRefresh: time.Hour,
		Authenticator: func(loginid string, password string, c *gin.Context) (string, bool) {

			// Lowercase email or username
			login := strings.ToLower(loginid)

			var q bson.M

			if _, err := mail.ParseAddress(login); err == nil {
				q = bson.M{"email": login}

			} else {
				q = bson.M{"username": login}
			}

			var user models.User

			if err := db.Users().Find(q).One(&user); err != nil {
				log.Warningln("Auth: User not found", q)
				return "", false
			}

			if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
				log.Warningln("Auth: PasswordHash mismach")
				return "", false
			}

			return user.ID.Hex(), true

		},
		Authorizator: func(userID string, c *gin.Context) bool {
			var user models.User
			if err := db.Users().FindId(bson.ObjectIdHex(userID)).One(&user); err != nil {
				log.Warningln("Auth: User not found", userID)
				return false
			}

			// set user to gin context
			c.Set("user", user)
			return true
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "cookie:<name>"
		TokenLookup: "header:Authorization",
		// TokenLookup: "query:token",
		// TokenLookup: "cookie:token",
	}
}
