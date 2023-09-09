package cookie

import (
	"time"
)

type CookieSameSite string

const (
	// CookieSameSiteLaxMode sets the SameSite flag with the "Lax" parameter
	CookieSameSiteLaxMode = "Lax"
	// CookieSameSiteStrictMode sets the SameSite flag with the "Strict" parameter
	CookieSameSiteStrictMode = "Strict"
	// CookieSameSiteNoneMode sets the SameSite flag with the "None" parameter
	// see https://tools.ietf.org/html/draft-west-cookie-incrementalism-00
	CookieSameSiteNoneMode = "None"
)

type Cookie struct {
	Key      string
	Value    string
	Expire   time.Time
	MaxAge   int
	Domain   string
	Path     string
	HttpOnly bool
	Secure   bool
	SameSite CookieSameSite
}
