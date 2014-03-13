package main

import(
	"html"
	"html/template"
	"crypto/md5"
	"io"
	"encoding/base32"
	"strings"
	"errors"
	"net"
	"net/url"
	"net/http"
	"net/mail"
	"math/rand"
	"strconv"
	"regexp"
	"fmt"

	"time"
	
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

	if args == nil {
		args = make(map[string]string)
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

func showDisabled(ctx *web.Context){
	commonTemplate(ctx, 
				   "generic.html", 
				   map[string]string{"title_text":"Link Disabled", 
									 "body_text":"The link you are trying to access has been disabled.",
									 })
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



	if devMode {
		time.Sleep(1000 * time.Millisecond)
	}


	
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
	
	sendEmailToAdmins("Redu.se Internal Error", err.Error())
	
	ctx.WriteHeader(500)
	
	commonTemplate(ctx,
				   "generic.html",
				   map[string]string{"title_text":"500 Internal Server Error",
				 					 "body_text":err.Error(),
				 					 })
}



func contactPage(ctx *web.Context) {
	// CAPTCHA length will be in [CAPTCHA_MIN_LENGTH, CAPTCHA_MIN_LENGTH + CAPTCHA_VARIANCE]
	captchaId := captcha.NewLen(CAPTCHA_MIN_LENGTH + rand.Intn(CAPTCHA_VARIANCE + 1))
	
	
	commonTemplate(ctx,
				   "contact.html",
				   map[string]string{"title_text":"Contact Us",
									 "captcha_id":captchaId, 
									 "captcha_soln_min_length":strconv.Itoa(CAPTCHA_MIN_LENGTH),
									 "captcha_soln_max_length":strconv.Itoa(CAPTCHA_MIN_LENGTH + CAPTCHA_VARIANCE),
									 "user_url":ctx.Params["url"],
									 })
}

func contactPageError(ctx *web.Context, capId string, errorMsg string, comment string, usrEmail string, usrName string){
	commonTemplate(ctx,
					   "contact.html",
					   map[string]string{"title_text":"Contact Us",
										 "captcha_id":capId, 
										 "captcha_soln_min_length":strconv.Itoa(CAPTCHA_MIN_LENGTH),
										 "captcha_soln_max_length":strconv.Itoa(CAPTCHA_MIN_LENGTH + CAPTCHA_VARIANCE),
										 "error_msg":errorMsg,
										 "user_comment":comment,
									 	 "user_email":usrEmail,
									 	 "user_name":usrName,
										 })
}

func submitContact(ctx *web.Context) {
	capId := ctx.Params["captcha_id"]
	capSoln := ctx.Params["captcha_soln"]
	
	usrNameStr := ctx.Params["contact_user_name"]
	usrEmailStr := ctx.Params["contact_user_email"]
	comment := ctx.Params["contact_comment"]
	
	//Make sure the user filled out the form
	if usrNameStr == "" {
		contactPageError(ctx, capId, "Please Enter your name", comment, usrEmailStr, usrNameStr)
	} else if usrEmailStr == "" {
		contactPageError(ctx, capId, "You must provide your email address", comment, usrEmailStr, usrNameStr)
		return
	} else if comment == "" {
		contactPageError(ctx, capId, "You must provide a comment as to why you are contacting us", comment, usrEmailStr, usrNameStr)
		return
	} else if capSoln == "" {
		contactPageError(ctx, capId, "You must provide a solution to the CAPTCHA", comment, usrEmailStr, usrNameStr)
		return
	}
	
	//Verify the user's CAPTCHA solution
	goodCapSoln, reason, err := goodCaptchaSolution(ctx, capId, capSoln)
	if err != nil {
		internalError(ctx, err)
		return
	} else if !goodCapSoln {
		captchaId := captcha.NewLen(CAPTCHA_MIN_LENGTH + rand.Intn(CAPTCHA_VARIANCE + 1))
		contactPageError(ctx, captchaId, reason, comment, usrEmailStr, usrNameStr)
		return
	}


	//verify the user's email address:
	emailAddr, err := mail.ParseAddress(usrEmailStr)

	if err != nil {
		internalError(ctx, err)
		return
	} else if emailAddr == nil || emailAddr.Address != usrEmailStr {
		captchaId := captcha.NewLen(CAPTCHA_MIN_LENGTH + rand.Intn(CAPTCHA_VARIANCE + 1))
		contactPageError(ctx, captchaId, "The email address you provided appears to be invalid.", comment, usrEmailStr, usrNameStr)
		return
	}


	subject := "Contact Request to Redu.se Admins"
	body := "<strong>User Name:</strong> " + escapeHTML(usrNameStr) + "<br/>"
	body += "<strong>User Email:</strong> " + escapeHTML(emailAddr.String()) + "<br/>"
	body += "<strong>User Comment:</strong><div style=\"padding-left:15px;\">" + escapeHTML(comment) + "</div>"

	err = sendHTMLEmailToAdmins(subject, body)
	if err != nil {
		internalError(ctx, err)
		return
	}

	commonTemplate(ctx, "generic.html", map[string]string{"title_text":"Thank You", "body_text":"Your contact request was submitted"})

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
	if strings.ToLower(u.Scheme) == "http" || strings.ToLower(u.Scheme) == "https"{	//Accepted URL schemes that end with "://"
		//check to make sure the URL contains a host
		if u.Host == ""{
			return URL_EMPTY_HOST, false, nil
		}
		
		//Check if the URL actually exists (ie. it points to a real webserver)
		resp, err := http.Get(u.String())
		if err != nil {
			logger.Printf("Could not verify existence of URL: %v\n", err)
			return URL_DOES_NOT_EXIST, false, nil
		}
		//close the response's body on return
		defer resp.Body.Close()
		
		//if we did not get a '200 OK' response, reject the URL
		if resp.StatusCode != http.StatusOK {
			logger.Printf("Response for URL existence check: %v\n", resp.StatusCode)
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

	//make the domian & scheme all lowercase (consolidate equivalent URLs to the same hash)
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)
	
	return u.String(), true, nil
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
		string:	A string containing the error text as to why the solution was not accepted, or nil
		error:	Any error that was encountered
*/
func goodCaptchaSolution(ctx *web.Context, id, soln string) (bool, string, error) {
	//make sure we were given a non-empty ID
	if id == "" {
		return false, "INTERNAL ERROR", errors.New("Attempting to verify CAPTCHA with empty ID")
	} else if soln == "" {		//Make sure they actually answered the CAPTCHA
		return false, "You must enter a solution to the CAPTCHA to generate a short link", nil
	} else if !captcha.VerifyString(ctx.Params["captcha_id"], soln) {	//They didn't give a correct solution
		return false, "The solution to the CAPTCHA that you entered was incorrect", nil
	}
	//The user gave us a correct solution to the CAPTCHA
	return true, "", nil
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
	
	urlStr := ctx.Params["url"]

	//Verify the user's CAPTCHA solution
	capId := ctx.Params["captcha_id"]
	capSoln := ctx.Params["captcha_soln"]
	goodCapSoln, reason, err := goodCaptchaSolution(ctx, capId, capSoln)
	if err != nil {
		internalError(ctx, err)
		return
	} else if !goodCapSoln {
		captchaId := captcha.NewLen(CAPTCHA_MIN_LENGTH + rand.Intn(CAPTCHA_VARIANCE + 1))
		commonTemplate(ctx,
					   "home.html",
					   map[string]string{"title_text":"",
										 "captcha_id":captchaId, 
										 "captcha_soln_min_length":strconv.Itoa(CAPTCHA_MIN_LENGTH),
										 "captcha_soln_max_length":strconv.Itoa(CAPTCHA_MIN_LENGTH + CAPTCHA_VARIANCE),
										 "error_msg":reason,
										 "user_url":urlStr,
										 })

		return
	}
	
	
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
		val, _, exists, err := db_linkForHash(testHash)
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
	link, numReports, exists, err := db_linkForHash(upperHash)
	if err != nil {
		//There was an error in the database
		internalError(ctx, errors.New("Database Error: "+err.Error()))
	} else if exists {
		//Check to see if the link has been flagged for review
		if numReports >= NUM_REPORTS_TO_FLAG {
			
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
			
			flaggedLink(ctx, hash, redir)
			return
			
		} else {
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
		}	
	} else {
		//No link exists for the hash, so serve a '404 Not Found' error page
		error404(ctx, hash)
	}
}

/*
	handles when a user attempts to visit a link that has been reported enough times by other users to
	be flagged - ie. serves a page telling the user that the link has been flagged
	Parameters:
		ctx:	The context of the request
		hash:	The hash of the link that the user was trying to access
		target:	The target URL of the link
*/
func flaggedLink(ctx *web.Context, hash string, target string){
	
	commonTemplate(ctx,
				   "flaggedLink.html",
				   map[string]string{"link_hash":hash,
				   					 "destination_url":target,
				 					 })
}

/*
	Serves a page with the form for users to report a link
*/
func reportLink(ctx *web.Context){
	// CAPTCHA length will be in [CAPTCHA_MIN_LENGTH, CAPTCHA_MIN_LENGTH + CAPTCHA_VARIANCE]
	captchaId := captcha.NewLen(CAPTCHA_MIN_LENGTH + rand.Intn(CAPTCHA_VARIANCE + 1))
	
	commonTemplate(ctx,
				   "report.html",
				   map[string]string{"title_text":"Report A Link",
									 "captcha_id":captchaId, 
									 "captcha_soln_min_length":strconv.Itoa(CAPTCHA_MIN_LENGTH),
									 "captcha_soln_max_length":strconv.Itoa(CAPTCHA_MIN_LENGTH + CAPTCHA_VARIANCE),
									 "user_url":ctx.Params["url"],
									 })
}

/*
	trims an IP address of any extras (ie. port, square brackets on an IPv6)
*/
func trimIPAddress(rawIP string) (string, error) {
	
	raw := rawIP
	
	//check to see if it matches the format of an IPv6 address
	isIPv6, err := regexp.MatchString("\\[([0-9a-fA-F]{0,4}:){1,7}[0-9a-fA-F]{0,4}\\].*", raw)
	if err != nil {
		return "", err
	}
	if isIPv6 {
		
		//check if the IPv6 addr also has a port
		hasPort, err := regexp.MatchString("\\[.+\\]:.+", raw)
		if err != nil {
			return "", err
		}
		if hasPort {
			//and remove the port
			raw = strings.TrimRight(raw, ":1234567890")
		}
		
		//remove the brackets surrounding the actual IP address
		raw = strings.TrimPrefix(raw, "[")
		raw = strings.TrimSuffix(raw, "]")
		
		
	} else {	
		//check for IPv4 address, with stict range checking as well
		isIPv4, err := regexp.MatchString("(([0-9]|[0-9][0-9]|[01][0-9][0-9]|2[0-5][0-5])\\.){3}([0-9]|[0-9][0-9]|[01][0-9][0-9]|2[0-5][0-5])(:[0-9]+)?", raw)
		if err != nil {
			return "", err
		}
		if isIPv4 {
			//less strict IP range check, but see if there is a :PORT on the end as well
			hasPort, err := regexp.MatchString("([0-9]{1,3}\\.){3}[0-9]{1,3}:[0-9]+", raw)
			if err != nil {
				return "", err
			}
			if hasPort {
				raw = strings.TrimRight(raw, "1234567890")	//trim the port number off the right
				raw = strings.TrimSuffix(raw, ":")			//and trim the colon (must do seperately, otherwise part of the ip addr will be trimmed too)
			}
		}
	}
	
	
	return raw, nil
}


func submitReportUserError(ctx *web.Context, capId string, linkId string, comment string, reason string){
	commonTemplate(ctx,
				   "report.html",
				   map[string]string{"title_text":"Report A Link",
									 "captcha_id":capId, 
									 "captcha_soln_min_length":strconv.Itoa(CAPTCHA_MIN_LENGTH),
									 "captcha_soln_max_length":strconv.Itoa(CAPTCHA_MIN_LENGTH + CAPTCHA_VARIANCE),
									 "user_url":linkId,
									 "user_comment":comment,
									 "error_msg":reason,
									 })
}

/*
	recieves a link report submission, verifies the CAPTCHA, makes a report struct, and attempts to add it to the database
*/
func submitReport(ctx *web.Context){
	
	capId := ctx.Params["captcha_id"]
	capSoln := ctx.Params["captcha_soln"]
	
	linkId := ctx.Params["linkId"]
	reportTypeString := ctx.Params["reportReason"]
	comment := ctx.Params["report_comment"]
	
	//Make sure the user filled out the form
	if linkId == "" {
		submitReportUserError(ctx, capId, linkId, comment, "You must provide a link to report")
		return
	} else if reportTypeString == "" {
		submitReportUserError(ctx, capId, linkId, comment, "You must select a reason that you are reporting this link")
		return
	} else if comment == "" {
		submitReportUserError(ctx, capId, linkId, comment, "You must provide a comment as to why you are reporting this link")
		return
	} else if capSoln == "" {
		submitReportUserError(ctx, capId, linkId, comment, "You must provide a solution to the CAPTCHA")
		return
	}
	
	//Verify the user's CAPTCHA solution
	goodCapSoln, reason, err := goodCaptchaSolution(ctx, capId, capSoln)
	if err != nil {
		internalError(ctx, err)
		return
	} else if !goodCapSoln {
		captchaId := captcha.NewLen(CAPTCHA_MIN_LENGTH + rand.Intn(CAPTCHA_VARIANCE + 1))
		submitReportUserError(ctx, captchaId, linkId, comment, reason)
		return
	}
	
	//make the hash all uppercase
	upperHash := strings.ToUpper(linkId)
	
	
	// attempt to parse the IP address of the user that made this report
	var rawIP string
	
	if herokuProduction {	// because of Heroku's reverse router system, we need to grab the user's IP from the X-Forwarded-For header
		forwardSlice := ctx.Request.Header["X-Forwarded-For"]	//The client's IP is guaranteed to be the last element
		rawIP = forwardSlice[len(forwardSlice)-1]
	} else {	//otherwise we can just grab the IP from the request
		rawIP = string(ctx.Request.RemoteAddr)
	}
	
	//trim the IP address of any extra stuff (whitespace, portnumber, etc.)
	trimmedIP, err := trimIPAddress(rawIP)
	if err != nil {
		internalError(ctx, err)
		return
	}
	//attempt to parse the IP
	ip := net.ParseIP(trimmedIP)
	if ip == nil {
		internalError(ctx, errors.New("Unable to parse client IP address"))
		return
	}
	ipStr := ip.String()
	
	
	//Generate a new report struct to add to the database
	rep := NewReport(upperHash, ipStr, ReportTypeForString(reportTypeString), comment)
	
	//attempt to add the report to the database
	numReports, exists, err := db_addReport(rep)
	if _, isREE := err.(ReportExistsError); isREE {
		//A report for this link already exists from the user's IP address
		commonTemplate(ctx,
					   "generic.html",
					   map[string]string{"title_text":"Report Exists",
			 							 "body_text":"A report for that link already exists from your IP address.",
			 							 })
	} else if err != nil {
		//any other errors
		internalError(ctx, err)
	} else if !exists {
		//The link doens't exist
		bStr := "The link redu.se/" + linkId + " does not exist." 
		captchaId := captcha.NewLen(CAPTCHA_MIN_LENGTH + rand.Intn(CAPTCHA_VARIANCE + 1))
		submitReportUserError(ctx, captchaId, linkId, comment, bStr)
		return
	}	
	
	
	//If the number of reports has increased over the flag point, send an email to the admins
	if numReports >= NUM_REPORTS_TO_FLAG {
		emailBody := "The following link has been reported by users:<br/>"
		emailBody += "<strong>LinkID:</strong> " + escapeHTML(linkId) + "<br/>"

		target, _, _, err := db_linkForHash(upperHash)
		if err != nil {
			internalError(ctx, err)
			return
		}

		emailBody += "<strong>Target URL:</strong> <a href=\"" + target + "\">" + escapeHTML(target) + "</a><br/><br/>"

		reports, err := db_reportsForHash(upperHash)
		if err != nil {
			internalError(ctx, err)
			return
		}

		for i, v := range reports {
			emailBody += fmt.Sprintf("Report %v of %v:<br/>", i+1, len(reports))
			emailBody += "<div style=\"padding-lefT:15px;\">" + escapeHTML(v.String()) + "</div><br/>"
		}



		err = sendHTMLEmailToAdmins("Link Reported", emailBody)
		if err != nil{
			internalError(ctx, err)
			return
		}
	}
	
	//Tell the user that their report has been recieved
	commonTemplate(ctx, "generic.html", map[string]string{"title_text":"Thank You", "body_text":"Your report was submitted"})
}


func escapeHTML(str string) string {
	str = html.EscapeString(str)

	str = strings.Replace(str, "\n", "<br/>", -1)
	str = strings.Replace(str, "\t", "&nbsp;&nbsp;&nbsp;&nbsp", -1)

	return str
}
