package main

import (
	"github.com/martini-contrib/sessionauth"
	"github.com/juju/errors"
	"fmt"
)

type CodisUser struct {
	Id            string `form:"id" db:"id"`
	Username      string `form:"name" db:"username"`
	Password      string `form:"password" db:"password"`
	authenticated bool   `form:"-" db:"-"`
}

// GetAnonymousUser should generate an anonymous user model
// for all sessions. This should be an unauthenticated 0 value struct.
func GenerateAnonymousUser() sessionauth.User {
	return &CodisUser{}
}

// Login will preform any actions that are required to make a user model
// officially authenticated.
func (u *CodisUser) Login() {
	// Update last login time
	// Add to logged-in user's list
	// etc ...
	u.authenticated = true
}

// Logout will preform any actions that are required to completely
// logout a user.
func (u *CodisUser) Logout() {
	// Remove from logged-in user's list
	// etc ...
	u.authenticated = false
}

func (u *CodisUser) IsAuthenticated() bool {
	return u.authenticated
}

func (u *CodisUser) UniqueId() interface{} {
	return u.Id
}

// GetById will populate a user object from a database model with
// a matching id.
func (u *CodisUser) GetById(id interface{}) error {
	password, ok := globalEnv.AuthUsers()[id.(string)]
	if !ok {
		return errors.UserNotFoundf(fmt.Sprintf("%s", id))
	}
	u.Id = id.(string)
	u.Username = id.(string)
	u.Password = password

	return nil
}