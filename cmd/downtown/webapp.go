package main

import (
	"downtown.zigdev.com/ui"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
)

type WebApp struct {
	App       *App
	Templates map[string]*template.Template
}

func NewWebApp(app *App) *WebApp {
	return &WebApp{
		App: app,
	}
}

func (a *WebApp) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", a.home)
	mux.HandleFunc("GET /login", a.loginPage)
	mux.HandleFunc("POST /login", a.login)
	mux.HandleFunc("GET /tasks", a.tasks)
	return mux
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
	ts := a.Templates["login.html"]
	err := ts.ExecuteTemplate(w, "base.html", nil)
	if err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
	}
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

//func (a *WebApp) requireAuthentication(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		_, err := r.Cookie("sid")
//		if err != nil {
//			http.Redirect(w, r, "/login", http.StatusFound)
//			return
//		}
//		w.Header().Add("Cache-Control", "no-store")
//		next.ServeHTTP(w, r)
//	})
//}

func (a *WebApp) tasks(w http.ResponseWriter, r *http.Request) {
	sid := requireSid(w, r)
	if sid == "" {
		return
	}

	var tasksResponse Response[TasksData]
	err := a.App.Client.GetTasks(r.Context(), sid, &tasksResponse)
	if err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
	ts := a.Templates["tasks.html"]
	err = ts.ExecuteTemplate(w, "base.html", tasksResponse.Data)
	if err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
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

func LoadTemplates() map[string]*template.Template {
	pages, err := fs.Glob(ui.Files, "html/pages/*.html")
	cache := make(map[string]*template.Template, len(pages))
	if err != nil {
		panic("Can't load templates: " + err.Error())
	}
	for _, page := range pages {
		name := filepath.Base(page)
		patterns := []string{
			"html/base.html",
			page,
		}
		ts, err := template.New(name).ParseFS(ui.Files, patterns...)
		if err != nil {
			panic("Can't load templates: " + err.Error())
		}
		cache[name] = ts
	}
	return cache
}
