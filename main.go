package main

import (
	"encoding/json"
	"flag"
	xj "github.com/basgys/goxml2json"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type Fuel struct {
	Type         string `json:"-TYPE"`
	Usage        string `json:"-VAL"`
	Percent      string `json:"-PCT"`
	Interconnect string `json:"-IC"`
}

type Inst struct {
	CalculatedTime string `json:"-AT"`
	TotalUsage     string `json:"-TOTAL"`
	FuelUsage      []Fuel `json:"FUEL"`
}

type PowerUsage struct {
	CurrentUsage Inst `json:"INST"`
}

type FuelTable struct {
	FuelTypeTable PowerUsage `json:"GENERATION_BY_FUEL_TYPE_TABLE"`
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

		/*xmlFile, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			w.Write([]byte("# couldnt read!!"))
			return
		}*/

		jsonFile, err := xj.Convert(resp.Body)

		if err != nil {
			panic(err)
		}

		var data FuelTable
		err = json.Unmarshal([]byte(jsonFile.String()), &data)

		if err != nil {
			panic(err)
		}

		prometheus := ""

		prometheus += "# HELP elexon_uk_energy_use_megawatts UK current power usage\n"
		prometheus += "# TYPE elexon_uk_energy_use_megawatts gague\n"

		for _, fuel := range data.FuelTypeTable.CurrentUsage.FuelUsage {
			prometheus += "elexon_uk_energy_use_megawatts{fuel=\"" + fuel.Type + "\"} " + fuel.Usage + "\n"
		}

		prometheus += "\n"
		prometheus += "# HELP elexon_uk_energy_use_percentage UK current power usage percentage\n"
		prometheus += "# TYPE elexon_uk_energy_use_percentage gague\n"

		for _, fuel := range data.FuelTypeTable.CurrentUsage.FuelUsage {
			prometheus += "elexon_uk_energy_use_percentage{fuel=\"" + fuel.Type + "\"} " + fuel.Percent + "\n"
		}

		w.Write([]byte(prometheus))

	})

	http.ListenAndServe("localhost:9991", nil)

}
