package main

import(
	"html/template"
	"crypto/md5"
	"io"
	"encoding/base32"
	"strings"
	"errors"
	"net/url"
	// "net/http"
	"math/rand"
	
	"github.com/hoisie/web"
	"github.com/dchest/captcha"
)

const(
	URL_EMPTY = "URL_EMPTY"
	URL_EMPTY_HOST = "URL_EMPTY_HOST"
	URL_NOT_ABSOLUTE = "URL_NOT_ABSOLUTE"
	URL_INVALID_SCHEME = "URL_INVALID_SCHEME"
	URL_VALIDATED_INEQUIVALENT = "URL_VALIDATED_INEQUIVALENT"
)


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
	
	// CAPTCHA length will be in [CAPTCHA_MIN_LENGTH, CAPTCHA_MIN_LENGNTH + CAPTCHA_VARIANCE]
	captchaId := captcha.NewLen(CAPTCHA_MIN_LENGTH + rand.Intn(CAPTCHA_VARIANCE + 1))
	
	
	templatePage(ctx, 
				 map[string]string{"template_name":"home.html", 
								   "template_file":"templatePages/home.html",
								   }, 
				 map[string]string{"captcha_id":captchaId})
}


func serveCaptcha(ctx *web.Context, id string){
	err := captcha.WriteImage(ctx, id, captcha.StdWidth, captcha.StdHeight)
	if err != nil {
		logger.Println("Error, could not write CAPTCHA image")
		logger.Println(err.Error())
	}
}



func error404(ctx *web.Context, urlStr string){
	logger.Printf("404 Error for URL: %v\n", urlStr)
	
	
	bodyStr := "Could not locate \"" + urlStr + "\" on this server"
	
	ctx.WriteHeader(404)
	
	templatePage(ctx,
				 map[string]string{"template_name":"error.html",
				 				   "template_file":"templatePages/error.html",
				 				   },
				 map[string]string{"title_text":"404 Page Not Found",
				 				   "body_text":bodyStr,
				 				   })
}

func internalError(ctx *web.Context, err error){
	logger.Printf("500 Internal Server Error: %v\n", err.Error())
	
	ctx.WriteHeader(500)
	templatePage(ctx,
				 map[string]string{"template_name":"error.html",
				 				   "template_file":"templatePages/error.html",
				 				   },
				 map[string]string{"title_text":"500 Internal Server Error",
				 				   "body_text":err.Error(),
				 				   })
}


func blacklistedPage(ctx *web.Context, urlStr string){
	bodyStr := "A link cannot be generated for \"" + urlStr + "\" because that domain has been blacklisted."
	templatePage(ctx,
				 map[string]string{"template_name":"blacklisted.html",
				 				   "template_file":"templatePages/blacklisted.html",
				 				   },
				 map[string]string{"title_text":"Blacklisted URL",
				 				   "body_text":bodyStr,
				 				   })
}


func invalidURLPage(ctx *web.Context, reason string) {
	
	params := make(map[string]string)
	
	
	params["title_text"] = "Invalid URL"
	params["body_text"] = "The given URL to shorten was invalid."
	

	switch reason {
		case URL_EMPTY:
			params["url_empty"] = "true"
		case URL_EMPTY_HOST:
			params["url_empty_host"] = "true"
		case URL_NOT_ABSOLUTE:
			params["url_not_absolute"] = "true"
		case URL_INVALID_SCHEME:
			params["url_invalid_scheme"] = "true"
		case URL_VALIDATED_INEQUIVALENT:
			params["url_validated_inequivalent"] = "true"
	}
	
	
	
	templatePage(ctx,
				 map[string]string{"template_name":"invalidURL.html",
				 				   "template_file":"templatePages/invalidURL.html",
				 				   },
				 params)
}



func dbTest(urlStr string) string{
	//make the hash all uppercase
	upperHash := strings.ToUpper(urlStr)
	
	link, exists, err := db_linkForHash(upperHash)
	
	if err != nil {
		return err.Error()
	} else if !exists {
		return "No link exists"
	}
	
	return upperHash +" : "+ link
	
}


func isBlacklisted(urlStr string) (bool, error) {
	u, err := url.Parse(urlStr)
	
	if err != nil {
		return false, err
	}
	
	host := u.Host
	
	logger.Println("Host: "+host)
	
	
	
	return false, nil
}


/*
	Validates a URL - ie. checks to make sure it is valid (absolute, uses http or https scheme, etc)
	Parameters:
		urlStr:	The URL (in a string) that is to be validated
	Returns:
		string:	A valid URL that is equivalent to the given one, or a message as to why the URL is invalid (when bool = false), 
					or nil (if an error was encountered)
		bool:	true if the URL is can be validated, false otherwise (including when an error occurs)
		error:	Any error that was encounterd, or nil
*/
func validateURL(urlStr string) (string, bool, error){
	//Make sure we weren't given an empty URL
	if urlStr == "" {
		return URL_EMPTY, false, nil
	}
	
	//Parse the URL
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", false, err
	}
	
	logger.Printf("%v has host: %v\n", urlStr, u.Host)
	
	//check to make sure the URL contains a host
	if u.Host == ""{
		return URL_EMPTY_HOST, false, nil
	}
	
	//Check to make sure it is using a vaild scheme (http or https)
	var needsScheme = false
	/*if u.Scheme == "" {
		needsScheme = true
		u.Scheme = "http"
	} else*/ if u.Scheme != "http" && u.Scheme != "https" {
		return URL_INVALID_SCHEME, false, nil
	}
	
	//Check if the URL is not absolute (relative URLs would not work anyways)
	if !u.IsAbs() {
		return URL_NOT_ABSOLUTE, false, nil
	}
	
	//Check to make sure the validated URL is equivalent to the given one
	validStr := u.String()
	if needsScheme {
		urlStr = "http://" + urlStr
	}
	if validStr != urlStr {
		return URL_VALIDATED_INEQUIVALENT, false, nil
	}
	
	return validStr, true, nil
}

func goodCaptchaSolution(ctx *web.Context, id, soln string) bool {
	//make sure we were given a non-empty ID
	if id == "" {
		internalError(ctx, errors.New("Attempting to verify CAPTCHA with empty ID"))
		return false
	} else if soln == "" {		//Make sure they actually answered the CAPTCHA
		templatePage(ctx,
				 map[string]string{"template_name":"wrongCaptcha.html",
				 				   "template_file":"templatePages/wrongCaptcha.html",
				 				   },
				 map[string]string{"title_text":"Incorrect CAPTCHA",
				 				   "body_text":"You must enter a solution to the CAPTCHA to generate a short link",
				 				   })
		return false
	} else if !captcha.VerifyString(ctx.Params["captcha_id"], soln) {	//They didn't give a correct solution
		templatePage(ctx,
				 map[string]string{"template_name":"wrongCaptcha.html",
				 				   "template_file":"templatePages/wrongCaptcha.html",
				 				   },
				 map[string]string{"title_text":"Incorrect CAPTCHA",
				 				   "body_text":"The solution to the CAPTCHA that you entered was incorrect",
				 				   })
		return false
	}
	//The user gave us a correct solution to the CAPTCHA
	return true
}


func generate(ctx *web.Context){
	
	//Verify the user's CAPTCHA solution
	capId := ctx.Params["captcha_id"]
	capSoln := ctx.Params["captcha_soln"]
	if !goodCaptchaSolution(ctx, capId, capSoln) {
		return
	}
	
	urlStr := ctx.Params["url"]
	
	//Check to make sure we were given a valid URL
	validURL, isValid, err := validateURL(urlStr)
	if err != nil {
		internalError(ctx, errors.New("Error validating URL: "+err.Error()))
		return
	} else if !isValid {
		logger.Println(validURL)
		invalidURLPage(ctx, validURL)
		return
	}
	urlStr = validURL
	
	
	//Check if the URL actually goes somewhere
	
	// resp, err := http.Get(urlStr)
	// if err != nil {
		
	// }
	// if 
	
	
	
	
	
	
	//TODO: Check the domain against the blacklist
	// blacklisted, err := isBlacklisted(urlStr)
	// if err != nil {
	// 	internalError(ctx, errors.New("Could not check URL against blacklist. ~ " + err.Error()))
	// 	return
	// } else if blacklisted {
	// 	blacklistedPage(ctx, urlStr)
	// 	return
	// }
	
	
	
	//Generate a new MD5 hasher, and hash the urlStr
	hasher := md5.New()
	io.WriteString(hasher, urlStr)
	hashBytes := hasher.Sum(nil)
	hashStr := base32.StdEncoding.EncodeToString(hashBytes)
	
	//Check for collisions (ie. different links resulting in the same short-hash), and fix them
		//(by adding the next character from the full hash to the short hash, and checking for another collision)
	var testHash string
	var collision bool = true
	var alreadyExists = false
	for i := LINK_START_LENGTH; i <= len(hashStr) && collision; i++ {
		testHash = hashStr[:i]
		
		//Check if this shorthash already exists in the database
		val, exists, err := db_linkForHash(testHash)
		if err != nil {
			internalError(ctx, errors.New("Database Error: "+err.Error()))
			return
		}
		
		if !exists {
			//No link exists for this short hash, so there is no collision
			collision = false
		} else if val == urlStr {
			//This short has is used already, but for the same URL
			collision = false
			alreadyExists = true
		}
		//otherwise, there was a collision, so check the short-hash of one char longer
	}
	
	//if we have hit the maximum length of the hash, and there is still a collision, throw an error
	if collision {
		internalError(ctx, errors.New("Could not resolve collision. Hash: " + hashStr + "    Link: " + urlStr))
		return
	}
	
	finalHash := testHash
	
	//if the link did not already exist (Optimization: db_addLink checks this too, but we've already done it here, so why do it again?)
	if !alreadyExists {
		//Save the link to the link table
		err := db_addLink(finalHash, urlStr)
		if err != nil {
			internalError(ctx, errors.New("Database Error: could not add link to database. \""+err.Error()+"\""))
			return
		}
	}
	
	//Give user output webpage
	body := "Generate short link for " + urlStr
	templatePage(ctx,
				 map[string]string{"template_name":"generate.html",
				 				   "template_file":"templatePages/generate.html",
				 				   },
				 map[string]string{"title_text":"Generate URL",
				 				   "body_text":body,
				 				   "link_hash":strings.ToLower(finalHash),
				 				   })
}

func serveLink(ctx *web.Context, hash string){
	serveLinkWithExtras(ctx, hash, "")
}

func serveLinkWithExtras(ctx *web.Context, hash string, extras string){
	//make the hash all uppercase
	upperHash := strings.ToUpper(hash)
	
	
	
	//Check to see if a link exists for the given hash
	link, exists, err := db_linkForHash(upperHash)
	if err != nil {
		//There was an error in the database
		internalError(ctx, errors.New("Database Error: "+err.Error()))
	} else if exists {
		redir := link
		
		//If there were any URL extras passed to us, append them to the redir link
		if extras != "" {
			redir += "/" + extras
		}
		
		//If there are any GET parameters being passed to us, append them to the redir link
		if len(ctx.Params) > 0 {
			params := "?"
			for k, v := range ctx.Params {
				params += k + "=" + v + "&"
			}
			//remove the trailing ampersand and append to the redir link
			redir += strings.TrimSuffix(params, "&")
		}
		
		//if the hash exists in the link table, issue a '302 Moved Permanently' to the client with the link URL
		ctx.Redirect(302, redir)	
	} else {
		//No link exists for the hash, so serve a 404
		error404(ctx, hash)
	}
}

func listLinks() string{
	var ret = ""
	
	links, err := db_getLinkTable()
	if err != nil {
		return "Internal Server/Database Error: " + err.Error()
	}
	
	for key, val := range links {
		key = strings.ToLower(key)
		ret += "<a href=\"" + siteBaseURL + key + "\">redu.se/" + key + "</a> : <a href=\"" + val + "\">" + val + "</a><br/>"
	}
	
	return ret
}
