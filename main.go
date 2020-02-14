package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/csrf"
	"go-imagecloud/controllers"
	"go-imagecloud/middleware"
	"go-imagecloud/models"
	"go-imagecloud/rand"
	"golang.org/x/oauth2"

	//"html/template"
	"net/http"

	"github.com/gorilla/mux"
)


func main() {

	//boolPtr := flag.Bool("prod", true, "Provide this flag "+
	//	"in production. This ensures that a .config file is "+
	//	"provided before the application starts.")
	boolPtr := flag.Bool("prod", false, "Provide this flag "+
		"in production. This ensures that a .config file is "+
		"provided before the application starts.")
	flag.Parse()

	cfg := LoadConfig(*boolPtr)
	fmt.Println(cfg.Dropbox)
	dbCfg := cfg.Database
	services, err := models.NewServices(
		models.WithGorm(dbCfg.Dialect(), dbCfg.ConnectionInfo()),
		// only log when not in prod
		models.WithLogMode(!cfg.IsProd()),
		// We want each of these services, but if we didn't need
		// one of them we could possibly skip that config func
		models.WithUser(cfg.Pepper, cfg.HMACKey),
		models.WithGallery(),
		models.WithImage(),
	)
	if err != nil {
		panic(err)
	}
	defer services.Close()
	services.AutoMigrate()

	// not set up
	//mgCfg := cfg.Mailgun
	//emailer := email.NewClient(
	//	email.WithSender("ImageCloud.com Support", "support@"+mgCfg.Domain),
	//	email.WithMailgun(mgCfg.Domain, mgCfg.APIKey, mgCfg.PublicAPIKey),
	//)

	r := mux.NewRouter()
	staticC := controllers.NewStatic()
	// emailer is nil because we arent using it now
	usersC := controllers.NewUsers(services.User, nil)
	galleriesC := controllers.NewGalleries(services.Gallery, services.Image, r)

	//services.DestructiveReset()

	// Middleware
	// csrf middleware
	b, err := rand.Bytes(32)
	if err != nil {
		panic(err)
	}
	// csrfMw := csrf.Protect(b) //, csrf.Secure(true)) // , csrf.Secure(cfg.IsProd()))
	// csrfMw := csrf.Protect(b, csrf.Secure(false)) //cfg.IsProd()))
	csrfMw := csrf.Protect(b, csrf.Secure(cfg.IsProd()))
	userMw := middleware.User{
		UserService: services.User,
	}
	requireUserMw := middleware.RequireUser{
		User: userMw,
	}

	// db
	dbxOAuth := &oauth2.Config{
		ClientID: cfg.Dropbox.ID,
		ClientSecret: cfg.Dropbox.Secret,
		Endpoint: oauth2.Endpoint{
			AuthURL :   cfg.Dropbox.AuthURL,
			TokenURL :  cfg.Dropbox.TokenURL,
		},
		RedirectURL: "http://localhost:3000/oauth/dropbox/callback",
	}

	dbxRedirect := func(w http.ResponseWriter, r *http.Request) {
		state :=  csrf.Token(r)
		url := dbxOAuth.AuthCodeURL(state)
		fmt.Println(state)
		http.Redirect(w, r, url, http.StatusFound)
	}
	r.HandleFunc("/oauth/dropbox/connect", dbxRedirect)
	dbxCallback := func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		fmt.Println(w, "code: ",  r.FormValue("code"), " state: ", r.FormValue("state"))
	}
	r.HandleFunc("/oauth/dropbox/callback", dbxCallback)

	r.Handle("/", staticC.Home).Methods("GET")
	r.Handle("/contact", staticC.Contact).Methods("GET")
	r.Handle("/faq", staticC.Faq).Methods("GET")
	r.HandleFunc("/signup", usersC.New).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")
	r.Handle("/login", usersC.LoginView).Methods("GET")
	r.HandleFunc("/login", usersC.Login).Methods("POST")
	r.Handle("/logout", requireUserMw.ApplyFn(usersC.Logout)).Methods("POST")
	r.HandleFunc("/cookietest", usersC.CookieTest).Methods("GET")

	// This will assign the page to the nor found handler
	var h http.Handler = http.Handler(staticC.NotFound)
	r.NotFoundHandler = h


	// Gallery routes

	r.Handle("/galleries",
		requireUserMw.ApplyFn(galleriesC.Index)).
		Methods("GET").
		Name(controllers.IndexGalleries)
	r.Handle("/galleries/new",
		requireUserMw.Apply(galleriesC.New)).
		Methods("GET")
	r.Handle("/galleries",
		requireUserMw.ApplyFn(galleriesC.Create)).
		Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}",
		galleriesC.Show).
		Methods("GET").
		Name(controllers.ShowGallery)
	r.HandleFunc("/galleries/{id:[0-9]+}/edit",
		galleriesC.Edit).
		Methods("GET").
		Name(controllers.EditGallery)
	r.HandleFunc("/galleries/{id:[0-9]+}/update",
		requireUserMw.ApplyFn(galleriesC.Update)).
		Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/delete",
		requireUserMw.ApplyFn(galleriesC.Delete)).
		Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/images",
		requireUserMw.ApplyFn(galleriesC.ImageUpload)).
		Methods("POST")


	// Image routes
	imageHandler := http.FileServer(http.Dir("./images/"))
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/",imageHandler))

	r.HandleFunc("/galleries/{id:[0-9]+}/images/{filename}/delete",
		requireUserMw.ApplyFn(galleriesC.ImageDelete)).
		Methods("POST")

	// Assets
	assetHandler := http.FileServer(http.Dir("./assets"))
	assetHandler = http.StripPrefix("/assets/", assetHandler)
	r.PathPrefix("/assets/").Handler(assetHandler)


	fmt.Println("Starting the server on :%d...", cfg.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), csrfMw(userMw.Apply(r)))
	// userMw.Apply(r)) //
}