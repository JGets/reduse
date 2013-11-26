package main

import(
	"errors"
	"database/sql"
	// "strings"
	_ "github.com/go-sql-driver/mysql"
)

const(
	TINYINT_MAX = 255
)

var dsn string

/*
	Initializes Database information & test for connectivity
	Note: will return an error if a connection to the database cannot be established
	
	Parameters:
		dbname:		The name of the database
		address:	The address of the database
		username:	The username to use with the database
		password:	The password to use with the database
	Returns:
		error:		Any error that was encountered
*/
func initDatabase(dbname, address, username, password string) error {
	//Set up the DSN (Data Source Name) for the database
	dsn = username + ":" + password + "@(" + address + ":3306)/" + dbname
	
	//Open the database
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()
	
	logger.Println("Database opened successfully")
	
	//validate the database connection
	err = db.Ping()
	if err != nil{
		return err
	} else {
		logger.Println("Database connection validated")
	}
	
	return nil
}

/*
	Opens an interface to the database
	Returns:
		*sql.DB:	A pointer to the database interface, or nil if an error was encountered
		error:		Any error that was encountered, or nil
*/
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
		int:	The number of reports against the link, or -1 if it does not exist or an error was encountered
		bool:	false only when there is no row for that hash in the DB, true otherwise (Note: is true even when an error is encountered)
		error:	Any error that was encountered, or nil
*/
func db_linkForHash(hash string) (string, int, bool, error){
	//open the database
	db, err := openDB()
	if err != nil {
		return "", -1, true, err
	}
	defer db.Close()
	
	//call the helper with the database pointer
	return db_linkForHashHelper(db, hash)
}

/*
	Get the link for a given hash from the given database interface.
	NOTE: Check for non-nil returned error BEFORE checking the boolean 'exists' return value
	
	Parameters:
		db:		A pointer to the database interface to use
		hash:	The hash that we are to look for in the DB
	Returns:
		string: The link, or an empty string (if there is no entry for the has in the DB, or an error was encountered)
		int:	The number of reports against the link, or -1 if it does not exist or an error was encountered
		bool:	false only when there is no row for that hash in the DB, true otherwise (Note: is true even when an error is encountered)
		error:	Any error that was encountered, or nil
*/
func db_linkForHashHelper(db *sql.DB, hash string) (string, int, bool, error){
	var link string
	var numReports int
	err := db.QueryRow("SELECT link, numReports FROM links WHERE hash=?", hash).Scan(&link, &numReports)
	
	if err != nil {
		if err == sql.ErrNoRows {
				//If we got an error say there was no row, return that no entry exists for this hash, but don't return the error
				return "", -1, false, nil
			} else {
				//any other error, just return an error
				return "", -1, true, err
			}
	}
	
	return link, numReports, true, nil
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

/*
	Add a link to the given database.
	
	Parameters:
		db:		A pointer to the database to add to
		hash:	The short-hash of the link to be added to the DB
		url:	The URL that the link is to redirect to
	Returns:
		error:	Any error that was encountered, or nil
*/
func db_addLinkHelper(db *sql.DB, hash string, url string) error {
	//Check to make sure we aren't trying to add a conflicting link row to the database
	exLink, _, exists, err := db_linkForHashHelper(db, hash)
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
	
	logger.Printf("Added %v => %v to DB\n", hash, url)
	
	return nil
}

// /*
// 	Checks to see if the given domain is blacklisted (ie. contained in the domain table)
// */
// func db_isDomainBlacklisted(domain string) (bool, error){
// 	db, err := openDB()
// 	if err != nil {
// 		return false, err
// 	}
// 	defer db.Close()
//	
// 	return db_isDomainBlacklistedHelper(db, domain)
// }
//
// func db_isDomainBlacklistedHelper(db *sql.DB, domain string) (bool, error){
//	
// 	domain = strings.ToLower(domain)
//	
// 	var bd_domain string
// 	err := db.QueryRow("SELECT domain FROM domain_blacklist WHERE domain=?", domain).Scan(&bd_domain)
//	
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 				//If we got an error say there was no row, return that no entry exists for this domain, but don't return the error
// 				return false, nil
// 			} else {
// 				//any other error, just return an error
// 				return false, err
// 			}
// 	}
//	
// 	return false, nil
// }


/*
	Increments the number of reports against a link in the database
	Parameters:
		hash:	The short hash of the link that was reported
	Returns:
		int:	The new number of reports against the link, or -1 if an error was encountered
		error:	Any error that was encountered, or nil
*/
func db_incrementReportCount(hash string) (int, error){
	db, err := openDB()
	if err != nil {
		return -1, err
	}
	defer db.Close()
	
	return db_incrementReportCountHelper(db, hash)
}

/*
	Increments the number of reports against a link in the given database
	Parameters:
		db:		The database interface to use
		hash:	The short hash of the link that was reported
	Returns:
		int:	The new number of reports against the link, or -1 if an error was encountered
		error:	Any error that was encountered, or nil
*/
func db_incrementReportCountHelper(db *sql.DB, hash string) (int, error){
	_, numReports, exists, err := db_linkForHashHelper(db, hash)
	if err != nil{
		return -1, err
	} else if !exists {
		return -1, errors.New("Trying to increment report count for link that does not exist")
	}
	
	//make sure not to overflow the numReports value in the DB (ie. don't increment it if it's already at the max)
	if numReports >= TINYINT_MAX-1 {
		return -1, nil
	}
	
	//Prepare the update query
	stmt, err := db.Prepare("UPDATE links SET numReports=? WHERE hash=?")
	if err != nil {
		return -1, err
	}
	
	//excecute the insert query
	_, err = stmt.Exec(numReports+1, hash)
	if err != nil {
		return -1, err
	}
	
	return numReports+1, nil
}

