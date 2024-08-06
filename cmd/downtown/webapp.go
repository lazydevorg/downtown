package main

import (
	"downtown.zigdev.com/ui"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
)

type TemplateCache map[string]*template.Template

var templateFunctions = template.FuncMap{
	"humanSize": HumanizeSize,
}

type WebApp struct {
	App       *App
	Templates TemplateCache
}

func (a *WebApp) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", a.home)
	mux.HandleFunc("GET /login", a.loginPage)
	mux.HandleFunc("POST /login", a.login)
	mux.HandleFunc("GET /tasks", a.tasks)
	mux.HandleFunc("/", a.notFound)
	return mux
}

func (a *WebApp) renderTemplate(w http.ResponseWriter, name string, data any) {
	ts := a.Templates[name]
	err := ts.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		panic("error rendering template: " + err.Error())
	}
}

func (a *WebApp) renderError(w http.ResponseWriter, serverError error) {
	w.WriteHeader(http.StatusInternalServerError)
	ts := a.Templates["error.html"]
	err := ts.ExecuteTemplate(w, "base.html", serverError)
	if err != nil {
		panic("error rendering template: " + err.Error())
	}
}

func (a *WebApp) home(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("sid")
	w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
	} else {
		http.Redirect(w, r, "/tasks", http.StatusFound)
	}
}

func (a *WebApp) loginPage(w http.ResponseWriter, _ *http.Request) {
	a.renderTemplate(w, "login.html", nil)
}

func (a *WebApp) login(w http.ResponseWriter, r *http.Request) {
	user := r.FormValue("user")
	pass := r.FormValue("pass")
	response, err := a.App.Client.Login(r.Context(), LoginRequest{
		user: user,
		pass: pass,
	})
	if err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "sid",
		Value:  response.Data.SID,
		Secure: true,
	})
	http.Redirect(w, r, "/", http.StatusFound)
}

func (a *WebApp) tasks(w http.ResponseWriter, r *http.Request) {
	sid := requireSid(w, r)
	if sid == "" {
		return
	}

	var tasksResponse Response[TasksData]
	err := a.App.Client.GetTasks(r.Context(), sid, &tasksResponse)
	if err != nil {
		a.renderError(w, err)
		return
	}

	w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
	a.renderTemplate(w, "tasks.html", tasksResponse.Data)
}

func (a *WebApp) notFound(w http.ResponseWriter, r *http.Request) {
	err := fmt.Errorf("Page %s not found", r.URL.Path)
	w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
	a.renderError(w, err)
}

func requireSid(w http.ResponseWriter, r *http.Request) string {
	sidCookie, err := r.Cookie("sid")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		w.(http.Flusher).Flush()
		return ""
	}
	return sidCookie.Value
}

func LoadTemplates() TemplateCache {
	pages, err := fs.Glob(ui.Files, "html/pages/*.html")
	if err != nil {
		panic("Can't load templates: " + err.Error())
	}
	cache := make(TemplateCache, len(pages))

	for _, page := range pages {
		name := filepath.Base(page)
		patterns := []string{
			"html/base.html",
			page,
		}
		ts, err := template.New(name).Funcs(templateFunctions).ParseFS(ui.Files, patterns...)
		if err != nil {
			panic("Can't load templates: " + err.Error())
		}
		cache[name] = ts
	}
	return cache
}
