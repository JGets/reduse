package main

import(
	"os"
	"errors"
	"bufio"
	"strings"
)

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
	
	
	linkTable, err = initLinkTable("reduse_LinkTable.txt")
	
	if err != nil {
		return false, err
	}
	
	
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
	
	if devMode {
		serverAddress = configs["dev_server_address"]
		siteBaseURL = configs["dev_base_url"]
	} else {
		serverAddress = configs["server_address"]
		siteBaseURL = configs["base_url"]
	}
}
