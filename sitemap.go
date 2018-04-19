package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"

	link "github.com/msiadak/gophercises-link"
)

const maxDepth = 3

type URL struct {
	XMLName xml.Name `xml:"url"`
	Loc     string   `xml:"loc"`
}

type urlSorter struct {
	URLs []URL
}

func (us *urlSorter) Len() int {
	return len(us.URLs)
}

func (us *urlSorter) Less(i, j int) bool {
	return us.URLs[i].Loc < us.URLs[j].Loc
}

func (us *urlSorter) Swap(i, j int) {
	us.URLs[i], us.URLs[j] = us.URLs[j], us.URLs[i]
}

func CrawlPageBFS(domain string, path string, sitemap map[string]bool, depth int) error {
	fmt.Printf("Crawling '%s'", path)
	domainURL, err := url.Parse(domain)
	if err != nil {
		return err
	}

	resp, err := http.Get(domain)
	if err != nil {
		return err
	}

	links, err := link.ExtractLinks(resp.Body)
	if err != nil {
		return err
	}

	toCrawl := make([]string, 0, len(links))

	for _, link := range links {
		linkURL, err := domainURL.Parse(link.HREF)
		if err != nil {
			return err
		}

		linkURL.RawQuery = ""
		linkURL.Fragment = ""

		if _, ok := sitemap[linkURL.String()]; !ok && domainURL.Hostname() == linkURL.Hostname() && depth <= maxDepth {
			sitemap[linkURL.String()] = true
			if depth+1 <= maxDepth {
				toCrawl = append(toCrawl, linkURL.String())
			}
		}
	}

	for _, u := range toCrawl {
		err := CrawlPageBFS(domain, u, sitemap, depth+1)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	urlString := os.Args[1]

	rootURL, err := url.Parse(urlString)
	if err != nil {
		log.Fatalf("Couldn't parse URL: '%s'\n%s", urlString, err)
	}

	sitemap := make(map[string]bool)
	sitemap[rootURL.String()] = true

	err = CrawlPageBFS(rootURL.String(), rootURL.String(), sitemap, 0)
	if err != nil {
		log.Fatalf("Couldn't crawl URL: '%s'\n%s", rootURL, err)
	}

	urls := make([]URL, len(sitemap))
	i := 0
	for link := range sitemap {
		urls[i].Loc = link
		i++
	}

	sort.Sort(&urlSorter{urls})

	f, err := os.Create("sitemap.xml")
	if err != nil {
		log.Fatalln("Couldn't create file: 'sitemap.xml'")
	}
	defer f.Close()

	io.WriteString(f, xml.Header)
	io.WriteString(f, `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`+"\n")

	e := xml.NewEncoder(f)
	e.Indent("  ", "  ")
	e.Encode(urls)

	io.WriteString(f, "\n"+`</urlset>`+"\n")
}
