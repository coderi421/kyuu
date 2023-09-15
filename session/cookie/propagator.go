package cookie

import (
	"net/http"
)

type PropagatorOption func(propagator *Propagator)

type Propagator struct {
	cookieName string
	cookieOpt  func(c *http.Cookie)
}

// WithCookieOption creates a PropagatorOption that sets the cookie option.
func WithCookieOption(opt func(c *http.Cookie)) PropagatorOption {
	return func(propagator *Propagator) {
		propagator.cookieOpt = opt
	}
}

// NewPropagator creates a new instance of the Propagator struct.
// It takes in a cookieName string and a variadic number of PropagatorOption functions.
// It returns a pointer to the newly created Propagator.
func NewPropagator(cookieName string, opts ...PropagatorOption) *Propagator {
	// Create a new instance of the Propagator struct with the given cookieName.
	res := &Propagator{
		cookieName: cookieName,
		cookieOpt:  func(c *http.Cookie) {},
	}

	// Apply each option function to the newly created Propagator.
	for _, opt := range opts {
		opt(res)
	}

	// Return the newly created Propagator.
	return res
}

// Inject sets a cookie with the provided session ID in the response writer.
// It only sets the basic attributes of the cookie, while other properties
// are set using the cookie option.
func (p *Propagator) Inject(id string, writer http.ResponseWriter) error {
	// Create the cookie object with the session ID.
	cookie := &http.Cookie{
		Name:  p.cookieName,
		Value: id, // id -> session id
		//Path:       "",
		//Domain:     "",
		//Expires:    time.Time{},
		//RawExpires: "",
		//MaxAge:     0,
		//Secure:     false,
		//HttpOnly:   false,
		//SameSite:   0,
		//Raw:        "",
		//Unparsed:   nil,
	}
	// Set additional properties of the cookie using the cookie option.
	p.cookieOpt(cookie)

	// Set the cookie in the response writer.
	http.SetCookie(writer, cookie)
	return nil
}

// Extract extracts the value of a cookie with the given name from the provided HTTP request.
// If the cookie is not found, it returns an empty string and an error.
func (p *Propagator) Extract(req *http.Request) (string, error) {
	// Retrieve the cookie with the given name from the request
	c, err := req.Cookie(p.cookieName)
	if err != nil {
		return "", err
	}

	// Return the value of the cookie
	return c.Value, nil
}

// Remove removes the cookie from the response writer by setting it to expire.
func (p *Propagator) Remove(writer http.ResponseWriter) error {
	// Create a new cookie with the same name and set its MaxAge to -1 to make it expire.
	cookie := &http.Cookie{
		Name:   p.cookieName,
		MaxAge: -1,
	}

	// Apply any additional cookie options.
	p.cookieOpt(cookie)

	// Set the cookie in the response writer.
	http.SetCookie(writer, cookie)

	return nil
}
