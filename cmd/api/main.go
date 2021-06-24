package main

import (
	"database/sql"
	"log"
	"net/http"
	"net/url"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/kelseyhightower/envconfig"
	"github.com/raphael-trzpit/covid-analytics/internal"
)

// Colonne,Type ,Description_FR,Description_EN,Exemple
// dep,String,Departement,State,01
// jour,Date,Jour,Day,2020-05-13
// t,integer,Nombre de test réalisés,Number of tests performed,2141
// cl_age90,integer,Classe d'age,Age class,09
// p,integer,Nombre de test positifs,Number of positive tests,34

type Config struct {
	DataGouvURL string `envconfig:"DATAGOUV_URL"`
	DBDSN       string `envconfig:"DB_DSN"`
	HTTPPort    int    `envconfig:"HTTP_PORT"`
}

func main() {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		log.Fatal(err)
	}

	dataGouvURL, err := url.Parse(config.DataGouvURL)
	if err != nil {
		log.Fatalf("invalid datagouv url (%s)", config.DataGouvURL)
	}

	dataGouvSource := internal.NewDataGouvSource(dataGouvURL)

	db, err := sql.Open("mysql", config.DBDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	sqlStorage := internal.NewStorageSQL(db)

	importHandler := internal.NewImportHandler(dataGouvSource, sqlStorage)
	analyticsHandler := internal.NewAnalyticsHandler(sqlStorage)

	http.HandleFunc("/import", importHandler.ImportHandler)
	http.HandleFunc("/departments", analyticsHandler.DepartmentListHandler)
	http.HandleFunc("/age_categories", analyticsHandler.AgeCategoryListHandler)
	http.HandleFunc("/days", analyticsHandler.DaysLimitHandler)
	http.HandleFunc("/data", analyticsHandler.DataHandler)
	http.HandleFunc("/national", analyticsHandler.NationalReportsHandler)
	http.HandleFunc("/department", analyticsHandler.DepartmentResumeHandler)
	http.HandleFunc("/daily_top5", analyticsHandler.DailyTop5Handler)

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.HTTPPort), nil))
}
