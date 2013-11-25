package main

import(
	"html/template"
	"crypto/md5"
	"io"
	"encoding/base32"
	"strings"
	"errors"
	"net/url"
	"net/http"
	"math/rand"
	"strconv"
	
	"github.com/hoisie/web"
	"github.com/dchest/captcha"
)

//define constants for URL validation errors
const(
	URL_EMPTY = "URL_EMPTY"
	URL_EMPTY_HOST = "URL_EMPTY_HOST"
	URL_NOT_ABSOLUTE = "URL_NOT_ABSOLUTE"
	URL_INVALID_SCHEME = "URL_INVALID_SCHEME"
	URL_VALIDATED_INEQUIVALENT = "URL_VALIDATED_INEQUIVALENT"
	URL_DOES_NOT_EXIST = "URL_DOES_NOT_EXIST"
)


/*
	Parses, executes, and writes the common page, with the given (template) page inserted as content
	Parameters:
		ctx:	The context of the http request
		contentFile:
				The filename of the template file to be used as content (must be in the commonTemplate/content/ directory)
		args:	A map containing all the arguments to be excecuted within the template
*/
func commonTemplate(ctx *web.Context, contentFile string, args map[string]string){
	t, err := template.New("commonTemplate.html").ParseFiles("commonTemplate/commonTemplate.html", "commonTemplate/content/"+contentFile)
	
	if err != nil{
		logger.Println("ERROR: ", err.Error())
	}
	
	args["content_file"] = "commonTemplate/content/" + contentFile
	
	//Check if a base url has been passed in. If not, set it to the default base url
	_, baseExists := args["base_url"]
	if !baseExists {
		args["base_url"] = siteBaseURL
	}
	
	_, appNameExists := args["app_name"]
	if !appNameExists {
		args["app_name"] = appName
	}
	
	_, shortAddrExists := args["short_addr"]
	if !shortAddrExists {
		args["short_addr"] = siteShortAddr
	}

	err = t.Execute(ctx, args)

	if err != nil{
		logger.Println("ERROR: ", err.Error())
	}
}



/*
	Serve the homepage
	Parameters:
		ctx:	the context of the http request
*/
func home(ctx *web.Context){
	// CAPTCHA length will be in [CAPTCHA_MIN_LENGTH, CAPTCHA_MIN_LENGTH + CAPTCHA_VARIANCE]
	captchaId := captcha.NewLen(CAPTCHA_MIN_LENGTH + rand.Intn(CAPTCHA_VARIANCE + 1))
	

	commonTemplate(ctx, 
				   "home.html", 
				   map[string]string{"captcha_id":captchaId, 
									 "user_url":ctx.Params["url"],
									 "captcha_soln_min_length":strconv.Itoa(CAPTCHA_MIN_LENGTH),
									 "captcha_soln_max_length":strconv.Itoa(CAPTCHA_MIN_LENGTH + CAPTCHA_VARIANCE),
									 })
}

func showTerms(ctx *web.Context){
	commonTemplate(ctx, "terms.html", map[string]string{})
}


/*
	Serves a CAPTCHA image
	Parameters:
		ctx:	The context of the http request
		id:		The ID of the captcha to serve
*/
func serveCaptchaImage(ctx *web.Context, id string){
	
	width, err := strconv.Atoi(ctx.Params["width"])
	if err != nil {
		logger.Printf("Error: could not parse captcha image width of '%v'\n%v\n", ctx.Params["width"], err.Error())
		width = captcha.StdWidth
	}
	
	height, err := strconv.Atoi(ctx.Params["height"])
	if err != nil {
		logger.Printf("Error: could not parse captcha image height of '%v'\n%v\n", ctx.Params["height"], err.Error())
		height = captcha.StdHeight
	}
	
	//tell the user's browser not to cache the image file
	ctx.SetHeader("Cache-Control", "no-cache", true)
	
	err = captcha.WriteImage(ctx, id, width, height)
	if err != nil {
		logger.Println("Error, could not write CAPTCHA image")
		logger.Println(err.Error())
	}
}


/*
	Reloads the CAPTCHA with the given ID, and returns a .png image representation to be solved
	Parameters:
		ctx:	The context of the http request
		id:		The ID of the captcha to reload and serve
*/
func reloadCaptchaImage(ctx *web.Context, id string){
	exists := captcha.Reload(id)
	if !exists {
		logger.Println("Error, trying to reload non-existent CAPTCHA")
	}
	
	serveCaptchaImage(ctx, id)
}



func serveCaptchaAudio(ctx *web.Context, id string){
	//tell the user's browser not to cache the audio file
		//(would cause old audio file to be used even if user has reloaded the CAPTCHA)
	ctx.SetHeader("Cache-Control", "no-cache", true)
	
	err := captcha.WriteAudio(ctx, id, "english")
	if err != nil {
		logger.Println("Error, could not write CAPTCHA audio")
		logger.Println(err.Error())
	}
}




/*
	Serve a 404 Not Found error page
	Parameters:
		ctx:	the context of the http request
		urlStr:	the URL that the request was trying to access
*/
func error404(ctx *web.Context, urlStr string){
	logger.Printf("404 Error for URL: %v\n", urlStr)
	
	bodyStr := "Could not locate \"" + urlStr + "\" on this server"
	
	ctx.WriteHeader(404)
	
	commonTemplate(ctx,
				   "generic.html", 
				   map[string]string{"title_text":"404 Page Not Found",
				   					 "body_text":bodyStr,
				   					 })
}

/*
	Server a 500 Internal Error page
	Parameters:
		ctx:	the context of the http request
		err:	the error that was encountered
*/
func internalError(ctx *web.Context, err error){
	logger.Printf("500 Internal Server Error: %v\n", err.Error())
	
	ctx.WriteHeader(500)
	
	commonTemplate(ctx,
				   "generic.html",
				   map[string]string{"title_text":"500 Internal Server Error",
				 					 "body_text":err.Error(),
				 					 })
}


// func blacklistedPage(ctx *web.Context, urlStr string){
// 	bodyStr := "A link cannot be generated for \"" + urlStr + "\" because that domain has been blacklisted."
// 	templatePage(ctx,
// 				 map[string]string{"template_name":"blacklisted.html",
// 				 				   "template_file":"templatePages/blacklisted.html",
// 				 				   },
// 				 map[string]string{"title_text":"Blacklisted URL",
// 				 				   "body_text":bodyStr,
// 				 				   })
// }


/*
	Serve a page telling the user that the URL they gave is not valid
	Parameters:
		ctx:	the context of the http request
		reason:	a string either representing one of the pre-defined URL validation errors, or a string containing a 
					reason as to why the URL is invalid
*/
func invalidURLPage(ctx *web.Context, reason string) {
	
	params := make(map[string]string)
	
	params["user_url"] = ctx.Params["url"]
	params["show_try_again"] = "true"
	
	params["title_text"] = "Invalid URL"
	params["body_text"] = "The given URL to shorten was invalid."
	
	//check if it is one of the pre-defined validation errors
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
		case URL_DOES_NOT_EXIST:
			params["url_does_not_exist"] = "true"
		default:
			//not one of the pre-defined, so just pass on the reason string
			params["other_reason"] = reason
	}
	
	commonTemplate(ctx, "invalidURL.html", params)
}


// func isBlacklisted(urlStr string) (bool, error) {
// 	u, err := url.Parse(urlStr)
//	
// 	if err != nil {
// 		return false, err
// 	}
//	
// 	host := u.Host
//	
// 	logger.Println("Host: "+host)
//	
// 	return false, nil
// }


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
	
	//Check if the URL is not absolute (relative URLs would not work anyways)
	if !u.IsAbs() {
		return URL_NOT_ABSOLUTE, false, nil
	}
	
	//Make sure we wer given an accepted scheme
	if u.Scheme == "http" || u.Scheme == "https"{	//Accepted URL schemes that end with "://"
		//check to make sure the URL contains a host
		if u.Host == ""{
			return URL_EMPTY_HOST, false, nil
		}
		
		//Check if the URL actually exists (ie. it points to a real webserver)
		resp, err := http.Get(u.String())
		if err != nil {
			return URL_DOES_NOT_EXIST, false, nil
		}
		//close the response's body on return
		defer resp.Body.Close()
		
		//if we did not get a '200 OK' response, reject the URL
		if resp.StatusCode != http.StatusOK {
			return URL_DOES_NOT_EXIST, false, nil
		}
		
	/*} else if u.Scheme == "mailto" {				//Accepted URL schemes that end with only ":"
		//check to make sure we were given a host
		if u.Opaque == "" {
			return URL_EMPTY_HOST, false, nil
		}*/
	} else {
		//The URL does not have a scheme, or it is not an accepted scheme
		return URL_INVALID_SCHEME, false, nil
	}
	
	
	//Check to make sure the validated URL is equivalent to the given one
	validStr := u.String()
	if validStr != urlStr {
		return URL_VALIDATED_INEQUIVALENT, false, nil
	}
	
	return validStr, true, nil
}


/*
	Checks to make sure the user gave a valid CAPTCHA solution. 
	(Note: if false is returned, this function takes care of serving a webpage to the user)
	Parameters:
		ctx:	the context of the http request
		id:		the id string for the captcha we are to check the solution against
		soln:	the solution the user submitted to the CAPTCHA
	Returns:
		bool:	true if the user entered a correct solution, false otherwise.
*/
func goodCaptchaSolution(ctx *web.Context, id, soln string) bool {
	//make sure we were given a non-empty ID
	if id == "" {
		internalError(ctx, errors.New("Attempting to verify CAPTCHA with empty ID"))
		return false
	} else if soln == "" {		//Make sure they actually answered the CAPTCHA
		commonTemplate(ctx,
					   "generic.html",
					   map[string]string{"title_text":"Incorrect CAPTCHA",
			 							 "body_text":"You must enter a solution to the CAPTCHA to generate a short link",
			 							 "show_try_again":"true",
			 							 "user_url":ctx.Params["url"],
			 							 })
		return false
	} else if !captcha.VerifyString(ctx.Params["captcha_id"], soln) {	//They didn't give a correct solution
		commonTemplate(ctx,
					   "generic.html",
					   map[string]string{"title_text":"Incorrect CAPTCHA",
			 							 "body_text":"The solution to the CAPTCHA that you entered was incorrect",
			 							 "show_try_again":"true",
			 							 "user_url":ctx.Params["url"],
			 							 })
		return false
	}
	//The user gave us a correct solution to the CAPTCHA
	return true
}

/*
	Generates a link for the URL the user entered and serves a page with the link, in the following order
		- Checks that the user entered a correct solution to the CAPTCHA
		- Checks that the user entered a valid URL
		- Generates a full hash string of the URL, and attempts to find an unused short-hash:
			- If the short-hash is used for a different URL, add on the next character from the full hash & check again
			- If the short-hash is used for the same URL, serve the user a page with the link
			- If the short-hash is unused, attempt to add it to the database, and then serve the user a page with the link
*/
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
		invalidURLPage(ctx, validURL)
		return
	}
	urlStr = validURL

	
	//TODO: Check the domain against a blacklist
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
	
	commonTemplate(ctx,
				   "generate.html",
				   map[string]string{"title_text":"Generated Link",
				 					 "dest_url":urlStr,
				 					 "link_hash":strings.ToLower(finalHash),
				 					 })
}

/*
	serve a link with no extras (ie. no path relative to the link, or any GET parameters)
*/
func serveLink(ctx *web.Context, hash string){
	serveLinkWithExtras(ctx, hash, "")
}

/*
	serve a link with extras (a path relative to the short-link and/or GET parameters)
	Parameters:
		ctx:	the context of the http request
		hash:	the short-hash of the link
		extras:	the extra path component
*/
func serveLinkWithExtras(ctx *web.Context, hash string, extras string){
	//make the hash all uppercase
	upperHash := strings.ToUpper(hash)
	
	//Check to see if a link exists for the given hash
	link, exists, err := db_linkForHash(upperHash)
	if err != nil {
		//There was an error in the database
		internalError(ctx, errors.New("Database Error: "+err.Error()))
	} else if exists {
		//The hash=>link exists
		redir := link
		
		//If there were any path extras passed to us, append them to the redir link
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
		
		//if the hash exists in the link table, issue a '302 Found' to the client with the link URL
		ctx.Redirect(302, redir)	
	} else {
		//No link exists for the hash, so serve a '404 Not Found' error page
		error404(ctx, hash)
	}
}


func reportLink(ctx *web.Context){
	commonTemplate(ctx, "report.html", map[string]string{"title_text":"Report A Link"})
}

func submitReport(ctx *web.Context){
	//TODO: implement actual reporting/flagging/emailing of reported link
	commonTemplate(ctx, "generic.html", map[string]string{"title_text":"Report A Link", "body_text":"This functionality is yet to be implemented"})
}
