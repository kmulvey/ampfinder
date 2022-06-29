package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/gocolly/colly"
	log "github.com/sirupsen/logrus"
)

const staticBool = false

const startURL = "https://www.decware.com/cgi-bin/yabb22/YaBB.pl?board=classifieds"

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	log.SetLevel(log.InfoLevel)

	// open the skip file
	log.Info("reading log file")
	processedLog, err := os.OpenFile("scraped.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		log.Fatalf("error opening scraped file: %s", err.Error())
	}
	defer func() {
		var err = processedLog.Close()
		if err != nil {
			log.Fatalf("error closing scraped file: %s", err.Error())
		}
	}()

	var skipMap = getSkipMap(processedLog)

	// Instantiate default collector
	c := colly.NewCollector(
		colly.AllowedDomains("decware.com"),
		colly.AllowedDomains("www.decware.com"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.0.0 Safari/537.36"),
		colly.IgnoreRobotsTxt(),
	)

	// On every a element which has href attribute call callback
	c.OnHTML("html body div#maincontainer div#container center div.seperator table.bordercolor tbody tr > td:nth-child(3)", func(e *colly.HTMLElement) {
		var title = strings.ToLower(strings.TrimSpace(e.Text))
		var _, found = skipMap[title]
		if strings.Contains(title, "taboo") && !found {
			log.Infof("found taboo: %s", title)

			var cmdStr = fmt.Sprintf("notify-send \"%s\" \"%s\" --icon=dialog-information", "Found an amp!", title)
			_, err := exec.Command("bash", "-c", cmdStr).Output()
			if err != nil {
				log.Fatalf("error notify-send: %s", err.Error())
			}

			_, err = processedLog.WriteString(title + "\n")
			if err != nil {
				log.Fatalf("error writing to cache file: %s", err.Error())
			}
		}
	})

	c.OnResponse(func(r *colly.Response) {
		if r.StatusCode != 200 {
			log.Errorf("Something went wrong: status: %d", r.StatusCode)
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Errorf("Something went wrong: status: %d, error: %s", r.StatusCode, err.Error())
	})

	// seed link
	err = c.Visit(startURL)
	if err != nil {
		log.Fatalf("error visting: %s, error: %s", startURL, err.Error())
	}
}

// getSkipMap read the log from the last time this was run and
// puts those filenames in a map so we dont have to process them again
// If you want to reprocess, just delete the file
func getSkipMap(processedImages *os.File) map[string]bool {

	var scanner = bufio.NewScanner(processedImages)
	scanner.Split(bufio.ScanLines)
	var compressedFiles = make(map[string]bool)

	for scanner.Scan() {
		compressedFiles[scanner.Text()] = staticBool
	}

	return compressedFiles
}
