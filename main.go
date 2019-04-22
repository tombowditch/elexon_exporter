package main

import (
	"encoding/xml"
	"flag"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type GenFuelType struct {
	XMLName xml.Name `xml:GENERATION_BY_FUEL_TYPE_TABLE"`
	Tags    []INST   `xml:"INST"`
}

type INST struct {
	AT    string `xml:"AT,attr"`
	Total string `xml:"TOTAL,attr"`
	Value []FUEL `xml:"FUEL"`
}

type FUEL struct {
	TYPE  string `xml:"TYPE,attr"`
	IC    string `xml:"IC,attr"`
	VAL   string `xml:"VAL,attr"`
	PCT   string `xml:"PCT,attr"`
	Value string `xml:",chardata"`
}

var myClient = &http.Client{Timeout: 3 * time.Second}
var apiKey *string

func init() {
	apiKey = flag.String("key", "", "Elexon scripting key")
	flag.Parse()
}

func main() {
	log.Info("hello!")

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		//xmlFile, _ := os.Open("a.xml")
		//defer xmlFile.Close()

		resp, err := http.Get("https://downloads.elexonportal.co.uk/fuel/download/latest?key=" + *apiKey)
		if err != nil {
			w.Write([]byte("# aaa couldnt get url"))
			return
		}

		defer resp.Body.Close()

		xmlFile, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Warnf("Unable to read from API: %s", err.Error())
			http.Error(w, "Unable to read off API", http.StatusInternalServerError)
			return
		}

		var q GenFuelType
		err = xml.Unmarshal(xmlFile, &q)
		if err != nil {
			log.Warnf("Unable to parse XML from API: %s", err.Error())
			http.Error(w, "Unable to parse off API", http.StatusInternalServerError)
			return
		}

		prometheus := ""

		prometheus += "# HELP elexon_uk_energy_use_megawatts UK current power usage\n"
		prometheus += "# TYPE elexon_uk_energy_use_megawatts gauge\n"

		for _, f := range q.Tags[0].Value {
			prometheus += "elexon_uk_energy_use_megawatts{fuel=\"" + f.TYPE + "\"} " + f.VAL + "\n"
		}

		prometheus += "\n"
		prometheus += "# HELP elexon_uk_energy_use_percentage UK current power usage percentage\n"
		prometheus += "# TYPE elexon_uk_energy_use_percentage gauge\n"

		for _, f := range q.Tags[0].Value {
			prometheus += "elexon_uk_energy_use_percentage{fuel=\"" + f.TYPE + "\"} " + f.PCT + "\n"
		}

		w.Write([]byte(prometheus))

	})

	http.ListenAndServe(":9991", nil)

}
