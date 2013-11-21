package main

import(
	"os"
	"errors"
	"bufio"
	"strings"
)

/*
	startup the app
	Returns:
		bool:	true if startup went off without a hitch, false otherwise
		error:	any error that was encountered, or nil
*/
func startup() (bool, error){
	//open and parse the config file
	configs, err := parseConfigFile()
	if err != nil {
		return false, err
	}
	
	//populate the default values for anything not over-ridden in the config file
	configs = populateDefaults(configs)
	if err != nil {
		return false, err
	}
	
	//check to make sure the configs are valid
	err = checkConfigs(configs)
	if err != nil {
		return false, err
	}
	
	//set the config values to their repsective global variables
	setConfigVars(configs)
	
	return true, nil
}

/*
	Parses the config file into a map
	Returns:
		map[string]string:
				A map of the config values, or nil if an error was encountered
		error:	any error that was encountered, or nil
*/
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

/*
	Populates the default values for any keys in the config map that have empty values
	Parameters:
		configs:	The map of config values that is to populated with any missing defaults
	Returns:
		map[string]string:
				The config map with any missing default keys and values
*/
func populateDefaults(configs map[string]string) map[string]string{
	//Go through the configs map and poplate any empty values that have a default value
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
	//go through and add any default missing keys (and their values) into the configs map
	for k, v := range DEFAULTS {
		_, inConfig := configs[k]
		if !inConfig {
			configs[k] = v
		}
	}
	
	return configs
}

/*
	Checks specific config values for any issues
	Parameters:
		configs:	the config map to be checked
	Returns:
		error:		any error that was encountered, or nil
*/
func checkConfigs(configs map[string]string) error {
	//The base URL MUST end with a /
	if !strings.HasSuffix(configs["base_url"], "/") {
		return errors.New("Error: config base url '" + configs["base_url"] + "' does not end with a slash '/'")
	}
	return nil
}

/*
	sets the config values to their associated global variables
	Parameters:
		congigs:	The map of config values
*/
func setConfigVars(configs map[string]string) {
	appName = configs["app_name"]
	if devMode {
		serverAddress = configs["dev_server_address"]
		siteBaseURL = configs["dev_base_url"]
		siteShortAddr = configs["dev_short_addr"]
	} else {
		serverAddress = configs["server_address"]
		siteBaseURL = configs["base_url"]
		siteShortAddr = configs["short_addr"]
	}
}
