package main

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/lazydevorg/downtown/ui"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
)

type TemplateCache map[string]*template.Template

var templateFunctions = template.FuncMap{
	"humanSize":          HumanizeSize,
	"progressPercentage": ProgressPercentage,
}

type WebApp struct {
	App       *App
	Logger    *slog.Logger
	Templates TemplateCache
}

func (a *WebApp) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(a.logRequests)

	r.Get("/", a.home)
	r.Get("/login", a.loginPage)
	r.Post("/login", a.login)
	r.Get("/logout", a.logout)
	r.Get("/up", a.health)

	r.Group(func(r chi.Router) {
		r.Use(authenticated)
		r.Get("/tasks", a.tasks)
		r.Post("/tasks", a.newTask)
		r.Delete("/tasks/{id}", a.deleteTask)
		r.Put("/tasks/{id}/pause", a.pauseTask)
		r.Put("/tasks/{id}/resume", a.resumeTask)
	})

	r.NotFound(a.notFound)

	return r
}

func (a *WebApp) renderTemplate(w http.ResponseWriter, name string, data any) {
	ts := a.Templates[name]
	err := ts.ExecuteTemplate(w, name, data)
	if err != nil {
		panic("error rendering template: " + err.Error())
	}
}

func (a *WebApp) renderError(w http.ResponseWriter, serverError error) {
	w.WriteHeader(http.StatusInternalServerError)
	ts := a.Templates["error.html"]
	err := ts.ExecuteTemplate(w, "error.html", serverError)
	if err != nil {
		panic("error rendering template: " + err.Error())
	}
}

func (a *WebApp) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.Logger.Debug("request received", "method", r.Method, "uri", r.URL.Path)
		next.ServeHTTP(w, r)
	})
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
		a.Logger.Error("login error", "error", err)
		a.renderError(w, err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:  "sid",
		Value: response.Data.SID,
	})
	http.Redirect(w, r, "/", http.StatusFound)
}

func (a *WebApp) logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
	http.SetCookie(w, &http.Cookie{
		Name:   "sid",
		Value:  "",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/", http.StatusFound)
}

func (a *WebApp) tasks(w http.ResponseWriter, r *http.Request) {
	sid := r.Context().Value("sid").(string)
	var tasksResponse Response[TasksData]
	err := a.App.Client.GetTasks(r.Context(), sid, &tasksResponse)
	if err != nil {
		a.Logger.Error("tasks error", "error", err)
		a.renderError(w, err)
		return
	}

	w.Header().Add("Cache-Control", "max-age=5")
	//w.WriteHeader(401)
	a.renderTemplate(w, "tasks.html", tasksResponse.Data)
}

func (a *WebApp) newTask(w http.ResponseWriter, r *http.Request) {
	sid := r.Context().Value("sid").(string)
	url := r.FormValue("url")
	response, err := a.App.Client.CreateTask(r.Context(), sid, TaskCreateRequest{Uri: url})
	if err != nil {
		a.Logger.Error("new task error", "error", err)
		a.renderError(w, err)
		return
	}
	if !response.Success {
		a.Logger.Error("new task error", "code", response.Error.Code)
		a.renderError(w, fmt.Errorf("new task error: %d", response.Error.Code))
		return
	}
	http.Redirect(w, r, "/tasks", http.StatusFound)
}

func (a *WebApp) deleteTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sid := r.Context().Value("sid").(string)
	err := a.App.Client.DeleteTask(r.Context(), sid, id)
	if err != nil {
		a.Logger.Error("delete task error", "error", err)
		a.renderError(w, err)
		return
	}
	http.Redirect(w, r, "/tasks", http.StatusFound)
}

func (a *WebApp) pauseTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sid := r.Context().Value("sid").(string)
	err := a.App.Client.PauseTask(r.Context(), sid, id)
	if err != nil {
		a.Logger.Error("pause task error", "error", err)
		a.renderError(w, err)
		return
	}
	http.Redirect(w, r, "/tasks", http.StatusFound)
}

func (a *WebApp) resumeTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sid := r.Context().Value("sid").(string)
	err := a.App.Client.ResumeTask(r.Context(), sid, id)
	if err != nil {
		a.Logger.Error("resume task error", "error", err)
		a.renderError(w, err)
		return
	}
	http.Redirect(w, r, "/tasks", http.StatusFound)
}

func (a *WebApp) notFound(w http.ResponseWriter, r *http.Request) {
	err := fmt.Errorf("Page %s not found", r.URL.Path)
	w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
	a.Logger.Warn("page not found", "url", r.URL.Path)
	a.renderError(w, err)
}

func (a *WebApp) health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func authenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sidCookie, err := r.Cookie("sid")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		ctx := context.WithValue(r.Context(), "sid", sidCookie.Value)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func LoadTemplates() TemplateCache {
	base := template.Must(template.New("base.html").Funcs(templateFunctions).ParseFS(ui.Files, "html/base.html"))
	pages, err := fs.Glob(ui.Files, "html/pages/*.html")
	if err != nil {
		panic("Can't load templates: " + err.Error())
	}
	cache := make(TemplateCache, len(pages))
	for _, page := range pages {
		name := filepath.Base(page)
		ts := template.Must(base.Clone())
		ts = template.Must(ts.New(name).ParseFS(ui.Files, page))
		cache[name] = ts
	}
	return cache
}
