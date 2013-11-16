package main

import(
	"html/template"
	"crypto/md5"
	"io"
	"encoding/base32"
	
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

func generate(ctx *web.Context){
	url := ctx.Params["url"]
	
	
	hasher := md5.New()
	
	io.WriteString(hasher, url)
	
	hashBytes := hasher.Sum(nil)
	
	hashStr := base32.StdEncoding.EncodeToString(hashBytes)
	
	hashStr = hashStr[:5]
	
	
	//TODO
	//need to save to database & make actual link work
	
	
	
	body := "Generate short url for " + url
	hashText := "Hash: " + hashStr
	
	templatePage(ctx,
				 map[string]string{"template_name":"generate.html",
				 				   "template_file":"templatePages/generate.html",
				 				   },
				 map[string]string{"title_text":"Generate URL",
				 				   "body_text":body,
				 				   "hash_text":hashText,
				 				   })
	
	
}

func serveLink(ctx *web.Context, identifier string){
	var invalidLink = true //CHANGE THIS to false (when TODO below is implemented)
	
	//TODO
	//get link from databse & redirect to it
	
	
	if invalidLink {
		error404(ctx, identifier)
	}
	
	
}
