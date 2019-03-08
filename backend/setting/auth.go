package setting

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"strings"
)

type Auth struct {
	Username string
	Password string
}

// Negroni compatible interface
func (c Auth) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)

	if len(auth) != 2 || auth[0] != "MD5" {
		http.Error(w, "authorization failed", http.StatusUnauthorized)
		return
	}

	payload, _ := base64.StdEncoding.DecodeString(auth[1])
	pair := strings.SplitN(string(payload), ":", 2)

	if len(pair) != 2 || !c.validate(pair[0], pair[1]) {
		http.Error(w, "authorization failed", http.StatusUnauthorized)
		return
	}
	next(w,r)
}

func (c Auth) validate(username, password string) bool {
	if username == c.Username && password == c.Password {
		return true
	}
	return false
}

func getMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func (c *Auth) Hash () {
	c.Username = getMD5Hash(c.Username)
	c.Password = getMD5Hash(c.Password)
}

func HandleDeny(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "Permission denied!", http.StatusForbidden)
}
