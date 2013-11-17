package main

import(
	"html/template"
	"crypto/md5"
	"io"
	"encoding/base32"
	"strings"
	"errors"
	
	"github.com/hoisie/web"
)



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
				 map[string]string{})
}

func error404(ctx *web.Context, url string){
	logger.Printf("404 Error for URL: %v\n", url)
	
	
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


func generate(ctx *web.Context){
	url := ctx.Params["url"]
	
	//TODO: check given url against blacklist
	//TODO: link validation (ie. make sure it is a valid URL)
	
	//logger.Printf("Generating for: %v\n", url)
	
	//link must start with http:// or https://
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
		
		//logger.Println("URL was missing http://")
	}
	
	//Generate a new MD5 hasher, and has the url
	hasher := md5.New()
	io.WriteString(hasher, url)
	hashBytes := hasher.Sum(nil)
	hashStr := base32.StdEncoding.EncodeToString(hashBytes)
	
	
	
	//Check for collisions (ie. different links resulting in the same short-hash), and fix them (by making the )
	var testHash string
	var collision bool = true
	for i := 3; i <= len(hashStr) && collision; i++ {
		testHash = hashStr[:i]
		val, exists := linkTable.linkForHash(testHash)
		
		//If a link does not exist for that hash, OR, a link exists for the hash, but it is the same link, there is no collision
		if !exists || val == url {
			collision = false
		}
	}
	
	//if we have hit the maximum length of the hash, and there is still a collision, throw an error
	if collision {
		internalError(ctx, errors.New("Could not resolve collision. Hash: " + testHash + "    Link: " + url))
		return
	}
	
	finalHash := testHash
	
	//Save the link to the link table
	err := linkTable.addLink(finalHash, url)
	
	
	if err != nil {
		internalError(ctx, err)
		return
	} else {
		body := "Generate short url for " + url
		
		templatePage(ctx,
					 map[string]string{"template_name":"generate.html",
					 				   "template_file":"templatePages/generate.html",
					 				   },
					 map[string]string{"title_text":"Generate URL",
					 				   "body_text":body,
					 				   "link_hash":finalHash,
					 				   })
	}
	
	
	
	
	
	
}

func serveLink(ctx *web.Context, hash string){
	//make the hash all uppercase
	upperHash := strings.ToUpper(hash)
	
	
	link, exists := linkTable.linkForHash(upperHash)
	
	if exists {
		logger.Printf("servering redirect for link: %v", link)
		//if the hash exists in the link table, issue a '302 Moved Permanently' to the client with the link url
		ctx.Redirect(302, link)
		
	} else {
		error404(ctx, hash)
	}
	
}

func listLinks() string{
	var ret = ""
	for key, val := range linkTable.getTable() {
		ret += key + " : " + val + "<br/>"
	}
	
	return ret
}
