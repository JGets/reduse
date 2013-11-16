package main

import(
	"os"
	"html/template"
	"log"
	"errors"
	"bufio"
	"strings"
	
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
	
	templatePage(ctx,
				 map[string]string{"template_name":"error.html",
				 				   "template_file":"templatePages/error.html",
				 				   },
				 map[string]string{"title_text":"404 Page Not Found",
				 				   "body_text":bodyStr,
				 				   })
}

func startup() (bool, error){
	
	//TODO
	//get the server address / base url from user config file
	
	configs, err := parseConfigFile()
	
	if err != nil {
		return false, err
	}
	
	configs, err = populateDefaults(configs)
	
	if err != nil {
		return false, err
	}
	
	err = checkConfigs(configs)
	if err != nil {
		return false, err
	}
	
	
	setConfigVars(configs)
	
	return true, nil
}

func parseConfigFile() (map[string]string, error){
	//Open and read in the config file
	
	file, err := os.Open(CONFIG_FILE_NAME)
	if err != nil {
		return nil, errors.New("Error: unable to open config file '" + CONFIG_FILE_NAME + "'")
	}
	defer file.Close()
	
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}
	
	//parse the lines of the config file & put them into a map[string]string
	ret := make(map[string]string)
	
	for _, val := range lines {
		if val != "" {
			spl := strings.Split(val, "=")
			if len(spl) == 2 {
				//Parse the config value
				ret[spl[0]] = spl[1]
			} else if len(spl) == 1 {
				//no config value used, so initialize with an empty string
				ret[spl[0]] = ""
			} else {
				logger.Printf("Warning: Could not parse '%v' from config file. (Too many arguments?)\n", val)
			}
		}
	}
	
	return ret, nil
}


func populateDefaults(configs map[string]string) (map[string]string, error){
	
	//TO DO
	// if configs["server_address"] == "" {
	// 	configs["server_address"] = DEFAULT_SERVER_ADDRESS
	// }
	
	for k, v := range configs {
		if v == "" {
			def, inDef := DEFAULTS[k]
			if inDef {
				configs[k] = def
			} else {
				logger.Printf("Warning: Empty config key '%v' with no default value, ignoring\n", k)
			}
		}
	}
	
	
	return configs, nil
}


func checkConfigs(configs map[string]string) error {
	
	if !strings.HasSuffix(configs["base_url"], "/") {
		return errors.New("Error: config base url '" + configs["base_url"] + "' does not end with a backslash '/'")
	}
	
	
	return nil
	
}

func setConfigVars(configs map[string]string) {
	appName = configs["app_name"]
	serverAddress = configs["server_address"]
	siteBaseURL = configs["base_url"]
}



func main() {
	logfile, err := os.Create("log.txt")
	
	if err != nil {
		log.Fatal("Error: Could not open logfile")
	}
	
	
	logger = log.New(logfile, "", log.Ldate | log.Ltime)
	
	//logger = log.New(os.Stdout, "", log.Ldate | log.Ltime)
	
	web.SetLogger(logger)
	
	started, err := startup()
	
	if !started {
		logger.Println("Oops, looks like something went wrong with startup.")
		logger.Panic(err)
		return
	}
	
	port := os.Getenv("PORT")
	
	serverAddressWithPort := serverAddress + port
	
	web.Get("/", home)
	web.Get("/(.+)", error404)	//Catch any other URL as unrecognized (regex '(.+)' = any single character 1 or more times)
	web.Run(serverAddressWithPort)
	
}
