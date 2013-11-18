package main

import(
	"html/template"
	"crypto/md5"
	"io"
	"encoding/base32"
	"strings"
	"errors"
	"net/url"
	
	"github.com/hoisie/web"
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
	templatePage(ctx, 
				 map[string]string{"template_name":"home.html", 
								   "template_file":"templatePages/home.html",
								   }, 
				 map[string]string{})
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


func generate(ctx *web.Context){
	urlStr := ctx.Params["url"]
	
	//TODO: link validation (ie. make sure it is a valid URL)
	
	
	
	//link must start with http:// or https://
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "http://" + urlStr
	}
	
	
	//Check the domain against the blacklist
	blacklisted, err := isBlacklisted(urlStr)
	if err != nil {
		internalError(ctx, errors.New("Could not check URL against blacklist. ~ " + err.Error()))
		return
	} else if blacklisted {
		blacklistedPage(ctx, urlStr)
		return
	}
	
	
	
	
	
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
	for i := 3; i <= len(hashStr) && collision; i++ {
		testHash = hashStr[:i]
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
	//make the hash all uppercase
	upperHash := strings.ToUpper(hash)
	
	//Check to see if a link exists for the given hash
	link, exists, err := db_linkForHash(upperHash)
	
	if err != nil {
		//There was an error in the database
		internalError(ctx, errors.New("Database Error: "+err.Error()))
	} else if exists {
		//if the hash exists in the link table, issue a '302 Moved Permanently' to the client with the link URL
		ctx.Redirect(302, link)	
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
