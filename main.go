package main

import (
	"github.com/antchfx/htmlquery"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	err                       error
	textBrokerAuthorUsername  string
	textBrokerAuthorPassword  string
	minimumAmountOrder        float64
	userAgent                 string
	alreadyReadOrdersFileName string
	alreadyReadOrders         map[string]bool
)

func init() {
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	textBrokerAuthorUsername = os.Getenv("TEXTBROKER_AUTHOR_USERNAME")
	textBrokerAuthorPassword = os.Getenv("TEXTBROKER_AUTHOR_PASSWORD")
	minimumAmountOrder, err = strconv.ParseFloat(os.Getenv("MINIMUM_EUROS_ORDER"), 8)
	if err != nil {
		log.Fatalln("Error while parsing MINIMUM_EUROS_ORDER in .env : " + err.Error())
	}
	userAgent = os.Getenv("USER_AGENT")
	alreadyReadOrdersFileName = os.Getenv("ALREADY_READ_ORDERS_FILE")
	loadAlreadyReadOrders()
}

func main() {
	var loginParams = map[string]string{
		"action":           "/login/login/ajax_context:1/",
		"params[0][name]":  "email",
		"params[0][value]": textBrokerAuthorUsername,
		"params[1][name]":  "password",
		"params[1][value]": textBrokerAuthorPassword,
		"params[2][name]":  "userType",
		"params[2][value]": "author",
	}

	collector := colly.NewCollector()
	collector.AllowURLRevisit = true
	collector.UserAgent = userAgent

	collector.OnHTML("div.box-wrapper", func(element *colly.HTMLElement) {
		if element.Response.Ctx.Get("mode") == "checkHighestPrice" {
			list := htmlquery.Find(element.DOM.Nodes[0], "//tr[contains(@id, 'tr_')]")
			if len(list) == 0 {
				log.Println("No orders found :)")
				time.Sleep(350 * time.Millisecond)
				queryOrders(collector)
			} else {
				orderID, maxOrderPrice := getMostExpensiveOrder(list)
				log.Printf("The biggest order value is %.2f euros", maxOrderPrice)
				if maxOrderPrice > minimumAmountOrder {
					ctx := colly.NewContext()
					ctx.Put("mode", "pickOrder")
					ctx.Put("orderId", orderID)
					headers := http.Header{}
					headers.Add("X-Requested-With", "XMLHttpRequest")
					err = collector.Request("POST", "https://intern.textbroker.fr/a/inc/headlines_common/show_headline.php", createFormReader(map[string]string{
						"id": orderID,
					}), ctx, headers)
					if err != nil {
						log.Fatal(err)
					}
				} else {
					queryOrders(collector)
				}
			}
		}
	})

	collector.OnResponse(func(r *colly.Response) {
		mode := r.Ctx.Get("mode")
		switch mode {
		case "login":
			log.Println("Login OK")
			queryOrders(collector)
			break
		case "pickOrder":
			addOrderToFile(r.Ctx.Get("orderId"))
			playTone()
			os.Exit(1)
		}

	})

	ctx := colly.NewContext()
	ctx.Put("mode", "login")
	err = collector.Request("POST", "https://intern.textbroker.fr/login/login/ajax_context:1/", createFormReader(loginParams), ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func getMostExpensiveOrder(nodes []*html.Node) (orderId string, orderPrice float64) {
	for _, hNode := range nodes {
		elem_price := htmlquery.FindOne(hNode, ".//td[@id=\"earnings\"]/strong/text()")
		elem_id := htmlquery.FindOne(hNode, ".//a[@class=\"headline_prev\"]/@id")
		if elem_id != nil {
			tmpOrderId := htmlquery.InnerText(elem_id)
			if _, ok := alreadyReadOrders[tmpOrderId]; !ok {
				priceString := strings.TrimSpace(strings.Split(strings.Split(htmlquery.InnerText(elem_price), "-")[1], "€")[0])
				priceFloat, err := strconv.ParseFloat(priceString, 8)
				if err != nil {
					log.Fatal(err)
				}
				if orderPrice < priceFloat {
					orderPrice = priceFloat
					orderId = htmlquery.InnerText(elem_id)
				}
				log.Printf("Found one order with price being %.2f", priceFloat)
			} else {
				log.Println("Found one order already read!")
			}
		} else {
			log.Println("elem_id is null")
		}
	}

	return
}

func queryOrders(collector *colly.Collector) {
	headers := http.Header{}
	headers.Add("X-Requested-With", "XMLHttpRequest")
	ctx := colly.NewContext()
	ctx.Put("mode", "checkHighestPrice")
	err := collector.Request("GET", "https://intern.textbroker.fr/a/order-search.ajax.php?search_headline=1&q=&client_id=&search_cat=0&date_from=&date_through=&which_date=0&order_type_open_order_2=1&order_type_open_order_3=1&order_type_open_order_4=1&fields=0&narrowness=0&=D%C3%A9buter+la+recherche", nil, ctx, headers)
	if err != nil {
		log.Fatal(err)
	}
}

func createFormReader(data map[string]string) io.Reader {
	form := url.Values{}
	for k, v := range data {
		form.Add(k, v)
	}
	return strings.NewReader(form.Encode())
}

func playTone() {
	var streamer beep.StreamSeekCloser
	var format beep.Format

	f, err := os.Open("tone.wav")
	if err != nil {
		log.Fatal(err)
	}
	streamer, format, err = wav.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()
	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		log.Fatal(err)
	}
	speaker.Play(streamer)
	time.Sleep(1 * time.Second)
}

func loadAlreadyReadOrders() {
	alreadyReadOrders = map[string]bool{}
	b, err := ioutil.ReadFile(alreadyReadOrdersFileName)
	if err != nil {
		log.Fatalln(err)
	}

	bString := string(b)

	for _, orderID := range strings.Split(bString, "\n") {
		if orderID != "" {
			alreadyReadOrders[orderID] = true
		}
	}
}

func addOrderToFile(orderID string) {
	var builder strings.Builder

	b, err := ioutil.ReadFile(alreadyReadOrdersFileName)
	if err != nil {
		log.Fatalln(err)
	}

	builder.Write(b)
	builder.WriteString("\n")
	builder.WriteString(orderID)

	err = ioutil.WriteFile(alreadyReadOrdersFileName, []byte(builder.String()), 0644)
	if err != nil {
		log.Fatalln(err)
	}
}
