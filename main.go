package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	mapset "github.com/deckarep/golang-set/v2"

	"github.com/gocolly/colly"
)

type Creator struct {
	ID            int
	Name          string
	Platform      string
	DoB           string
	Age           int
	RelatedToURLs []string
	RelatedTo     []int
	Rank          int
	URL           string
}

func main() {
	start := time.Now()
	seeders := []string{
		"https://www.famousbirthdays.com/people/jakenbakelive.html",
		"https://www.famousbirthdays.com/people/tyler-blevins.html",
		"https://www.famousbirthdays.com/people/gigguk.html",
		"https://www.famousbirthdays.com/people/felix-kjellberg.html",
		"https://www.famousbirthdays.com/people/mr-beast.html",
		"https://www.famousbirthdays.com/people/jake-paul.html",
		"https://www.famousbirthdays.com/people/sssniperwolf.html",
		"https://www.famousbirthdays.com/people/david-dobrik.html",
		"https://www.famousbirthdays.com/people/logan-paul.html",
	}

	parseCreators(seeders)
	fmt.Printf("Completed the code process, took: %f seconds\n", time.Since(start).Seconds())
}

func isElementExist(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func parseCreators(seeders []string) {
	log.Println("Parsing creators")
	queue := make(chan string, 30)
	seen := mapset.NewSet[string]()
	creatorChan := make(chan Creator, 30)

	finalCreatorIndex := []Creator{}

	var wg sync.WaitGroup

	maxLimit := 100

	globalCount := 0

	for _, seeder := range seeders {
		queue <- seeder
	}

	go func() {
		for {
			newCreator := <-creatorChan
			newCreator.ID = globalCount
			globalCount++
			// fmt.Println(newCreator)
			finalCreatorIndex = append(finalCreatorIndex, newCreator)
			fmt.Println(globalCount)
		}
	}()

	for {
		if globalCount >= maxLimit {

			fmt.Println("Done. Wrapping up...")
			for _, c := range finalCreatorIndex {
				fmt.Println(c)
			}
			fmt.Println("Final length :", len(finalCreatorIndex))

			return
		}
		url := <-queue
		if !seen.Contains(url) {
			wg.Add(1)
			go parseCreator(url, queue, creatorChan, &wg)
		}
		seen.Add(url)
	}

}

func parseCreator(person_url string, queue chan string, creatorChan chan Creator, wg *sync.WaitGroup) {
	defer (*wg).Done()

	var creator Creator
	creator.URL = person_url
	allowed := []string{"YouTube", "TikTok", "Instagram", "Twitch"}

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 11_2_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.182 Safari/537.36"),
		colly.AllowedDomains("www.famousbirthdays.com"),
		colly.Async(true),
	)

	// Set max Parallelism and introduce a Random Delay
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1,
	})
	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL.String())

	})

	c.OnHTML(".main-info h1", func(e *colly.HTMLElement) {
		creator.Name = e.Text //this is the "stage" name?
	})

	c.OnHTML(".person-title a", func(e *colly.HTMLElement) {
		creator.Platform = strings.Split(e.Text, " ")[1]

	})

	c.OnHTML("a.btn-rank:first-child", func(e *colly.HTMLElement) {

		// log.Println(e.Text)

		//get popularity
		if strings.Contains(e.Text, "Most Popular") {
			arr := strings.Split(e.Text, "#")
			rank, _ := strconv.Atoi(arr[1])
			creator.Rank = rank
		}

		//get age
		if strings.Contains(e.Text, "Year Old") {
			arr := strings.Split(e.Text, " ")
			age, _ := strconv.Atoi(arr[0])
			creator.Age = age
		}

	})

	c.OnHTML("div.also-viewed div.row a", func(e *colly.HTMLElement) {
		url := e.Attr("href")
		// log.Println(url)
		creator.RelatedToURLs = append(creator.RelatedToURLs, url)
		queue <- url

	})

	c.Visit(person_url)

	c.Wait()
	if isElementExist(allowed, creator.Platform) {
		creatorChan <- creator
	}
}
