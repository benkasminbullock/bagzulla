package bagzullaDb

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Bug struct {
	BugId       int64
	Title       int64
	Description int64
	ProjectId   int64
	PartId      int64
	Entered     time.Time
	Owner       int64
	Status      int64
	Priority    int64
	Changed     time.Time
	Estimate    int64
}

var SelFi = "bug_id, title, description, project_id, part_id, entered, owner, status, priority, changed, estimate"

type Project struct {
	ProjectId   int64
	Name        string
	Directory   string
	Description int64
	Owner       int64
	Status      int64
}
type Part struct {
	PartId      int64
	Name        string
	Description int64
	ProjectId   int64
}
type Gitcommit struct {
	GitcommitId int64
	Githash     string
	ProjectId   int64
}
type Comment struct {
	CommentId int64
	TxtId     int64
	BugId     int64
	PersonId  int64
}
type Image struct {
	ImageId  int64
	File     string
	BugId    int64
	PersonId int64
}
type Person struct {
	PersonId int64
	Name     string
	Email    string
	Password string
}
type Dependency struct {
	DependencyId int64
	Cause        int64
	Effect       int64
}
type Duplicate struct {
	DuplicateId int64
	Original    int64
	Duplicate   int64
}
type Txt struct {
	TxtId   int64
	Entered time.Time
	Content string
}
type bdb struct {
	Sql  string
	Stmt *sql.Stmt
}

var bugFromId = bdb{
	Sql: `
SELECT bug_id, title, description, project_id, part_id, entered, owner, status, priority, changed, estimate
FROM bug
WHERE bug.bug_id = ?
`,
}

func BugFromId(db *sql.DB, bugId int64) (bug Bug, err error) {
	if bugFromId.Stmt == nil {
		bugFromId.Stmt, err = db.Prepare(bugFromId.Sql)
		if err != nil {
			return
		}
	}
	rows, err := bugFromId.Stmt.Query(bugId)
	if err != nil {
		return
	}
	defer rows.Close()
	found := false
	for rows.Next() {
		var dummy int64
		var estimate sql.NullInt64
		if found {
			return bug, errors.New(fmt.Sprintf("Duplicate bugs for id %d",
				bugId))
		}
		err = rows.Scan(&dummy, &bug.Title, &bug.Description, &bug.ProjectId, &bug.PartId, &bug.Entered, &bug.Owner, &bug.Status, &bug.Priority, &bug.Changed, &estimate)
		if err != nil {
			return bug, err
		}
		if estimate.Valid {
			bug.Estimate = estimate.Int64
		}
		found = true
	}
	if !found {
		err = errors.New(fmt.Sprintf("bug with id %d not found",
			bugId))
		return bug, err
	}
	bug.BugId = bugId
	return bug, nil
}

var insertBug = bdb{
	Sql: `
INSERT INTO bug(title, description, project_id, part_id, entered, owner, status, priority, changed, estimate) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`,
}

func InsertBug(db *sql.DB, bug Bug) (bug_id int64, err error) {
	if insertBug.Stmt == nil {
		insertBug.Stmt, err = db.Prepare(insertBug.Sql)
		if err != nil {
			return
		}
	}
	result, err := insertBug.Stmt.Exec(bug.Title, bug.Description, bug.ProjectId, bug.PartId, bug.Entered, bug.Owner, bug.Status, bug.Priority, bug.Changed, bug.Estimate)
	if err != nil {
		return
	}
	bug_id, err = result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return bug_id, nil
}

var allBugs = bdb{
	Sql: `
SELECT bug_id, title, description, project_id, part_id, entered, owner, status, priority, changed, estimate FROM bug
`,
}

func AllBugs(db *sql.DB) (bugs []Bug, err error) {
	if allBugs.Stmt == nil {
		allBugs.Stmt, err = db.Prepare(allBugs.Sql)
		if err != nil {
			return
		}
	}
	rows, err := allBugs.Stmt.Query()
	if err != nil {
		return bugs, err
	}
	return BugsFromRows(rows)
}

var bugFromTitle = bdb{
	Sql: `
SELECT bug_id, title, description, project_id, part_id, entered, owner, status, priority, changed, estimate
FROM bug
WHERE bug.title = ?
`,
}

func BugsFromTitle(db *sql.DB, title int64) (bugs []Bug, err error) {
	if bugFromTitle.Stmt == nil {
		bugFromTitle.Stmt, err = db.Prepare(bugFromTitle.Sql)
		if err != nil {
			return
		}
	}
	rows, err := bugFromTitle.Stmt.Query(title)
	if err != nil {
		return
	}
	defer rows.Close()
	return BugsFromRows(rows)
}

var bugUpdateTitle = bdb{
	Sql: `
UPDATE bug
SET title = ?
WHERE bug_id = ?
`,
}

func UpdateTitleForBug(db *sql.DB, title int64, bugId int64) (err error) {
	if bugUpdateTitle.Stmt == nil {
		bugUpdateTitle.Stmt, err = db.Prepare(bugUpdateTitle.Sql)
		if err != nil {
			return
		}
	}
	result, err := bugUpdateTitle.Stmt.Exec(title, bugId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to bug id %d", rows, title, bugId))
	}
	return nil
}

var bugFromDescription = bdb{
	Sql: `
SELECT bug_id, title, description, project_id, part_id, entered, owner, status, priority, changed, estimate
FROM bug
WHERE bug.description = ?
`,
}

func BugsFromDescription(db *sql.DB, description int64) (bugs []Bug, err error) {
	if bugFromDescription.Stmt == nil {
		bugFromDescription.Stmt, err = db.Prepare(bugFromDescription.Sql)
		if err != nil {
			return
		}
	}
	rows, err := bugFromDescription.Stmt.Query(description)
	if err != nil {
		return
	}
	defer rows.Close()
	return BugsFromRows(rows)
}

var bugUpdateDescription = bdb{
	Sql: `
UPDATE bug
SET description = ?
WHERE bug_id = ?
`,
}

func UpdateDescriptionForBug(db *sql.DB, description int64, bugId int64) (err error) {
	if bugUpdateDescription.Stmt == nil {
		bugUpdateDescription.Stmt, err = db.Prepare(bugUpdateDescription.Sql)
		if err != nil {
			return
		}
	}
	result, err := bugUpdateDescription.Stmt.Exec(description, bugId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to bug id %d", rows, description, bugId))
	}
	return nil
}

var bugFromProjectId = bdb{
	Sql: `SELECT ` + SelFi + `
FROM bug
WHERE bug.project_id = ?
`,
}

func BugsFromProjectId(db *sql.DB, project_id int64) (bugs []Bug, err error) {
	if bugFromProjectId.Stmt == nil {
		bugFromProjectId.Stmt, err = db.Prepare(bugFromProjectId.Sql)
		if err != nil {
			return
		}
	}
	rows, err := bugFromProjectId.Stmt.Query(project_id)
	if err != nil {
		return
	}
	defer rows.Close()
	return BugsFromRows(rows)
}

var bugUpdateProjectId = bdb{
	Sql: `
UPDATE bug
SET project_id = ?
WHERE bug_id = ?
`,
}

func UpdateProjectIdForBug(db *sql.DB, project_id int64, bugId int64) (err error) {
	if bugUpdateProjectId.Stmt == nil {
		bugUpdateProjectId.Stmt, err = db.Prepare(bugUpdateProjectId.Sql)
		if err != nil {
			return
		}
	}
	result, err := bugUpdateProjectId.Stmt.Exec(project_id, bugId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to bug id %d", rows, project_id, bugId))
	}
	return nil
}

var bugFromPartId = bdb{
	Sql: `
SELECT bug_id, title, description, project_id, part_id, entered, owner, status, priority, changed, estimate
FROM bug
WHERE bug.part_id = ?
`,
}

func BugsFromPartId(db *sql.DB, part_id int64) (bugs []Bug, err error) {
	if bugFromPartId.Stmt == nil {
		bugFromPartId.Stmt, err = db.Prepare(bugFromPartId.Sql)
		if err != nil {
			return
		}
	}
	rows, err := bugFromPartId.Stmt.Query(part_id)
	if err != nil {
		return
	}
	defer rows.Close()
	return BugsFromRows(rows)
}

var bugUpdatePartId = bdb{
	Sql: `
UPDATE bug
SET part_id = ?
WHERE bug_id = ?
`,
}

func UpdatePartIdForBug(db *sql.DB, part_id int64, bugId int64) (err error) {
	if bugUpdatePartId.Stmt == nil {
		bugUpdatePartId.Stmt, err = db.Prepare(bugUpdatePartId.Sql)
		if err != nil {
			return
		}
	}
	result, err := bugUpdatePartId.Stmt.Exec(part_id, bugId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to bug id %d", rows, part_id, bugId))
	}
	return nil
}

var bugFromEntered = bdb{
	Sql: `
SELECT bug_id, title, description, project_id, part_id, entered, owner, status, priority, changed, estimate
FROM bug
WHERE bug.entered = ?
`,
}

func BugsFromEntered(db *sql.DB, entered time.Time) (bugs []Bug, err error) {
	if bugFromEntered.Stmt == nil {
		bugFromEntered.Stmt, err = db.Prepare(bugFromEntered.Sql)
		if err != nil {
			return
		}
	}
	rows, err := bugFromEntered.Stmt.Query(entered)
	defer rows.Close()
	if err != nil {
		return
	}
	return BugsFromRows(rows)
}

var bugUpdateEntered = bdb{
	Sql: `
UPDATE bug
SET entered = ?
WHERE bug_id = ?
`,
}

func UpdateEnteredForBug(db *sql.DB, entered time.Time, bugId int64) (err error) {
	if bugUpdateEntered.Stmt == nil {
		bugUpdateEntered.Stmt, err = db.Prepare(bugUpdateEntered.Sql)
		if err != nil {
			return
		}
	}
	result, err := bugUpdateEntered.Stmt.Exec(entered, bugId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to bug id %d", rows, entered, bugId))
	}
	return nil
}

var bugFromOwner = bdb{
	Sql: `
SELECT bug_id, title, description, project_id, part_id, entered, owner, status, priority, changed, estimate
FROM bug
WHERE bug.owner = ?
`,
}

func BugsFromRows(rows *sql.Rows) (bugs []Bug, err error) {
	for rows.Next() {
		var bug Bug
		var estimate sql.NullInt64
		err = rows.Scan(&bug.BugId, &bug.Title, &bug.Description, &bug.ProjectId, &bug.PartId, &bug.Entered, &bug.Owner, &bug.Status, &bug.Priority, &bug.Changed, &estimate)
		if err != nil {
			return bugs, err
		}
		if estimate.Valid {
			bug.Estimate = estimate.Int64
		}
		bugs = append(bugs, bug)
	}
	return bugs, nil
}

func BugsFromOwner(db *sql.DB, owner int64) (bugs []Bug, err error) {
	if bugFromOwner.Stmt == nil {
		bugFromOwner.Stmt, err = db.Prepare(bugFromOwner.Sql)
		if err != nil {
			return bugs, err
		}
	}
	rows, err := bugFromOwner.Stmt.Query(owner)
	if err != nil {
		return bugs, err
	}
	defer rows.Close()
	return BugsFromRows(rows)
}

var bugUpdateOwner = bdb{
	Sql: `
UPDATE bug
SET owner = ?
WHERE bug_id = ?
`,
}

func UpdateOwnerForBug(db *sql.DB, owner int64, bugId int64) (err error) {
	if bugUpdateOwner.Stmt == nil {
		bugUpdateOwner.Stmt, err = db.Prepare(bugUpdateOwner.Sql)
		if err != nil {
			return
		}
	}
	result, err := bugUpdateOwner.Stmt.Exec(owner, bugId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to bug id %d", rows, owner, bugId))
	}
	return nil
}

var bugFromStatus = bdb{
	Sql: `
SELECT bug_id, title, description, project_id, part_id, entered, owner, priority, changed, estimate
FROM bug
WHERE bug.status = ?
`,
}

func BugsFromStatus(db *sql.DB, status int64) (bugs []Bug, err error) {
	if bugFromStatus.Stmt == nil {
		bugFromStatus.Stmt, err = db.Prepare(bugFromStatus.Sql)
		if err != nil {
			return
		}
	}
	rows, err := bugFromStatus.Stmt.Query(status)
	defer rows.Close()
	if err != nil {
		return
	}
	return BugsFromRows(rows)
}

var bugUpdateStatus = bdb{
	Sql: `
UPDATE bug
SET status = ?
WHERE bug_id = ?
`,
}

func UpdateStatusForBug(db *sql.DB, status int64, bugId int64) (err error) {
	if bugUpdateStatus.Stmt == nil {
		bugUpdateStatus.Stmt, err = db.Prepare(bugUpdateStatus.Sql)
		if err != nil {
			return
		}
	}
	result, err := bugUpdateStatus.Stmt.Exec(status, bugId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to bug id %d", rows, status, bugId))
	}
	return nil
}

var bugFromPriority = bdb{
	Sql: `
SELECT bug_id, title, description, project_id, part_id, entered, owner, status, changed, estimate
FROM bug
WHERE bug.priority = ?
`,
}

func BugsFromPriority(db *sql.DB, priority int64) (bugs []Bug, err error) {
	if bugFromPriority.Stmt == nil {
		bugFromPriority.Stmt, err = db.Prepare(bugFromPriority.Sql)
		if err != nil {
			return
		}
	}
	rows, err := bugFromPriority.Stmt.Query(priority)
	defer rows.Close()
	if err != nil {
		return
	}
	return BugsFromRows(rows)
}

var bugUpdatePriority = bdb{
	Sql: `
UPDATE bug
SET priority = ?
WHERE bug_id = ?
`,
}

func UpdatePriorityForBug(db *sql.DB, priority int64, bugId int64) (err error) {
	if bugUpdatePriority.Stmt == nil {
		bugUpdatePriority.Stmt, err = db.Prepare(bugUpdatePriority.Sql)
		if err != nil {
			return
		}
	}
	result, err := bugUpdatePriority.Stmt.Exec(priority, bugId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to bug id %d", rows, priority, bugId))
	}
	return nil
}

var bugFromChanged = bdb{
	Sql: `
SELECT bug_id, title, description, project_id, part_id, entered, owner, status, priority
FROM bug
WHERE bug.changed = ?
`,
}

func BugsFromChanged(db *sql.DB, changed time.Time) (bugs []Bug, err error) {
	if bugFromChanged.Stmt == nil {
		bugFromChanged.Stmt, err = db.Prepare(bugFromChanged.Sql)
		if err != nil {
			return
		}
	}
	rows, err := bugFromChanged.Stmt.Query(changed)
	defer rows.Close()
	if err != nil {
		return
	}
	return BugsFromRows(rows)
}

var bugUpdateChanged = bdb{
	Sql: `
UPDATE bug
SET changed = ?
WHERE bug_id = ?
`,
}

func UpdateChangedForBug(db *sql.DB, changed time.Time, bugId int64) (err error) {
	if bugUpdateChanged.Stmt == nil {
		bugUpdateChanged.Stmt, err = db.Prepare(bugUpdateChanged.Sql)
		if err != nil {
			return
		}
	}
	result, err := bugUpdateChanged.Stmt.Exec(changed, bugId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to bug id %d", rows, changed, bugId))
	}
	return nil
}

var projectFromId = bdb{
	Sql: `
SELECT name, directory, description, owner, status
FROM project
WHERE project.project_id = ?
`,
}

func ProjectFromId(db *sql.DB, projectId int64) (project Project, err error) {
	if projectFromId.Stmt == nil {
		projectFromId.Stmt, err = db.Prepare(projectFromId.Sql)
		if err != nil {
			return
		}
	}
	rows, err := projectFromId.Stmt.Query(projectId)
	defer rows.Close()
	if err != nil {
		return
	}
	found := false
	for rows.Next() {
		if found {
			return project, errors.New(fmt.Sprintf("Duplicate projects for id %d",
				projectId))
		}
		err = rows.Scan(&project.Name, &project.Directory, &project.Description, &project.Owner, &project.Status)
		if err != nil {
			return
		}
		found = true
	}
	if !found {
		err = errors.New(fmt.Sprintf("project with id %d not found",
			projectId))
		return
	}
	project.ProjectId = projectId
	return project, nil
}

var insertProject = bdb{
	Sql: `
INSERT INTO project(name, directory, description, owner, status) VALUES (?, ?, ?, ?, ?)
`,
}

func InsertProject(db *sql.DB, project Project) (project_id int64, err error) {
	if insertProject.Stmt == nil {
		insertProject.Stmt, err = db.Prepare(insertProject.Sql)
		if err != nil {
			return
		}
	}
	result, err := insertProject.Stmt.Exec(project.Name, project.Directory, project.Description, project.Owner, project.Status)
	if err != nil {
		return
	}
	project_id, err = result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return project_id, nil
}

var allProjects = bdb{
	Sql: `
SELECT project_id, name, directory, description, owner, status FROM project
`,
}

func AllProjects(db *sql.DB) (projects []Project, err error) {
	if allProjects.Stmt == nil {
		allProjects.Stmt, err = db.Prepare(allProjects.Sql)
		if err != nil {
			return
		}
	}
	rows, err := allProjects.Stmt.Query()
	for rows.Next() {
		var project Project
		err = rows.Scan(&project.ProjectId, &project.Name, &project.Directory, &project.Description, &project.Owner, &project.Status)
		if err != nil {
			return
		}
		projects = append(projects, project)
	}
	return projects, nil
}

var projectFromName = bdb{
	Sql: `
SELECT project_id, directory, description, owner, status
FROM project
WHERE project.name = ?
`,
}

func ProjectFromName(db *sql.DB, name string) (project Project, err error) {
	if projectFromName.Stmt == nil {
		projectFromName.Stmt, err = db.Prepare(projectFromName.Sql)
		if err != nil {
			return
		}
	}
	rows, err := projectFromName.Stmt.Query(name)
	defer rows.Close()
	if err != nil {
		return
	}
	found := false
	for rows.Next() {
		if found {
			return project, errors.New(fmt.Sprintf("Duplicate projects for name %v",
				name))
		}
		err = rows.Scan(&project.ProjectId, &project.Directory, &project.Description, &project.Owner, &project.Status)
		if err != nil {
			return
		}
	}
	return project, nil
}

var projectUpdateName = bdb{
	Sql: `
UPDATE project
SET name = ?
WHERE project_id = ?
`,
}

func UpdateNameForProject(db *sql.DB, name string, projectId int64) (err error) {
	if projectUpdateName.Stmt == nil {
		projectUpdateName.Stmt, err = db.Prepare(projectUpdateName.Sql)
		if err != nil {
			return
		}
	}
	result, err := projectUpdateName.Stmt.Exec(name, projectId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to project id %d", rows, name, projectId))
	}
	return nil
}

var projectFromDirectory = bdb{
	Sql: `
SELECT project_id, name, description, owner, status
FROM project
WHERE project.directory = ?
`,
}

func ProjectsFromDirectory(db *sql.DB, directory string) (projects []Project, err error) {
	if projectFromDirectory.Stmt == nil {
		projectFromDirectory.Stmt, err = db.Prepare(projectFromDirectory.Sql)
		if err != nil {
			return
		}
	}
	rows, err := projectFromDirectory.Stmt.Query(directory)
	defer rows.Close()
	if err != nil {
		return
	}
	for rows.Next() {
		var project Project
		err = rows.Scan(&project.ProjectId, &project.Name, &project.Description, &project.Owner, &project.Status)
		if err != nil {
			return
		}
		projects = append(projects, project)
	}
	return projects, nil
}

var projectUpdateDirectory = bdb{
	Sql: `
UPDATE project
SET directory = ?
WHERE project_id = ?
`,
}

func UpdateDirectoryForProject(db *sql.DB, directory string, projectId int64) (err error) {
	if projectUpdateDirectory.Stmt == nil {
		projectUpdateDirectory.Stmt, err = db.Prepare(projectUpdateDirectory.Sql)
		if err != nil {
			return
		}
	}
	result, err := projectUpdateDirectory.Stmt.Exec(directory, projectId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to project id %d", rows, directory, projectId))
	}
	return nil
}

var projectFromDescription = bdb{
	Sql: `
SELECT project_id, name, directory, owner, status
FROM project
WHERE project.description = ?
`,
}

func ProjectsFromDescription(db *sql.DB, description int64) (projects []Project, err error) {
	if projectFromDescription.Stmt == nil {
		projectFromDescription.Stmt, err = db.Prepare(projectFromDescription.Sql)
		if err != nil {
			return
		}
	}
	rows, err := projectFromDescription.Stmt.Query(description)
	defer rows.Close()
	if err != nil {
		return
	}
	for rows.Next() {
		var project Project
		err = rows.Scan(&project.ProjectId, &project.Name, &project.Directory, &project.Owner, &project.Status)
		if err != nil {
			return
		}
		projects = append(projects, project)
	}
	return projects, nil
}

var projectUpdateDescription = bdb{
	Sql: `
UPDATE project
SET description = ?
WHERE project_id = ?
`,
}

func UpdateDescriptionForProject(db *sql.DB, description int64, projectId int64) (err error) {
	if projectUpdateDescription.Stmt == nil {
		projectUpdateDescription.Stmt, err = db.Prepare(projectUpdateDescription.Sql)
		if err != nil {
			return
		}
	}
	result, err := projectUpdateDescription.Stmt.Exec(description, projectId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to project id %d", rows, description, projectId))
	}
	return nil
}

var projectFromOwner = bdb{
	Sql: `
SELECT project_id, name, directory, description, status
FROM project
WHERE project.owner = ?
`,
}

func ProjectsFromOwner(db *sql.DB, owner int64) (projects []Project, err error) {
	if projectFromOwner.Stmt == nil {
		projectFromOwner.Stmt, err = db.Prepare(projectFromOwner.Sql)
		if err != nil {
			return
		}
	}
	rows, err := projectFromOwner.Stmt.Query(owner)
	defer rows.Close()
	if err != nil {
		return
	}
	for rows.Next() {
		var project Project
		err = rows.Scan(&project.ProjectId, &project.Name, &project.Directory, &project.Description, &project.Status)
		if err != nil {
			return
		}
		projects = append(projects, project)
	}
	return projects, nil
}

var projectUpdateOwner = bdb{
	Sql: `
UPDATE project
SET owner = ?
WHERE project_id = ?
`,
}

func UpdateOwnerForProject(db *sql.DB, owner int64, projectId int64) (err error) {
	if projectUpdateOwner.Stmt == nil {
		projectUpdateOwner.Stmt, err = db.Prepare(projectUpdateOwner.Sql)
		if err != nil {
			return
		}
	}
	result, err := projectUpdateOwner.Stmt.Exec(owner, projectId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to project id %d", rows, owner, projectId))
	}
	return nil
}

var projectFromStatus = bdb{
	Sql: `
SELECT project_id, name, directory, description, owner
FROM project
WHERE project.status = ?
`,
}

func ProjectsFromStatus(db *sql.DB, status int64) (projects []Project, err error) {
	if projectFromStatus.Stmt == nil {
		projectFromStatus.Stmt, err = db.Prepare(projectFromStatus.Sql)
		if err != nil {
			return
		}
	}
	rows, err := projectFromStatus.Stmt.Query(status)
	defer rows.Close()
	if err != nil {
		return
	}
	for rows.Next() {
		var project Project
		err = rows.Scan(&project.ProjectId, &project.Name, &project.Directory, &project.Description, &project.Owner)
		if err != nil {
			return
		}
		projects = append(projects, project)
	}
	return projects, nil
}

var projectUpdateStatus = bdb{
	Sql: `
UPDATE project
SET status = ?
WHERE project_id = ?
`,
}

func UpdateStatusForProject(db *sql.DB, status int64, projectId int64) (err error) {
	if projectUpdateStatus.Stmt == nil {
		projectUpdateStatus.Stmt, err = db.Prepare(projectUpdateStatus.Sql)
		if err != nil {
			return
		}
	}
	result, err := projectUpdateStatus.Stmt.Exec(status, projectId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to project id %d", rows, status, projectId))
	}
	return nil
}

var partFromId = bdb{
	Sql: `
SELECT name, description, project_id
FROM part
WHERE part.part_id = ?
`,
}

func PartFromId(db *sql.DB, partId int64) (part Part, err error) {
	if partFromId.Stmt == nil {
		partFromId.Stmt, err = db.Prepare(partFromId.Sql)
		if err != nil {
			return
		}
	}
	rows, err := partFromId.Stmt.Query(partId)
	defer rows.Close()
	if err != nil {
		return
	}
	found := false
	for rows.Next() {
		if found {
			return part, errors.New(fmt.Sprintf("Duplicate parts for id %d",
				partId))
		}
		err = rows.Scan(&part.Name, &part.Description, &part.ProjectId)
		if err != nil {
			return
		}
		found = true
	}
	if !found {
		err = errors.New(fmt.Sprintf("part with id %d not found",
			partId))
		return
	}
	part.PartId = partId
	return part, nil
}

var insertPart = bdb{
	Sql: `
INSERT INTO part(name, description, project_id) VALUES (?, ?, ?)
`,
}

func InsertPart(db *sql.DB, part Part) (part_id int64, err error) {
	if insertPart.Stmt == nil {
		insertPart.Stmt, err = db.Prepare(insertPart.Sql)
		if err != nil {
			return
		}
	}
	result, err := insertPart.Stmt.Exec(part.Name, part.Description, part.ProjectId)
	if err != nil {
		return
	}
	part_id, err = result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return part_id, nil
}

var allParts = bdb{
	Sql: `
SELECT part_id, name, description, project_id FROM part
`,
}

func AllParts(db *sql.DB) (parts []Part, err error) {
	if allParts.Stmt == nil {
		allParts.Stmt, err = db.Prepare(allParts.Sql)
		if err != nil {
			return
		}
	}
	rows, err := allParts.Stmt.Query()
	for rows.Next() {
		var part Part
		err = rows.Scan(&part.PartId, &part.Name, &part.Description, &part.ProjectId)
		if err != nil {
			return
		}
		parts = append(parts, part)
	}
	return parts, nil
}

var partFromName = bdb{
	Sql: `
SELECT part_id, description, project_id
FROM part
WHERE part.name = ?
`,
}

func PartsFromName(db *sql.DB, name string) (parts []Part, err error) {
	if partFromName.Stmt == nil {
		partFromName.Stmt, err = db.Prepare(partFromName.Sql)
		if err != nil {
			return
		}
	}
	rows, err := partFromName.Stmt.Query(name)
	defer rows.Close()
	if err != nil {
		return
	}
	for rows.Next() {
		var part Part
		err = rows.Scan(&part.PartId, &part.Description, &part.ProjectId)
		if err != nil {
			return
		}
		parts = append(parts, part)
	}
	return parts, nil
}

var partUpdateName = bdb{
	Sql: `
UPDATE part
SET name = ?
WHERE part_id = ?
`,
}

func UpdateNameForPart(db *sql.DB, name string, partId int64) (err error) {
	if partUpdateName.Stmt == nil {
		partUpdateName.Stmt, err = db.Prepare(partUpdateName.Sql)
		if err != nil {
			return
		}
	}
	result, err := partUpdateName.Stmt.Exec(name, partId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to part id %d", rows, name, partId))
	}
	return nil
}

var partFromDescription = bdb{
	Sql: `
SELECT part_id, name, project_id
FROM part
WHERE part.description = ?
`,
}

func PartsFromDescription(db *sql.DB, description int64) (parts []Part, err error) {
	if partFromDescription.Stmt == nil {
		partFromDescription.Stmt, err = db.Prepare(partFromDescription.Sql)
		if err != nil {
			return
		}
	}
	rows, err := partFromDescription.Stmt.Query(description)
	defer rows.Close()
	if err != nil {
		return
	}
	for rows.Next() {
		var part Part
		err = rows.Scan(&part.PartId, &part.Name, &part.ProjectId)
		if err != nil {
			return
		}
		parts = append(parts, part)
	}
	return parts, nil
}

var partUpdateDescription = bdb{
	Sql: `
UPDATE part
SET description = ?
WHERE part_id = ?
`,
}

func UpdateDescriptionForPart(db *sql.DB, description int64, partId int64) (err error) {
	if partUpdateDescription.Stmt == nil {
		partUpdateDescription.Stmt, err = db.Prepare(partUpdateDescription.Sql)
		if err != nil {
			return
		}
	}
	result, err := partUpdateDescription.Stmt.Exec(description, partId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to part id %d", rows, description, partId))
	}
	return nil
}

var partFromProjectId = bdb{
	Sql: `
SELECT part_id, name, description
FROM part
WHERE part.project_id = ?
`,
}

func PartsFromProjectId(db *sql.DB, project_id int64) (parts []Part, err error) {
	if partFromProjectId.Stmt == nil {
		partFromProjectId.Stmt, err = db.Prepare(partFromProjectId.Sql)
		if err != nil {
			return
		}
	}
	rows, err := partFromProjectId.Stmt.Query(project_id)
	defer rows.Close()
	if err != nil {
		return
	}
	for rows.Next() {
		var part Part
		err = rows.Scan(&part.PartId, &part.Name, &part.Description)
		if err != nil {
			return
		}
		parts = append(parts, part)
	}
	return parts, nil
}

var partUpdateProjectId = bdb{
	Sql: `
UPDATE part
SET project_id = ?
WHERE part_id = ?
`,
}

func UpdateProjectIdForPart(db *sql.DB, project_id int64, partId int64) (err error) {
	if partUpdateProjectId.Stmt == nil {
		partUpdateProjectId.Stmt, err = db.Prepare(partUpdateProjectId.Sql)
		if err != nil {
			return
		}
	}
	result, err := partUpdateProjectId.Stmt.Exec(project_id, partId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to part id %d", rows, project_id, partId))
	}
	return nil
}

var gitcommitFromId = bdb{
	Sql: `
SELECT githash, project_id
FROM gitcommit
WHERE gitcommit.gitcommit_id = ?
`,
}

func GitcommitFromId(db *sql.DB, gitcommitId int64) (gitcommit Gitcommit, err error) {
	if gitcommitFromId.Stmt == nil {
		gitcommitFromId.Stmt, err = db.Prepare(gitcommitFromId.Sql)
		if err != nil {
			return
		}
	}
	rows, err := gitcommitFromId.Stmt.Query(gitcommitId)
	defer rows.Close()
	if err != nil {
		return
	}
	found := false
	for rows.Next() {
		if found {
			return gitcommit, errors.New(fmt.Sprintf("Duplicate gitcommits for id %d",
				gitcommitId))
		}
		err = rows.Scan(&gitcommit.Githash, &gitcommit.ProjectId)
		if err != nil {
			return
		}
		found = true
	}
	if !found {
		err = errors.New(fmt.Sprintf("gitcommit with id %d not found",
			gitcommitId))
		return
	}
	gitcommit.GitcommitId = gitcommitId
	return gitcommit, nil
}

var insertGitcommit = bdb{
	Sql: `
INSERT INTO gitcommit(githash, project_id) VALUES (?, ?)
`,
}

func InsertGitcommit(db *sql.DB, gitcommit Gitcommit) (gitcommit_id int64, err error) {
	if insertGitcommit.Stmt == nil {
		insertGitcommit.Stmt, err = db.Prepare(insertGitcommit.Sql)
		if err != nil {
			return
		}
	}
	result, err := insertGitcommit.Stmt.Exec(gitcommit.Githash, gitcommit.ProjectId)
	if err != nil {
		return
	}
	gitcommit_id, err = result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return gitcommit_id, nil
}

var allGitcommits = bdb{
	Sql: `
SELECT gitcommit_id, githash, project_id FROM gitcommit
`,
}

func AllGitcommits(db *sql.DB) (gitcommits []Gitcommit, err error) {
	if allGitcommits.Stmt == nil {
		allGitcommits.Stmt, err = db.Prepare(allGitcommits.Sql)
		if err != nil {
			return
		}
	}
	rows, err := allGitcommits.Stmt.Query()
	for rows.Next() {
		var gitcommit Gitcommit
		err = rows.Scan(&gitcommit.GitcommitId, &gitcommit.Githash, &gitcommit.ProjectId)
		if err != nil {
			return
		}
		gitcommits = append(gitcommits, gitcommit)
	}
	return gitcommits, nil
}

var gitcommitFromGithash = bdb{
	Sql: `
SELECT gitcommit_id, project_id
FROM gitcommit
WHERE gitcommit.githash = ?
`,
}

func GitcommitsFromGithash(db *sql.DB, githash string) (gitcommits []Gitcommit, err error) {
	if gitcommitFromGithash.Stmt == nil {
		gitcommitFromGithash.Stmt, err = db.Prepare(gitcommitFromGithash.Sql)
		if err != nil {
			return
		}
	}
	rows, err := gitcommitFromGithash.Stmt.Query(githash)
	defer rows.Close()
	if err != nil {
		return
	}
	for rows.Next() {
		var gitcommit Gitcommit
		err = rows.Scan(&gitcommit.GitcommitId, &gitcommit.ProjectId)
		if err != nil {
			return
		}
		gitcommits = append(gitcommits, gitcommit)
	}
	return gitcommits, nil
}

var gitcommitUpdateGithash = bdb{
	Sql: `
UPDATE gitcommit
SET githash = ?
WHERE gitcommit_id = ?
`,
}

func UpdateGithashForGitcommit(db *sql.DB, githash string, gitcommitId int64) (err error) {
	if gitcommitUpdateGithash.Stmt == nil {
		gitcommitUpdateGithash.Stmt, err = db.Prepare(gitcommitUpdateGithash.Sql)
		if err != nil {
			return
		}
	}
	result, err := gitcommitUpdateGithash.Stmt.Exec(githash, gitcommitId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to gitcommit id %d", rows, githash, gitcommitId))
	}
	return nil
}

var gitcommitFromProjectId = bdb{
	Sql: `
SELECT gitcommit_id, githash
FROM gitcommit
WHERE gitcommit.project_id = ?
`,
}

func GitcommitsFromProjectId(db *sql.DB, project_id int64) (gitcommits []Gitcommit, err error) {
	if gitcommitFromProjectId.Stmt == nil {
		gitcommitFromProjectId.Stmt, err = db.Prepare(gitcommitFromProjectId.Sql)
		if err != nil {
			return
		}
	}
	rows, err := gitcommitFromProjectId.Stmt.Query(project_id)
	defer rows.Close()
	if err != nil {
		return
	}
	for rows.Next() {
		var gitcommit Gitcommit
		err = rows.Scan(&gitcommit.GitcommitId, &gitcommit.Githash)
		if err != nil {
			return
		}
		gitcommits = append(gitcommits, gitcommit)
	}
	return gitcommits, nil
}

var gitcommitUpdateProjectId = bdb{
	Sql: `
UPDATE gitcommit
SET project_id = ?
WHERE gitcommit_id = ?
`,
}

func UpdateProjectIdForGitcommit(db *sql.DB, project_id int64, gitcommitId int64) (err error) {
	if gitcommitUpdateProjectId.Stmt == nil {
		gitcommitUpdateProjectId.Stmt, err = db.Prepare(gitcommitUpdateProjectId.Sql)
		if err != nil {
			return
		}
	}
	result, err := gitcommitUpdateProjectId.Stmt.Exec(project_id, gitcommitId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to gitcommit id %d", rows, project_id, gitcommitId))
	}
	return nil
}

var commentFromId = bdb{
	Sql: `
SELECT txt_id, bug_id, person_id
FROM comment
WHERE comment.comment_id = ?
`,
}

func CommentFromId(db *sql.DB, commentId int64) (comment Comment, err error) {
	if commentFromId.Stmt == nil {
		commentFromId.Stmt, err = db.Prepare(commentFromId.Sql)
		if err != nil {
			return
		}
	}
	rows, err := commentFromId.Stmt.Query(commentId)
	defer rows.Close()
	if err != nil {
		return
	}
	found := false
	for rows.Next() {
		if found {
			return comment, errors.New(fmt.Sprintf("Duplicate comments for id %d",
				commentId))
		}
		err = rows.Scan(&comment.TxtId, &comment.BugId, &comment.PersonId)
		if err != nil {
			return
		}
		found = true
	}
	if !found {
		err = errors.New(fmt.Sprintf("comment with id %d not found",
			commentId))
		return
	}
	comment.CommentId = commentId
	return comment, nil
}

var insertComment = bdb{
	Sql: `
INSERT INTO comment(txt_id, bug_id, person_id) VALUES (?, ?, ?)
`,
}

func InsertComment(db *sql.DB, comment Comment) (comment_id int64, err error) {
	if insertComment.Stmt == nil {
		insertComment.Stmt, err = db.Prepare(insertComment.Sql)
		if err != nil {
			return
		}
	}
	result, err := insertComment.Stmt.Exec(comment.TxtId, comment.BugId, comment.PersonId)
	if err != nil {
		return
	}
	comment_id, err = result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return comment_id, nil
}

var allComments = bdb{
	Sql: `
SELECT comment_id, txt_id, bug_id, person_id FROM comment
`,
}

func AllComments(db *sql.DB) (comments []Comment, err error) {
	if allComments.Stmt == nil {
		allComments.Stmt, err = db.Prepare(allComments.Sql)
		if err != nil {
			return
		}
	}
	rows, err := allComments.Stmt.Query()
	for rows.Next() {
		var comment Comment
		err = rows.Scan(&comment.CommentId, &comment.TxtId, &comment.BugId, &comment.PersonId)
		if err != nil {
			return
		}
		comments = append(comments, comment)
	}
	return comments, nil
}

var commentFromTxtId = bdb{
	Sql: `
SELECT comment_id, bug_id, person_id
FROM comment
WHERE comment.txt_id = ?
`,
}

func CommentsFromTxtId(db *sql.DB, txt_id int64) (comments []Comment, err error) {
	if commentFromTxtId.Stmt == nil {
		commentFromTxtId.Stmt, err = db.Prepare(commentFromTxtId.Sql)
		if err != nil {
			return
		}
	}
	rows, err := commentFromTxtId.Stmt.Query(txt_id)
	defer rows.Close()
	if err != nil {
		return
	}
	for rows.Next() {
		var comment Comment
		err = rows.Scan(&comment.CommentId, &comment.BugId, &comment.PersonId)
		if err != nil {
			return
		}
		comments = append(comments, comment)
	}
	return comments, nil
}

var commentUpdateTxtId = bdb{
	Sql: `
UPDATE comment
SET txt_id = ?
WHERE comment_id = ?
`,
}

func UpdateTxtIdForComment(db *sql.DB, txt_id int64, commentId int64) (err error) {
	if commentUpdateTxtId.Stmt == nil {
		commentUpdateTxtId.Stmt, err = db.Prepare(commentUpdateTxtId.Sql)
		if err != nil {
			return
		}
	}
	result, err := commentUpdateTxtId.Stmt.Exec(txt_id, commentId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to comment id %d", rows, txt_id, commentId))
	}
	return nil
}

var commentFromBugId = bdb{
	Sql: `
SELECT comment_id, txt_id, person_id
FROM comment
WHERE comment.bug_id = ?
`,
}

func CommentsFromBugId(db *sql.DB, bug_id int64) (comments []Comment, err error) {
	if commentFromBugId.Stmt == nil {
		commentFromBugId.Stmt, err = db.Prepare(commentFromBugId.Sql)
		if err != nil {
			return
		}
	}
	rows, err := commentFromBugId.Stmt.Query(bug_id)
	defer rows.Close()
	if err != nil {
		return
	}
	for rows.Next() {
		var comment Comment
		err = rows.Scan(&comment.CommentId, &comment.TxtId, &comment.PersonId)
		if err != nil {
			return
		}
		comments = append(comments, comment)
	}
	return comments, nil
}

var commentUpdateBugId = bdb{
	Sql: `
UPDATE comment
SET bug_id = ?
WHERE comment_id = ?
`,
}

func UpdateBugIdForComment(db *sql.DB, bug_id int64, commentId int64) (err error) {
	if commentUpdateBugId.Stmt == nil {
		commentUpdateBugId.Stmt, err = db.Prepare(commentUpdateBugId.Sql)
		if err != nil {
			return
		}
	}
	result, err := commentUpdateBugId.Stmt.Exec(bug_id, commentId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to comment id %d", rows, bug_id, commentId))
	}
	return nil
}

var commentFromPersonId = bdb{
	Sql: `
SELECT comment_id, txt_id, bug_id
FROM comment
WHERE comment.person_id = ?
`,
}

func CommentsFromPersonId(db *sql.DB, person_id int64) (comments []Comment, err error) {
	if commentFromPersonId.Stmt == nil {
		commentFromPersonId.Stmt, err = db.Prepare(commentFromPersonId.Sql)
		if err != nil {
			return
		}
	}
	rows, err := commentFromPersonId.Stmt.Query(person_id)
	defer rows.Close()
	if err != nil {
		return
	}
	for rows.Next() {
		var comment Comment
		err = rows.Scan(&comment.CommentId, &comment.TxtId, &comment.BugId)
		if err != nil {
			return
		}
		comments = append(comments, comment)
	}
	return comments, nil
}

var commentUpdatePersonId = bdb{
	Sql: `
UPDATE comment
SET person_id = ?
WHERE comment_id = ?
`,
}

func UpdatePersonIdForComment(db *sql.DB, person_id int64, commentId int64) (err error) {
	if commentUpdatePersonId.Stmt == nil {
		commentUpdatePersonId.Stmt, err = db.Prepare(commentUpdatePersonId.Sql)
		if err != nil {
			return
		}
	}
	result, err := commentUpdatePersonId.Stmt.Exec(person_id, commentId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to comment id %d", rows, person_id, commentId))
	}
	return nil
}

var imageFromId = bdb{
	Sql: `
SELECT file, bug_id, person_id
FROM image
WHERE image.image_id = ?
`,
}

func ImageFromId(db *sql.DB, imageId int64) (image Image, err error) {
	if imageFromId.Stmt == nil {
		imageFromId.Stmt, err = db.Prepare(imageFromId.Sql)
		if err != nil {
			return
		}
	}
	rows, err := imageFromId.Stmt.Query(imageId)
	defer rows.Close()
	if err != nil {
		return
	}
	found := false
	for rows.Next() {
		if found {
			return image, errors.New(fmt.Sprintf("Duplicate images for id %d",
				imageId))
		}
		err = rows.Scan(&image.File, &image.BugId, &image.PersonId)
		if err != nil {
			return
		}
		found = true
	}
	if !found {
		err = errors.New(fmt.Sprintf("image with id %d not found",
			imageId))
		return
	}
	image.ImageId = imageId
	return image, nil
}

var insertImage = bdb{
	Sql: `
INSERT INTO image(file, bug_id, person_id) VALUES (?, ?, ?)
`,
}

func InsertImage(db *sql.DB, image Image) (image_id int64, err error) {
	if insertImage.Stmt == nil {
		insertImage.Stmt, err = db.Prepare(insertImage.Sql)
		if err != nil {
			return
		}
	}
	result, err := insertImage.Stmt.Exec(image.File, image.BugId, image.PersonId)
	if err != nil {
		return
	}
	image_id, err = result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return image_id, nil
}

var allImages = bdb{
	Sql: `
SELECT image_id, file, bug_id, person_id FROM image
`,
}

func AllImages(db *sql.DB) (images []Image, err error) {
	if allImages.Stmt == nil {
		allImages.Stmt, err = db.Prepare(allImages.Sql)
		if err != nil {
			return
		}
	}
	rows, err := allImages.Stmt.Query()
	for rows.Next() {
		var image Image
		err = rows.Scan(&image.ImageId, &image.File, &image.BugId, &image.PersonId)
		if err != nil {
			return
		}
		images = append(images, image)
	}
	return images, nil
}

var imageFromFile = bdb{
	Sql: `
SELECT image_id, bug_id, person_id
FROM image
WHERE image.file = ?
`,
}

func ImageFromFile(db *sql.DB, file string) (image Image, err error) {
	if imageFromFile.Stmt == nil {
		imageFromFile.Stmt, err = db.Prepare(imageFromFile.Sql)
		if err != nil {
			return
		}
	}
	rows, err := imageFromFile.Stmt.Query(file)
	defer rows.Close()
	if err != nil {
		return
	}
	found := false
	for rows.Next() {
		if found {
			return image, errors.New(fmt.Sprintf("Duplicate images for file %v",
				file))
		}
		err = rows.Scan(&image.ImageId, &image.BugId, &image.PersonId)
		if err != nil {
			return
		}
	}
	return image, nil
}

var imageUpdateFile = bdb{
	Sql: `
UPDATE image
SET file = ?
WHERE image_id = ?
`,
}

func UpdateFileForImage(db *sql.DB, file string, imageId int64) (err error) {
	if imageUpdateFile.Stmt == nil {
		imageUpdateFile.Stmt, err = db.Prepare(imageUpdateFile.Sql)
		if err != nil {
			return
		}
	}
	result, err := imageUpdateFile.Stmt.Exec(file, imageId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to image id %d", rows, file, imageId))
	}
	return nil
}

var imageFromBugId = bdb{
	Sql: `
SELECT image_id, file, person_id
FROM image
WHERE image.bug_id = ?
`,
}

func ImagesFromBugId(db *sql.DB, bug_id int64) (images []Image, err error) {
	if imageFromBugId.Stmt == nil {
		imageFromBugId.Stmt, err = db.Prepare(imageFromBugId.Sql)
		if err != nil {
			return
		}
	}
	rows, err := imageFromBugId.Stmt.Query(bug_id)
	defer rows.Close()
	if err != nil {
		return
	}
	for rows.Next() {
		var image Image
		err = rows.Scan(&image.ImageId, &image.File, &image.PersonId)
		if err != nil {
			return
		}
		images = append(images, image)
	}
	return images, nil
}

var imageUpdateBugId = bdb{
	Sql: `
UPDATE image
SET bug_id = ?
WHERE image_id = ?
`,
}

func UpdateBugIdForImage(db *sql.DB, bug_id int64, imageId int64) (err error) {
	if imageUpdateBugId.Stmt == nil {
		imageUpdateBugId.Stmt, err = db.Prepare(imageUpdateBugId.Sql)
		if err != nil {
			return
		}
	}
	result, err := imageUpdateBugId.Stmt.Exec(bug_id, imageId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to image id %d", rows, bug_id, imageId))
	}
	return nil
}

var imageFromPersonId = bdb{
	Sql: `
SELECT image_id, file, bug_id
FROM image
WHERE image.person_id = ?
`,
}

func ImagesFromPersonId(db *sql.DB, person_id int64) (images []Image, err error) {
	if imageFromPersonId.Stmt == nil {
		imageFromPersonId.Stmt, err = db.Prepare(imageFromPersonId.Sql)
		if err != nil {
			return
		}
	}
	rows, err := imageFromPersonId.Stmt.Query(person_id)
	defer rows.Close()
	if err != nil {
		return
	}
	for rows.Next() {
		var image Image
		err = rows.Scan(&image.ImageId, &image.File, &image.BugId)
		if err != nil {
			return
		}
		images = append(images, image)
	}
	return images, nil
}

var imageUpdatePersonId = bdb{
	Sql: `
UPDATE image
SET person_id = ?
WHERE image_id = ?
`,
}

func UpdatePersonIdForImage(db *sql.DB, person_id int64, imageId int64) (err error) {
	if imageUpdatePersonId.Stmt == nil {
		imageUpdatePersonId.Stmt, err = db.Prepare(imageUpdatePersonId.Sql)
		if err != nil {
			return
		}
	}
	result, err := imageUpdatePersonId.Stmt.Exec(person_id, imageId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to image id %d", rows, person_id, imageId))
	}
	return nil
}

var personFromId = bdb{
	Sql: `
SELECT name, email, password
FROM person
WHERE person.person_id = ?
`,
}

func PersonFromId(db *sql.DB, personId int64) (person Person, err error) {
	if personFromId.Stmt == nil {
		personFromId.Stmt, err = db.Prepare(personFromId.Sql)
		if err != nil {
			return
		}
	}
	rows, err := personFromId.Stmt.Query(personId)
	defer rows.Close()
	if err != nil {
		return
	}
	found := false
	for rows.Next() {
		if found {
			return person, errors.New(fmt.Sprintf("Duplicate persons for id %d",
				personId))
		}
		err = rows.Scan(&person.Name, &person.Email, &person.Password)
		if err != nil {
			return
		}
		found = true
	}
	if !found {
		err = errors.New(fmt.Sprintf("person with id %d not found",
			personId))
		return
	}
	person.PersonId = personId
	return person, nil
}

var insertPerson = bdb{
	Sql: `
INSERT INTO person(name, email, password) VALUES (?, ?, ?)
`,
}

func InsertPerson(db *sql.DB, person Person) (person_id int64, err error) {
	if insertPerson.Stmt == nil {
		insertPerson.Stmt, err = db.Prepare(insertPerson.Sql)
		if err != nil {
			return
		}
	}
	result, err := insertPerson.Stmt.Exec(person.Name, person.Email, person.Password)
	if err != nil {
		return
	}
	person_id, err = result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return person_id, nil
}

var allPersons = bdb{
	Sql: `
SELECT person_id, name, email, password FROM person
`,
}

func AllPersons(db *sql.DB) (persons []Person, err error) {
	if allPersons.Stmt == nil {
		allPersons.Stmt, err = db.Prepare(allPersons.Sql)
		if err != nil {
			return
		}
	}
	rows, err := allPersons.Stmt.Query()
	for rows.Next() {
		var person Person
		err = rows.Scan(&person.PersonId, &person.Name, &person.Email, &person.Password)
		if err != nil {
			return
		}
		persons = append(persons, person)
	}
	return persons, nil
}

var personFromName = bdb{
	Sql: `
SELECT person_id, email, password
FROM person
WHERE person.name = ?
`,
}

func PersonFromName(db *sql.DB, name string) (person Person, err error) {
	if personFromName.Stmt == nil {
		personFromName.Stmt, err = db.Prepare(personFromName.Sql)
		if err != nil {
			return
		}
	}
	rows, err := personFromName.Stmt.Query(name)
	defer rows.Close()
	if err != nil {
		return
	}
	found := false
	for rows.Next() {
		if found {
			return person, errors.New(fmt.Sprintf("Duplicate persons for name %v",
				name))
		}
		err = rows.Scan(&person.PersonId, &person.Email, &person.Password)
		if err != nil {
			return
		}
	}
	return person, nil
}

var personUpdateName = bdb{
	Sql: `
UPDATE person
SET name = ?
WHERE person_id = ?
`,
}

func UpdateNameForPerson(db *sql.DB, name string, personId int64) (err error) {
	if personUpdateName.Stmt == nil {
		personUpdateName.Stmt, err = db.Prepare(personUpdateName.Sql)
		if err != nil {
			return
		}
	}
	result, err := personUpdateName.Stmt.Exec(name, personId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to person id %d", rows, name, personId))
	}
	return nil
}

var personFromEmail = bdb{
	Sql: `
SELECT person_id, name, password
FROM person
WHERE person.email = ?
`,
}

func PersonFromEmail(db *sql.DB, email string) (person Person, err error) {
	if personFromEmail.Stmt == nil {
		personFromEmail.Stmt, err = db.Prepare(personFromEmail.Sql)
		if err != nil {
			return
		}
	}
	rows, err := personFromEmail.Stmt.Query(email)
	defer rows.Close()
	if err != nil {
		return
	}
	found := false
	for rows.Next() {
		if found {
			return person, errors.New(fmt.Sprintf("Duplicate persons for email %v",
				email))
		}
		err = rows.Scan(&person.PersonId, &person.Name, &person.Password)
		if err != nil {
			return
		}
	}
	return person, nil
}

var personUpdateEmail = bdb{
	Sql: `
UPDATE person
SET email = ?
WHERE person_id = ?
`,
}

func UpdateEmailForPerson(db *sql.DB, email string, personId int64) (err error) {
	if personUpdateEmail.Stmt == nil {
		personUpdateEmail.Stmt, err = db.Prepare(personUpdateEmail.Sql)
		if err != nil {
			return
		}
	}
	result, err := personUpdateEmail.Stmt.Exec(email, personId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to person id %d", rows, email, personId))
	}
	return nil
}

var personFromPassword = bdb{
	Sql: `
SELECT person_id, name, email
FROM person
WHERE person.password = ?
`,
}

func PersonsFromPassword(db *sql.DB, password string) (persons []Person, err error) {
	if personFromPassword.Stmt == nil {
		personFromPassword.Stmt, err = db.Prepare(personFromPassword.Sql)
		if err != nil {
			return
		}
	}
	rows, err := personFromPassword.Stmt.Query(password)
	defer rows.Close()
	if err != nil {
		return
	}
	for rows.Next() {
		var person Person
		err = rows.Scan(&person.PersonId, &person.Name, &person.Email)
		if err != nil {
			return
		}
		persons = append(persons, person)
	}
	return persons, nil
}

var personUpdatePassword = bdb{
	Sql: `
UPDATE person
SET password = ?
WHERE person_id = ?
`,
}

func UpdatePasswordForPerson(db *sql.DB, password string, personId int64) (err error) {
	if personUpdatePassword.Stmt == nil {
		personUpdatePassword.Stmt, err = db.Prepare(personUpdatePassword.Sql)
		if err != nil {
			return
		}
	}
	result, err := personUpdatePassword.Stmt.Exec(password, personId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to person id %d", rows, password, personId))
	}
	return nil
}

var dependencyFromId = bdb{
	Sql: `
SELECT cause, effect
FROM dependency
WHERE dependency.dependency_id = ?
`,
}

func DependencyFromId(db *sql.DB, dependencyId int64) (dependency Dependency, err error) {
	if dependencyFromId.Stmt == nil {
		dependencyFromId.Stmt, err = db.Prepare(dependencyFromId.Sql)
		if err != nil {
			return
		}
	}
	rows, err := dependencyFromId.Stmt.Query(dependencyId)
	defer rows.Close()
	if err != nil {
		return
	}
	found := false
	for rows.Next() {
		if found {
			return dependency, errors.New(fmt.Sprintf("Duplicate dependencys for id %d",
				dependencyId))
		}
		err = rows.Scan(&dependency.Cause, &dependency.Effect)
		if err != nil {
			return
		}
		found = true
	}
	if !found {
		err = errors.New(fmt.Sprintf("dependency with id %d not found",
			dependencyId))
		return
	}
	dependency.DependencyId = dependencyId
	return dependency, nil
}

var insertDependency = bdb{
	Sql: `
INSERT INTO dependency(cause, effect) VALUES (?, ?)
`,
}

func InsertDependency(db *sql.DB, dependency Dependency) (dependency_id int64, err error) {
	if insertDependency.Stmt == nil {
		insertDependency.Stmt, err = db.Prepare(insertDependency.Sql)
		if err != nil {
			return
		}
	}
	result, err := insertDependency.Stmt.Exec(dependency.Cause, dependency.Effect)
	if err != nil {
		return
	}
	dependency_id, err = result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return dependency_id, nil
}

var allDependencys = bdb{
	Sql: `
SELECT dependency_id, cause, effect FROM dependency
`,
}

func AllDependencys(db *sql.DB) (dependencys []Dependency, err error) {
	if allDependencys.Stmt == nil {
		allDependencys.Stmt, err = db.Prepare(allDependencys.Sql)
		if err != nil {
			return
		}
	}
	rows, err := allDependencys.Stmt.Query()
	for rows.Next() {
		var dependency Dependency
		err = rows.Scan(&dependency.DependencyId, &dependency.Cause, &dependency.Effect)
		if err != nil {
			return
		}
		dependencys = append(dependencys, dependency)
	}
	return dependencys, nil
}

var dependencyFromCause = bdb{
	Sql: `
SELECT dependency_id, effect
FROM dependency
WHERE dependency.cause = ?
`,
}

func DependencysFromCause(db *sql.DB, cause int64) (dependencys []Dependency, err error) {
	if dependencyFromCause.Stmt == nil {
		dependencyFromCause.Stmt, err = db.Prepare(dependencyFromCause.Sql)
		if err != nil {
			return
		}
	}
	rows, err := dependencyFromCause.Stmt.Query(cause)
	defer rows.Close()
	if err != nil {
		return
	}
	for rows.Next() {
		var dependency Dependency
		err = rows.Scan(&dependency.DependencyId, &dependency.Effect)
		if err != nil {
			return
		}
		dependencys = append(dependencys, dependency)
	}
	return dependencys, nil
}

var dependencyUpdateCause = bdb{
	Sql: `
UPDATE dependency
SET cause = ?
WHERE dependency_id = ?
`,
}

func UpdateCauseForDependency(db *sql.DB, cause int64, dependencyId int64) (err error) {
	if dependencyUpdateCause.Stmt == nil {
		dependencyUpdateCause.Stmt, err = db.Prepare(dependencyUpdateCause.Sql)
		if err != nil {
			return
		}
	}
	result, err := dependencyUpdateCause.Stmt.Exec(cause, dependencyId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to dependency id %d", rows, cause, dependencyId))
	}
	return nil
}

var dependencyFromEffect = bdb{
	Sql: `
SELECT dependency_id, cause
FROM dependency
WHERE dependency.effect = ?
`,
}

func DependencysFromEffect(db *sql.DB, effect int64) (dependencys []Dependency, err error) {
	if dependencyFromEffect.Stmt == nil {
		dependencyFromEffect.Stmt, err = db.Prepare(dependencyFromEffect.Sql)
		if err != nil {
			return
		}
	}
	rows, err := dependencyFromEffect.Stmt.Query(effect)
	defer rows.Close()
	if err != nil {
		return
	}
	for rows.Next() {
		var dependency Dependency
		err = rows.Scan(&dependency.DependencyId, &dependency.Cause)
		if err != nil {
			return
		}
		dependencys = append(dependencys, dependency)
	}
	return dependencys, nil
}

var dependencyUpdateEffect = bdb{
	Sql: `
UPDATE dependency
SET effect = ?
WHERE dependency_id = ?
`,
}

func UpdateEffectForDependency(db *sql.DB, effect int64, dependencyId int64) (err error) {
	if dependencyUpdateEffect.Stmt == nil {
		dependencyUpdateEffect.Stmt, err = db.Prepare(dependencyUpdateEffect.Sql)
		if err != nil {
			return
		}
	}
	result, err := dependencyUpdateEffect.Stmt.Exec(effect, dependencyId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to dependency id %d", rows, effect, dependencyId))
	}
	return nil
}

var duplicateFromId = bdb{
	Sql: `
SELECT original, duplicate
FROM duplicate
WHERE duplicate.duplicate_id = ?
`,
}

func DuplicateFromId(db *sql.DB, duplicateId int64) (duplicate Duplicate, err error) {
	if duplicateFromId.Stmt == nil {
		duplicateFromId.Stmt, err = db.Prepare(duplicateFromId.Sql)
		if err != nil {
			return
		}
	}
	rows, err := duplicateFromId.Stmt.Query(duplicateId)
	defer rows.Close()
	if err != nil {
		return
	}
	found := false
	for rows.Next() {
		if found {
			return duplicate, errors.New(fmt.Sprintf("Duplicate duplicates for id %d",
				duplicateId))
		}
		err = rows.Scan(&duplicate.Original, &duplicate.Duplicate)
		if err != nil {
			return
		}
		found = true
	}
	if !found {
		err = errors.New(fmt.Sprintf("duplicate with id %d not found",
			duplicateId))
		return
	}
	duplicate.DuplicateId = duplicateId
	return duplicate, nil
}

var insertDuplicate = bdb{
	Sql: `
INSERT INTO duplicate(original, duplicate) VALUES (?, ?)
`,
}

func InsertDuplicate(db *sql.DB, duplicate Duplicate) (duplicate_id int64, err error) {
	if insertDuplicate.Stmt == nil {
		insertDuplicate.Stmt, err = db.Prepare(insertDuplicate.Sql)
		if err != nil {
			return
		}
	}
	result, err := insertDuplicate.Stmt.Exec(duplicate.Original, duplicate.Duplicate)
	if err != nil {
		return
	}
	duplicate_id, err = result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return duplicate_id, nil
}

var allDuplicates = bdb{
	Sql: `
SELECT duplicate_id, original, duplicate FROM duplicate
`,
}

func AllDuplicates(db *sql.DB) (duplicates []Duplicate, err error) {
	if allDuplicates.Stmt == nil {
		allDuplicates.Stmt, err = db.Prepare(allDuplicates.Sql)
		if err != nil {
			return
		}
	}
	rows, err := allDuplicates.Stmt.Query()
	for rows.Next() {
		var duplicate Duplicate
		err = rows.Scan(&duplicate.DuplicateId, &duplicate.Original, &duplicate.Duplicate)
		if err != nil {
			return
		}
		duplicates = append(duplicates, duplicate)
	}
	return duplicates, nil
}

var duplicateFromOriginal = bdb{
	Sql: `
SELECT duplicate_id, duplicate
FROM duplicate
WHERE duplicate.original = ?
`,
}

func DuplicatesFromOriginal(db *sql.DB, original int64) (duplicates []Duplicate, err error) {
	if duplicateFromOriginal.Stmt == nil {
		duplicateFromOriginal.Stmt, err = db.Prepare(duplicateFromOriginal.Sql)
		if err != nil {
			return
		}
	}
	rows, err := duplicateFromOriginal.Stmt.Query(original)
	defer rows.Close()
	if err != nil {
		return
	}
	for rows.Next() {
		var duplicate Duplicate
		err = rows.Scan(&duplicate.DuplicateId, &duplicate.Duplicate)
		if err != nil {
			return
		}
		duplicates = append(duplicates, duplicate)
	}
	return duplicates, nil
}

var duplicateUpdateOriginal = bdb{
	Sql: `
UPDATE duplicate
SET original = ?
WHERE duplicate_id = ?
`,
}

func UpdateOriginalForDuplicate(db *sql.DB, original int64, duplicateId int64) (err error) {
	if duplicateUpdateOriginal.Stmt == nil {
		duplicateUpdateOriginal.Stmt, err = db.Prepare(duplicateUpdateOriginal.Sql)
		if err != nil {
			return
		}
	}
	result, err := duplicateUpdateOriginal.Stmt.Exec(original, duplicateId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to duplicate id %d", rows, original, duplicateId))
	}
	return nil
}

var duplicateFromDuplicate = bdb{
	Sql: `
SELECT duplicate_id, original
FROM duplicate
WHERE duplicate.duplicate = ?
`,
}

func DuplicatesFromDuplicate(db *sql.DB, duplicate int64) (duplicates []Duplicate, err error) {
	if duplicateFromDuplicate.Stmt == nil {
		duplicateFromDuplicate.Stmt, err = db.Prepare(duplicateFromDuplicate.Sql)
		if err != nil {
			return
		}
	}
	rows, err := duplicateFromDuplicate.Stmt.Query(duplicate)
	defer rows.Close()
	if err != nil {
		return
	}
	for rows.Next() {
		var duplicate Duplicate
		err = rows.Scan(&duplicate.DuplicateId, &duplicate.Original)
		if err != nil {
			return
		}
		duplicates = append(duplicates, duplicate)
	}
	return duplicates, nil
}

var duplicateUpdateDuplicate = bdb{
	Sql: `
UPDATE duplicate
SET duplicate = ?
WHERE duplicate_id = ?
`,
}

func UpdateDuplicateForDuplicate(db *sql.DB, duplicate int64, duplicateId int64) (err error) {
	if duplicateUpdateDuplicate.Stmt == nil {
		duplicateUpdateDuplicate.Stmt, err = db.Prepare(duplicateUpdateDuplicate.Sql)
		if err != nil {
			return
		}
	}
	result, err := duplicateUpdateDuplicate.Stmt.Exec(duplicate, duplicateId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to duplicate id %d", rows, duplicate, duplicateId))
	}
	return nil
}

var txtFromId = bdb{
	Sql: `
SELECT entered, content
FROM txt
WHERE txt.txt_id = ?
`,
}

func TxtFromId(db *sql.DB, txtId int64) (txt Txt, err error) {
	if txtFromId.Stmt == nil {
		txtFromId.Stmt, err = db.Prepare(txtFromId.Sql)
		if err != nil {
			return
		}
	}
	rows, err := txtFromId.Stmt.Query(txtId)
	defer rows.Close()
	if err != nil {
		return
	}
	found := false
	for rows.Next() {
		if found {
			return txt, errors.New(fmt.Sprintf("Duplicate txts for id %d",
				txtId))
		}
		err = rows.Scan(&txt.Entered, &txt.Content)
		if err != nil {
			return
		}
		found = true
	}
	if !found {
		err = errors.New(fmt.Sprintf("txt with id %d not found",
			txtId))
		return
	}
	txt.TxtId = txtId
	return txt, nil
}

var insertTxt = bdb{
	Sql: `
INSERT INTO txt(entered, content) VALUES (?, ?)
`,
}

func InsertTxt(db *sql.DB, txt Txt) (txt_id int64, err error) {
	if insertTxt.Stmt == nil {
		insertTxt.Stmt, err = db.Prepare(insertTxt.Sql)
		if err != nil {
			return
		}
	}
	result, err := insertTxt.Stmt.Exec(txt.Entered, txt.Content)
	if err != nil {
		return
	}
	txt_id, err = result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return txt_id, nil
}

var allTxts = bdb{
	Sql: `
SELECT txt_id, entered, content FROM txt
`,
}

func AllTxts(db *sql.DB) (txts []Txt, err error) {
	if allTxts.Stmt == nil {
		allTxts.Stmt, err = db.Prepare(allTxts.Sql)
		if err != nil {
			return
		}
	}
	rows, err := allTxts.Stmt.Query()
	for rows.Next() {
		var txt Txt
		err = rows.Scan(&txt.TxtId, &txt.Entered, &txt.Content)
		if err != nil {
			return
		}
		txts = append(txts, txt)
	}
	return txts, nil
}

var txtFromEntered = bdb{
	Sql: `
SELECT txt_id, content
FROM txt
WHERE txt.entered = ?
`,
}

func TxtsFromEntered(db *sql.DB, entered time.Time) (txts []Txt, err error) {
	if txtFromEntered.Stmt == nil {
		txtFromEntered.Stmt, err = db.Prepare(txtFromEntered.Sql)
		if err != nil {
			return
		}
	}
	rows, err := txtFromEntered.Stmt.Query(entered)
	defer rows.Close()
	if err != nil {
		return
	}
	for rows.Next() {
		var txt Txt
		err = rows.Scan(&txt.TxtId, &txt.Content)
		if err != nil {
			return
		}
		txts = append(txts, txt)
	}
	return txts, nil
}

var txtUpdateEntered = bdb{
	Sql: `
UPDATE txt
SET entered = ?
WHERE txt_id = ?
`,
}

func UpdateEnteredForTxt(db *sql.DB, entered time.Time, txtId int64) (err error) {
	if txtUpdateEntered.Stmt == nil {
		txtUpdateEntered.Stmt, err = db.Prepare(txtUpdateEntered.Sql)
		if err != nil {
			return
		}
	}
	result, err := txtUpdateEntered.Stmt.Exec(entered, txtId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to txt id %d", rows, entered, txtId))
	}
	return nil
}

var txtFromContent = bdb{
	Sql: `
SELECT txt_id, entered
FROM txt
WHERE txt.content = ?
`,
}

func TxtsFromContent(db *sql.DB, content string) (txts []Txt, err error) {
	if txtFromContent.Stmt == nil {
		txtFromContent.Stmt, err = db.Prepare(txtFromContent.Sql)
		if err != nil {
			return
		}
	}
	rows, err := txtFromContent.Stmt.Query(content)
	defer rows.Close()
	if err != nil {
		return
	}
	for rows.Next() {
		var txt Txt
		err = rows.Scan(&txt.TxtId, &txt.Entered)
		if err != nil {
			return
		}
		txts = append(txts, txt)
	}
	return txts, nil
}

var txtUpdateContent = bdb{
	Sql: `
UPDATE txt
SET content = ?
WHERE txt_id = ?
`,
}

func UpdateContentForTxt(db *sql.DB, content string, txtId int64) (err error) {
	if txtUpdateContent.Stmt == nil {
		txtUpdateContent.Stmt, err = db.Prepare(txtUpdateContent.Sql)
		if err != nil {
			return
		}
	}
	result, err := txtUpdateContent.Stmt.Exec(content, txtId)
	if err != nil {
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if rows != 1 {
		return errors.New(fmt.Sprintf("%d rows returned assigning %v to txt id %d", rows, content, txtId))
	}
	return nil
}
