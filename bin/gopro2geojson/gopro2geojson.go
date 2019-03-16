package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	geojson "github.com/paulmach/go.geojson"
	"github.com/stilldavid/gopro-utils/telemetry"
)

func main() {
	inName := flag.String("i", "", "Required: telemetry file to read")
	outName := flag.String("o", "", "Required: geo json file to write")
	flag.Parse()

	if *inName == "" {
		flag.Usage()
		return
	}

	telemFile, err := os.Open(*inName)
	if err != nil {
		fmt.Printf("Cannot access telemetry file %s.\n", *inName)
		os.Exit(1)
	}
	defer telemFile.Close()

	t := &telemetry.TELEM{}
	t_prev := &telemetry.TELEM{}

	var coordinates [][]float64

	for {
		t, err = telemetry.Read(telemFile)
		if err != nil && err != io.EOF {
			fmt.Println("Error reading telemetry file", err)
			os.Exit(1)
		} else if err == io.EOF || t == nil {
			break
		}

		// first full, guess it's about a second
		if t_prev.IsZero() {
			*t_prev = *t
			t.Clear()
			continue
		}

		// process until t.Time
		t_prev.FillTimes(t.Time.Time)
		telems := t_prev.ShitJson()

		for i, _ := range telems {
			if telems[i].GpsAccuracy == 9999 {
				// Invalid GPS accuracy
				continue
			}

			longLat := []float64{telems[i].Longitude, telems[i].Latitude}
			coordinates = append(coordinates, longLat)
		}

		*t_prev = *t
		t = &telemetry.TELEM{}
	}

	jsonFile, err := os.Create(*outName)
	if err != nil {
		fmt.Printf("Cannot make output file %s.\n", *outName)
		os.Exit(1)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("Cannot close json file %s: %s", file.Name(), err)
			os.Exit(1)
		}
	}(jsonFile)

	g := geojson.NewLineStringFeature(coordinates)
	if err := json.NewEncoder(jsonFile).Encode(g); err != nil {
		fmt.Println("Error encoding output json", err)
		os.Exit(1)
	}
}
