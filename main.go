package main

import (
	"database/sql"
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
	"github.com/justinas/nosurf"
	_ "github.com/mattn/go-sqlite3"
	"github.com/yosssi/gold"
	"html/template"
	"log"
	"net/http"
)

type ConfigStructDatabase struct {
	Type string
	Conn string
}

type ConfigStructMail struct {
	Username string
	Password string
	Server   string
	Port     int64
}

type ConfigStruct struct {
	TemplateDir   string
	SessionSecret string
	BaseDomain    string
	Database      ConfigStructDatabase
	Mail          ConfigStructMail
}

var (
	Config       ConfigStruct
	SessionStore *sessions.CookieStore
	gDb          gorm.DB
	Db           *sql.DB
	Red          redis.Conn
	Template     *gold.Generator

	TmplLogin     *template.Template
	TmplRegister  *template.Template
	TmplOverview  *template.Template
	TmploAuthAuth *template.Template
)

func main() {
	flag.Parse()

	// load config
	configFile := flag.String("c", "config.toml", "specify location of config file")
	if _, err := toml.DecodeFile(*configFile, &Config); err != nil {
		log.Fatal(err)
	}
	Template = gold.NewGenerator(true).SetBaseDir(Config.TemplateDir)
	SessionStore = sessions.NewCookieStore([]byte(Config.SessionSecret))
	TmplLogin, _ = Template.ParseFile("login.gold")
	TmplRegister, _ = Template.ParseFile("register.gold")
	TmplOverview, _ = Template.ParseFile("overview.gold")
	TmploAuthAuth, _ = Template.ParseFile("oauth-authorization.gold")

	// open database connection
	if d, err := gorm.Open(Config.Database.Type, Config.Database.Conn); err == nil {
		gDb = d
		Db = d.DB()
		InitModels()
	} else {
		log.Fatal(err)
	}

	// open redis connection
	if r, err := redis.Dial("tcp", ":6379"); err == nil {
		Red = r
		defer r.Close()
	} else {
		log.Fatal(err)
	}

	// set routes
	r := mux.NewRouter()
	r.HandleFunc("/register", Register)
	r.HandleFunc("/confirm/{uuid}", Confirm)
	r.HandleFunc("/login", Login)
	r.HandleFunc("/logout", Logout)
	r.HandleFunc("/overview", Overview)

	r.HandleFunc("/o/authorization", oAuthAuthorization)
	r.HandleFunc("/o/token", oAuthToken)

	// csrf protection
	csrfHandler := nosurf.New(r)
	csrfHandler.ExemptPath("/o/token")

	http.ListenAndServe(":3000", context.ClearHandler(csrfHandler))
}
