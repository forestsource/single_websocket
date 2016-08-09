package main

import (
	"flag"
	"log"
    "strings"
    "io/ioutil"
	"net/http"
    "encoding/json"
    "html/template"
	"github.com/gorilla/websocket"
)
type location struct{
    lat float32
    long float32
}
type IPInfo struct{
    IP string
    HostName string
    City string
    Region string
    Country string
    Location string `json:"loc"`
    Organization string `json:"org"`
    Postal int
}
var addr = flag.String("addr", "127.0.0.1:8080", "http service address")

var upgrader = websocket.Upgrader{} // use default options

func echo(w http.ResponseWriter, r *http.Request) {
    ip := strings.Split(r.Host,":")[0]
    //日本以外からのアクセスを拒否
    country := getLocationFromIP(ip);
    if country != "JP"{
        return
    }
    //start
	c, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func home(w http.ResponseWriter, r *http.Request) {
    html_str := readHtml("./response.html")
    var homeTemplate = template.Must(template.New("").Parse(html_str))
    log.Printf("From:"+r.Host)
	homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
}
func getLocationFromIP(ip string) string{
    //setting
    baseURL := "http://ipinfo.io/"
    //ip = "8.8.8.8"
    URL := baseURL + ip
    log.Println("URL: "+URL)
    req, _ := http.NewRequest("GET",URL, nil)
    req.Header.Add("User-Agent", "curl/7.43.0")// to get json,not html.
    //start
    client := new(http.Client)
    res, _ := client.Do(req)
    defer res.Body.Close()
    body, _ := ioutil.ReadAll(res.Body)
    //log.Println(string(body))
    //decode json
    var d IPInfo
    json.Unmarshal([]byte(body), &d)
    log.Println("ip: ",d.IP)
    log.Println("Country: ",d.Country)
    return d.Country
}

func readHtml(filepath string)string {
    data, err := ioutil.ReadFile(filepath)
    if err != nil {
        log.Println("ERROR: Can't read file ",filepath)
    }
    //log.Print(string(data))
    return string(data)
}

func main() {
	flag.Parse()
	log.SetFlags(0)
    readHtml("./response.html")
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
