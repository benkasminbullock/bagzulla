// Bagzulla is a bug-tracking application

// This is the main program of Bagzulla.

// Database-specific access code goes into "database.go".

// Authentication goes into "auth.go".

// There is also an automatically-generated file
// bagzullaDb/bagzullaDb.go.

package main

import (
	"bagzulla/bagzullaDb"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/benkasminbullock/gologin/login"
	"github.com/benkasminbullock/gologin/store"
)

// The various priorities that a bug may have. The default value is
// "unknown".
var priorities = []string{
	"unknown",
	"top",
	"high",
	"medium",
	"low",
	"unimportant",
}

// A bug with no project number
const ProjectNone = 1

// Write a compressed response.
type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

// Use the Writer part of gzipResponseWriter to write the output.
func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// Holder for application information.
type Bagapp struct {
	store login.LoginStore
	login login.Login
	// The connection to the database.
	db *sql.DB
	// All the templates after reading in.
	templates *template.Template
	port      string
	TopURL    string
	CSS       string
	// Application to display a directory or file
	DisplayDir string
	Cancel     context.CancelFunc
	Context    context.Context
	Server     *http.Server
}

// Holder for an individual interaction with the bug tracker.
type Bagreply struct {
	App *Bagapp
	w   http.ResponseWriter
	r   *http.Request
	// The user's identification, nil if unidentified.
	User *bagzullaDb.Person
	// The title of the page
	Title string
}

// Any handler.
type BagFunc func(*Bagreply)

// Convert a string like "open" to a status number.
func stringToStatus(statusString string) (int64, error) {
	for i, s := range statuses {
		if statusString == s {
			return int64(i), nil
		}
	}
	return -1, fmt.Errorf("Unknown status %s", statusString)
}

// Convert a string like "high" to a priority number.
func stringToPriority(priorityString string) (int64, error) {
	for i, s := range priorities {
		if priorityString == s {
			return int64(i), nil
		}
	}
	return -1, fmt.Errorf("Unknown priority %s", priorityString)
}

// Get an array value. This is used by the template construction.
func GetArray(counts []int64, id int64) int64 {
	return counts[id]
}

// For an individual bug's page, a comment on the bug.
type ListComment struct {
	Comment bagzullaDb.Comment
	Txt     bagzullaDb.Txt
	Person  string
}

// This type contains information about related bugs, such as the bug
// which this bug depends on, or the bug which this blocks.
type RelatedBug struct {
	Id     int64
	Status int64
}

// A bug as displayed in the user's browser. This contains data taken
// from various database fields and processed to make it more
// understandable. For example the priority in the database is just a
// number, but here it is a meaningful word.
type ListBug struct {
	Bug         bagzullaDb.Bug
	Description string
	// Description with e.g. urls changed to links.
	DisplayDescription string
	// Estimate of completion time
	Estimate string
	// The title of the bug
	Title        string
	DisplayTitle string
	// The status of the bug.
	Status string
	// The priority of the bug.
	Priority string
	// The name of the project this belongs to.
	ProjectName string
	// The ID of the project this bug belongs to.
	ProjectId int64
	// The name of the part of the project which this bug belongs to.
	PartName string
	// The name of the owner of this bug.
	Owner string
	// The comments made regarding this bug.
	Comments []ListComment
	// Duplicates of this bug
	Duplicates []RelatedBug
	// Bugs which this bug duplicates
	Originals []RelatedBug
	// Upstream bugs caused by this bug
	DependsOn []RelatedBug
	// Bugs which this blocks
	Blocks []RelatedBug
	// The possible values for the status field of the bug's form.
	Statuses []string
	// The possible values for the priority field of the bug's form.
	Priorities []string
	// The possible values for the project field of the bug's form.
	Projects []bagzullaDb.Project
	Images   []bagzullaDb.Image
	User     *bagzullaDb.Person
}

// A structure which contains a list of bugs. For example a search
// result.
type ListBugPage struct {
	Title string
	Bugs  []ListBug
}

// A cache of the project names
var projectNames = make(map[int64]string)

// Given the ID number of a project, get its name. If the name cannot
// be found, an error page is produced for the user, and the second
// return value is "false".
func getProjectName(b *Bagreply, projectId int64) (string, bool) {
	if projectId == 0 {
		return "None", true
	}
	// Look in the cache of names first
	name, ok := projectNames[projectId]
	if ok {
		return name, true
	}
	// The name was not in the cache, look in the database.
	project, err := bagzullaDb.ProjectFromId(b.App.db, projectId)
	if err != nil {
		b.errorPage("Error retrieving project with id %d from database: %s",
			projectId, err.Error())
		return "", false
	}
	name = project.Name
	// Save this name to the cache.
	projectNames[projectId] = name
	return name, true
}

// Given a project name, find the corresponding project ID.
func (b *Bagreply) projectIdFromName(projectName string) (int64, bool) {
	project, err := bagzullaDb.ProjectFromName(b.App.db, projectName)
	if err != nil {
		b.errorPage("Error getting project with name '%s' from database: %s",
			projectName, err.Error())
		return 0, false
	}
	return project.ProjectId, true
}

// Cache of part names so that we don't re-read the database over and
// over.
var partNames = make(map[int64]string)

// Given a part ID, get the name of the part.
func getPartName(b *Bagreply, partId int64) (string, bool) {
	if partId == 0 {
		return "None", true
	}
	name, ok := partNames[partId]
	if ok {
		return name, true
	}
	part, err := bagzullaDb.PartFromId(b.App.db, partId)
	if err != nil {
		b.errorPage("Error retrieving part with id %d from database: %s", partId, err.Error())
		return "", false
	}
	name = part.Name
	partNames[partId] = name
	return name, true
}

// Cache of person names
var personNames = make(map[int64]string)

// Given the ID of a person, get their name.
func getPersonName(b *Bagreply, personId int64) (string, bool) {
	if personId == 0 {
		return "None", true
	}
	name, ok := personNames[personId]
	if ok {
		return name, true
	}
	person, err := bagzullaDb.PersonFromId(b.App.db, personId)
	if err != nil {
		b.errorPage("Error retrieving person with id %d from database: %s", personId, err.Error())
		return "", false
	}
	name = person.Name
	personNames[personId] = name
	return name, true
}

var finalNum = regexp.MustCompile("/([0-9]+)$")

// Get the final number at the end of the URL. "ok" is true if the
// number was successfully parsed, false otherwise. Errors are handled
// here, so the caller does not need to do further processing if ok is
// false.
func getFinalNum(b *Bagreply) (num int64, ok bool) {
	m := finalNum.FindStringSubmatch(b.r.URL.Path)
	if m == nil {
		b.errorPage("Url '%s' should end in a number", b.r.URL.Path)
		return 0, false
	}
	num, err := strconv.ParseInt(m[1], 10, 64)
	if err != nil {
		b.errorPage("Could not get number from %s: %s",
			m[1], err.Error())
		return 0, false
	}
	if num == 0 {
		b.errorPage("ID cannot be zero in %s", b.r.URL.Path)
		return 0, false
	}
	return num, true
}

// Insert a piece of text into the text-storing place of the database.
func insertText(b *Bagreply, text string) (id int64, ok bool) {
	var txt = bagzullaDb.Txt{
		Content: text,
		Entered: time.Now(),
	}
	id, err := bagzullaDb.InsertTxt(b.App.db, txt)
	if err != nil {
		b.errorPage("Error inserting text %s: %s", text, err.Error())
		return 0, false
	}
	return id, true
}

func (b *Bagreply) GetText(id int64) (text string, ok bool) {
	if id == 0 {
		return "", true
	}
	dbtext, dbok := getText(b, id)
	if !dbok {
		return text, false
	}
	return html.EscapeString(dbtext.Content), true
}

// Retrieve a piece of text from the text-storing place of the
// database. The input is the ID of the text. The return values are
// the text and whether or not it was found.
func getText(b *Bagreply, id int64) (text bagzullaDb.Txt, ok bool) {
	text, err := bagzullaDb.TxtFromId(b.App.db, id)
	if err != nil {
		b.errorPage("Error retrieving text with ID %d from database: %s",
			id, err.Error())
		return text, false
	}
	return text, true
}

// Return the project specified by the number at the end of the
// current URL.
func getProject(b *Bagreply) (project bagzullaDb.Project, ok bool) {
	projectId, ok := getFinalNum(b)
	if !ok {
		return
	}
	if projectId == 0 {
		b.errorPage("Project %d does not have a description", projectId)
		return project, false
	}
	project, err := bagzullaDb.ProjectFromId(b.App.db, projectId)
	if err != nil {
		b.errorPage("Error retrieving project with id %d from database: %s",
			projectId, err.Error())
		return project, false
	}
	return project, true
}

// Return the part specified by the current URL.
func getPart(b *Bagreply) (part bagzullaDb.Part, ok bool) {
	partId, ok := getFinalNum(b)
	if !ok {
		return
	}
	part, err := bagzullaDb.PartFromId(b.App.db, partId)
	if err != nil {
		b.errorPage("Error getting part with ID %d from database: %s",
			partId, err.Error())
		return part, false
	}
	return part, true
}

// Get information on a bug to fill out the page on the bug. If ok is
// false, the caller can return safely assuming that an error has
// already been handled.
func getBugInfo(b *Bagreply, bug bagzullaDb.Bug) (lb ListBug, ok bool) {
	lb.Bug = bug
	lb.User = b.User
	title, ok := getText(b, bug.Title)
	if !ok {
		return lb, false
	}
	lb.Title = title.Content
	lb.DisplayTitle = html.EscapeString(lb.Title)
	description, ok := getText(b, bug.Description)
	if !ok {
		return lb, false
	}
	lb.Description = description.Content
	if bug.ProjectId != ProjectNone {
		var ok bool
		lb.ProjectName, ok = getProjectName(b, bug.ProjectId)
		if !ok {
			return lb, false
		}
		lb.ProjectId = bug.ProjectId
	}
	if bug.PartId != 0 {
		lb.PartName, ok = getPartName(b, bug.PartId)
		if !ok {
			return lb, false
		}
	}
	lb.Status = statuses[bug.Status]
	lb.Priority = priorities[bug.Priority]
	lb.Owner, ok = getPersonName(b, bug.Owner)
	lb.Estimate = "unknown"
	if bug.Estimate > 0 {
		lb.Estimate = fmt.Sprintf("%d minutes", bug.Estimate)
	}
	if !ok {
		return lb, false
	}
	return lb, true
}

// Put information about the bugs specified by "bugs" into a
// ListBugPage structure "p".
func getBugsInfo(b *Bagreply, p *ListBugPage, bugs []bagzullaDb.Bug) bool {
	for _, bug := range bugs {
		var lb ListBug
		lb, ok := getBugInfo(b, bug)
		if !ok {
			return false
		}
		p.Bugs = append(p.Bugs, lb)
	}
	return true
}

// Run a template with error handling if the template fails to process.
func (b *Bagreply) runATemplate(name string, data interface{}) bool {
	t := b.App.templates.Lookup(name)
	err := t.Execute(b.w, data)
	if err != nil {
		b.errorPage("Error executing template %s: %s", name, err.Error())
		return false
	}
	return true
}

// Run all the templates to make a page, with error recovery.
func (b *Bagreply) runTemplate(name string, data interface{}) {
	ok := b.runATemplate("top.html", b)
	if !ok {
		return
	}
	ok = b.runATemplate(name, data)
	if !ok {
		return
	}
	ok = b.runATemplate("bottom.html", b)
	if !ok {
		return
	}
}

// Print an error page.
type ErrorPage struct {
	Text string
}

func (b *Bagreply) errorPage(format string, a ...interface{}) {
	errorText := fmt.Sprintf(format, a...)
	var ep = ErrorPage{
		Text: errorText,
	}
	b.runTemplate("error.html", ep)
}

type bugsById []ListBug

func (bid bugsById) Len() int {
	return len(bid)
}

func (bid bugsById) Swap(i, j int) {
	bid[i], bid[j] = bid[j], bid[i]
}

func (bid bugsById) Less(i, j int) bool {
	return bid[i].Bug.BugId > bid[j].Bug.BugId
}

func sortBugsById(bugs []ListBug) {
	sort.Sort(bugsById(bugs))
}

type bugsByPriority []ListBug

func (bpr bugsByPriority) Len() int {
	return len(bpr)
}

func (bpr bugsByPriority) Swap(i, j int) {
	bpr[i], bpr[j] = bpr[j], bpr[i]
}

func (bpr bugsByPriority) Less(i, j int) bool {
	ip := bpr[i].Bug.Priority
	jp := bpr[j].Bug.Priority
	if ip == 0 {
		return false
	}
	if jp == 0 {
		return true
	}
	return ip < jp
}

func sortBugsByPriority(bugs []ListBug) {
	sort.Sort(bugsByPriority(bugs))
}

var sortOrder = regexp.MustCompile("/(priority|id|changed)$")

func sortBugs(b *Bagreply, p ListBugPage) (sortType string) {
	m := sortOrder.FindStringSubmatch(b.r.URL.Path)
	if m != nil {
		order := m[1]
		switch order {
		case "id":
			sortBugsById(p.Bugs)
		case "priority":
			sortBugsByPriority(p.Bugs)
		case "changed":
		default:
			// No sorting needs to be done.
		}
		return order
	}
	return ""
}

// Output the page of all open bugs
func openBugsHandler(b *Bagreply) {
	bugs, ok := StatusBugs(b, 0)
	if !ok {
		return
	}
	var p ListBugPage
	ok = getBugsInfo(b, &p, bugs)
	if !ok {
		return
	}
	sortType := sortBugs(b, p)
	p.Title = "Open bugs"
	switch sortType {
	case "priority":
		p.Title += " by priority"
	case "id":
		p.Title += " sorted by ID"
	}
	b.Title = p.Title
	b.runTemplate("bugs.html", p)
}

// Output the page of all bugs, regardless of status
func allBugsHandler(b *Bagreply) {
	bugs, err := bagzullaDb.AllBugs(b.App.db)
	if err != nil {
		b.errorPage("Error getting a list of all bugs: %s",
			err.Error())
		return
	}
	var p ListBugPage
	getBugsInfo(b, &p, bugs)
	b.Title = "All bugs - Bagzulla"
	p.Title = "All bugs"
	b.runTemplate("bugs.html", p)
}

func redirectToProject(b *Bagreply, projectId int64) {
	projectUrl := fmt.Sprintf("%s/project/%d", b.App.TopURL, projectId)
	http.Redirect(b.w, b.r, projectUrl, http.StatusFound)
}

// Update the time of the most recent change of the bug.
func (b *Bagreply) updateChanged(bugId int64) bool {
	err := bagzullaDb.UpdateChangedForBug(b.App.db, time.Now(), bugId)
	if err != nil {
		b.errorPage("Error updating change time for %d: %s", bugId, err.Error())
		return false
	}
	return true
}

func editProjectName(b *Bagreply) {
	project, ok := getProject(b)
	if !ok {
		return
	}
	projectName := b.r.FormValue("project-name")
	if len(projectName) > 0 {
		bagzullaDb.UpdateNameForProject(b.App.db, projectName, project.ProjectId)
		projectNames[project.ProjectId] = projectName
		redirectToProject(b, project.ProjectId)
		return
	}
	var pp ProjectPage
	pp.Project = project
	b.Title = fmt.Sprintf("Edit name of %s", project.Name)
	b.runTemplate("edit-project-name.html", pp)
}

type PartPage struct {
	Part        bagzullaDb.Part
	Description string
	Bugs        []ListBug
}

func editPartName(b *Bagreply) {
	part, ok := getPart(b)
	if !ok {
		return
	}
	partName := b.r.FormValue("part-name")
	if len(partName) > 0 {
		bagzullaDb.UpdateNameForPart(b.App.db, partName, part.PartId)
		partNames[part.PartId] = partName
		redirectToPart(b, part.PartId)
		return
	}
	var pp PartPage
	pp.Part = part
	b.Title = fmt.Sprintf("Edit name of %s", part.Name)
	b.runTemplate("edit-part-name.html", pp)
}

// Return the form text corresponding to "key" as "value". If there is
// no valid text, valid is set to false. If valid is true, value may
// be the empty string.
func (b *Bagreply) FormText(key string) (value string, valid bool) {
	value = b.r.FormValue(key)
	if len(value) == 0 {
		if _, ok := b.r.Form[key]; ok {
			return "", true
		}
		return "", false
	}
	return value, true
}

func editProjectDescription(b *Bagreply) {
	project, ok := getProject(b)
	if !ok {
		return
	}
	description, valid := b.FormText("description")
	if valid {
		var descriptionId = int64(0)
		if len(description) > 0 {
			var ok bool
			descriptionId, ok = insertText(b, description)
			if !ok {
				return
			}
		}
		if !b.DeleteText(project.Description) {
			return
		}
		bagzullaDb.UpdateDescriptionForProject(b.App.db, descriptionId, project.ProjectId)
		redirectToProject(b, project.ProjectId)
		return
	}
	var pp ProjectPage
	pp.Project = project
	d, ok := b.GetText(project.Description)
	if !ok {
		return
	}
	pp.Description = d
	b.Title = fmt.Sprintf("Edit description of %s", project.Name)
	b.runTemplate("edit-project-description.html", pp)
}

// Given b Bagreply, find the associated bug assuming that it is the
// final number in the URL.
func getBug(b *Bagreply) (bug bagzullaDb.Bug, ok bool) {
	bugid, ok := getFinalNum(b)
	if !ok {
		return
	}
	bug, err := bagzullaDb.BugFromId(b.App.db, bugid)
	if err != nil {
		b.errorPage(fmt.Sprintf("Error retrieving bug with id %d from database: %s",
			bugid, err.Error()))
		return bug, false
	}
	return bug, true
}

func editBugDescription(b *Bagreply) {
	if b.NotLoggedIn() {
		return
	}
	bug, ok := getBug(b)
	if !ok {
		return
	}
	description := b.r.FormValue("description")
	if len(description) > 0 {
		descriptionId, ok := insertText(b, description)
		if !ok {
			return
		}
		if !b.DeleteText(bug.Description) {
			return
		}
		bagzullaDb.UpdateDescriptionForBug(b.App.db, descriptionId, bug.BugId)
		ok = b.updateChanged(bug.BugId)
		if !ok {
			return
		}
		b.redirectToBug(bug.BugId)
		return
	}
	lb, ok := getBugInfo(b, bug)
	if !ok {
		return
	}
	b.Title = fmt.Sprintf("Edit bug description for %s", lb.Title)
	b.runTemplate("edit-bug-description.html", lb)
}

func redirectToPart(b *Bagreply, partId int64) {
	partUrl := fmt.Sprintf("%s/part/%d", b.App.TopURL, partId)
	http.Redirect(b.w, b.r, partUrl, http.StatusFound)
}

type PartDesc struct {
	Part        bagzullaDb.Part
	Description string
}

// Edit the part description of the current page.
func editPartDescription(b *Bagreply) {
	part, ok := getPart(b)
	if !ok {
		return
	}
	description := b.r.FormValue("description")
	if len(description) > 0 {
		partId := part.PartId
		descriptionId, ok := insertText(b, description)
		if !ok {
			return
		}
		if !b.DeleteText(part.Description) {
			return
		}
		bagzullaDb.UpdateDescriptionForPart(b.App.db, descriptionId, partId)
		redirectToPart(b, partId)
		return
	}
	var pd PartDesc
	pd.Part = part
	d, ok := getText(b, part.Description)
	if !ok {
		return
	}
	pd.Description = d.Content
	b.runTemplate("edit-part-description.html", pd)
}

//  _     _     _                     _           _
// | |   (_)___| |_   _ __  _ __ ___ (_) ___  ___| |_ ___
// | |   | / __| __| | '_ \| '__/ _ \| |/ _ \/ __| __/ __|
// | |___| \__ \ |_  | |_) | | | (_) | |  __/ (__| |_\__ \
// |_____|_|___/\__| | .__/|_|  \___// |\___|\___|\__|___/
//                   |_|           |__/

type listProjectPage struct {
	Projects    []bagzullaDb.Project
	OpenBugs    []int64
	DisplayDir  string
	ProjectNone int64
}

type pros []bagzullaDb.Project

func (p pros) Len() int {
	return len(p)
}

func (p pros) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p pros) Less(i, j int) bool {
	iname := strings.ToLower(p[i].Name)
	jname := strings.ToLower(p[j].Name)
	if iname == "none" {
		return true
	}
	if jname == "none" {
		return false
	}
	return iname < jname
}

func sortProjects(projects []bagzullaDb.Project) {
	sort.Sort(pros(projects))
}

type prts []bagzullaDb.Part

func (p prts) Len() int {
	return len(p)
}

func (p prts) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p prts) Less(i, j int) bool {
	iname := strings.ToLower(p[i].Name)
	jname := strings.ToLower(p[j].Name)
	if iname == "none" {
		return true
	}
	if jname == "none" {
		return false
	}
	return iname < jname
}

func sortParts(parts []bagzullaDb.Part) {
	sort.Sort(prts(parts))
}

func listProjects(b *Bagreply) {
	var lpp = listProjectPage{
		ProjectNone: ProjectNone,
		DisplayDir:  b.App.DisplayDir,
	}
	projects, err := bagzullaDb.AllProjects(b.App.db)
	if err != nil {
		b.errorPage("Error making list of pages: %s", err.Error())
		return
	}
	var openProjects []bagzullaDb.Project
	for _, p := range projects {
		if p.Status == 0 {
			openProjects = append(openProjects, p)
		}
	}
	projects = openProjects
	sortProjects(projects)
	lpp.Projects = projects
	lpp.OpenBugs, err = getOpenBugs(b, projects)
	if err != nil {
		b.errorPage("Error getting list of open bugs: %s", err.Error())
		return
	}
	b.runTemplate("list-projects.html", lpp)
}

//  ____  _                                             _           _
// / ___|| |__   _____      __   __ _   _ __  _ __ ___ (_) ___  ___| |_
// \___ \| '_ \ / _ \ \ /\ / /  / _` | | '_ \| '__/ _ \| |/ _ \/ __| __|
//  ___) | | | | (_) \ V  V /  | (_| | | |_) | | | (_) | |  __/ (__| |_
// |____/|_| |_|\___/ \_/\_/    \__,_| | .__/|_|  \___// |\___|\___|\__|
//                                     |_|           |__/

type partPage struct {
	Part        bagzullaDb.Part
	Description string
	Project     bagzullaDb.Project
	Bugs        []ListBug
	OpenOnly    bool
	User        *bagzullaDb.Person
}

func getPartInfo(b *Bagreply) (pp partPage, ok bool) {
	part, ok := getPart(b)
	if !ok {
		return pp, false
	}
	pp.Part = part
	var err error
	pp.Project, err = bagzullaDb.ProjectFromId(b.App.db, pp.Part.ProjectId)
	if err != nil {
		b.errorPage("Error retrieving project with id %d for part id %d: %s",
			pp.Part.ProjectId, part.PartId, err.Error())
		return pp, false
	}
	description, ok := getText(b, part.Description)
	if !ok {
		return pp, false
	}
	pp.Description = description.Content
	return pp, true
}

func getPartBugs(b *Bagreply, pp *partPage) (ok bool) {
	partId := pp.Part.PartId
	var bugs []bagzullaDb.Bug
	var err error
	if pp.OpenOnly {
		bugs, err = openBugsFromPartId(b.App.db, partId)
	} else {
		bugs, err = bagzullaDb.BugsFromPartId(b.App.db, partId)
	}
	if err != nil {
		b.errorPage("Error retrieving bugs with part id %d: %s",
			partId, err.Error())
		return false
	}
	for _, bug := range bugs {
		lb, ok := getBugInfo(b, bug)
		if !ok {
			return false
		}
		pp.Bugs = append(pp.Bugs, lb)
	}
	pp.User = b.User
	return true
}

func showPart(b *Bagreply) {
	pp, ok := getPartInfo(b)
	if !ok {
		return
	}
	pp.OpenOnly = true
	ok = getPartBugs(b, &pp)
	if !ok {
		return
	}
	b.runTemplate("part.html", pp)
}

func showPartAll(b *Bagreply) {
	pp, ok := getPartInfo(b)
	if !ok {
		return
	}
	pp.OpenOnly = false
	ok = getPartBugs(b, &pp)
	if !ok {
		return
	}
	b.runTemplate("part.html", pp)
}

type personPage struct {
	Person bagzullaDb.Person
	Bugs   []ListBug
}

func getPerson(b *Bagreply) (person bagzullaDb.Person, ok bool) {
	personId, ok := getFinalNum(b)
	if !ok {
		return
	}
	if personId == 0 {
		b.errorPage("Person %d does not have a description", personId)
		return person, false
	}
	person, err := bagzullaDb.PersonFromId(b.App.db, personId)
	if err != nil {
		b.errorPage("Error getting person with ID %d from database: %s", personId, err.Error())
		return person, false
	}
	return person, true
}

func showPerson(b *Bagreply) {
	person, ok := getPerson(b)
	if !ok {
		return
	}
	var pp personPage
	pp.Person = person
	var err error
	bugs, err := bagzullaDb.BugsFromOwner(b.App.db, pp.Person.PersonId)
	if err != nil {
		b.errorPage("Error retrieving bugs with person id %d: %s",
			person.PersonId, err.Error())
		return
	}
	for _, bug := range bugs {
		lb, ok := getBugInfo(b, bug)
		if !ok {
			return
		}
		pp.Bugs = append(pp.Bugs, lb)
	}
	b.runTemplate("person.html", pp)
}

func projectParts(b *Bagreply) {
	projectid, ok := getFinalNum(b)
	if !ok {
		return
	}
	parts, err := bagzullaDb.PartsFromProjectId(b.App.db, projectid)
	if err != nil {
		b.errorPage("Error getting parts for project %d", projectid)
		return
	}
	b.w.Header().Set("Content-Type", "application/json")
	var jout []byte
	if len(parts) > 0 {
		sortParts(parts)
		jout, err = json.Marshal(parts)
		if err != nil {
			b.errorPage("Error marshalling part list: %s", err)
			return
		}
	} else {
		jout = []byte("[]")
	}
	fmt.Fprintf(b.w, string(jout))
}

// Structure for putting into project page listing bugs.
type ProjectPage struct {
	Project     bagzullaDb.Project
	Description string
	Parts       []bagzullaDb.Part
	Bugs        []ListBug
	DisplayDir  string
}

func showProject(b *Bagreply) {
	projectid, ok := getFinalNum(b)
	if !ok {
		return
	}
	project, err := bagzullaDb.ProjectFromId(b.App.db, projectid)
	if err != nil {
		b.errorPage("Error finding project with ID %d: %s", projectid, err.Error())
		return
	}
	var pp ProjectPage
	pp.DisplayDir = b.App.DisplayDir
	pp.Project = project
	description, ok := b.GetText(project.Description)
	if !ok {
		return
	}
	pp.Description = b.urlsToLinks(description)
	parts, err := bagzullaDb.PartsFromProjectId(b.App.db, projectid)
	if err != nil {
		errorText := fmt.Sprintf("Error retrieving project id %d: %s",
			projectid, err.Error())
		b.errorPage(errorText)
		return
	}
	pp.Parts = parts
	sortParts(pp.Parts)
	bugs, err := openBugsFromProjectId(b.App.db, projectid)
	if err != nil {
		b.errorPage("Error getting list of bugs for project with id %d: %s", projectid, err.Error())
		return
	}
	for _, bug := range bugs {
		bp, ok := getBugInfo(b, bug)
		if !ok {
			return
		}
		pp.Bugs = append(pp.Bugs, bp)
	}
	b.Title = fmt.Sprintf("%s project bugs", project.Name)
	b.runTemplate("project.html", pp)
}

func showProjectAllBugs(b *Bagreply) {
	projectid, ok := getFinalNum(b)
	if !ok {
		return
	}
	project, err := bagzullaDb.ProjectFromId(b.App.db, projectid)
	if err != nil {
		b.errorPage("Error finding project with ID %d: %s", projectid, err.Error())
		return
	}
	var pp ProjectPage
	pp.Project = project
	description, ok := getText(b, project.Description)
	if !ok {
		return
	}
	pp.Description = description.Content
	parts, err := bagzullaDb.PartsFromProjectId(b.App.db, projectid)
	if err != nil {
		errorText := fmt.Sprintf("Error retrieving project id %d: %s",
			projectid, err.Error())
		b.errorPage(errorText)
		return
	}
	pp.Parts = parts
	sortParts(pp.Parts)
	bugs, err := bagzullaDb.BugsFromProjectId(b.App.db, projectid)
	if err != nil {
		b.errorPage("Error getting list of bugs for project with id %d: %s", projectid, err.Error())
		return
	}
	for _, bug := range bugs {
		bp, ok := getBugInfo(b, bug)
		if !ok {
			return
		}
		pp.Bugs = append(pp.Bugs, bp)
	}
	b.Title = fmt.Sprintf("All bugs for %s", project.Name)
	b.runTemplate("project-all.html", pp)
}

// Display the page for adding a new project.
func addProjectPage(b *Bagreply) {
	b.runTemplate("add-project.html", interface{}(nil))
}

// Actually add a new project.
func addNewProject(b *Bagreply) {
	var p bagzullaDb.Project
	p.Name = b.r.FormValue("name")
	p.Name = strings.TrimSpace(p.Name)
	p.Directory = b.r.FormValue("directory")
	descriptionId, ok := insertText(b, b.r.FormValue("description"))
	if !ok {
		return
	}
	p.Description = descriptionId
	projectid, err := bagzullaDb.InsertProject(b.App.db, p)
	if err != nil {
		b.errorPage("Error adding new project with name %s: %s",
			p.Name, err.Error())
	} else {
		redirectToProject(b, projectid)
	}
}

func addProjectHandler(b *Bagreply) {
	if b.NotLoggedIn() {
		return
	}
	name := b.r.FormValue("name")
	if len(name) > 0 {
		addNewProject(b)
		return
	}
	addProjectPage(b)
}

func randomOpen(b *Bagreply) {
	var bugs []bagzullaDb.Bug
	open, ok := StatusBugs(b, 0)
	if !ok {
		return
	}
	bugs = append(bugs, open...)
	if len(bugs) == 0 {
		b.errorPage("There are no open bugs. Congratulations.")
		return
	}
	randomBug := rand.Intn(len(bugs))
	b.redirectToBug(bugs[randomBug].BugId)
	return
}

func openProjects(b *Bagreply) (projects []bagzullaDb.Project, ok bool) {
	allp, ok := allProjects(b)
	if !ok {
		return projects, false
	}
	for _, p := range allp {
		if p.Status == 0 {
			projects = append(projects, p)
		}
	}
	return projects, true
}

func allProjects(b *Bagreply) (projects []bagzullaDb.Project, ok bool) {
	projects, err := bagzullaDb.AllProjects(b.App.db)
	if err != nil {
		b.errorPage("Error getting all projects: %s",
			err.Error())
		return projects, false
	}
	sortProjects(projects)
	return projects, true
}

type AddBugPage struct {
	Project     bagzullaDb.Project
	Part        bagzullaDb.Part
	Choices     []bagzullaDb.Project
	Title       string
	Description string
}

// Display the page for adding a new bug.
func addBugPage(b *Bagreply) {
	var abp AddBugPage
	var ok bool
	abp.Choices, ok = openProjects(b)
	if !ok {
		return
	}
	b.runTemplate("add-bug.html", abp)
}

// Add a bug with some of the fields filled in.  This is for the case
// that we want to link to the bug reporting page from other projects.
func addAutoBugHandler(b *Bagreply) {
	auto := b.r.FormValue("auto")
	if len(auto) > 0 {
		addNewBug(b)
		return
	}
	var abp AddBugPage
	var ok bool
	abp.Title = html.EscapeString(b.r.FormValue("title"))
	abp.Description = b.r.FormValue("description")
	projectString := b.r.FormValue("project")
	partString := b.r.FormValue("part")
	if len(partString) > 0 {
		partId, err := strconv.ParseInt(partString, 10, 64)
		part, err := bagzullaDb.PartFromId(b.App.db, partId)
		if err != nil {
			b.errorPage("Error getting part with ID %d from database: %s",
				partId, err.Error())
			return
		}
		abp.Part = part
		abp.Project, ok = projectFromId(b, part.ProjectId)
		if !ok {
			return
		}
		b.runTemplate("add-bug-to-part.html", abp)
	} else if len(projectString) > 0 {
		projectId, err := strconv.ParseInt(projectString, 10, 64)
		if err != nil {
			b.errorPage("Error getting ID from project string %s: %s", projectString, err.Error())
			return
		}
		project, err := bagzullaDb.ProjectFromId(b.App.db, projectId)
		if err != nil {
			b.errorPage("Error retrieving project with id %d from database: %s",
				projectId, err.Error())
			return
		}
		abp.Project = project
		b.runTemplate("add-bug-to-project.html", abp)
	} else {
		abp.Choices, ok = openProjects(b)
		if !ok {
			return
		}
		b.runTemplate("add-bug.html", abp)
	}
}

// Add a bug to any project.
func addBugHandler(b *Bagreply) {
	if b.NotLoggedIn() {
		return
	}
	project := b.r.FormValue("project")
	if len(project) > 0 {
		addNewBug(b)
		return
	}
	b.Title = "New bug"
	addBugPage(b)
}

func FormPart(b *Bagreply) (partId int64, ok bool) {
	part := b.r.FormValue("part")
	if len(part) > 0 {
		var err error
		partId, err = strconv.ParseInt(part, 10, 64)
		if err != nil {
			b.errorPage("Error parsing part ID %s: %s.\n", part, err)
			return partId, false
		}
	}
	return partId, true
}

func addBugToProjectHandler(b *Bagreply) {
	if b.NotLoggedIn() {
		return
	}
	project, ok := getProject(b)
	if !ok {
		return
	}
	title := b.r.FormValue("title")
	description := b.r.FormValue("description")
	partId, ok := FormPart(b)
	if !ok {
		return
	}
	if len(title) > 0 || len(description) > 0 {
		owner := b.User.PersonId
		bugid, ok := newbug(b, title, description, project.ProjectId,
			partId, owner)
		if !ok {
			return
		}
		b.redirectToBug(bugid)
		return
	}
	var abip AddBugPage
	abip.Project = project
	b.runTemplate("add-bug-to-project.html", abip)
}

func (b *Bagreply) redirectToBug(bugid int64) {
	abs := b.AbsRef(b.r.Referer())
	re := fmt.Sprintf("%s/bug/%d", abs, bugid)
	http.Redirect(b.w, b.r, re, http.StatusFound)
}

func projectFromId(b *Bagreply, projectid int64) (project bagzullaDb.Project, ok bool) {
	project, err := bagzullaDb.ProjectFromId(b.App.db, projectid)
	if err != nil {
		b.errorPage("Error retreiving project with id %d: %s",
			projectid, err.Error())
		return project, false
	}
	return project, true
}

func addBugToPartHandler(b *Bagreply) {
	if b.NotLoggedIn() {
		return
	}
	part, ok := getPart(b)
	title := b.r.FormValue("title")
	if len(title) > 0 {
		description := b.r.FormValue("description")
		owner := b.User.PersonId
		bugid, ok := newbug(b, title, description, part.ProjectId, part.PartId, owner)
		if !ok {
			return
		}
		b.redirectToBug(bugid)
		return
	}
	var projectPart AddBugPage
	projectPart.Part = part
	projectPart.Project, ok = projectFromId(b, part.ProjectId)
	if !ok {
		return
	}
	b.runTemplate("add-bug-to-part.html", projectPart)
}

func newbug(b *Bagreply, title string, description string, projectid int64, partid int64, owner int64) (bugid int64, ok bool) {
	_, ok = projectFromId(b, projectid)
	if !ok {
		return 0, false
	}
	var bug bagzullaDb.Bug
	titleId, ok := insertText(b, title)
	if !ok {
		return
	}
	bug.Title = titleId
	descriptionId, ok := insertText(b, description)
	if !ok {
		return
	}
	bug.Description = descriptionId
	bug.ProjectId = projectid
	bug.PartId = partid
	bug.Owner = owner
	bug.Entered = time.Now()
	bug.Changed = bug.Entered
	bugid, err := bagzullaDb.InsertBug(b.App.db, bug)
	if err != nil {
		b.errorPage("Error inserting bug with title %s: %s",
			title, err.Error())
		return 0, false
	}
	return bugid, true
}

func addNewBug(b *Bagreply) {
	projectString := b.r.FormValue("project")
	title := b.r.FormValue("title")
	description := b.r.FormValue("description")
	projectId, err := strconv.ParseInt(projectString, 10, 64)
	if err != nil {
		b.errorPage("addNewBug: Error getting ID from project string %s: %s", projectString, err.Error())
		return
	}
	partId, ok := FormPart(b)
	if !ok {
		return
	}
	owner := b.User.PersonId
	bugid, ok := newbug(b, title, description, projectId, partId, owner)
	if !ok {
		return
	}
	b.redirectToBug(bugid)
}

func getRelatedBugStatuses(b *Bagreply, rb []RelatedBug) bool {
	for i, _ := range rb {
		id := rb[i].Id
		bug, err := bagzullaDb.BugFromId(b.App.db, id)
		if err != nil {
			b.errorPage("Error getting bug information for bug with id %d from database: %s", id, err.Error())
			return false
		}
		rb[i].Status = bug.Status
	}
	return true
}

func getDependsOn(b *Bagreply, bugId int64) (dependsOn []RelatedBug, ok bool) {
	deps, err := bagzullaDb.DependencysFromEffect(b.App.db, bugId)
	if err != nil {
		b.errorPage("Error getting dependent bugs for bug with id %d from database: %s", bugId, err.Error())
		return dependsOn, false
	}
	for _, dep := range deps {
		dependsOn = append(dependsOn, RelatedBug{Id: dep.Cause})
	}
	ok = getRelatedBugStatuses(b, dependsOn)
	if !ok {
		return dependsOn, false
	}
	return dependsOn, true
}

func getDuplicates(b *Bagreply, bugId int64) (duplicates []RelatedBug, ok bool) {
	dups, err := bagzullaDb.DuplicatesFromOriginal(b.App.db, bugId)
	if err != nil {
		b.errorPage("Error getting dependent bugs for bug with id %d from database: %s", bugId, err.Error())
		return duplicates, false
	}
	for _, dup := range dups {
		duplicates = append(duplicates, RelatedBug{Id: dup.Duplicate})
	}
	ok = getRelatedBugStatuses(b, duplicates)
	if !ok {
		return duplicates, false
	}
	return duplicates, true
}

func getOriginals(b *Bagreply, bugId int64) (originals []RelatedBug, ok bool) {
	originals, err := OriginalsFromDuplicate(b.App.db, bugId)
	if err != nil {
		b.errorPage("Error getting dependent bugs for bug with id %d from database: %s", bugId, err.Error())
		return originals, false
	}
	ok = getRelatedBugStatuses(b, originals)
	if !ok {
		return originals, false
	}
	return originals, true
}

func getBlocks(b *Bagreply, bugId int64) (blocks []RelatedBug, ok bool) {
	deps, err := bagzullaDb.DependencysFromCause(b.App.db, bugId)
	if err != nil {
		b.errorPage("Error getting blocking bugs for bug with id %d from database: %s", bugId, err.Error())
		return blocks, false
	}
	for _, dep := range deps {
		blocks = append(blocks, RelatedBug{Id: dep.Effect})
	}
	ok = getRelatedBugStatuses(b, blocks)
	if !ok {
		return blocks, false
	}
	return blocks, true
}

func setBugStatus(b *Bagreply, newStatus int64, bugId int64) bool {
	if b.NotLoggedIn() {
		return false
	}
	err := bagzullaDb.UpdateStatusForBug(b.App.db, newStatus, bugId)
	if err != nil {
		b.errorPage(fmt.Sprintf("Error updating status for bug with id %d to status %d: %s",
			bugId, newStatus, err.Error()))
		return false
	}
	return true
}

// Allow /one/ or /one/123 or /hone/abc but not anything more.
var AbsToRel = regexp.MustCompile("(^.*?/[^/]+?)(/[^/]+((?:/[^/]*)?))$")

func (b *Bagreply) AbsRef(r string) string {
	abs := AbsToRel.FindStringSubmatch(r)
	if len(abs) > 0 {
		return abs[1]
	}
	return "/"
}

func (b *Bagreply) RelRef(r string) string {
	abs := AbsToRel.FindStringSubmatch(r)
	if len(abs) > 0 {
		r = abs[2]
	}
	return r
}

func (b *Bagreply) NotLoggedIn() bool {
	if b.User == nil {
		b.ErrorLogin()
		return true
	}
	return false
}

// Handle /bug/%d requests, including those which post new comments to
// the bug.

func bugHandler(b *Bagreply) {
	// Was there input in the form?
	changed := false
	bug, ok := getBug(b)
	if !ok {
		return
	}
	projectname := b.r.FormValue("project")
	if len(projectname) > 0 {
		if b.NotLoggedIn() {
			return
		}
		// Deal with user input.
		var projectid int64
		if projectname == "None" {
			projectid = ProjectNone
		} else {
			projectid, ok = b.projectIdFromName(projectname)
			if !ok {
				return
			}
		}
		if projectid != bug.ProjectId {
			b.assignProjectToBug(bug, projectid)
		}
		changed = true
	}

	// If the user has input a new comment, add that to the database.
	comment_text := b.r.FormValue("comment-text")
	if len(comment_text) == 0 {
		if _, ok := b.r.Form["comment-text"]; ok {
			b.errorPage("Empty comment text")
			return
		}
	}
	if len(comment_text) > 0 {
		if b.NotLoggedIn() {
			return
		}
		var comment bagzullaDb.Comment
		var ok bool
		comment.TxtId, ok = insertText(b, comment_text)
		if !ok {
			return
		}
		comment.BugId = bug.BugId
		comment.PersonId = b.User.PersonId
		bagzullaDb.InsertComment(b.App.db, comment)
		changed = true
		newStatusString := b.r.FormValue("bug-status")
		if len(newStatusString) > 0 {
			newStatus, err := stringToStatus(newStatusString)
			if err != nil {
				b.errorPage("Error with %s: %s", newStatusString, err.Error())
				return
			}
			if newStatus == 3 {
				b.errorPage("Use Edit duplicates to mark duplicates")
				return
			}
			oldStatus := bug.Status
			if oldStatus != newStatus {
				ok := setBugStatus(b, newStatus, bug.BugId)
				if !ok {
					return
				}
				if !b.updateChanged(bug.BugId) {
					return
				}

				changed = true
			}
		}
	}
	if changed {
		ok := b.updateChanged(bug.BugId)
		if !ok {
			return
		}
		b.redirectToBug(bug.BugId)
		return
	}
	bp, ok := getBugInfo(b, bug)
	if !ok {
		return
	}
	bp.DisplayDescription = b.urlsToLinks(bp.Description)
	projects, err := bagzullaDb.AllProjects(b.App.db)
	if err != nil {
		b.errorPage("Error making list of pages: %s", err.Error())
		return
	}
	sortProjects(projects)
	bp.Projects = projects
	images, err := bagzullaDb.ImagesFromBugId(b.App.db, bug.BugId)
	if err != nil {
		b.errorPage(err.Error())
		return
	}
	bp.Images = images
	comments, err := bagzullaDb.CommentsFromBugId(b.App.db, bug.BugId)
	if err != nil {
		b.errorPage(err.Error())
		return
	}
	bp.Statuses = statuses
	bp.Priorities = priorities
	for _, comment := range comments {
		var lc ListComment
		lc.Comment = comment
		var ok bool
		lc.Txt, ok = getText(b, comment.TxtId)
		if !ok {
			return
		}
		// Substitute URLs with linked URLs.
		lc.Txt.Content = b.urlsToLinks(lc.Txt.Content)

		lc.Person, ok = getPersonName(b, comment.PersonId)
		if !ok {
			return
		}
		bp.Comments = append(bp.Comments, lc)
	}
	bp.Originals, ok = getOriginals(b, bug.BugId)
	if !ok {
		return
	}
	bp.Duplicates, ok = getDuplicates(b, bug.BugId)
	if !ok {
		return
	}
	bp.DependsOn, ok = getDependsOn(b, bug.BugId)
	if !ok {
		return
	}
	bp.Blocks, ok = getBlocks(b, bug.BugId)
	if !ok {
		return
	}
	b.Title = html.EscapeString(fmt.Sprintf("%s - %s", bp.Title, bp.ProjectName))
	b.runTemplate("bug.html", bp)
}

func topHandler(b *Bagreply) {
	ru := b.r.URL.RequestURI()
	if ru == "/" {
		http.Redirect(b.w, b.r, b.App.TopURL+"/open-bugs/", http.StatusFound)
	}
}

func assignDirToProject(b *Bagreply, projectId int64, directory string) (err error) {
	return bagzullaDb.UpdateDirectoryForProject(b.App.db, directory, projectId)
}

// Assign the given part ID to the bug specified.
func assignPartToBug(b *Bagreply, bug bagzullaDb.Bug, partid int64) {
	err := bagzullaDb.UpdatePartIdForBug(b.App.db, partid, bug.BugId)
	if err != nil {
		b.errorPage("Error assigning part with id %d to bug with id %d: %s", partid, bug.BugId, err.Error())
		return
	}
	ok := b.updateChanged(bug.BugId)
	if !ok {
		return
	}
	b.redirectToBug(bug.BugId)
}

// Assign the project ID to the bug specified.
func (b *Bagreply) assignProjectToBug(bug bagzullaDb.Bug, projectid int64) {
	if b.NotLoggedIn() {
		return
	}
	err := bagzullaDb.UpdateProjectIdForBug(b.App.db, projectid, bug.BugId)
	if err != nil {
		b.errorPage("Error assigning project with id %d to bug with id %d: %s",
			projectid, bug.BugId, err.Error())
		return
	}
	// Set the part to "None".
	err = bagzullaDb.UpdatePartIdForBug(b.App.db, 0, bug.BugId)
	if err != nil {
		b.errorPage("Error resetting part for bug with id %d: %s",
			bug.BugId, err.Error())
		return
	}
	if !b.updateChanged(bug.BugId) {
		return
	}
	b.redirectToBug(bug.BugId)
}

// Given a project id and a part name, return the part id and true or
// false if found or not found.
func partIdFromName(b *Bagreply, projectId int64, partName string) (int64, bool, error) {
	parts, err := bagzullaDb.PartsFromProjectId(b.App.db, projectId)
	if err != nil {
		return 0, false, err
	}
	for _, part := range parts {
		if part.Name == partName {
			return part.PartId, true, nil
		}
	}
	return 0, false, nil
}

func (b *Bagreply) ChangeBug() (bugid int64, bug bagzullaDb.Bug, ok bool) {
	if b.NotLoggedIn() {
		return bugid, bug, false
	}
	bugid, ok = getFinalNum(b)
	if !ok {
		return bugid, bug, false
	}
	var err error
	bug, err = bagzullaDb.BugFromId(b.App.db, bugid)
	if err != nil {
		b.errorPage(fmt.Sprintf("Error looking for bug with ID %d: %s",
			bugid, err.Error()))
		return bugid, bug, false
	}
	return bugid, bug, true
}

type ChangeBugEstimate struct {
	Bug          bagzullaDb.Bug
	Title        string
	Project      bagzullaDb.Project
	ProjectParts []bagzullaDb.Part
}

func changeBugEstimate(b *Bagreply) {
	var cbe ChangeBugEstimate
	var ok bool
	var bugid int64
	bugid, cbe.Bug, ok = b.ChangeBug()
	if !ok {
		return
	}
	cbe.Title, ok = b.GetText(cbe.Bug.Title)
	if !ok {
		return
	}
	fmt.Fprintf(b.w, "Change estimate of %d %s\n", bugid, cbe.Title)
}

type ChangeBugPart struct {
	Bug          bagzullaDb.Bug
	Title        string
	Project      bagzullaDb.Project
	ProjectParts []bagzullaDb.Part
}

// Change the part of a project to which a bug belongs.
func changeBugPartHandler(b *Bagreply) {
	if b.NotLoggedIn() {
		return
	}
	var cbp ChangeBugPart
	bugid, ok := getFinalNum(b)
	if !ok {
		return
	}
	var err error
	cbp.Bug, err = bagzullaDb.BugFromId(b.App.db, bugid)
	if err != nil {
		b.errorPage(fmt.Sprintf("Error looking for bug with ID %d: %s",
			bugid, err.Error()))
		return
	}
	if cbp.Bug.ProjectId != ProjectNone {
		var ok bool
		cbp.Project, ok = projectFromId(b, cbp.Bug.ProjectId)
		if !ok {
			return
		}
	} else {
		b.errorPage("Bug id %d is not currently associated with any project; <a href='/set-bug-project/%d'>please pick a project for this bug</a>.", bugid, bugid)
		return
	}
	newPartName := b.r.FormValue("new-part")
	if len(newPartName) > 0 {
		url := fmt.Sprintf("%s/add-part-to-project/%d?part-name=%s&bug-id=%d",
			b.App.TopURL, cbp.Bug.ProjectId, newPartName, bugid)
		http.Redirect(b.w, b.r, url, http.StatusFound)
		return
	}
	title, ok := getText(b, cbp.Bug.Title)
	if !ok {
		return
	}
	cbp.Title = html.EscapeString(title.Content)
	partname := b.r.FormValue("part")
	if len(partname) > 0 {
		// Deal with user input.
		var partid int64
		if partname == "None" {
			partid = 0
		} else {
			var found bool
			partid, found, err = partIdFromName(b, cbp.Bug.ProjectId, partname)
			if !found {
				b.errorPage("No such part %s", partname)
				return
			}
			if err != nil {
				b.errorPage("Error looking for part with name %s: %s", partname, err.Error())
				return
			}
		}
		assignPartToBug(b, cbp.Bug, partid)
		return
	}
	// There was no user input, so print the form.
	cbp.ProjectParts, err = bagzullaDb.PartsFromProjectId(b.App.db, cbp.Bug.ProjectId)
	if err != nil {
		b.errorPage("Error getting parts for project %d: %s",
			cbp.Bug.ProjectId, err.Error())
		return
	}
	sortParts(cbp.ProjectParts)
	b.Title = fmt.Sprintf("Change part of %s", cbp.Title)
	b.runTemplate("change-bug-part.html", cbp)
}

type ChangeBugProject struct {
	Bug      bagzullaDb.Bug
	Title    string
	Projects []bagzullaDb.Project
}

// Change the part of a project to which a bug belongs.
func changeBugProjectHandler(b *Bagreply) {
	if b.NotLoggedIn() {
		return
	}
	var cbp ChangeBugProject
	bugid, ok := getFinalNum(b)
	if !ok {
		return
	}
	var err error
	cbp.Bug, err = bagzullaDb.BugFromId(b.App.db, bugid)
	if err != nil {
		b.errorPage(fmt.Sprintf("Error looking for bug with ID %d: %s",
			bugid, err.Error()))
		return
	}
	title, ok := getText(b, cbp.Bug.Title)
	if !ok {
		return
	}
	cbp.Title = html.EscapeString(title.Content)
	projectname := b.r.FormValue("project")
	if len(projectname) > 0 {
		// Deal with user input.
		var projectid int64
		if projectname == "None" {
			projectid = ProjectNone
		} else {
			projectid, ok = b.projectIdFromName(projectname)
			if !ok {
				return
			}
		}
		b.assignProjectToBug(cbp.Bug, projectid)
		return
	}
	// There was no user input, so print the form.
	cbp.Projects, err = bagzullaDb.AllProjects(b.App.db)
	if err != nil {
		b.errorPage("Error getting all projects: %s", err.Error())
		return
	}
	sortProjects(cbp.Projects)
	b.runTemplate("change-bug-project.html", cbp)
}

// Add a part to a project based on the form input.
func addPartToProjectHandler(b *Bagreply) {
	if b.NotLoggedIn() {
		return
	}
	project, ok := getProject(b)
	if !ok {
		return
	}
	if project.ProjectId == ProjectNone {
		b.errorPage("Can't add a part to project 'None': choose a project first")
		return
	}
	if len(b.r.FormValue("name")) > 0 {
		var p bagzullaDb.Part
		p.Name = b.r.FormValue("name")
		if strings.EqualFold(p.Name, "none") {
			b.errorPage("Part cannot be called 'none'")
			return
		}
		parts, err := bagzullaDb.PartsFromProjectId(b.App.db, project.ProjectId)
		if err != nil {
			b.errorPage("Error retrieving existing parts: %s", err)
			return
		}
		for _, q := range parts {
			if strings.EqualFold(p.Name, q.Name) {
				b.errorPage("There is already a part '%s'", q.Name)
				return
			}
		}
		description := b.r.FormValue("description")
		descriptionId, ok := insertText(b, description)
		if !ok {
			return
		}
		p.Description = descriptionId
		p.ProjectId = project.ProjectId
		p.PartId, err = bagzullaDb.InsertPart(b.App.db, p)
		if err != nil {
			b.errorPage(fmt.Sprintf("Error creating part %s for project with id %d: %s",
				p.Name, project.ProjectId, err.Error()))
		} else {
			bugIdStr := b.r.FormValue("bug-id")
			if len(bugIdStr) > 0 {
				bugId, err := strconv.ParseInt(bugIdStr, 10, 64)
				if err != nil {
					b.errorPage("Error getting bug id number from %s",
						bugIdStr)
					return
				}
				if bugId != 0 {
					err = bagzullaDb.UpdatePartIdForBug(b.App.db, p.PartId, bugId)
					if err != nil {
						b.errorPage("Error adding part id %d for bug with id %d",
							p.PartId, bugId)
						return
					}
					ok := b.updateChanged(bugId)
					if !ok {
						return
					}
				}
			}
			redirectToPart(b, p.PartId)
		}
		return
	}
	var addPart struct {
		PartName string
		BugId    int64
		Project  bagzullaDb.Project
	}
	addPart.Project = project
	addPart.PartName = b.r.FormValue("part-name")
	bugId := b.r.FormValue("bug-id")
	if len(bugId) > 0 {
		var err error
		addPart.BugId, err = strconv.ParseInt(bugId, 10, 64)
		if err != nil {
			b.errorPage("Error getting bug id number from %s", bugId)
			return
		}
		if addPart.BugId == 0 {
			b.errorPage("Bug ID for initial bug cannot be zero")
			return
		}
	}
	b.runTemplate("add-part-to-project.html", addPart)
}

// Change the directory associated with the project specified in the
// URL.
func changeProjectDirectory(b *Bagreply) {
	if b.NotLoggedIn() {
		return
	}
	// Get the project information.
	project, ok := getProject(b)
	if !ok {
		return
	}
	// Get the user's requested directory.
	dir := b.r.FormValue("dir")
	if len(dir) > 0 {
		// Respond to user input.
		projectid := project.ProjectId
		err := bagzullaDb.UpdateDirectoryForProject(b.App.db, dir, projectid)
		if err != nil {
			b.errorPage(fmt.Sprintf("Error updating directory for project %d to %s: %s",
				projectid, dir, err.Error()))
			return
		}
		// Send user back to project page with the updated directory.
		redirectToProject(b, projectid)
		return
	}
	// If the user has no requested directory, print the form.
	b.runTemplate("change-project-directory.html", project)
}

type bugStatusPage struct {
	Bug      ListBug
	Statuses []string
}

type bugPriorityPage struct {
	Bug        ListBug
	Priorities []string
}

func changeBugStatus(b *Bagreply) {
	if b.NotLoggedIn() {
		return
	}
	bug, ok := getBug(b)
	if !ok {
		return
	}
	newStatusString := b.r.FormValue("status")
	if len(newStatusString) > 0 {
		newStatus, err := stringToStatus(newStatusString)
		if err != nil {
			b.errorPage("Error with %s: %s", newStatusString, err.Error())
			return
		}
		oldStatus := bug.Status
		if oldStatus != newStatus {
			ok := setBugStatus(b, newStatus, bug.BugId)
			if !ok {
				return
			}
			if !b.updateChanged(bug.BugId) {
				return
			}
			b.redirectToBug(bug.BugId)
		}
	}
	var bsp bugStatusPage
	bsp.Bug, ok = getBugInfo(b, bug)
	if !ok {
		return
	}
	bsp.Bug.Status = statuses[bug.Status]
	bsp.Statuses = statuses
	b.runTemplate("change-bug-status.html", bsp)
}

func changeBugPriority(b *Bagreply) {
	if b.NotLoggedIn() {
		return
	}
	bug, ok := getBug(b)
	if !ok {
		return
	}
	newPriorityString := b.r.FormValue("priority")
	if len(newPriorityString) > 0 {
		newPriority, err := stringToPriority(newPriorityString)
		if err != nil {
			b.errorPage("Error with %s: %s", newPriorityString, err.Error())
			return
		}
		oldPriority := bug.Priority
		if oldPriority != newPriority {
			err := bagzullaDb.UpdatePriorityForBug(b.App.db, newPriority, bug.BugId)
			if err != nil {
				b.errorPage(fmt.Sprintf("Error updating priority for bug with id %d to priority %s (%d): %s",
					bug.BugId, newPriorityString, newPriority, err.Error()))
				return
			}
			if !b.updateChanged(bug.BugId) {
				return
			}
			b.redirectToBug(bug.BugId)
		}
	}
	var bsp bugPriorityPage
	bsp.Bug, ok = getBugInfo(b, bug)
	if !ok {
		return
	}
	bsp.Bug.Priority = priorities[bug.Priority]
	bsp.Priorities = priorities
	b.runTemplate("change-bug-priority.html", bsp)
}

// Partner with edit to save a changed bug

func save(b *Bagreply) {
	if b.NotLoggedIn() {
		return
	}
	bug, ok := getBug(b)
	if !ok {
		return
	}
	var lb ListBug
	lb, ok = getBugInfo(b, bug)
	if !ok {
		return
	}
	newTitle := b.r.FormValue("title")
	var changed bool
	if newTitle != lb.Title {
		newTitleId, ok := insertText(b, newTitle)
		if !ok {
			return
		}
		bagzullaDb.UpdateTitleForBug(b.App.db, newTitleId, bug.BugId)
		changed = true
	}
	if changed {
		ok := b.updateChanged(bug.BugId)
		if !ok {
			return
		}
	}
	b.redirectToBug(bug.BugId)
}

// Edit the title of a bug

func edit(b *Bagreply) {
	if b.NotLoggedIn() {
		return
	}
	bug, ok := getBug(b)
	if !ok {
		return
	}
	var lb ListBug
	lb, ok = getBugInfo(b, bug)
	b.runTemplate("edit.html", lb)
}

type searchResult struct {
	SearchTerm string
	Ids        []text
	Bugs       ListBug
}

func search(b *Bagreply) {
	var s searchResult
	s.SearchTerm = b.r.FormValue("searchterm")
	if len(s.SearchTerm) > 0 {
		var ok bool
		s.Ids, ok = searchText(b, s.SearchTerm)
		if !ok {
			return
		}
	}
	for i, c := range s.Ids {
		c.Content = strings.Replace(c.Content, "&", "&amp;", -1)
		c.Content = strings.Replace(c.Content, "<", "&lt;", -1)
		c.Content = strings.Replace(c.Content, ">", "&gt;", -1)
		s.Ids[i] = c
	}
	b.runTemplate("search.html", s)
}

// Show recent changes in the bug tracker.
func recent(b *Bagreply) {
	// Default to twenty recent bugs.
	var max int64 = 20
	// Change the maximum based on the final number in the URL if
	// there is one.
	m := finalNum.FindStringSubmatch(b.r.URL.Path)
	if m != nil {
		var err error
		max, err = strconv.ParseInt(m[1], 10, 64)
		if err != nil {
			b.errorPage("Could not parse number in %s", b.r.URL.Path)
			return
		}
	}
	bugs, ok := BugsByChange(b, max)
	if !ok {
		return
	}
	var p ListBugPage
	ok = getBugsInfo(b, &p, bugs)
	if !ok {
		return
	}
	p.Title = "Recently changed bugs"
	b.Title = p.Title
	b.runTemplate("bugs.html", p)
}

// Find a comment based on the number.
func findComment(b *Bagreply) (comment bagzullaDb.Comment, ok bool) {
	commentid, ok := getFinalNum(b)
	if !ok {
		return
	}
	comment, err := bagzullaDb.CommentFromId(b.App.db, commentid)
	if err != nil {
		b.errorPage(fmt.Sprintf("Error retrieving comment with id %d from database: %s",
			commentid, err.Error()))
		return comment, false
	}
	return comment, true
}

type commentToEdit struct {
	Id   int64
	Text string
}

// Edit a comment.
func editComment(b *Bagreply) {
	if b.NotLoggedIn() {
		return
	}
	comment, ok := findComment(b)
	if !ok {
		return
	}
	text, ok := getText(b, comment.TxtId)
	if !ok {
		return
	}
	commentText := b.r.FormValue("comment-text")
	if len(commentText) > 0 {
		if commentText != text.Content {
			commentTextId, ok := insertText(b, commentText)
			if !ok {
				return
			}
			ok = UpdateCommentTextId(b, comment.CommentId, commentTextId)
			if !ok {
				return
			}
		}
		b.redirectToBug(comment.BugId)
	}
	var c commentToEdit
	c.Id = comment.CommentId
	c.Text = text.Content
	b.runTemplate("edit-comment.html", c)
}

// Edit bugs which are caused by (blocked by) this bug.
func editDependencies(b *Bagreply) {
	if b.NotLoggedIn() {
		return
	}
	bug, ok := getBug(b)
	if !ok {
		return
	}
	var lb ListBug
	lb, ok = getBugInfo(b, bug)
	if !ok {
		return
	}
	currentBlocks, ok := getBlocks(b, bug.BugId)
	if !ok {
		return
	}
	lb.Blocks = currentBlocks
	currentDependsOn, ok := getDependsOn(b, bug.BugId)
	if !ok {
		return
	}
	lb.DependsOn = currentDependsOn
	// Has anything changed?
	changed := false
	blocks := b.r.FormValue("blocks")
	if len(blocks) > 0 {
		// Get the blocks of the bugs by splitting into numbers.
		numbers := strings.Fields(blocks)
		var blocks []int64
		for i, nStr := range numbers {
			block, err := strconv.ParseInt(nStr, 10, 64)
			if err != nil {
				b.errorPage(fmt.Sprintf("Entry %d (%s) is not a number", i, nStr))
				return
			}
			if block == bug.BugId {
				b.errorPage(fmt.Sprintf("Bug cannot block itself"))
				return
			}
			known := false
			for _, c := range currentBlocks {
				if c.Id == block {
					known = true
				}
			}
			if !known {
				var d bagzullaDb.Dependency
				d.Cause = bug.BugId
				d.Effect = block
				bagzullaDb.InsertDependency(b.App.db, d)
				changed = true
			}
			blocks = append(blocks, block)
		}
		for _, c := range currentBlocks {
			exists := false
			for _, e := range blocks {
				if c.Id == e {
					exists = true
					break
				}
			}
			if !exists {
				// delete it
			}
		}
	}
	dependsOn := b.r.FormValue("depends-on")
	if len(dependsOn) > 0 {
		// Get the blocks of the bugs by splitting into numbers.
		numbers := strings.Fields(dependsOn)
		var dependsOns []int64
		for i, nStr := range numbers {
			block, err := strconv.ParseInt(nStr, 10, 64)
			if err != nil {
				b.errorPage(fmt.Sprintf("Entry %d (%s) is not a number", i, nStr))
				return
			}
			if block == bug.BugId {
				b.errorPage(fmt.Sprintf("Bug cannot depend on itself"))
				return
			}
			known := false
			for _, c := range currentDependsOn {
				if c.Id == block {
					known = true
				}
			}
			if !known {
				var d bagzullaDb.Dependency
				d.Cause = block
				d.Effect = bug.BugId
				bagzullaDb.InsertDependency(b.App.db, d)
				changed = true
			}
			dependsOns = append(dependsOns, block)
		}
		for _, c := range currentDependsOn {
			exists := false
			for _, e := range dependsOns {
				if c.Id == e {
					exists = true
					break
				}
			}
			if !exists {
				// delete it
			}
		}
	}
	if changed {
		ok := b.updateChanged(bug.BugId)
		if !ok {
			return
		}
		b.redirectToBug(bug.BugId)
		return
	}
	// Print a page where we can enter bugs which are blocked by this
	// bug.
	b.runTemplate("edit-dependencies.html", lb)
}

// Edit bugs which are caused by (blocked by) this bug.

func editDuplicates(b *Bagreply) {
	if b.NotLoggedIn() {
		return
	}
	bug, ok := getBug(b)
	if !ok {
		return
	}
	var lb ListBug
	lb, ok = getBugInfo(b, bug)
	if !ok {
		return
	}
	currentOriginals, ok := getOriginals(b, bug.BugId)
	if !ok {
		return
	}
	lb.Originals = currentOriginals
	currentDuplicates, ok := getDuplicates(b, bug.BugId)
	if !ok {
		return
	}
	lb.Duplicates = currentDuplicates
	// Has anything changed?
	changed := false
	// This is bogus, there should only be one original!
	originals := b.r.FormValue("originals")
	if len(originals) > 0 {
		// Get the originals of the bugs by splitting into numbers.
		numbers := strings.Fields(originals)
		if len(numbers) > 1 {
			b.errorPage(fmt.Sprintf("Too many originals for %d, can only have one", bug.BugId))
			return
		}
		var originals []int64
		for i, nStr := range numbers {
			original, err := strconv.ParseInt(nStr, 10, 64)
			if err != nil {
				b.errorPage(fmt.Sprintf("Entry %d (%s) is not a number", i, nStr))
				return
			}
			if original == bug.BugId {
				b.errorPage(fmt.Sprintf("%d is the current bug entry, cannot be a duplicate of itself", original))
			}
			known := false
			for _, c := range currentOriginals {
				if c.Id == original {
					known = true
				}
			}
			if !known {
				var d bagzullaDb.Duplicate
				d.Duplicate = bug.BugId
				d.Original = original
				bagzullaDb.InsertDuplicate(b.App.db, d)
				ok := setBugStatus(b, 3, bug.BugId)
				if !ok {
					return
				}
				changed = true
			}
			originals = append(originals, original)
		}
		for _, c := range currentOriginals {
			exists := false
			for _, e := range originals {
				if c.Id == e {
					exists = true
					break
				}
			}
			if !exists {
				removeDuplicateOrig(b.App.db, c.Id)
			}
		}
	}
	duplicates := b.r.FormValue("duplicates")
	if len(duplicates) > 0 {
		// Get the blocks of the bugs by splitting into numbers.
		numbers := strings.Fields(duplicates)
		var duplicates []int64
		for i, nStr := range numbers {
			duplicate, err := strconv.ParseInt(nStr, 10, 64)
			if err != nil {
				b.errorPage(fmt.Sprintf("Entry %d (%s) is not a number", i, nStr))
				return
			}
			known := false
			for _, c := range currentDuplicates {
				if c.Id == duplicate {
					known = true
				}
			}
			if !known {
				var d bagzullaDb.Duplicate
				d.Original = bug.BugId
				d.Duplicate = duplicate
				bagzullaDb.InsertDuplicate(b.App.db, d)
				setBugStatus(b, 3, d.Duplicate)
				changed = true
			}
			duplicates = append(duplicates, duplicate)
		}
		for _, c := range currentDuplicates {
			exists := false
			for _, e := range duplicates {
				if c.Id == e {
					exists = true
					break
				}
			}
			if !exists {
				removeDuplicate(b.App.db, c.Id)
			}
		}
	}
	if changed {
		ok := b.updateChanged(bug.BugId)
		if !ok {
			return
		}
		b.redirectToBug(bug.BugId)
		return
	}
	// Print a page where we can enter bugs which are blocked by this
	// bug.
	b.runTemplate("edit-duplicates.html", lb)
}

func Bin() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatalf("Error getting path of binary: %s\n", err)
	}
	return dir
}

var topDir string

// The uploaded image files go into this directory.

var bugimages = "bugimages"
var fileDir string = topDir + "/" + bugimages

var ntou = regexp.MustCompile(fileDir + `/(.*)`)

var filename = regexp.MustCompile("/([^/]+)$")

func fileNameToUrl(name string) (url string) {
	url = ntou.ReplaceAllString(name, "/image/$1")
	return url
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	h := w.Header()
	file := filename.FindStringSubmatch(r.URL.Path)
	if file == nil {
		fmt.Fprintf(w, "File is not in URL path")
		return
	}
	fileBytes, err := ioutil.ReadFile(fileDir + "/" + file[1])
	if err != nil {
		fmt.Fprintf(w, "Error reading file: %s.\n", err)
		return
	}
	mime_type := http.DetectContentType(fileBytes)
	if strings.Contains(mime_type, "text/xml") {
		mime_type = "image/svg+xml; charset=utf-8"
	}
	h.Set("Content-Type", mime_type)
	w.Write(fileBytes)
}

func upload(b *Bagreply) {
	if b.NotLoggedIn() {
		return
	}
	r := b.r
	r.ParseMultipartForm(10 << 20)
	file, _, err := r.FormFile("bug-image")
	if err != nil {
		b.errorPage("Error retrieving file from form-data: %s.\n", err)
		return
	}
	defer file.Close()
	tempFile, err := ioutil.TempFile(fileDir, "upload*")
	if err != nil {
		b.errorPage("Error creating temp file: %s.\n", err)
		return
	}
	defer tempFile.Close()
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		b.errorPage("Error reading file: %s.\n", err)
		return
	}
	tempFile.Write(fileBytes)

	// Insert the image into the database.

	bugid := r.FormValue("bug-id")
	if bugid == "" {
		b.errorPage("Could not get bug ID from form inputs")
		return
	}
	bugnum, err := strconv.Atoi(bugid)
	if err != nil {
		b.errorPage("Error parsing bug ID %s: %s.\n", bugid, err)
		return
	}
	name := tempFile.Name()
	name = ntou.ReplaceAllString(name, "$1")
	personId := int64(0)
	if b.User != nil {
		personId = b.User.PersonId
	}
	var image = bagzullaDb.Image{
		File:     name,
		BugId:    int64(bugnum),
		PersonId: personId,
	}
	_, err = bagzullaDb.InsertImage(b.App.db, image)
	if err != nil {
		b.errorPage("Error inserting image into database: %s.\n", err)
		return
	}
	b.redirectToBug(int64(bugnum))
}

func deleteImage(b *Bagreply) {
	if b.NotLoggedIn() {
		return
	}
	r := b.r
	file := filename.FindStringSubmatch(r.URL.Path)
	if file == nil {
		b.errorPage("File to delete not supplied")
	}
	removeImage(b.App.db, file[1])
	err := os.Remove(fileDir + "/" + file[1])
	if err != nil {
		b.errorPage("Error removing %s: %s", err, file[1])
	}
	b.errorPage("File %s removed", file[1])
}

func getId(b *Bagreply, s string) (r int64) {
	r, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		b.errorPage("Error converting %s: %s", s, err)
	}
	return r
}

func deleteDependency(b *Bagreply) {
	if b.NotLoggedIn() {
		return
	}
	cause := b.r.FormValue("cause")
	effect := b.r.FormValue("effect")
	bug := b.r.FormValue("bug")
	if len(cause) > 0 && len(effect) > 0 {
		causeId := getId(b, cause)
		effectId := getId(b, effect)
		removeDependencyCause(b.App.db, causeId, effectId)
	}
	if len(bug) > 0 {
		bugId := getId(b, bug)
		if bugId != 0 {
			b.redirectToBug(bugId)
		}
	} else {
		b.errorPage("No bug specified")
	}
}

// Set this bug's status to "open" (0) again, after removing its
// "duplicate" status. This is a helper for deleteDuplicate.
func openBug(b *Bagreply, bugId int64) {
	if bugId != 0 {
		ok := setBugStatus(b, 0, bugId)
		if !ok {
			return
		}
	}
}

// Delete a duplicate from the database.
func deleteDuplicate(b *Bagreply) {
	if b.NotLoggedIn() {
		return
	}
	original := b.r.FormValue("original")
	duplicate := b.r.FormValue("duplicate")
	bug := b.r.FormValue("bug")
	bugId := int64(0)
	if len(bug) > 0 {
		bugId = getId(b, bug)
	}
	if len(original) > 0 {
		originalId := getId(b, original)
		removeDuplicateOrig(b.App.db, originalId)
		openBug(b, bugId)
	}
	if len(duplicate) > 0 {
		duplicateId := getId(b, duplicate)
		removeDuplicate(b.App.db, duplicateId)
		openBug(b, duplicateId)
	}
	if bugId != 0 {
		b.redirectToBug(bugId)
	} else {
		b.errorPage("No bug specified")
	}
}

type BagControl struct {
	b        *Bagreply
	Stopping bool
}

func controls(b *Bagreply) {
	cb := BagControl{
		b: b,
	}
	stop := b.r.FormValue("stop")
	if len(stop) > 0 && stop != "0" {
		cb.Stopping = true
		b.App.Cancel()
	}
	b.runTemplate("control.html", &cb)
}

// makeHandler makes a handler which responds to HTTP requests out of
// a BagFunc, "fn".  A BagFunc takes one argument, (b *Bagreply).
func makeHandler(ba *Bagapp, fn BagFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b := Bagreply{
			App: ba,
			w:   w,
			r:   r,
		}
		user, found, ok := b.getSession()
		if !ok {
			return
		}
		if found && user.PersonId != 0 {
			if debugLogin {
				log.Printf("User %s found\n", user.Name)
			}
			b.User = &user
		} else {
			if debugLogin {
				log.Printf("User not found\n")
			}
			b.User = nil
		}
		b.Title = "Bagzulla"
		h := w.Header()
		h.Set("Content-Type", "text/html")
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			// The client cannot accept it, so return the output
			// uncompressed.
			h.Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(w)
			defer gz.Close()
			b.w = gzipResponseWriter{Writer: gz, ResponseWriter: w}
		}
		fn(&b)
	}
}

func (b *Bagapp) Init() {
	defaultPort := "8000"
	defaultURL := "localhost"
	defaultDisplayDir := ""
	var err error
	topDir, err = filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatalf("Error getting directory: %s", err)
	}
	defaultDatabase := topDir + "/bagzulla.db"
	portPtr := flag.String("port", defaultPort, "port to listen on")
	database := flag.String("database", defaultDatabase, "database file to use")
	url := flag.String("url", defaultURL, "URL")
	display := flag.String("display", defaultDisplayDir, "Application to display directory contents")
	flag.Parse()
	b.port = *portPtr
	b.db, err = sql.Open("sqlite3", *database)
	if err != nil {
		log.Fatalf("Error connecting to database: %s", err)
	}
	b.TopURL = *url
	b.DisplayDir = *display
	loginFile := false
	if loginFile {
		s := store.Store{}
		s.Init(topDir)
		if debugLogin {
			s.Verbose = true
		}
		b.store = &s
	} else {
		bu := baguser{
			b: b,
		}
		b.store = &bu
	}
	b.login.Init(b.store, cookieName, cookiePath)
	if debugLogin {
		b.login.Verbose = true
	}
	b.templates = template.New("bagzulla")
	customFunctions := template.FuncMap{"GetArray": GetArray}
	b.templates.Funcs(customFunctions)
	tmplDir := topDir + "/tmpl/"
	template.Must(b.templates.ParseGlob(tmplDir + "*.html"))
	b.Context, b.Cancel = context.WithCancel(context.Background())
	b.Server = &http.Server{Addr: ":" + b.port}
}

type hand struct {
	path   string
	handle func(b *Bagreply)
}

var hands = []hand{
	{"/", topHandler},
	{"/add-auto-bug/", addAutoBugHandler},
	{"/add-bug-to-part/", addBugToPartHandler},
	{"/add-bug-to-project/", addBugToProjectHandler},
	{"/add-bug/", addBugHandler},
	{"/add-part-to-project/", addPartToProjectHandler},
	{"/add-project/", addProjectHandler},
	{"/bug/", bugHandler},
	{"/bugs/", allBugsHandler},
	{"/change-bug-estimate/", changeBugEstimate},
	{"/change-bug-part/", changeBugPartHandler},
	{"/change-bug-priority/", changeBugPriority},
	{"/change-bug-project/", changeBugProjectHandler},
	{"/change-bug-status/", changeBugStatus},
	{"/change-project-directory/", changeProjectDirectory},
	{"/controls/", controls},
	{"/delete-dependency/", deleteDependency},
	{"/delete-duplicate/", deleteDuplicate},
	{"/delete-image/", deleteImage},
	{"/delete-part/", deletePart},
	{"/edit-bug-description/", editBugDescription},
	{"/edit-comment/", editComment},
	{"/edit-dependencies/", editDependencies},
	{"/edit-duplicates/", editDuplicates},
	{"/edit-part-description/", editPartDescription},
	{"/edit-part-name/", editPartName},
	{"/edit-project-description/", editProjectDescription},
	{"/edit-project-name/", editProjectName},
	{"/edit/", edit},
	{"/login/", loginHandler},
	{"/logout/", logoutHandler},
	{"/open-bugs/", openBugsHandler},
	{"/part-all/", showPartAll},
	{"/part/", showPart},
	{"/person/", showPerson},
	{"/project-all/", showProjectAllBugs},
	{"/project-parts/", projectParts},
	{"/project/", showProject},
	{"/projects/", listProjects},
	{"/random-open/", randomOpen},
	{"/recent/", recent},
	{"/save/", save},
	{"/search/", search},
	{"/upload/", upload},
}

var debugLogin = false

func main() {
	var b Bagapp
	b.Init()
	defer b.db.Close()
	for _, h := range hands {
		http.HandleFunc(h.path, makeHandler(&b, h.handle))
	}
	// This does not serve gzip content or text/html content, so it
	// does not use the "makeHandler" subroutine.
	http.HandleFunc("/image/", imageHandler)
	http.Handle("/static/", http.FileServer(http.Dir(topDir)))
	go func() {
		err := b.Server.ListenAndServe()
		switch err {
		case http.ErrServerClosed:
			log.Printf("The server has shut down.\n")
		default:
			log.Printf("Error from server: %s\n", err)
		}
	}()
	<-b.Context.Done()
	err := b.Server.Shutdown(context.Background())
	if err != nil {
		log.Printf("Error shutting down server: %s\n", err)
	}
}
