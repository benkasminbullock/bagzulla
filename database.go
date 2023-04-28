// bagzullaDb contains generic, generated code which handles most of
// the simple ways that one might want to access the database. This
// file handles other database statements which are not generic but
// crafted to get a specific result, like a list of open bugs rather
// than just a list of bugs.

package main

import (
	"bagzulla/bagzullaDb"
	"database/sql"
	"errors"
	"fmt"
)

// Prepare an SQL statement.

func PrepareSql(b *Bagreply, sql string) (*sql.Stmt, bool) {
	stmt, err := b.App.db.Prepare(sql)
	if err != nil {
		b.errorPage(fmt.Sprintf("Error preparing SQL %s: %s",
			sql, err.Error()))
		return stmt, false
	}
	return stmt, true
}

// Scan the rows of a list of bugs returned by a query to the database
// into "bugs".

func scanRows(b *Bagreply, rows *sql.Rows) (bugs []bagzullaDb.Bug, ok bool) {
	bugs, err := bagzullaDb.BugsFromRows(rows)
	if err != nil {
		b.errorPage("Error scanning for bugs: %s", err.Error())
		return bugs, false
	}
	return bugs, true
}

// Make a list of bugs with the specified status.

var statusBugsSql = `
SELECT * FROM bug WHERE bug.status=? ORDER BY bug.changed DESC 
`
var statusBugsStmt *sql.Stmt

func StatusBugs(b *Bagreply, status int) (bugs []bagzullaDb.Bug, ok bool) {
	if statusBugsStmt == nil {
		statusBugsStmt, ok = PrepareSql(b, statusBugsSql)
		if !ok {
			return bugs, false
		}
	}
	rows, err := statusBugsStmt.Query(status)
	defer rows.Close()
	if err != nil {
		b.errorPage("Error looking for status %d bugs: %s", status, err.Error())
		return bugs, false
	}
	bugs, ok = scanRows(b, rows)
	return bugs, ok
}

// Delete a part of a project

var setNoPartSQL = `
UPDATE bug SET part_id=0 WHERE part_id=?
`
var setNoPartStmt *sql.Stmt

func deletePart(b *Bagreply) {
	part, ok := getPart(b)
	if !ok {
		return
	}
	setNoPartStmt, ok := PrepareSql(b, setNoPartSQL)
	if !ok {
		return
	}
	_, err := setNoPartStmt.Exec(part.PartId)
	if err != nil {
		b.errorPage("Error changing existing parts: %s", err.Error())
	}
	err = bagzullaDb.UpdateProjectIdForPart(b.App.db, 0, part.PartId)
	if err != nil {
		b.errorPage("Error setting project id to 0 for part %d: %s",
			part.PartId, err.Error())
	}
	// Redirect user back to main project.
	redirectToProject(b, part.ProjectId)
}

// Get a count of open bugs by project

var openBugCounts = `
SELECT count(*), project_id
FROM bug
WHERE status=0
GROUP BY project_id
`

var openBugCountsStmt *sql.Stmt

func getOpenBugs(b *Bagreply, projects []bagzullaDb.Project) (openBugs []int64, err error) {
	var maxProjectId int64
	for _, project := range projects {
		if project.ProjectId > maxProjectId {
			maxProjectId = project.ProjectId
		}
	}
	if openBugCountsStmt == nil {
		openBugCountsStmt, err = b.App.db.Prepare(openBugCounts)
		if err != nil {
			return
		}
	}
	rows, err := openBugCountsStmt.Query()
	openBugs = make([]int64, maxProjectId+1)
	defer rows.Close()
	if err != nil {
		return
	}
	for rows.Next() {
		var count int64
		var id int64
		err = rows.Scan(&count, &id)
		if err != nil {
			return
		}
		if id > maxProjectId {
			err = errors.New(fmt.Sprintf("Id is too big %d >= %d", id, maxProjectId))
			return openBugs, err
		}
		openBugs[id] = count
	}
	return openBugs, nil
}

// Get a list of open bugs with a particular project ID

var openBugFromProjectId = "SELECT " + bagzullaDb.SelFi + ` FROM bug
WHERE bug.project_id = ?
AND bug.status=0
ORDER BY bug.changed DESC
`

var openBugFromProjectIdStmt *sql.Stmt

func openBugsFromProjectId(db *sql.DB, project_id int64) (bugs []bagzullaDb.Bug, err error) {
	if openBugFromProjectIdStmt == nil {
		openBugFromProjectIdStmt, err = db.Prepare(openBugFromProjectId)
		if err != nil {
			return
		}
	}
	rows, err := openBugFromProjectIdStmt.Query(project_id)
	if err != nil {
		return
	}
	defer rows.Close()
	return bagzullaDb.BugsFromRows(rows)
}

// Get a list of open bugs with a particular part-of-project ID

var openBugFromPartId = "SELECT " + bagzullaDb.SelFi + ` FROM bug
WHERE bug.part_id = ?
AND (bug.status=0)
ORDER BY bug.changed DESC
`

var openBugFromPartIdStmt *sql.Stmt

func openBugsFromPartId(db *sql.DB, part_id int64) (bugs []bagzullaDb.Bug, err error) {
	if openBugFromPartIdStmt == nil {
		openBugFromPartIdStmt, err = db.Prepare(openBugFromPartId)
		if err != nil {
			return
		}
	}
	rows, err := openBugFromPartIdStmt.Query(part_id)
	if err != nil {
		return
	}
	defer rows.Close()
	return bagzullaDb.BugsFromRows(rows)
}

// Get the user ID number corresponding to the name and password
// given.

var userIdSql = `
SELECT person_id FROM person WHERE name = ? AND password = ?
`

var userIdStmt *sql.Stmt

func userId(b *Bagreply) (int64, bool) {
	var p bagzullaDb.Person
	p.Name = b.r.FormValue("name")
	p.Password = b.r.FormValue("password")
	var err error
	if userIdStmt == nil {
		userIdStmt, err = b.App.db.Prepare(userIdSql)
		if err != nil {
			b.errorPage(fmt.Sprintf("Error preparing SQL %s: %s",
				userIdSql, err.Error()))
			return 0, false
		}
	}
	rows, err := userIdStmt.Query(p.Name, p.Password)
	defer rows.Close()
	if err != nil {
		b.errorPage("Error retrieving user id from database: %s",
			err.Error())
		return 0, false
	}
	found := false
	for rows.Next() {
		if found {
			b.errorPage("Duplicate user ids for this name/password combination")
			return 0, false
		}
		err = rows.Scan(&p.PersonId)
		if err != nil {
			b.errorPage("Error retrieving ID: %s", err.Error())
		}
		found = true
	}
	if found {
		return p.PersonId, true
	}
	b.errorPage("User with name <b>%s</b> and password <b>%s</b> was not found", p.Name, p.Password)
	return 0, false
}

// https://devtidbits.com/2020/08/03/go-sql-error-converting-null-to-string-is-unsupported/
type text struct {
	Content string
	TxtId   int64
	Type    sql.NullString
	OtherId sql.NullInt64
	BugId   int64
}

// Search the database for "string".

func searchText(b *Bagreply, searchTerm string) (txtIds []text, ok bool) {
	var textSearchSql = "SELECT " +
		"content, txt_id, txttype, other_id " +
		"FROM txt WHERE content LIKE '%" + searchTerm +
		"%' AND txttype IS NOT 'deleted'"
	var textSearchStmt *sql.Stmt
	if false {
		b.errorPage("%s\n", textSearchSql)
		return txtIds, false
	}
	textSearchStmt, ok = PrepareSql(b, textSearchSql)
	if !ok {
		return txtIds, false
	}
	rows, err := textSearchStmt.Query()
	if err != nil {
		b.errorPage("Error searching for '%s': %s", searchTerm, err.Error())
		return txtIds, false
	}
	defer rows.Close()
	for rows.Next() {
		var r text
		err = rows.Scan(&r.Content, &r.TxtId, &r.Type, &r.OtherId)
		if err != nil {
			b.errorPage("Error scanning: %s\n", err)
			return txtIds, false
		}
		switch r.Type.String {
		case "comment":
			var c bagzullaDb.Comment
			c, err = bagzullaDb.CommentFromId(b.App.db, r.OtherId.Int64)
			r.BugId = c.BugId
		case "title", "description":
			r.BugId = r.OtherId.Int64
		default:
			r.BugId = 0
		}
		txtIds = append(txtIds, r)
	}
	return txtIds, true
}

var bugsByChangeSql = `
SELECT * FROM bug ORDER BY bug.changed DESC limit ?
`

var bugsByChangeStmt *sql.Stmt

func BugsByChange(b *Bagreply, m int64) (bugs []bagzullaDb.Bug, ok bool) {
	if bugsByChangeStmt == nil {
		bugsByChangeStmt, ok = PrepareSql(b, bugsByChangeSql)
		if !ok {
			return bugs, false
		}
	}
	rows, err := bugsByChangeStmt.Query(m)
	defer rows.Close()
	if err != nil {
		b.errorPage("Error getting bugs by change time: %s", err.Error())
		return bugs, false
	}
	bugs, ok = scanRows(b, rows)
	return bugs, ok
}

/* Update the text associated with a particular comment by putting the
   new text's id number into the database. */

var updateCommentTextIdSql = `
UPDATE comment SET txt_id=? WHERE comment_id=?
`

var updateCommentTextIdStmt *sql.Stmt

func UpdateCommentTextId(b *Bagreply, c int64, t int64) (ok bool) {
	if updateCommentTextIdStmt == nil {
		updateCommentTextIdStmt, ok = PrepareSql(b, updateCommentTextIdSql)
		if !ok {
			return false
		}
	}
	_, err := updateCommentTextIdStmt.Exec(t, c)
	if err != nil {
		b.errorPage("Error changing text content of comment %d to %d: %s",
			c, t, err.Error())
		return false
	}
	return true
}

var effectToCausesSql = `
SELECT cause_id WHERE effect_id=?
`

var effectToCausesStmt *sql.Stmt

func EffectToCauses(b *Bagreply, effect int64) (causes []int64, ok bool) {
	if effectToCausesStmt == nil {
		effectToCausesStmt, ok = PrepareSql(b, effectToCausesSql)
		if !ok {
			return causes, false
		}
	}
	rows, err := effectToCausesStmt.Query(effect)
	defer rows.Close()
	for rows.Next() {
		var cause int64
		err = rows.Scan(&cause)
		if err != nil {
			return causes, false
		}
		causes = append(causes, cause)
	}
	return causes, true
}

var causeToEffectsSql = `
SELECT effect_id WHERE cause_id=?
`

var causeToEffectsStmt *sql.Stmt

func CauseToEffects(b *Bagreply, cause int64) (effects []int64, ok bool) {
	if causeToEffectsStmt == nil {
		causeToEffectsStmt, ok = PrepareSql(b, causeToEffectsSql)
		if !ok {
			return effects, false
		}
	}
	rows, err := causeToEffectsStmt.Query(cause)
	defer rows.Close()
	for rows.Next() {
		var effect int64
		err = rows.Scan(&effect)
		if err != nil {
			return effects, false
		}
		effects = append(effects, effect)
	}
	return effects, true
}

var originalFromDuplicateSql = `
SELECT original
FROM duplicate
WHERE duplicate.duplicate = ?
`
var originalFromDuplicateStmt *sql.Stmt

func OriginalsFromDuplicate(db *sql.DB, duplicate int64) (originals []RelatedBug, err error) {
	if originalFromDuplicateStmt == nil {
		originalFromDuplicateStmt, err = db.Prepare(originalFromDuplicateSql)
		if err != nil {
			return originals, err
		}
	}
	rows, err := originalFromDuplicateStmt.Query(duplicate)
	defer rows.Close()
	if err != nil {
		return originals, err
	}
	for rows.Next() {
		var original int64
		err = rows.Scan(&original)
		if err != nil {
			return originals, err
		}
		originals = append(originals, RelatedBug{Id: original})
	}
	return originals, nil
}

var deleteFileSql = `DELETE FROM image WHERE file = ?`

var deleteFileStmt *sql.Stmt

func removeImage(db *sql.DB, file string) (err error) {
	if deleteFileStmt == nil {
		deleteFileStmt, err = db.Prepare(deleteFileSql)
		if err != nil {
			return err
		}
	}
	_, err = deleteFileStmt.Exec(file)
	if err != nil {
		return err
	}
	return err
}

var deleteDuplicateSql = `DELETE FROM duplicate WHERE duplicate = ?`

var deleteDuplicateStmt *sql.Stmt

func removeDuplicate(db *sql.DB, duplicate int64) (err error) {
	if deleteDuplicateStmt == nil {
		deleteDuplicateStmt, err = db.Prepare(deleteDuplicateSql)
		if err != nil {
			return err
		}
	}
	_, err = deleteDuplicateStmt.Exec(duplicate)
	if err != nil {
		return err
	}
	return err
}

var deleteDuplicateOrigSql = `DELETE FROM duplicate WHERE original = ?`

var deleteDuplicateOrigStmt *sql.Stmt

func removeDuplicateOrig(db *sql.DB, orig int64) (err error) {
	if deleteDuplicateOrigStmt == nil {
		deleteDuplicateOrigStmt, err = db.Prepare(deleteDuplicateOrigSql)
		if err != nil {
			return err
		}
	}
	_, err = deleteDuplicateOrigStmt.Exec(orig)
	if err != nil {
		return err
	}
	return err
}

var deleteDependencyCauseSql = `DELETE FROM dependency WHERE cause = ? AND effect = ?`

var deleteDependencyCauseStmt *sql.Stmt

func removeDependencyCause(db *sql.DB, cause int64, effect int64) (err error) {
	if deleteDependencyCauseStmt == nil {
		deleteDependencyCauseStmt, err = db.Prepare(deleteDependencyCauseSql)
		if err != nil {
			return err
		}
	}
	_, err = deleteDependencyCauseStmt.Exec(cause, effect)
	if err != nil {
		return err
	}
	return err
}

// The sessions are not stored in the database but in a file called
// "logins.json". This should be updated when they are stored there.

var searchCookieSQL = "SELECT person_id FROM session WHERE cookie=?"
var searchCookieStmt *sql.Stmt

func SearchCookie(b *Bagapp, cookie string) (personId int64, found bool, err error) {
	if searchCookieStmt == nil {
		var err error
		searchCookieStmt, err = b.db.Prepare(searchCookieSQL)
		if err != nil {
			return 0, false, err
		}
	}
	rows, err := searchCookieStmt.Query(cookie)
	defer rows.Close()
	if err != nil {
		return 0, false, err
	}
	for rows.Next() {
		err := rows.Scan(&personId)
		if err != nil {
			return 0, false, err
		}
		if personId != 0 {
			return personId, true, nil
		}
	}
	return 0, false, nil
}

var txtDeleteSQL = `
UPDATE txt
SET txttype="deleted"
WHERE txt_id = ?
`

var txtDeleteStmt *sql.Stmt

func (b *Bagreply) DeleteText(id int64) (ok bool) {
	if txtDeleteStmt == nil {
		txtDeleteStmt, ok = PrepareSql(b, txtDeleteSQL)
		if !ok {
			return false
		}
	}
	_, err := txtDeleteStmt.Exec(id)
	if err != nil {
		b.errorPage("Error changing text to deleted status: %s", err.Error())
		return false
	}
	return true
}
