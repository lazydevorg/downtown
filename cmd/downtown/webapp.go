package main

import (
	"html/template"
	"net/http"
)

type WebApp struct {
	App *App
}

func NewWebApp(app *App) *WebApp {
	return &WebApp{
		App: app,
	}
}

func (a *WebApp) Start() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", a.home)
	mux.HandleFunc("GET /login", a.loginPage)
	mux.HandleFunc("POST /login", a.login)
	mux.HandleFunc("GET /tasks", a.tasks)
	_ = http.ListenAndServe("localhost:4000", mux)
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
	ts, err := template.ParseFiles(
		"./ui/html/pages/base.html",
		"./ui/html/pages/login.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = ts.Execute(w, nil)
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
	ts, err := template.ParseFiles(
		"./ui/html/pages/base.html",
		"./ui/html/pages/tasks.html")
	if err != nil {
		http.Error(w, "Internal Server Error "+err.Error(), http.StatusInternalServerError)
		return
	}

	sidCookie, err := r.Cookie("sid")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	var tasksResponse Response[TasksData]
	err = a.App.Client.GetTasks(r.Context(), sidCookie.Value, &tasksResponse)
	if err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
	err = ts.Execute(w, tasksResponse.Data)
	if err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

}
