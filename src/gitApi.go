//  synchrone: 25
//  asynchrone : 2.5


package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
    "fmt"
    "strconv"
    "sync"
)

// This program use the OAuth authentication.
// Please put your Github API token here
var token = ""

// global map to store languages name and number of line for each
//var Lmap = make(map[string]int)
var mapLocker = struct{
    sync.RWMutex
    Lmap map[string]int
} {Lmap: make(map[string]int)}

// if set to True, the display is activated
var print = true

// json object
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

    var i = 0
    var mark1 = -1
    var mark2 = -1
    var language_name string
    var number_line int
    var err error
    var m map[string]int

    m = make(map[string]int)

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

            // cast string to int
            number_line, err = strconv.Atoi(str[mark1:mark2])
            checkError(err)
            m[language_name] = number_line

            // make map safe for concurrent use
            mapLocker.Lock()
            mapLocker.Lmap[language_name] = mapLocker.Lmap[language_name] + number_line
            mapLocker.Unlock()

            mark1 = i + 2 // begining of the language name
            mark2 = -1
            i += 1 // shift in the string, restart after the '""'
        }

        i += 1
    }

    return m

}


/*
This function is a goroutine.
It makes a request with the given url to the github API.
Then it parses the response to get the languages name and the number of line
*/
func parseLanguageRouting(url string, i int, wg *sync.WaitGroup) {
    // request http api
    res := request(url)

    // read json data
    jsonResponce, err := ioutil.ReadAll(res.Body)
    res.Body.Close()
    checkError(err)

    // parse string
    repositories[i].Languages = parseStringLanguage(string(jsonResponce))

    // for the barrier
    wg.Done()

}

func main() {
	var err error

    // TODO : get 100 last public repositories
    var repository_url = "https://api.github.com/repositories"
    var nb int

	// do request to http api
    var res = request(repository_url)

	// read json data
	jsonResponce, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
    checkError(err)

	err = json.Unmarshal([]byte(jsonResponce), &repositories)
    checkError(err)

    // iterate over each repositories
    nb = len(repositories)

    // set the barrier for the nb goroutines
    var wg sync.WaitGroup
    wg.Add(nb)

    for i := 0; i < nb; i++ {
        if print {
            fmt.Printf("Name: %s\n", repositories[i].Name)
            fmt.Printf("Url: %s\n", repositories[i].Html_url)
            fmt.Printf("Description: %s\n", repositories[i].Description)
            fmt.Printf("Owner: %s\n", repositories[i].Owner.Login)
        }

        // run goroutine for request and parse responses
        go parseLanguageRouting(repositories[i].Languages_url, i, &wg)

        if print {
            fmt.Println("\n --------------------- \n");
        }

    }

    // wait for all goroutines
    wg.Wait()

    if print {
        fmt.Println("Number of lines of each language in each folder:")
    }

    // TODO optimize these loops
    for k := range mapLocker.Lmap { //each language
        if print {
            fmt.Printf("%s: %d lines\n", k, mapLocker.Lmap[k])
        }
        for i := 0; i < nb; i++ {   // each repository
            for kk := range repositories[i].Languages { // each language of each repository
                if kk == k { // test if the repository has the concerning language
                    if print {
                        fmt.Printf("\t - %s : ", repositories[i].Html_url)
                        fmt.Printf("%d lines\n", repositories[i].Languages[kk])
                    }
                }
            }
        }
    }


}
