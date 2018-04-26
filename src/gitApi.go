//print enable : 25
//print disable : 21

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
    "fmt"
    "strconv"
    //"sort"
)

// This program use the OAuth authentication.
// Please put your Github API token here
var token = ""

// global map to store languages name and number of line for each
var Lmap = make(map[string]int)

// if set to True, the display is activated
var print = true

/*
This function sends the request to the github api and returns the response
OAuth authentication is used
*/
func request(url string) *http.Response {
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        log.Fatal(err)
    }

    req.SetBasicAuth(token, "x-oauth-basic")

    client := http.Client{}
    res, err := client.Do(req)
    if err != nil {
        log.Println("StatusCode:", res.StatusCode)
        log.Fatal(err)
    }

    //fmt.Println(reflect.TypeOf(res))  //get type of the res variable. Need reflect import
    return res
}


/*
This function check if there is any error.
In case of error, it prints the error and exits the program
*/
func checkError(err error) {
    if err != nil {
        log.Fatal(err)
    }
}


/*
This function parses the string languages_url form the json response.
It maps the language name as key and the number of line as value
*/
func parseStringLanguage(str string)  map[string]int{

    if len(str) == 2 {  //avoid repo without any language
        return nil
    }

    var m map[string]int
    m = make(map[string]int)

    i := 0
    mark1 := -1
    mark2 := -1
    var language_name string
    var number_line int
    var err error
    for i < len(str) {
        c := str[i]

        if c == '"' && mark1 == -1 {  // begining of the language name
            mark1 = i + 1
        }

        if c == '"' && mark1 != -1 && mark2 == -1 && i > mark1 {  // end of the language name
            mark2 = i
            language_name = str[mark1:mark2]
            //fmt.Printf("\t%s: ", language_name)
            mark1 = i + 2 // begining of the number
            mark2 = -1
            i += 1 // shift in the string, restart after the ':'
        }

        if c == '}' || (c == ',' && mark1 != -1 && mark2 == -1 && i > mark1) {
            mark2 = i
            number_line, err = strconv.Atoi(str[mark1:mark2])
            checkError(err)
            //fmt.Printf("%d lines\n", number_line)
            m[language_name] = number_line
            Lmap[language_name] = Lmap[language_name] + number_line
            mark1 = i + 2 // begining of the language name
            mark2 = -1
            i += 1 // shift in the string, restart after the '""'
        }


        i += 1
    }

    return m

}



func main() {
	var err error

    // var url = "https://api.github.com/users/iAmoric"
    var repository_url = "https://api.github.com/repositories"
    var nb int
    //var repository_url = "https://api.github.com/user/repos?per_page=10"
    var languages_url string

	// do request to http api
    var res = request(repository_url)

	// read json data
	jsonResponce, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
    checkError(err)

	// parse json
    type Owner struct {
        Login string `json:"login"`
    }

	type Repository struct {
		Name string `json:"name"`
		Html_url string `json:"html_url"`
        Description string `json:"description"`
        Owner Owner `json:"owner"`
        Languages_url string `json:"languages_url"`
        Languages map[string]int
	}

    var repositories []Repository

	err = json.Unmarshal([]byte(jsonResponce), &repositories)
    checkError(err)

    // iterate over each repositories
    nb = len(repositories)
    for i := 0; i < nb; i++ {
        if print {
            fmt.Printf("Name: %s\n", repositories[i].Name)
            fmt.Printf("Url: %s\n", repositories[i].Html_url)
            fmt.Printf("Description: %s\n", repositories[i].Description)
            fmt.Printf("Owner: %s\n", repositories[i].Owner.Login)
        }

        languages_url = repositories[i].Languages_url

        // request http api
    	res = request(languages_url)

        // read json data
        jsonResponce, err = ioutil.ReadAll(res.Body)
    	res.Body.Close()
        checkError(err)

        // parse string and print languages map
        if print {
            fmt.Printf("Languages:\n")
        }
        repositories[i].Languages = parseStringLanguage(string(jsonResponce))
        for k := range repositories[i].Languages {
            if print {
                fmt.Printf("\t%s: %d\n", k, repositories[i].Languages[k])
            }
        }

        if print {
            fmt.Println("\n --------------------- \n");
        }

    }

    if print {
        fmt.Println("Number of lines of each language in each folder:")
    }
    // TODO optimize these loops
    for k := range Lmap {
        if print {
            fmt.Printf("%s: %d lines\n", k, Lmap[k])
        }
        for i := 0; i < nb; i++ {
            for kk := range repositories[i].Languages {
                if kk == k {
                    if print {
                        fmt.Printf("\t - %s : ", repositories[i].Html_url)
                        fmt.Printf("%d lines\n", repositories[i].Languages[kk])
                    }
                }
            }
        }
    }


}
