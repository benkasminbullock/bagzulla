package main

/* This handles the log in and log out of the user. */

import (
	"bagzulla/bagzullaDb"
	"log"
	"net/http"
)

// The name of the cookie which is used to authenticate the user.
var cookieName = "session"
var cookiePath = "/"

// Print an error page in case of authentication failure.
func (b *Bagreply) authError(format string, a ...interface{}) {
	format = "<b>Authentication error:</b> " + format
	b.errorPage(format, a...)
}

// Get the session
func (b *Bagreply) getSession() (user bagzullaDb.Person, found bool, ok bool) {
	login, err := b.App.login.User(b.w, b.r)
	if err != nil {
		b.errorPage("Error getting user from cookie: %s", err)
		return user, false, false
	}
	user, err = bagzullaDb.PersonFromName(b.App.db, login)
	if err != nil {
		b.errorPage("Error getting user details for name '%s': %s", login, err)
		return user, false, false
	}
	// For some reason bagzullaDb.go doesn't fill in the Name field of
	// the Person structure.
	user.Name = login
	return user, true, true
}

type loginform struct {
	Referer string
}

// Indicate that the requested action is not possible due to not being
// logged in.
func (b *Bagreply) ErrorLogin() {
	if !b.runATemplate("top.html", b) {
		return
	}
	var ep = ErrorPage{
		Text: "You are not logged in, and so cannot make changes.",
	}
	if !b.runATemplate("error.html", ep) {
		return
	}
	var lf loginform
	// We don't want to go back to the referrer, we want to go back to the
	// caller function's URL.

	// The following rigamarole seems to be necessary.
	abs := b.AbsRef(b.r.Referer())
	ref := abs + b.r.RequestURI
	lf.Referer = ref
	if !b.runATemplate("login.html", lf) {
		return
	}
	if !b.runATemplate("bottom.html", b) {
		return
	}
}

// This is called from the form page /login/.
func loginHandler(b *Bagreply) {
	name := b.r.FormValue("name")
	if len(name) == 0 {
		// Show the login form with a hidden value containing the
		// referring page.
		var lf loginform
		lf.Referer = b.r.Referer()
		b.runTemplate("login.html", lf)
		return
	}
	if b.r.Method != "POST" {
		b.errorPage("Must be a POST request")
		return
	}
	password := b.r.FormValue("password")
	if len(password) == 0 {
		b.errorPage("No password")
		return
	}
	err := b.App.login.LogIn(b.w, b.r, name, password)
	if err != nil {
		b.errorPage("Login failed: %s", err)
		return
	}
	if debugLogin {
		log.Printf("Logged in.\n")
	}
	// Go back to the page where the login says it was referred from.
	referer := b.r.FormValue("referer")
	http.Redirect(b.w, b.r, referer, http.StatusFound)
}

// This is called directly by /logout/, so it is not from a form, so
// just get the referrer from b.r.
func logoutHandler(b *Bagreply) {
	b.App.login.LogOut(b.w, b.r)
	referer := b.r.Referer()
	http.Redirect(b.w, b.r, referer, http.StatusFound)
}
