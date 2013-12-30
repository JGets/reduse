package main

import(
	"os"
	"log"
	
	"github.com/hoisie/web"			/* http://webgo.io */
	"github.com/yvasiyarov/gorelic"
)

const(
	CONFIG_FILE_NAME = "config.txt"
	LINK_START_LENGTH = 5
	CAPTCHA_MIN_LENGTH = 6
	CAPTCHA_VARIANCE = 2	// CAPTCHA length will be in [CAPTCHA_MIN_LENGTH, CAPTCHA_MIN_LENGNTH + CAPTCHA_VARIANCE]
	NUM_REPORTS_TO_FLAG	= 1	// The number of reports required before a link becomes flagged for review
)

var DEFAULTS = map[string]string {
								  "app_name":"[APP_NAME]",
								  "server_address":"",
								  "base_url":"http://0.0.0.0:8080",
								  "short_addr":"http://0.0.0.0:8080",
								  "dev_port":"8080",
								  "dev_base_url":"http://0.0.0.0:8080/",
								  "dev_short_addr":"http://0.0.0.0:8080",
								  }

//Global app variables
var devMode, herokuProduction bool
var appName, serverAddress, siteBaseURL, siteShortAddr string
var logger *log.Logger


func main() {
	//Get environment variables
	devStr := os.Getenv("REDUSEDEVELOPMODE")
	devMode = (devStr == "true")
	
	herokuProdStr := os.Getenv("REDUSE_HEROKU_PRODUCTION")
	herokuProduction = (herokuProdStr == "true")
	
	port := os.Getenv("PORT")	//get the port that we are to run off of
	
	//get the database information from the environment
	dbName := os.Getenv("REDUSE_DB_NAME")
	dbAddress := os.Getenv("REDUSE_DB_ADDRESS")
	dbUsername := os.Getenv("REDUSE_DB_USERNAME")
	dbPassword := os.Getenv("REDUSE_DB_PASSWORD")
	
	//get the pertinent email information
	emailUsername := os.Getenv("REDUSE_EMAIL_USERNAME")
	emailPassword := os.Getenv("REDUSE_EMAIL_PASSWORD")
	adminEmails := os.Getenv("REDUSE_EMAIL_ADMIN_ADDRESSES")
	
	//Set up logging to stdout
	logger = log.New(os.Stdout, "", log.Lshortfile)
	web.SetLogger(logger)
	
	if devMode {
		logger.Println("Running in Develop mode")
	}
	
	if herokuProduction {
		logger.Println("Heroku Production flag set")
	}
	
	//Run startup code
	started, err := startup()
	if !started {
		logger.Println("Oops, looks like something went wrong with startup.")
		logger.Panic(err)
		return
	}
	
	//initialize the database (this also validates the database connection)
	err = initDatabase(dbName, dbAddress, dbUsername, dbPassword)
	if err != nil {
		logger.Println("Could not initialize database interface")
		logger.Panic(err.Error())
		return
	}
	
	//Initialize the email sending functionality
	err = initEmail(adminEmails, emailUsername, emailPassword)
	if err != nil{
		logger.Println("Could not initialize email")
		logger.Panic(err.Error())
		return
	}
	
	//Past this point, we should not have any panic()'s, rather any and all errors should be handled gracefully
	
	//don't do any of this stuff if we are in development mode (ie. production-only initialization goes here)
	if !devMode {
		//Set up the NewRelic agent
		agent := gorelic.NewAgent()
		agent.NewrelicLicense = os.Getenv("REDUSE_NEWRELIC_LICENSE_KEY")
		agent.NewrelicName = "Redu.se"
		agent.Run()
	}
	
	web.Get("/", home)
	web.Get("/page/disabled/?", showDisabled)
	web.Get("/page/terms/?", showTerms)
	web.Post("/page/generate/?", generate)
	web.Get("/page/report/?", reportLink)
	web.Post("/page/report/submit/?", submitReport)
	web.Get("/page/contact/?", contactPage)
	web.Post("/page/contact/submit/?", submitContact)
	web.Get("/rsrc/captcha/img/reload/(.+)\\.png", reloadCaptchaImage)
	web.Get("/rsrc/captcha/img/(.+)\\.png", serveCaptchaImage)
	//web.Get("/rsrc/captcha/audio/(.+)\\.wav", serveCaptchaAudio)
	web.Get("/(.+)/(.*)", serveLinkWithExtras)
	web.Get("/(.+)", serveLink)
	web.Run(":" + port)
}






