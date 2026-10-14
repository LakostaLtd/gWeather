package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Current struct {
	Temperature float64 `json:"temperature_2m"`
}

type Response struct {
	Current Current `json:"current"`
}

func main() {
	resp, err := http.Get("https://api.open-meteo.com/v1/forecast?latitude=43.24&longitude=76.92&current=temperature_2m,relative_humidity_2m,apparent_temperature,wind_speed_10m")

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(resp.Status)

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		fmt.Println(err)
		return
	}

	var data Response
	err = json.Unmarshal(body, &data)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(data.Current.Temperature)
}
