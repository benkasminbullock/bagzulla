// Store the user/cookie/session in the database.

package main

import (
	"bagzulla/bagzullaDb"
	"database/sql"
	"fmt"
)

// This implements the interface of login.LoginStore.
type baguser struct {
	b *Bagapp
}

func (bu *baguser) CheckPassword(user string, password string) (found bool) {
	person, err := bagzullaDb.PersonFromName(bu.b.db, user)
	if person.PersonId == 0 {
		return false
	}
	if err != nil {
		return false
	}
	if person.Password == password {
		return true
	}
	return false
}

var deleteCookieSQL = `DELETE FROM session WHERE cookie=?`
var deleteCookieStmt *sql.Stmt

func (bu *baguser) DeleteCookie(cookie string) (err error) {
	if len(cookie) == 0 {
		return fmt.Errorf("Empty cookie")
	}
	if deleteCookieStmt == nil {
		deleteCookieStmt, err = bu.b.db.Prepare(deleteCookieSQL)
		if err != nil {
			return err
		}
	}
	_, err = deleteCookieStmt.Exec(cookie)
	return err
}

func (bu *baguser) FindUser(user string) (found bool) {
	person, err := bagzullaDb.PersonFromName(bu.b.db, user)
	if person.PersonId == 0 {
		return false
	}
	if err != nil {
		return false
	}
	return true
}

func (bu *baguser) LookUpCookie(cookie string) (user string, found bool, err error) {
	var userID int64
	userID, found, err = SearchCookie(bu.b, cookie)
	if err != nil || !found {
		return "", found, err
	}
	person, err := bagzullaDb.PersonFromId(bu.b.db, userID)
	if err != nil {
		return "", false, err
	}
	return person.Name, true, nil
}

var storeLogin = `INSERT INTO session(person_id,cookie,start) VALUES(?,?,CURRENT_TIMESTAMP)`
var storeLoginStmt *sql.Stmt

func (bu *baguser) StoreLogin(user string, cookie string) (err error) {
	person, err := bagzullaDb.PersonFromName(bu.b.db, user)
	pid := person.PersonId
	if storeLoginStmt == nil {
		storeLoginStmt, err = bu.b.db.Prepare(storeLogin)
		if err != nil {
			return err
		}
	}
	_, err = storeLoginStmt.Exec(pid, cookie)
	return err
}
