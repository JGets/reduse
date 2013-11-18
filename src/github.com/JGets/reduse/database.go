package main

import(
	"errors"
	// "log"
	"database/sql"
	"strings"
	_ "github.com/go-sql-driver/mysql"
)

var dsn string

func initDatabase(dbname, address, username, password string) error {
	//Set up the DSN (Data Source Name) for the database
	dsn = username + ":" + password + "@(" + address + ":3306)/" + dbname
	
	//Open the database
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()
	
	logger.Println("db opened")
	
	//validate the database connection
	err = db.Ping()
	
	if err != nil{
		return err
	} else {
		logger.Println("Database connection validated")
	}
	
	
	return nil
}

func openDB() (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

/*
	Get a map of all the links in the database
	Returns:
		map[srtring]string:
				A map (hash => link) of all links in the database, or nil if an error is encountered
		error:	Any error that was encountered, or nil
*/
func db_getLinkTable() (map[string]string, error){
	//open the database
	db, err := openDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	
	//get all the rows from the "links" table
	rows, err := db.Query("SELECT * FROM links")
	if err != nil {
		return nil, err
	}
	
	//Initialize a map & populate it with all the rows from the table
	var ret = make(map[string]string)
	for rows.Next() {
		var hash, link string
		if err := rows.Scan(&hash, &link); err != nil{
			return nil, err
		}
		ret[hash] = link
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return ret, nil
	
}

/*
	Get the link for a given hash from the database.
	NOTE: Check for non-nil returned error BEFORE checking the boolean 'exists' return value
	
	Parameters:
		hash:	The hash that we are to look for in the DB
	Returns:
		string: The link, or an empty string (if there is no entry for the has in the DB, or an error was encountered)
		bool:	false only when there is no row for that hash in the DB, true otherwise (Note: is true even when an error is encountered)
		error:	Any error that was encountered, or nil
*/
func db_linkForHash(hash string) (string, bool, error){
	//open the database
	db, err := openDB()
	if err != nil {
		return "", true, err
	}
	defer db.Close()
	
	//call the helper with the database pointer
	return db_linkForHashHelper(db, hash)
}


func db_linkForHashHelper(db *sql.DB, hash string) (string, bool, error){
	var link string
	err := db.QueryRow("SELECT link FROM links WHERE hash=?", hash).Scan(&link)
	
	if err != nil {
		if err == sql.ErrNoRows {
				//If we got an error say there was no row, return that no entry exists for this hash, but don't return the error
				return "", false, nil
			} else {
				//any other error, just return an error
				return "", true, err
			}
	}
	
	return link, true, nil
}

/*
	Add a link to the database.
	
	Parameters:
		hash:	The short-hash of the link to be added to the DB
		url:	The URL that the link is to redirect to
	Returns:
		error:	Any error that was encountered, or nil
*/
func db_addLink(hash, url string) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()
	
	return db_addLinkHelper(db, hash, url)
}

func db_addLinkHelper(db *sql.DB, hash string, url string) error {
	//Check to make sure we aren't trying to add a conflicting link row to the database
	exLink, exists, err := db_linkForHashHelper(db, hash)
	if err != nil {
		return err
	} else if exists {
		if url != exLink {
			//A row for this hash exists already, but not with the same redirect link --> ERROR!
			return errors.New("Attempting to add different url for existing hash. Hash:"+hash+", existing URL:"+exLink+", new URL:"+url)
		} else {
			//The link already exists in the database, so no need to do anything
			return nil
		}
	}
	
	//Prepare the insert query
	stmt, err := db.Prepare("INSERT INTO links(hash, link) VALUES(?, ?)")
	if err != nil {
		return err
	}
	
	//excecute the insert query
	_, err = stmt.Exec(hash, url)
	if err != nil {
		return err
	}
	
	return nil
}

/*
	Checks to see if the given domain is blacklisted (ie. contained in the domain)
*/
func db_isDomainBlacklisted(domain string) (bool, error){
	db, err := openDB()
	if err != nil {
		return false, err
	}
	defer db.Close()
	
	return db_isDomainBlacklistedHelper(db, domain)
}

func db_isDomainBlacklistedHelper(db *sql.DB, domain string) (bool, error){
	
	domain = strings.ToLower(domain)
	
	var bd_domain string
	err := db.QueryRow("SELECT domain FROM domain_blacklist WHERE domain=?", domain).Scan(&bd_domain)
	
	if err != nil {
		if err == sql.ErrNoRows {
				//If we got an error say there was no row, return that no entry exists for this domain, but don't return the error
				return false, nil
			} else {
				//any other error, just return an error
				return false, err
			}
	}
	
	return false, nil
}

