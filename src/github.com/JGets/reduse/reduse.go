package main

import(
	"os"
	// "html/template"
	"log"
	// "errors"
	// "bufio"
	// "strings"
	
	"github.com/hoisie/web"			/* http://webgo.io */
)

const(
	CONFIG_FILE_NAME = "config.txt"
)

var DEFAULTS = map[string]string {
								  "app_name":"[APP_NAME]",
								  "server_address":"",
								  "base_url":"http://0.0.0.0:8080",
								  "dev_port":"8080",
								  "dev_base_url":"http://0.0.0.0:8080/",
								  }

var devMode bool
var appName, serverAddress, siteBaseURL string
var logger *log.Logger
var linkTable *LinkTable





func main() {
	//Get environment variables
	devStr := os.Getenv("REDUSEDEVELOPMODE")
	devMode = (devStr == "true")
	
	port := os.Getenv("PORT")
	
	// dbName := os.Getenv("REDUSE_DB_NAME")
	// dbAddress := os.Getenv("REDUSE_DB_ADDRESS")
	// dbUsername := os.Getenv("REDUSE_DB_USERNAME")
	// dbPassword := os.Getenv("REDUSE_DB_PASSWORD")
	
	
	// logfile, err := os.Create("log.txt")
	
	// if err != nil {
	// 	log.Fatal("Error: Could not open logfile")
	// }
	
	
	// logger = log.New(logfile, "", log.Ldate | log.Ltime)
	
	logger = log.New(os.Stdout, "", log.Lshortfile)
	
	web.SetLogger(logger)
	
	
	if devMode {
		logger.Println("Running in Develop mode")
	}
	
	
	started, err := startup()
	
	if !started {
		logger.Println("Oops, looks like something went wrong with startup.")
		logger.Panic(err)
		return
	}
	
	
	// err = initDatabase(dbName, dbAddress, dbUsername, dbPassword)
	
	// if err != nil {
	// 	logger.Println("Could not initialize database interface")
	// 	logger.Panic(err.Error())
	// 	return
	// }
	
	
	
	serverAddressWithPort := /*serverAddress +*/ ":" + port
	
	web.Get("/", home)
	web.Get("/generate/", generate)
	web.Get("/list/", listLinks)
	web.Get("/(.+)", serveLink)
	//web.Get("/(.+)", error404)	//Catch any other URL as unrecognized (regex '(.+)' = any single character 1 or more times)
	web.Run(serverAddressWithPort)
	
}
