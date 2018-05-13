package main // import "github.com/presbyter/dmhycolly"

import (
	"github.com/gocolly/colly"
	"gopkg.in/mgo.v2"
	"log"
	"colly/models"
	"github.com/gin-gonic/gin/json"
	"fmt"
)

func main() {
	c := colly.NewCollector(
		colly.Async(false),
		colly.MaxDepth(1),
		colly.AllowedDomains("share.dmhy.org"),
	)

	c.OnHTML("table#topic_list>tbody>tr", func(element *colly.HTMLElement) {
		title := element.ChildText("td.title>a")
		pageUrl := element.ChildAttr("td.title>a", "href")
		dbLink := element.ChildAttr("td:nth-child(4)>a", "href")
		writeToMongo(&models.Resource{Title: title, MagnetUrl: dbLink, PageUrl: element.Request.AbsoluteURL(pageUrl)})
	})

	for i := 1; i < 4000; i++ {
		println(i)
		c.Visit(fmt.Sprintf("https://share.dmhy.org/topics/list/page/%d", i))
	}

	c.Wait()
}

func writeToMongo(model *models.Resource) bool {
	jsonBytes, err := json.Marshal(model)
	if err != nil {
		panic(err)
	}
	println(string(jsonBytes))
	session, err := mgo.Dial("mongodb://mongoadmin:Pa$$w0rd@localhost:27017/")
	defer session.Close()
	if err != nil {
		log.Println("连接mongo失败\t", err.Error())
		return false
	}
	db := session.DB("dmhy")
	c := db.C("resources")
	index := mgo.Index{
		Key:        []string{"pageurl"},
		Unique:     true,
		DropDups:   true,
		Background: true, // See notes.
		Sparse:     true,
	}
	c.EnsureIndex(index)
	err = c.Insert(model)
	if err != nil {
		log.Println(err.Error())
	}
	return err != nil
}
