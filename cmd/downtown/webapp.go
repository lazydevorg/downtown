package main

import (
	"fmt"
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
	_ = http.ListenAndServe("localhost:4000", mux)
}

func (a *WebApp) home(w http.ResponseWriter, r *http.Request) {
	ts, err := template.ParseFiles(
		"./ui/html/pages/base.html",
		"./ui/html/pages/home.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = ts.Execute(w, nil)
	if err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
	}
}

func (a *WebApp) loginPage(w http.ResponseWriter, r *http.Request) {
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
	_, _ = fmt.Fprintf(w, "Login succeeded with SID: %s", response.Data.SID)
	w.(http.Flusher).Flush()
}
