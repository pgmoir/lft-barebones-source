package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

type Cms struct {
	PageTitle   string          `json:"pageTitle,omitempty"`
	Keywords    string          `json:"keywords,omitempty"`
	Description string          `json:"description,omitempty"`
	Template    string          `json:"template,omitempty"`
	Breadcrumbs []CmsBreadcrumb `json:"breadcrumbs"`
	Title       CmsTitle        `json:"title,omitempty"`
	Body        CmsBody         `json:"body,omitempty"`
	Rhnav       CmsRhnav        `json:"rhnav,omitempty"`
}

type CmsBreadcrumb struct {
	Style string `json:"style,omitempty"`
	Title string `json:"title,omitempty"`
	Url   string `json:"url,omitempty"`
	Icon  string `json:"icon,omitempty"`
}

type CmsTitle struct {
	Complex    bool    `json:"complex,omitempty"`
	Title      string  `json:"title,omitempty"`
	Icon       string  `json:"icon,omitempty"`
	Background string  `json:"background,omitempty"`
	Card       CmsCard `json:"card,omitempty"`
}

// CmsBody defines Primarys for 8-column width, Cards for 4-column width
type CmsBody struct {
	Primarys []CmsCard `json:"primarys"`
	Cards    []CmsCard `json:"cards"`
}

type CmsRhnav struct {
	Cards []CmsCard `json:"cards"`
}

type CmsCard struct {
	Template string       `json:"template,omitempty"`
	Style    string       `json:"style,omitempty"`
	Title    string       `json:"title,omitempty"`
	Texts    []string     `json:"texts,omitempty"`
	Links    []CmsLink    `json:"links,omitempty"`
	Link     CmsLink      `json:"link,omitempty"`
	Image    CmsCardImage `json:"image,omitempty"`
	Carousel []CmsCardImage `json:"carousel,omitempty"`
	Wrapped  bool         `json:"wrapped,oninitempty"`
}

type CmsLink struct {
	Title  string `json:"title,omitempty"`
	Url    string `json:"url,omitempty"`
	Icon   string `json:"icon,omitempty"`
	Size   string `json:"size,omitempty"`
	Button string `json:"button,omitempty"`
	Image  CmsCardImage `json:"image,omitempty"`
}

type CmsCardImage struct {
	Url string `json:"url,omitempty"`
	Alt string `json:"alt,omitempty"`
}

type CmsHeader struct {
	Template string `json:"template,omitempty"`
	Level    string `json:"level,omitempty"`
	Value    string `json:"value,omitempty"`
}

type CmsList struct {
	Template string        `json:"template,omitempty"`
	Items    []CmsListItem `json:"items,omitempty"`
}

type CmsListItem struct {
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
	Size        string `json:"size,omitempty"`
	Link        string `json:"link,omitempty"`
}

type CmsTable struct {
	Template string     `json:"template,omitempty"`
	Titles   []string   `json:"titles,omitempty"`
	Rows     [][]string `json:"rows,omitempty"`
}

type CmsBasicList struct {
	Template string   `json:"template,omitempty"`
	Items    []string `json:"items,omitempty"`
}

func main() {
	sourceDir := "tflviews/cms"
	destDir := "feeds"
	prepareDest(destDir)
	processSource(sourceDir)
	copybackExamples()
}

// createJson is primary method to create the output json feed file
func createJson(sourceDir, sourceFile string) {
	newDir := strings.Replace(sourceDir, "tflviews/cms", "feeds", -1)
	newFile := strings.Replace(sourceFile, ".cshtml", ".json", -1)
	filepath := path.Join(sourceDir, sourceFile)

	contents := readSource(filepath)
	//title := getTitle(filepath, contents)
	cms := getCms(contents)

	bcms, err := json.MarshalIndent(cms, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	jsonFile, err := os.Create(path.Join(newDir, newFile))
	jsonFile.Write(bcms)
}

func readSource(filepath string) string {
	filec, _ := ioutil.ReadFile(filepath)
	return string(filec)
}

func getViewBagPropertyValue(property, contents string) string {
	re := regexp.MustCompile(`@{ViewBag.` + property + ` = \"([a-zA-Z0-9-'",\.\s\&]*)\";}`)
	v := re.FindStringSubmatch(contents)
	if len(v) > 1 {
		return strings.TrimSpace(v[1])
	}
	return ""
}

func getIcon(src string) string {
	re := regexp.MustCompile(`/([a-zA-Z0-9]*)-partner.png`)
	v := re.FindStringSubmatch(src)
	if len(v) > 1 {
		return strings.TrimSpace(v[1])
	}
	return ""
}

func divHasClassMatch(d *html.Node, c string) bool {
	if d.Type != html.ElementNode || d.Data != "div" {
		return false
	}
	return hasClassMatch(c, d)
}

func hasClassMatch(a string, n *html.Node) bool {
	for _, b := range n.Attr {
		if b.Key == "class" {
			classes := strings.Split(b.Val, " ")
			for _, c := range classes {
				if c == a {
					return true
				}
			}
		}
	}
	return false
}

func hasClassSimilar(a string, n *html.Node) bool {
	for _, b := range n.Attr {
		if b.Key == "class" && strings.Contains(b.Val, a) {
			return true
		}
	}
	return false
}

func idOfNode(id string, n *html.Node) bool {
	for _, b := range n.Attr {
		if b.Key == "id" && b.Val == id {
			return true
		}
	}
	return false
}

func getCms(contents string) Cms {
	cms := Cms{}
	processContents(contents, &cms)
	return cms
}

// processContents walks through the html node tree, looking for specific elements
// and parsing them into the title, main body and right hand navigation
func processContents(contents string, cms *Cms) {
	d, err := html.Parse(strings.NewReader(contents))
	if err != nil {
		log.Fatal(err)
	}

	// preset
	cms.Title.Complex = false
	cms.PageTitle = getViewBagPropertyValue("Title", contents)
	cms.Keywords = getViewBagPropertyValue("Keywords", contents)
	cms.Description = getViewBagPropertyValue("Description", contents)
	cms.Template = "page-content"

	var f func(*html.Node)
	f = func(n *html.Node) {
		if divHasClassMatch(n, "breadcrumb-container") {
			processBreadcrumbs(n, cms)
		} else if divHasClassMatch(n, "SSP-landing") {
			cms.Template = "page-ssp"
			processHeroTitle(n, &cms.Title, cms.PageTitle)
		} else if divHasClassMatch(n, "headline-container") {
			processBasicTitle(n, &cms.Title)
		} else if divHasClassMatch(n, "section-image-container") {
			processComplexTitleImage(n, &cms.Title)
		} else if divHasClassMatch(n, "section-overview") {
			processComplexTitle(n, &cms.Title)
		} else if divHasClassMatch(n, "main") {
			processBody(n, cms)
		} else if divHasClassMatch(n, "aside") {
			processRhnav(n, &cms.Rhnav)
		} else {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
	}
	f(d)
}

func processBreadcrumbs(d *html.Node, c *Cms) {
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "li" {
			b := getBreadcrumb(n)
			if len(b.Title) > 0 {
				c.Breadcrumbs = append(c.Breadcrumbs, b)
			}
		} else {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
	}
	f(d)
}

func getBreadcrumb(d *html.Node) CmsBreadcrumb {
	bc := CmsBreadcrumb{}
	bc.Style = getClass(d)
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			bc.Url = getHrefFromAtag(n)
		}
		if n.Type == html.TextNode && len(strings.TrimSpace(n.Data)) > 0 {
			bc.Title = strings.TrimSpace(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(d)
	return bc
}

func getClass(d *html.Node) string {
	for _, a := range d.Attr {
		if a.Key == "class" {
			return a.Val
		}
	}
	return ""
}

func getPageTemplate(d *html.Node) string {
	if hasClassMatch("content-container", d) {
		return "page-content"
	}
	return "page-cards"
}

func processBody(d *html.Node, c *Cms) {
	c.Body = CmsBody{Primarys: []CmsCard{}, Cards: []CmsCard{}}

	var f func(*html.Node)
	f = func(n *html.Node) {
		//fmt.Println("pf", n.Type, n.Data, n.Attr)
		if divHasClassMatch(n, "news-teaser") {
			addNewsCard(n, &c.Body, true)
		} else if divHasClassMatch(n, "search-filter") {
			addSearchFilter(n, &c.Body, true)
		} else if divHasClassMatch(n, "article-teaser") {
			cd := getArticleTeaserCard(n, "intro")
			c.Body.Cards = append(c.Body.Cards, cd)
		} else if divHasClassMatch(n, "on-this-page") || divHasClassMatch(n, "content-container") {
			addContentToBody(n, &c.Body, "content")
		} else if divHasClassMatch(n, "section-overview") {
			addSectionOverview(n, &c.Body)
		} else if divHasClassMatch(n, "video-gallery-wrapper") {
			cd := CmsCard{Template: "video"}
			c.Body.Primarys = append(c.Body.Primarys, cd)
		} else if divHasClassMatch(n, "related-links") {
			cd := addListGroups(n)
			c.Body.Primarys = append(c.Body.Primarys, cd)
		} else if divHasClassMatch(n, "vertical-button-container") {
			cl := getListGroupCard(n, "listgroup", "nocard")
			if len(cl.Links) > 0 {
				c.Body.Primarys = append(c.Body.Primarys, cl)
			}
		} else if divHasClassMatch(n, "module-grid") {
			c.Template = "page-cards"
			processModuleGrid(n, c)
		} else if divHasClassMatch(n, "image-box") {
			cd := getImageCard(n)
			c.Body.Primarys = append(c.Body.Primarys, cd)
		} else {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
	}
	f(d)
}

func addSectionOverview(d *html.Node, c *CmsBody) {
	cd := CmsCard{}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "h2" {
			//fmt.Println("here3")
			t := n.FirstChild
			if t != nil {
				cd.Title = strings.TrimSpace(t.Data)
			}
		} else if n.Type == html.ElementNode && n.Data == "p" {
			p := getParagraph(n)
			if len(p) > 0 {
				cd.Texts = append(cd.Texts, p)
			}
		} else if n.Type == html.TextNode {
			p := strings.TrimSpace(n.Data)
			if len(p) > 0 {
				cd.Texts = append(cd.Texts, p)
			}
		} else {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
	}
	f(d)
	c.Primarys = append(c.Primarys, cd)
}

func processModuleGrid(d *html.Node, c *Cms) {
	var f func(*html.Node)
	f = func(n *html.Node) {
		if divHasClassMatch(n, "news-teaser") {
			addNewsCard(n, &c.Body, false)
		} else if divHasClassMatch(n, "call-to-action-button") {
			addCta(n, &c.Body)
		} else if divHasClassMatch(n, "related-links") {
			cd := addListGroups(n)
			c.Body.Cards = append(c.Body.Cards, cd)
		} else if divHasClassMatch(n, "vertical-button-container") {
			cl := getListGroupCard(n, "listgroup", "nocard")
			if len(cl.Links) > 0 {
				c.Body.Cards = append(c.Body.Cards, cl)
			}
		} else if divHasClassMatch(n, "follow-social") {
			cd := getSocialMediaCard(n)
			c.Body.Cards = append(c.Body.Cards, cd)	
		} else if n.Type == html.ElementNode && n.Data == "a" && hasClassMatch("twitter-timeline", n) {
			cd := getTwitterTimelineCard(n)
			c.Body.Cards = append(c.Body.Cards, cd)
		} else {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
	}
	f(d)
}

// <div class="call-to-action-button">
// 	<a href="/modes/driving/ultra-low-emission-zone/check-your-vehicle" class="secondary-button no-hover" >Check your vehicle</a>
// </div>
func addCta(d *html.Node, c *CmsBody) {
	cd := CmsCard{Template: "cta"}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			if hasClassMatch("secondary-button", n) {
				cd.Style = "outline-primary"
			}
			cd.Link.Url = getHrefFromAtag(n)
			cd.Link.Title = strings.TrimSpace(n.FirstChild.Data)
		} else {
			for cn := n.FirstChild; cn != nil; cn = cn.NextSibling {
				f(cn)
			}
		}
	}
	f(d)
	c.Cards = append(c.Cards, cd)
}

func processBasicTitle(n *html.Node, c *CmsTitle) {
	var f func(*html.Node)
	f = func(cn *html.Node) {
		if cn.Type == html.ElementNode && cn.Data == "h1" {
			fc := cn.FirstChild
			if fc != nil && fc.Type == html.TextNode && fc.Data != "" {
				c.Title = fc.Data
			}
		} else if cn.Type == html.ElementNode && cn.Data == "div" && hasClassMatch("heading-logo", cn) {
			fc := cn.FirstChild
			if fc != nil && (fc.Type != html.ElementNode || fc.Data == "img") {
				fc = fc.NextSibling
			}
			if fc != nil && fc.Type == html.ElementNode && fc.Data == "img" {
				for _, a := range fc.Attr {
					if a.Key == "src" {
						c.Icon = getIcon(a.Val)
						break
					}
				}
			}
		} else {
			for cx := cn.FirstChild; cx != nil; cx = cx.NextSibling {
				f(cx)
			}
		}
	}
	f(n)
}

func getTitleText(d *html.Node, t string) string {
	pt := t
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode && n.Data != "@ViewBag.Title" {
			pt = n.Data
		} else {
			for cn := n.FirstChild; cn != nil; cn = cn.NextSibling {
				f(cn)
			}
		}
	}
	f(d)
	return pt
}

func processHeroTitle(d *html.Node, c *CmsTitle, t string) {
	c.Complex = true
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "h1" {
			c.Title = getTitleText(n, t)
		} else if n.Type == html.ElementNode && n.Data == "div" && hasClassMatch("heading-logo", n) {
			fc := n.FirstChild
			if fc != nil && (fc.Type != html.ElementNode || fc.Data == "img") {
				fc = fc.NextSibling
			}
			if fc != nil && fc.Type == html.ElementNode && fc.Data == "img" {
				for _, a := range fc.Attr {
					if a.Key == "src" {
						c.Icon = getIcon(a.Val)
						break
					}
				}
			}
		} else {
			for cn := n.FirstChild; cn != nil; cn = cn.NextSibling {
				f(cn)
			}
		}
	}
	f(d)
}



func processComplexTitleImage(n *html.Node, c *CmsTitle) {
	c.Complex = true
	for _, a := range n.Attr {
		if a.Key == "data-highlight-image" {
			c.Background = "https://tfl.gov.uk" + a.Val
			break
		}
	}
}

func addHrefButton(d *html.Node) (string, string) {
	href := ""
	button := ""
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			href = getHrefFromAtag(n)
			button = strings.TrimSpace(n.FirstChild.Data)
		} else {
			for cn := n.FirstChild; cn != nil; cn = cn.NextSibling {
				f(cn)
			}
		}
	}
	f(d)
	return href, button
}



func processComplexTitle(d *html.Node, c *CmsTitle) {
	c.Complex = true
	cd := CmsCard{}
	cd.Style = addStyle(d)
	var f func(*html.Node)
	f = func(n *html.Node) {
		if divHasClassMatch(n, "main") {
			for mn := n.FirstChild; mn != nil; mn = mn.NextSibling {
				if mn.Type == html.ElementNode && mn.Data == "h2" {
					//fmt.Println("here4")
					cd.Title = strings.TrimSpace(mn.FirstChild.Data)
				} else if mn.Type == html.ElementNode && mn.Data == "p" {
					p := getParagraph(mn)
					if len(p) > 0 {
						cd.Texts = append(cd.Texts, p)
					}
				} else if mn.Type == html.TextNode {
					p := strings.TrimSpace(mn.Data)
					if len(p) > 0 {
						cd.Texts = append(cd.Texts, p)
					}
				}
			}
		} else if divHasClassMatch(n, "aside") {
			cd.Link = addLink(n)
		} else {
			for cn := n.FirstChild; cn != nil; cn = cn.NextSibling {
				f(cn)
			}
		}
	}
	f(d)
	c.Card = cd
}

func addLink(d *html.Node) CmsLink {
	l := CmsLink{}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if divHasClassMatch(n, "call-to-action-button") {
			l.Url, l.Button = addHrefButton(n)
		} else {
			for cn := n.FirstChild; cn != nil; cn = cn.NextSibling {
				f(cn)
			}
		}
	}
	f(d)
	return l
}

func addContentToBody(n *html.Node, c *CmsBody, t string) {
	card := CmsCard{Template: t}
	var f func(*html.Node)
	f = func(cn *html.Node) {
		if divHasClassMatch(cn, "article-teaser") {
			cd := getArticleTeaserCard(cn, "intro")
			b, _ := json.Marshal(cd)
			if len(string(b)) > 0 {
				card.Texts = append(card.Texts, string(b))
			}
		} else if divHasClassMatch(cn, "video-gallery-wrapper") {
			cd := getVideoCard(cn)
			b, _ := json.Marshal(cd)
			if len(string(b)) > 0 {
				card.Texts = append(card.Texts, string(b))
			}
		} else if divHasClassMatch(cn, "gallery-lite-wrap") {
			cd := getCarouselCard(cn)
			b, _ := json.Marshal(cd)
			if len(string(b)) > 0 {
				card.Texts = append(card.Texts, string(b))
			}
		} else if divHasClassMatch(cn, "image-box") {
			cd := getImageCard(cn)
			b, _ := json.Marshal(cd)
			if len(string(b)) > 0 {
				card.Texts = append(card.Texts, string(b))
			}
		} else if divHasClassMatch(cn, "multi-document-download-container") {
			cl := getListGroupCard(cn, "listgroup", "nocard")
			if len(cl.Links) > 0 {
				b, _ := json.Marshal(cl)
				if len(string(b)) > 0 {
					card.Texts = append(card.Texts, string(b))
				}
			}
		} else if divHasClassMatch(cn, "table-container") {
			card.Texts = append(card.Texts, AddTableToContents(cn))
		} else if cn.Type == html.ElementNode && (cn.Data == "h2" || cn.Data == "h3") {
			h := AddHeaderToContents(cn)
			if len(h) > 0 {
				card.Texts = append(card.Texts, h)
			}
		} else if cn.Type == html.ElementNode && cn.Data == "p" {
			t := getParagraph(cn)
			if len(t) > 0 {
				card.Texts = append(card.Texts, t)
			}
		} else if cn.Type == html.ElementNode && cn.Data == "ul" {
			card.Texts = append(card.Texts, addListToContents(cn))
		} else if divHasClassMatch(cn, "module-grid") {
			processModuleGridInContent(cn, &card)	
		} else {
			for cx := cn.FirstChild; cx != nil; cx = cx.NextSibling {
				f(cx)
			}
		}
	}
	f(n)
	c.Cards = append(c.Cards, card)
}

func processModuleGridInContent(d *html.Node, card *CmsCard) {
	sr := CmsCard{Template: "start-row"}
	sc, _ := json.Marshal(sr)
	card.Texts = append(card.Texts, string(sc))

	var f func(*html.Node)
	f = func(n *html.Node) {
		if divHasClassMatch(n, "news-teaser") {
			cd := getNewsCard(n, false)
			cd.Wrapped = true;
			b, _ := json.Marshal(cd)
			if len(string(b)) > 0 {
				card.Texts = append(card.Texts, string(b))
			}
		} else if divHasClassMatch(n, "call-to-action-button") {
			//addCta(n, &c.Body)
		} else if divHasClassMatch(n, "related-links") {
			cd := addListGroups(n)
			cd.Wrapped = true;
			b, _ := json.Marshal(cd)
			if len(string(b)) > 0 {
				card.Texts = append(card.Texts, string(b))
			}
		} else if divHasClassMatch(n, "vertical-button-container") {
			cd := getListGroupCard(n, "listgroup", "nocard")
			if len(cd.Links) > 0 {
				cd.Wrapped = true;
				b, _ := json.Marshal(cd)
				if len(string(b)) > 0 {
					card.Texts = append(card.Texts, string(b))
				}
			}
		} else if divHasClassMatch(n, "follow-social") {
			cd := getSocialMediaCard(n)
			cd.Wrapped = true;
			b, _ := json.Marshal(cd)
			if len(string(b)) > 0 {
				card.Texts = append(card.Texts, string(b))
			}
	} else if n.Type == html.ElementNode && n.Data == "a" && hasClassMatch("twitter-timeline", n) {
			cd := getTwitterTimelineCard(n)
			cd.Wrapped = true;
			b, _ := json.Marshal(cd)
			if len(string(b)) > 0 {
				card.Texts = append(card.Texts, string(b))
			}
	} else {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
	}
	f(d)

	er := CmsCard{Template: "end-row"}
	ec, _ := json.Marshal(er)
	card.Texts = append(card.Texts, string(ec))
}


func getParagraph(d *html.Node) string {
	t := ""
	// fmt.Println(d.Data, d.FirstChild.Data)
	// if d.FirstChild.NextSibling != nil {
	// 	fmt.Println("nsd", d.FirstChild.NextSibling.Data)
	// }
	// if d.FirstChild.FirstChild != nil {
	// 	fmt.Println("fcd", d.FirstChild.FirstChild.Data)
	// }
	for n := d.FirstChild; n != nil; n = n.NextSibling {
		if n.Type == html.ElementNode && n.Data == "p" {
			t += getParagraph(n)
		} else if n.Type == html.ElementNode && n.Data == "a" {
			t += GetAnchorTag(n)
		} else if n.Type == html.ElementNode && n.Data == "br" {
			t += "<br />"
		} else if n.Type == html.ElementNode && n.Data == "strong" {
			t += GetStrongTag(n)
		} else {
			t += strings.TrimSpace(n.Data)
		}
	}
	return t
}

func getParagraph2(d *html.Node) string {
	t := ""
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			t += GetAnchorTag(n)
		} else if n.Type == html.ElementNode && n.Data == "br" {
			t += "<br />"
		} else if n.Type == html.ElementNode && n.Data == "strong" {
			t += GetStrongTag(n)
		} else if strings.TrimSpace(n.Data) != "" {
			t += strings.TrimSpace(n.Data)
		} else {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
	}
	f(d)
	return t
}

func AddHeaderToContents(n *html.Node) string {
	if n.FirstChild == nil {
		return ""
	}
	l := "h3"
	if n.Data == "h3" {
		l = "h4"
	}
	fc := n.FirstChild
	if fc.Type != html.TextNode && fc.NextSibling != nil {
		fc = fc.NextSibling
	}
	value := strings.TrimSpace(fc.Data)
	cmsHeader := CmsHeader{Template: "htag", Value: value, Level: l}
	b, _ := json.Marshal(cmsHeader)
	return string(b)
}

func getListGroupCard(n *html.Node, t, s string) CmsCard {
	cd := CmsCard{Template: t, Style: s, Links: []CmsLink{}}
	var f func(*html.Node)
	f = func(cn *html.Node) {
		if cn.Type == html.ElementNode && cn.Data == "a" {
			getListGroup(cn, &cd)
		} else {
			for cx := cn.FirstChild; cx != nil; cx = cx.NextSibling {
				f(cx)
			}
		}
	}
	f(n)
	return cd
}

func getLinkIcon(d *html.Node, u string) string {
	if hasClassMatch("pdf", d) || hasClassMatch("xlsx", d) {
		return "download"
	} else if strings.HasPrefix(u, "http") {
		return "external"
	} else {
		return "internal"
	}
}

func getListGroup(n *html.Node, c *CmsCard) {
	l := CmsLink{Title: ""}
	l.Url = getHrefFromAtag(n)
	l.Icon = getLinkIcon(n, l.Url)

	var f func(*html.Node)
	f = func(cn *html.Node) {
		if cn.Type == html.TextNode && len(l.Title) == 0 {
			l.Title = strings.TrimSpace(cn.Data)
		} else if divHasClassMatch(cn, "document-download-text") {
			l.Icon = "download"
			l.Title = getDownloadProperty(cn)
		} else if divHasClassMatch(cn, "document-download-attachment") {
			l.Size = getDownloadProperty(cn)
		} else if divHasClassMatch(cn, "document-download-image") {
			fmt.Println(cn.Data, cn.Type)
			l.Image = getDownloadImage(cn)
		} else {
			for cx := cn.FirstChild; cx != nil; cx = cx.NextSibling {
				f(cx)
			}
		}
	}
	f(n)
	c.Links = append(c.Links, l)
}

func addListToContents(d *html.Node) string {
	c := CmsBasicList{Template: "bulletlist"}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "li" {
			i := getParagraph(n)
			if len(i) > 0 {
				c.Items = append(c.Items, i)
			}
		} else {
			for cn := n.FirstChild; cn != nil; cn = cn.NextSibling {
				f(cn)
			}
		}
	}
	f(d)
	b, _ := json.Marshal(c)
	return string(b)
}

func getListItemText(d *html.Node) string {
	n := d.FirstChild
	if n == nil {
		return ""
	}
	return strings.TrimSpace(n.Data)
}

func AddTableToContents(n *html.Node) string {
	cmsTable := CmsTable{Template: "table"}
	AddTitlesToTable(n, &cmsTable)
	var f func(*html.Node)
	f = func(cn *html.Node) {
		if cn.Type == html.ElementNode && cn.Data == "tr" {
			AddRowToTable(cn, &cmsTable)
		} else {
			for cx := cn.FirstChild; cx != nil; cx = cx.NextSibling {
				f(cx)
			}
		}
	}
	f(n)
	b, _ := json.Marshal(cmsTable)
	return string(b)
	//	return "{ \"template\": \"table\", \"titles\": [ \"Country\", \"Currency Code\", \"Currency\" ], \"rows\": [ [ \"Country\", \"Currency Code\", \"Currency\" ], [ \"Country\", \"Currency Code\", \"Currency\" ] ] }"
}

func AddTitlesToTable(n *html.Node, c *CmsTable) {
	var f func(*html.Node)
	f = func(cn *html.Node) {
		if cn.Type == html.ElementNode && cn.Data == "th" {
			t := cn.FirstChild
			if t.Type == html.ElementNode && t.Data == "strong" {
				t = t.FirstChild
			}
			c.Titles = append(c.Titles, strings.TrimSpace(t.Data))
		} else {
			for cx := cn.FirstChild; cx != nil; cx = cx.NextSibling {
				f(cx)
			}
		}
	}
	f(n)
}

func AddRowToTable(n *html.Node, c *CmsTable) {
	s := []string{}
	var f func(*html.Node)
	f = func(cn *html.Node) {
		if cn.Type == html.ElementNode && cn.Data == "td" {
			v := ""
			v = getParagraph(cn)
			s = append(s, v)
		} else {
			for cx := cn.FirstChild; cx != nil; cx = cx.NextSibling {
				f(cx)
			}
		}
	}
	f(n)
	c.Rows = append(c.Rows, s)
}

func getDownloadProperty(n *html.Node) string {
	for cx := n.FirstChild; cx != nil; cx = cx.NextSibling {
		if cx.Type == html.ElementNode && cx.Data == "p" {
			return getParagraph(cx)
		}
	}
	return ""
}

func getDownloadImage(n *html.Node) CmsCardImage {
	i := CmsCardImage{}
	for cx := n.FirstChild; cx != nil; cx = cx.NextSibling {
		fmt.Println(cx.Data, cx.Type)
		if cx.Type == html.ElementNode && cx.Data == "img" {
			for _, a := range cx.Attr {
				if a.Key == "src" {
					i.Url = "https://tfl.gov.uk" + a.Val
				} else if a.Key == "alt" {
					i.Alt = strings.TrimSpace(a.Val)
				}
			}
			return i
		}
	}
	return i
}

func getArticleTeaserCard(n *html.Node, t string) CmsCard {
	card := CmsCard{Template: t}
	var f func(*html.Node)
	f = func(cn *html.Node) {
		if cn.Type == html.TextNode {
			t := strings.TrimSpace(cn.Data)
			if len(t) > 0 {
				card.Texts = append(card.Texts, t)
			}
		} else {
			for cx := cn.FirstChild; cx != nil; cx = cx.NextSibling {
				f(cx)
			}
		}
	}
	f(n)
	return card
}

func getSocialMediaCard(d *html.Node) CmsCard {
	//fmt.Println(d.Type, d.Data, d.Attr)
	c := CmsCard{Template: "socialmedianotsure"}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if divHasClassMatch(n, "email") {
			c.Template = "emailupdates"
		} else if divHasClassMatch(n, "twitter") {
			c.Template = "twitterupdates"
			t := getTwitterHandle(n)
			c.Link = CmsLink{Title: t}
		} else {
			for cn := n.FirstChild; cn != nil; cn = cn.NextSibling {
				f(cn)
			}
		}
	}
	f(d)
	return c
}

func getTwitterHandle(d *html.Node) string {
	th := ""
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					th = strings.Replace(a.Val, "https://twitter.com/", "", -1)
					break
				}
			}
		} else {
			for cn := n.FirstChild; cn != nil; cn = cn.NextSibling {
				f(cn)
			}
		}
	}
	f(d)
	return "@" + th
}

func getVideoCard(d *html.Node) CmsCard {
	cd := CmsCard{Template: "video"}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "li" && hasClassMatch("gallery-thumb", n) {
			for _, a := range n.Attr {
				if a.Key == "data-youtubeid" {
					cd.Style = a.Val
				} else if a.Key == "data-video-caption" {
					cd.Title = a.Val
				}
			}
		} else {
			for cn := n.FirstChild; cn != nil; cn = cn.NextSibling {
				f(cn)
			}
		}
	}
	f(d)
	return cd
}

func getCarouselCard(d *html.Node) CmsCard {
	cd := CmsCard{Template: "carousel"}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "li" && hasClassMatch("gallery-item", n) {
			cd.Carousel = append(cd.Carousel, getCarouselImage(n))
		} else {
			for cn := n.FirstChild; cn != nil; cn = cn.NextSibling {
				f(cn)
			}
		}
	}
	f(d)
	return cd
}

func getCarouselImage(d *html.Node) CmsCardImage {
	cd := CmsCardImage{}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			for _, a := range n.Attr {
				if a.Key == "src" {
					cd.Url = "https://tfl.gov.uk" + a.Val
				} else if a.Key == "alt" {
					cd.Alt = strings.TrimSpace(a.Val)
				}
			}
		} else if n.Type == html.ElementNode && n.Data == "span" && hasClassMatch("figure-description", n) {
			t := getParagraph(n)
			if len(t) > 0 {
				cd.Alt = t
			}
		} else {
			for cn := n.FirstChild; cn != nil; cn = cn.NextSibling {
				f(cn)
			}
		}
	}
	f(d)
	return cd
}

func getImageCard(d *html.Node) CmsCard {
	cd := CmsCard{Template: "image"}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			for _, a := range n.Attr {
				if a.Key == "src" {
					cd.Image = CmsCardImage{Url: "https://tfl.gov.uk" + a.Val}
				} else if a.Key == "alt" {
					cd.Image.Alt = strings.TrimSpace(a.Val)
				}
			}
		} else {
			for cn := n.FirstChild; cn != nil; cn = cn.NextSibling {
				f(cn)
			}
		}
	}
	f(d)
	return cd
}

func GetStrongTag(cn *html.Node) string {
	fc := cn.FirstChild
	if fc == nil {
		return ""
	}
	s := "<strong>"
	s += cn.FirstChild.Data
	s += "</strong> "
	return s
}

func GetAnchorTag(cn *html.Node) string {
	fc := cn.FirstChild
	if fc == nil {
		return ""
	}
	s := " <a "
	for _, a := range cn.Attr {
		s += a.Key + "=\"" + a.Val + "\""
	}
	s += ">"
	s += cn.FirstChild.Data
	s += "</a> "
	return s
}

func CleanText(s string) string {
	return strings.TrimSpace(s)
}

func addStyle(d *html.Node) string {
	if hasClassMatch("visitor-centres", d) {
		return "card-visitor"
	} else if hasClassMatch("overground", d) {
		return "card-overground"
	} else if hasClassMatch("tube", d) {
		return "card-tube"
	} else if hasClassMatch("bus", d) {
		return "card-bus"
	}
	return ""
}

func addSearchFilter(d *html.Node, c *CmsBody, p bool) {
	cd := CmsCard{Template: "searchfilter"}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "h3" {
			cd.Title = n.FirstChild.Data
		} else if n.Type == html.ElementNode && n.Data == "p" {
			t := getParagraph(n)
			if len(t) > 0 {
				cd.Texts = append(cd.Texts, t)
			}
		} else if n.Type == html.TextNode && strings.HasPrefix(strings.TrimSpace(n.Data), "@") {
			cd.Link = getSearchFilterAction(n.Data)
		} else {
			for cn := n.FirstChild; cn != nil; cn = cn.NextSibling {
				f(cn)
			}
		}
	}
	f(d)
	if p {
		c.Primarys = append(c.Primarys, cd)
	} else {
		c.Cards = append(c.Cards, cd)
	}
}

func getSearchFilterAction(d string) CmsLink {
	return CmsLink{
		Url:    "https://tfl.gov.uk/disambiguation",
		Button: "Go"}
}

func getUrl(u string) string {
	re := regexp.MustCompile(`(intcmp=[0-9]*)`)
	v := re.FindStringSubmatch(u)
	if len(v) > 1 {
		u = strings.Replace(u, v[0], "", 1)
		ls := u[len(u)-1:]
		if ls == "?" || ls == "&" {
			u = u[0 : len(u)-1]
		}
	}
	return u
}

func addNewsCard(d *html.Node, c *CmsBody, p bool) {
	cd := getNewsCard(d, p)
	if p {
		c.Primarys = append(c.Primarys, cd)
	} else {
		c.Cards = append(c.Cards, cd)
	}
}

func getNewsCard(d *html.Node, p bool) CmsCard {
	ci := CmsCardImage{"", ""}
	cd := CmsCard{Image: ci}
	if p {
		cd.Style = "card-top "
	}	
	cd.Style = cd.Style + addStyle(d)
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					cd.Link.Url = getUrl(a.Val)
					if strings.HasPrefix(cd.Link.Url, "https://tfl.gov.uk") {
						cd.Link.Url = strings.Replace(cd.Link.Url, "https://tfl.gov.uk", "", -1)
					} else if !strings.HasPrefix(cd.Link.Url, "/") {
						cd.Link.Icon = "external"
					}
				}
				if a.Key == "data-img" {
					cd.Image.Url = "https://tfl.gov.uk" + a.Val
				}
			}
			for cn := n.FirstChild; cn != nil; cn = cn.NextSibling {
				if cn.Type == html.ElementNode && cn.Data == "img" {
					for _, a := range cn.Attr {
						if a.Key == "src" {
							if cd.Image.Url == "" {
								cd.Image.Url = "https://tfl.gov.uk" + a.Val
							}
						}
					}
				}
				if divHasClassMatch(cn, "text-link") {
					addTextLinkToCard(cn, &cd)
				}
			}
		}
		if divHasClassMatch(n, "text-link") {
			addTextLinkToCard(n, &cd)
		}
	}
	for cx := d.FirstChild; cx != nil; cx = cx.NextSibling {
		f(cx)
	}
	return cd
}

func addTextLinkToCard(d *html.Node, cd *CmsCard) {
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && (n.Data == "h2" || n.Data == "h3") {
			//fmt.Println("here2")
			cd.Title = n.FirstChild.Data
		} else if divHasClassMatch(n, "multi-document-download-container") {
			lc := getListGroupCard(n, "", "")
			for _, l := range lc.Links {
				cd.Links = append(cd.Links, l)
			}
		} else if n.Type == html.ElementNode && n.Data == "p" {
			t := getParagraph(n)
			if len(t) > 0 {
				cd.Texts = append(cd.Texts, t)
			}
		} else if n.Type == html.TextNode {
			t := strings.TrimSpace(n.Data)
			if len(t) > 0 {
				cd.Texts = append(cd.Texts, t)
			}
		} else {
			for cn := n.FirstChild; cn != nil; cn = cn.NextSibling {
				f(cn)
			}
		}
	}
	f(d)
}

func processRhnav(n *html.Node, c *CmsRhnav) {

	var f func(*html.Node)
	f = func(cn *html.Node) {
		if cn.Type == html.ElementNode && hasClassMatch("share-widget-wrapper", cn) {
			c.Cards = append(c.Cards, CmsCard{Template: "socialmediashare"})
		} else if divHasClassMatch(cn, "related-links") {
			cd := addListGroups(cn)
			c.Cards = append(c.Cards, cd)
		} else if cn.Type == html.ElementNode && cn.Data == "div" && hasClassSimilar("service-board", cn) {
			c.Cards = append(c.Cards, CmsCard{Template: "serviceboard"})
		} else if divHasClassMatch(cn, "journey-planner-widget") {
			c.Cards = append(c.Cards, CmsCard{Template: "planajourney"})
		} else if divHasClassMatch(cn, "contact-info-box") {
			c.Cards = append(c.Cards, CmsCard{Template: "contactus"})
		} else if divHasClassMatch(cn, "advert-tile") {
			c.Cards = append(c.Cards, CmsCard{Template: "advert"})
		} else if divHasClassMatch(cn, "advert-tile") {
			c.Cards = append(c.Cards, CmsCard{Template: "advert"})
		} else if divHasClassMatch(cn, "follow-social") {
			cd := getSocialMediaCard(n)
			c.Cards = append(c.Cards, cd)
		} else if divHasClassMatch(cn, "fact-box-wrapper") {
			addFact(cn, c)
		} else if cn.Type == html.ElementNode && cn.Data == "a" && hasClassMatch("twitter-timeline", cn) {
			cd := getTwitterTimelineCard(cn)
			c.Cards = append(c.Cards, cd)
		} else if cn.Type == html.ElementNode && idOfNode("right-hand-nav", cn) {
			addRhNavigation(cn, c)
		} else {
			for cnxt := cn.FirstChild; cnxt != nil; cnxt = cnxt.NextSibling {
				f(cnxt)
			}
		}
	}

	f(n)
}

// <div id="right-hand-nav" class="expandable-list moving-source-order" role="navigation" aria-labelledby="sub-menu-heading">
//     <a href="/modes/driving/ultra-low-emission-zone" class="heading">
//         <h2 id="sub-menu-heading">Ultra Low Emission Zone<span class="visually-hidden"> navigation</span></h2>
//     </a>
//     <ul>
//         <li><div class="link-wrapper"><a href="/modes/driving/ultra-low-emission-zone/ulez-where-and-when">ULEZ: Where and when</a></div></li>
//         <li><div class="link-wrapper"><a href="/modes/driving/ultra-low-emission-zone/why-we-need-ulez">Why we need the ULEZ</a></div></li>
//         <li><div class="link-wrapper"><a href="/modes/driving/ultra-low-emission-zone/ways-to-meet-the-standard">ULEZ standards</a></div></li>
//         <li class="parent"><div class="link-wrapper"><a href="/modes/driving/ultra-low-emission-zone/check-your-vehicle">Check your vehicle</a></div></li>
//         <li><div class="link-wrapper"><a href="/modes/driving/ultra-low-emission-zone/discounts-and-exemptions">Discounts & exemptions</a></div></li>
//         <li><div class="link-wrapper"><a href="/modes/driving/ultra-low-emission-zone/cars">Cars</a></div></li>
//         <li><div class="link-wrapper"><a href="/modes/driving/ultra-low-emission-zone/vans-minibuses-and-more">Vans, minibuses and more</a></div></li>
//         <li><div class="link-wrapper"><a href="/modes/driving/ultra-low-emission-zone/motorcycles-mopeds-and-more">Motorcycles, mopeds and more</a></div></li>
//         <li><div class="link-wrapper"><a href="/modes/driving/ultra-low-emission-zone/larger-vehicles">Lorries, coaches and larger vehicles</a></div></li>
//     </ul>
// </div>
// This does not take any account of parent nodes, etc.
func addRhNavigation(d *html.Node, c *CmsRhnav) {
	cd := CmsCard{Template: "listgroup", Style: "foldernav" }
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			cd.Link.Url = getHrefFromAtag(n)
		} else if n.Type == html.ElementNode && n.Data == "h2" {
			cd.Title = strings.TrimSpace(n.FirstChild.Data)
			cd.Link.Title = cd.Title
		} else if divHasClassMatch(n, "link-wrapper") {
			cl := CmsLink{}
			a := n.FirstChild
			if a != nil {
				cl.Url = getHrefFromAtag(a)
				cl.Title = strings.TrimSpace(a.FirstChild.Data)
				cd.Links = append(cd.Links, cl)
			}
		}
		for cn := n.FirstChild; cn != nil; cn = cn.NextSibling {
			f(cn)
		}
	}
	f(d)
	c.Cards = append(c.Cards, cd)
}

func addFact(d *html.Node, c *CmsRhnav) {
	cd := CmsCard{}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "span" {
			cd.Title = strings.TrimSpace(n.FirstChild.Data)
		} else if n.Type == html.ElementNode && n.Data == "p" {
			t := getParagraph(n)
			if len(t) > 0 {
				cd.Texts = append(cd.Texts, t)
			}
		} else {
			for cn := n.FirstChild; cn != nil; cn = cn.NextSibling {
				f(cn)
			}
		}
	}
	f(d)
	c.Cards = append(c.Cards, cd)
}

func getHrefFromAtag(n *html.Node) string {
	for _, a := range n.Attr {
		if a.Key == "href" {
			return strings.TrimSpace(a.Val)
		}
	}
	return ""
}

func getTwitterTimelineCard(n *html.Node) CmsCard {
	cd := CmsCard{Template: "twitter"}
	cd.Link.Url = getHrefFromAtag(n)
	tn := n.FirstChild
	if tn.Type == html.TextNode && len(tn.Data) > 0 {
		cd.Title = strings.TrimSpace(tn.Data)
	}
	return cd
}

func addListGroups(n *html.Node) CmsCard {
	cd := CmsCard{Template: "listgroup"}
	var f func(*html.Node)
	f = func(cn *html.Node) {
		if cn.Type == html.ElementNode && cn.Data == "h3" {
			tn := cn.FirstChild
			if tn.Type == html.TextNode && len(tn.Data) > 0 {
				cd.Title = strings.TrimSpace(tn.Data)
			}
		} else if divHasClassMatch(cn, "vertical-button-container") {
			addLinksToCard(cn, &cd)
		} else {
			for cnxt := cn.FirstChild; cnxt != nil; cnxt = cnxt.NextSibling {
				f(cnxt)
			}
		}
	}
	f(n)
	return cd
}

// 	cd.Links = append(cd.Links, CmsLink{Title: "Explore London tickets", Url: "/modes/river/explore-london-tickets", Icon: "internal" })
func addLinksToCard(n *html.Node, c *CmsCard) {
	for cn := n.FirstChild; cn != nil; cn = cn.NextSibling {
		l := CmsLink{}
		if cn.Type == html.ElementNode && cn.Data == "a" {
			l.Url = getHrefFromAtag(cn)
			l.Icon = getLinkIcon(cn, l.Url)
			t := cn.FirstChild
			if t.Type == html.TextNode && len(t.Data) > 0 {
				l.Title = strings.TrimSpace(t.Data)
			}
			c.Links = append(c.Links, l)
		}
	}
}

func prepareDest(destDir string) {
	os.RemoveAll(destDir)
	os.Mkdir(destDir, 0777)
}

func copybackExamples() {
	copybackExample("holdfeeds/index.json", "feeds/travel-information/visiting-london/index1.json")
	copybackExample("holdfeeds/santander-cycles.json", "feeds/modes/cycling/santander-cycles1.json")
	copybackExample("holdfeeds/paj-index.json", "feeds/plan-a-journey/index.json")
	copybackExample("holdfeeds/ti-ssp-index.json", "feeds/travel-information/stations-stops-and-piers/index.json")
}

func copybackExample(sourceFile, destFile string) {
	os.Remove(destFile)
	os.Link(sourceFile, destFile)
}

func processSource(dir string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if file.IsDir() {
			newDir := path.Join(dir, file.Name())
			createDir(newDir)
			processSource(newDir)
		} else {
			fmt.Println(dir, file.Name())
			createJson(dir, file.Name())
		}
	}
}

// createDir creates a new directory to match the source for the output feeds
func createDir(dir string) {
	newDir := strings.Replace(dir, "tflviews/cms", "feeds", -1)
	if _, err := os.Stat(newDir); os.IsNotExist(err) {
		os.Mkdir(newDir, 0777)
	}
}

// func renderNode(n *html.Node) string {
// 	var buf bytes.Buffer
// 	w := io.Writer(&buf)
// 	html.Render(w, n)
// 	return buf.String()
// }
