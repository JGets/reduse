package main

import(
	"os"
	"html/template"
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
								  "server_address":"0.0.0.0:8080",
								  "base_url":"http://0.0.0.0:8080",
								  }

var appName, serverAddress, siteBaseURL string
var logger *log.Logger



/*
	templ = {"template_name":"name_for_template", "template_file":"relative_file_location"}
	args = template arguments
*/
func templatePage(ctx *web.Context, templ map[string]string, args map[string]string){
	
	t, err := template.New(templ["template_name"]).ParseFiles(templ["template_file"])
	
	if err != nil{
		logger.Println("ERROR: ", err.Error())
	}
	
	//Check if a base url has been passed in. If not, set it to the default base url
	_, baseExists := args["base_url"]
	if !baseExists {
		args["base_url"] = siteBaseURL
	}
	
	_, appNameExists := args["app_name"]
	if !appNameExists {
		args["app_name"] = appName
	}
    
    err = t.Execute(ctx, args)
    
    if err != nil{
		logger.Println("ERROR: ", err.Error())
	}
}





func home(ctx *web.Context){
	templatePage(ctx, 
				 map[string]string{"template_name":"home.html", 
								   "template_file":"templatePages/home.html",
								   }, 
				 map[string]string{"title_text":"Homepage", 
				 				   "body_text":"Hello World!",
				 				   })
}

func error404(ctx *web.Context, url string){
	bodyStr := "Could not locate \"" + url + "\" on this server"
	
	ctx.WriteHeader(404)
	
	templatePage(ctx,
				 map[string]string{"template_name":"error.html",
				 				   "template_file":"templatePages/error.html",
				 				   },
				 map[string]string{"title_text":"404 Page Not Found",
				 				   "body_text":bodyStr,
				 				   })
}





func main() {
	// logfile, err := os.Create("log.txt")
	
	// if err != nil {
	// 	log.Fatal("Error: Could not open logfile")
	// }
	
	
	// logger = log.New(logfile, "", log.Ldate | log.Ltime)
	
	logger = log.New(os.Stdout, "", log.Ldate | log.Ltime)
	
	web.SetLogger(logger)
	
	started, err := startup()
	
	if !started {
		logger.Println("Oops, looks like something went wrong with startup.")
		logger.Panic(err)
		return
	}
	
	port := os.Getenv("PORT")
	
	serverAddressWithPort := /*serverAddress +*/ ":" + port
	
	web.Get("/", home)
	web.Get("/(.+)", error404)	//Catch any other URL as unrecognized (regex '(.+)' = any single character 1 or more times)
	web.Run(serverAddressWithPort)
	
}
