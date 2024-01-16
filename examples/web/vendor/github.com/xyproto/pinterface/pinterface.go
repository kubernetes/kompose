// Package pinterface provides interface types for the xyproto/simple* and xyproto/permission* packages
package pinterface

import "net/http"

// Version is the API version. The API is stable within the same major version number
const Version = 5.3

// Database interfaces

type IList interface {
	Add(value string) error
	All() ([]string, error)
	Clear() error
	LastN(n int) ([]string, error)
	Last() (string, error)
	Remove() error
}

type ISet interface {
	Add(value string) error
	All() ([]string, error)
	Clear() error
	Del(value string) error
	Has(value string) (bool, error)
	Remove() error
}

type IHashMap interface {
	All() ([]string, error)
	Clear() error
	DelKey(owner, key string) error
	Del(key string) error
	Exists(owner string) (bool, error)
	Get(owner, key string) (string, error)
	Has(owner, key string) (bool, error)
	Keys(owner string) ([]string, error)
	Remove() error
	Set(owner, key, value string) error
}

type IHashMap2 interface {
	All() ([]string, error)
	AllWhere(key, value string) ([]string, error)
	Clear() error
	Count() (int64, error)
	DelKey(owner, key string) error
	Del(key string) error
	Empty() (bool, error)
	Exists(owner string) (bool, error)
	GetMap(owner string, keys []string) (map[string]string, error)
	Get(owner, key string) (string, error)
	Has(owner, key string) (bool, error)
	Keys(owner string) ([]string, error)
	Remove() error
	SetLargeMap(all map[string]map[string]string) error
	SetMap(owner string, m map[string]string) error
	Set(owner, key, value string) error
}

type IKeyValue interface {
	Clear() error
	Del(key string) error
	Get(key string) (string, error)
	Inc(key string) (string, error)
	Remove() error
	Set(key, value string) error
}

// Interface for making it possible to depend on different versions of the permission package,
// or other packages that implement userstates.
type IUserState interface {
	AddUnconfirmed(username, confirmationCode string)
	AddUser(username, password, email string)
	AdminRights(req *http.Request) bool
	AllUnconfirmedUsernames() ([]string, error)
	AllUsernames() ([]string, error)
	AlreadyHasConfirmationCode(confirmationCode string) bool
	BooleanField(username, fieldname string) bool
	ClearCookie(w http.ResponseWriter)
	ConfirmationCode(username string) (string, error)
	ConfirmUserByConfirmationCode(confirmationcode string) error
	Confirm(username string)
	CookieSecret() string
	CookieTimeout(username string) int64
	CorrectPassword(username, password string) bool
	Email(username string) (string, error)
	FindUserByConfirmationCode(confirmationcode string) (string, error)
	GenerateUniqueConfirmationCode() (string, error)
	HashPassword(username, password string) string
	HasUser(username string) bool
	IsAdmin(username string) bool
	IsConfirmed(username string) bool
	IsLoggedIn(username string) bool
	Login(w http.ResponseWriter, username string) error
	Logout(username string)
	MarkConfirmed(username string)
	PasswordAlgo() string
	PasswordHash(username string) (string, error)
	RemoveAdminStatus(username string)
	RemoveUnconfirmed(username string)
	RemoveUser(username string)
	SetAdminStatus(username string)
	SetBooleanField(username, fieldname string, val bool)
	SetCookieSecret(cookieSecret string)
	SetCookieTimeout(cookieTime int64)
	SetLoggedIn(username string)
	SetLoggedOut(username string)
	SetMinimumConfirmationCodeLength(length int)
	SetPasswordAlgo(algorithm string) error
	SetPassword(username, password string)
	SetUsernameCookie(w http.ResponseWriter, username string) error
	UsernameCookie(req *http.Request) (string, error)
	Username(req *http.Request) string
	UserRights(req *http.Request) bool

	Creator() ICreator
	Host() IHost
	Users() IHashMap
}

// Data structure creator
type ICreator interface {
	NewHashMap(id string) (IHashMap, error)
	NewKeyValue(id string) (IKeyValue, error)
	NewList(id string) (IList, error)
	NewSet(id string) (ISet, error)
}

// Database host (or file)
type IHost interface {
	Close()
	Ping() error
}

// Redis host (implemented structures can also be an IHost, of course)
type IRedisHost interface {
	DatabaseIndex()
	Pool()
}

// Redis data structure creator
type IRedisCreator interface {
	SelectDatabase(dbindex int)
}

// Middleware for permissions
type IPermissions interface {
	AddAdminPath(prefix string)
	AddPublicPath(prefix string)
	AddUserPath(prefix string)
	Clear()
	DenyFunction() http.HandlerFunc
	Rejected(w http.ResponseWriter, req *http.Request) bool
	ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc)
	SetAdminPath(pathPrefixes []string)
	SetDenyFunction(f http.HandlerFunc)
	SetPublicPath(pathPrefixes []string)
	SetUserPath(pathPrefixes []string)
	UserState() IUserState
}
