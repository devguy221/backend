package objects

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/zekroTJA/myrunes/internal/auth"
	"github.com/zekroTJA/myrunes/internal/static"

	"github.com/bwmarrin/snowflake"
)

var userIDCLuster, _ = snowflake.NewNode(static.ClusterIDUsers)

var allowedUNameChars = regexp.MustCompile(`[\w_\-]+`)

var (
	ErrInvalidUsername = errors.New("invalid username")
)

type User struct {
	UID         snowflake.ID   `json:"uid"`
	Username    string         `json:"username"`
	DisplayName string         `json:"displayname"`
	PassHash    []byte         `json:"passhash,omitempty"`
	LastLogin   time.Time      `json:"lastlogin"`
	Created     time.Time      `json:"created"`
	Favorites   []string       `json:"favorites"`
	PageOrder   []snowflake.ID `json:"pageorder"`
}

func NewUser(username, password string, authMiddleware auth.Middleware) (*User, error) {
	now := time.Now()
	passHash, err := authMiddleware.CreateHash([]byte(password))
	if err != nil {
		return nil, err
	}

	user := &User{
		Created:     now,
		LastLogin:   now,
		PassHash:    passHash,
		UID:         userIDCLuster.Generate(),
		Username:    strings.ToLower(username),
		DisplayName: username,
		Favorites:   []string{},
	}

	return user, nil
}

func (u *User) Validate(acceptEmptyUsername bool) error {
	if (!acceptEmptyUsername && len(u.Username) < 3) ||
		len(allowedUNameChars.FindAllString(u.Username, -1)) > 1 {

		return ErrInvalidUsername
	}

	return nil
}
