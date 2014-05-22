package main

import (
	"database/sql"
	"flag"
	"github.com/BurntSushi/toml"
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
}

type ConfigStruct struct {
	TemplateDir   string
	SessionSecret string
	Database      ConfigStructDatabase
	Mail          ConfigStructMail
}

var (
	Config       ConfigStruct
	SessionStore *sessions.CookieStore
	gDb          gorm.DB
	Db           *sql.DB
	Template     *gold.Generator

	TmplRegister *template.Template
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
	TmplRegister, _ = Template.ParseFile("register.gold")

	// open database connection
	if d, err := gorm.Open(Config.Database.Type, Config.Database.Conn); err == nil {
		gDb = d
		Db = d.DB()
		InitModels()
	} else {
		log.Fatal(err)
	}

	// set routes
	r := mux.NewRouter()
	r.HandleFunc("/register", Register)

	http.ListenAndServe(":3000", nosurf.New(r))
}