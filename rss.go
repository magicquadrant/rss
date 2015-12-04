package rss

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"

	"appengine"
	"appengine/urlfetch"
)

// init is called before the application starts.
func init() {
	// Register a handler for / URLs.
	http.HandleFunc("/", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	client := urlfetch.Client(c)

	// pull later from request
	// parse form
	//"http://feeds.feedburner.com/udacity-linear-digressions"
	feedLink := "http://css-tricks.com/feed"

	resp, err := client.Get(feedLink)
	if err != nil {
		c.Errorf("Error on Feed Retrieval: %s", err)
		e := Error{Code: http.StatusInternalServerError, Message: "Sorry Charlie, Trix are for kids!"}
		message, _ := json.Marshal(e)
		http.Error(w, string(message), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	// check for response code
	if resp.StatusCode != 200 {
		c.Errorf("HTTP Status code other than 200 received: %s", resp.StatusCode)
		e := Error{Code: http.StatusInternalServerError, Message: "Stay Frosty!"}
		message, _ := json.Marshal(e)
		http.Error(w, string(message), http.StatusInternalServerError)
		return
	}

	XMLdata, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.Errorf("Error on Reading Request: %s", err)
		e := Error{Code: http.StatusInternalServerError, Message: "Game Over Man, GAME OVER!"}
		message, _ := json.Marshal(e)
		http.Error(w, string(message), http.StatusInternalServerError)
		return
	}

	rss := new(Rss)
	buffer := bytes.NewBuffer(XMLdata)
	decoded := xml.NewDecoder(buffer)

	err = decoded.Decode(rss)
	if err != nil {
		c.Errorf("Error on XML Parse: %s", err)
		e := Error{Code: http.StatusInternalServerError, Message: "Hey Vasquez, have you ever been mistaken for a man?"}
		message, _ := json.Marshal(e)
		http.Error(w, string(message), http.StatusInternalServerError)
	}

	feed := Feed{}
	feed.FeedUrl = feedLink
	feed.Title = rss.Channel.Title
	feed.Description = rss.Channel.Description
	feed.Link = rss.Channel.Link

	total := len(rss.Channel.Items)

	for i := 0; i < total; i++ {
		entry := Entry{}
		entry.Title = rss.Channel.Items[i].Title
		entry.Link = rss.Channel.Items[i].Link

		// strip HTML tags?
		contentSnippet := rss.Channel.Items[i].Description
		if len(contentSnippet) > 119 {
			contentSnippet = contentSnippet[:119]
		}
		entry.Snippet = contentSnippet

		feed.Entries = append(feed.Entries, entry)
	}

	root := Root{}
	root.Feed = feed

	json.NewEncoder(w).Encode(root)
}

type Rss struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
}

type Root struct {
	Error Error `json:"error"`
	Feed  Feed  `json:"feed"`
}

type Feed struct {
	FeedUrl     string  `json:"feedUrl"`
	Title       string  `json:"title"`
	Link        string  `json:"link"`
	Description string  `json:"description"`
	Author      string  `json:"author"`
	Entries     []Entry `json:"entries"`
}

type Entry struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"contentSnippet"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
