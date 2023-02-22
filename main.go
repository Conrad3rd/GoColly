package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

// Language, Serie, Name/ID
var lang string = "ger"
var idMovie string = "8081"
var baseURL string = "https://www.opensubtitles.org"

type tableData struct {
	Season     string
	Episode    string
	Title      string
	SubCounts  string
	SubListURL string
	IMDB       string
}

var subs []tableData

func main() {
	// extractData()
	// downloadFile2(extractData2(extractData()))
	fmt.Println("<<<", extractData2(extractData()))
	// extractData2(extractData())
}

// extractData extracts data from opensubtitles.org
func extractData() (urllist []string) {
	var se string
	var ep string
	c := colly.NewCollector(colly.AllowedDomains("www.opensubtitles.org"))

	c.OnHTML("#search_results > tbody", func(h *colly.HTMLElement) {

		h.ForEach("tr", func(_ int, el *colly.HTMLElement) {
			temp := el.ChildText("td:nth-child(1) > span")
			// fmt.Println(">>", temp)

			// Puts the episode to ep and the season to se
			if _, err := strconv.Atoi(temp); err == nil {
				ep = temp // --> 12

			} else {
				se = temp // --> Season 4
			}

			// separate the word from the number (Season 4) --> (4)
			if strings.Contains(se, "Season") {
				se = strings.Split(se, " ")[1]
				ep = "" // emyte the var to skip. The info come in the next table row
			}

			// add the info to the struct "tableData"
			if se != "" && ep != "" {
				if _, err := strconv.Atoi(temp); err == nil {
					tableData := tableData{
						Season:     se,
						Episode:    ep,
						Title:      el.ChildText("td:nth-child(1) > a"),
						SubListURL: el.ChildAttr("a", "href"),
						SubCounts:  el.ChildText("td:nth-child(2)"),
						IMDB:       el.ChildAttr("td:nth-child(4) > a", "href"),
					}
					if tableData.Title == "" {

						tableData.Title = "noData"
						tableData.SubListURL = "noData"
						tableData.SubCounts = "noData"
						tableData.IMDB = "noData"

					}
					subs = append(subs, tableData)
					// fmt.Println(tableData.SubListURL)

					if tableData.SubListURL == "noData" {
						fmt.Println("noData")

					} else {
						episodeURL, err := url.JoinPath(baseURL, tableData.SubListURL)
						if err != nil {
							log.Fatal(err)
						}
						// fmt.Println(episodeURL)
						urllist = append(urllist, episodeURL)
					}
					// return urllist
					// fmt.Println("<<<<<<<<", urllist)

				}

			}
		})
		// fmt.Println(">>>>", tableData)

		// writeToJson(subs)

	})

	startUrl := fmt.Sprintf("%s/de/ssearch/sublanguageid-%s/idmovie-%s", baseURL, lang, idMovie)
	c.Visit(startUrl)
	return
}

// writeToJson write Season, Episode, Title, SubCounts, SubList-URL, IMDB-URL
// to subs.json file
func writeToJson(subs []tableData) {
	content, err := json.Marshal(subs)
	if err != nil {
		fmt.Println(err.Error())
	}
	os.WriteFile("subs.json", content, 0644)
	fmt.Println("Total subs: ", len(subs))
}

// extracts available entries
var sID []string

func extractData2(url []string) []string {
	c := colly.NewCollector(colly.AllowedDomains("www.opensubtitles.org"))
	// #bt-dwl-bt
	// if only 1 entry available
	c.OnHTML("#bt-dwl-bt", func(h *colly.HTMLElement) {
		sID = nil

		path := h.Attr("href")
		// sID = strings.Split(path, "/")[4]
		sID = append(sID, strings.Split(path, "/")[4])
		fmt.Println(">>>", sID)
		// return sID
		// fmt.Println("only 1", path)
	})

	// #search_results > tbody
	// if more than 1 entry available
	c.OnHTML("#search_results > tbody", func(h *colly.HTMLElement) {
		h.ForEach("tr", func(_ int, el *colly.HTMLElement) {
			sID = nil
			// #main8544733 > strong > a
			// td:nth-child(1) > span
			temp := el.ChildAttr("a", "href")

			if temp != "" {
				// sID = strings.Split(temp, "/")[3]
				sID = append(sID, strings.Split(temp, "/")[3])
				fmt.Println("more then 1", sID)

			}
		})
	})

	// startUrl := fmt.Sprintf("%s/de/ssearch/sublanguageid-%s/idmovie-%s", baseURL, lang, idMovie)
	// c.Visit("https://www.opensubtitles.org/de/search/sublanguageid-ger/imdbid-697784")
	// c.Visit("https://www.opensubtitles.org/de/search/sublanguageid-ger/imdbid-98286")
	for _, v := range url {
		c.Visit(v)

		// fmt.Println(sID)
		// return
	}
	return sID
}

var (
	fileName    string
	fullURLFile string
)

func downloadFile2(id string) {
	// fullURLFile = "http://dl.opensubtitles.org/en/download/sub/7893558"
	fullURLFile = fmt.Sprintf("%s/en/download/sub/%s", baseURL, id)

	// Build fileName from fullPath
	fileURL, err := url.Parse(fullURLFile)
	if err != nil {
		log.Fatal(err)
	}
	path := fileURL.Path
	segments := strings.Split(path, "/")
	fileName = segments[len(segments)-1]

	local := "/home/conrad/Downloads/temp/subs/"
	fileName = filepath.Join(local, fileName)

	// Create blank file
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	// Put content on file
	resp, err := client.Get(fullURLFile)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	size, _ := io.Copy(file, resp.Body)

	defer file.Close()

	fmt.Printf("Downloaded a file %s with size %d\n", fileName, size)
	// fmt.Println(fmt.Sprintf("Downloaded a file %s with size %d", fileName, size))
}
