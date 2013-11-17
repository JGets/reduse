package main

import(
	"errors"
	"os"
	"bufio"
	"strings"
)

const(
	TABLE_FILE_DELIMITER="\t"
)


type LinkTable struct{
	filename string
	//file os.File
	table map[string]string
}

type Link struct{
	hashCode string
	url string
}


func initLinkTable(filename string) (*LinkTable, error){
	tbl := new(LinkTable)
	
	tbl.filename = filename
	
	tableFile, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	defer tableFile.Close()
	
	if err != nil {
		return nil, err
	}
	
	// tbl.file = tableFile
	
	tbl.table = make(map[string]string)
	
	
	err = tbl.populateLinkTable(tableFile)
	
	if err != nil {
		return nil, err
	}
	
	return tbl, nil	
}

func (lt *LinkTable) populateLinkTable(file *os.File) error{
	
	//read in the file, line by line
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if scanner.Err() != nil {
		return scanner.Err()
	}
	
	for _, val := range lines {
		if val != "" {
			spl := strings.Split(val, TABLE_FILE_DELIMITER)
			
			if len(spl) == 2 {
				lt.table[spl[0]] = spl[1]
			} else {
				logger.Printf("WARNING: Could not parse line from LinkTable file: '%v'", val)
			}
			
		}
	}
	
	return nil
}




func (lt *LinkTable) getTable() (map[string]string){
	return lt.table
}


func (lt *LinkTable) linkForHash(hash string) (string, bool){
	val, exists := lt.table[hash]
	return val, exists
}

func (lt *LinkTable) addLink(hash, url string) error{
	
	exUrl, exists := lt.linkForHash(hash)
	
	
	//Check to make sure we aren't trying to add a conflicting link
	if exists {
		if exUrl != url {
			return errors.New("Error: Attempting to add different url for existing hash. Hash:"+hash+", existing URL:"+exUrl+", new URL:"+url)
		} else {
			//the link already exists, so no need to do anything
			return nil
		}
	}
	//otherwise, no link exists for that hash, so we can add one, Yay!
	
	file, err := os.OpenFile(lt.filename, os.O_RDWR|os.O_APPEND, 0666)
	defer file.Close()
	
	if err != nil {
		return err
	}
	
	str := hash + TABLE_FILE_DELIMITER + url + "\n"
	
	//write the link entry to the persistent file
	_, err = file.WriteString(str)
	if err != nil {
		return err
	}
	
	//add the link entry to the in-memory map
	lt.table[hash] = url
	
	
	return nil
}


