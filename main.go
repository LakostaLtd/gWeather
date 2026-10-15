package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

const Version = "0.1.2"

type City struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Country   string  `json:"country"`
}

type GeocodingResponse struct {
	Results []City `json:"results"`
}

// Структура JSON возвращаемая сервером
type Current struct {
	Temperature  float64 `json:"temperature_2m"`
	Humidity     int     `json:"relative_humidity_2m"`
	ApparentTemp float64 `json:"apparent_temperature"`
	WindSpeed    float64 `json:"wind_speed_10m"`
}

type CurrentUnits struct {
	TemperatureUnit  string `json:"temperature_2m"`
	HumiditySign     string `json:"relative_humidity_2m"`
	ApparentTempSign string `json:"apparent_temperature"`
	WindSpeedSign    string `json:"wind_speed_10m"`
}

type Response struct {
	Current      Current      `json:"current"`
	CurrentUnits CurrentUnits `json:"current_units"`
}

func GetCity(name string) (City, error) {
	ApiURL := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1&language=ru&format=json", url.QueryEscape(name))

	rsp, e := http.Get(ApiURL)

	if e != nil {
		return City{}, e
	}

	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return City{}, fmt.Errorf("сервер вернул ответ: %s", rsp.Status)
	}

	body, err := io.ReadAll(rsp.Body)

	// Если ошибка при чтении - выход
	if err != nil {
		return City{}, err
	}

	var geo GeocodingResponse
	err = json.Unmarshal(body, &geo)

	if err != nil {
		return City{}, err
	}

	if len(geo.Results) == 0 {
		return City{}, fmt.Errorf("Город не найден")
	}

	return geo.Results[0], nil
}

func main() {
	fmt.Printf("gWeather v%s", Version)
	fmt.Println("")

	if len(os.Args) < 2 {
		fmt.Println("Использование: gweather <город>")
		return
	}
	city, e := GetCity(os.Args[1])
	if e != nil {
		fmt.Println(e)
		return
	}

	latitude := strconv.FormatFloat(city.Latitude, 'f', -1, 64)
	longitude := strconv.FormatFloat(city.Longitude, 'f', -1, 64)

	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%s&longitude=%s&current=temperature_2m,relative_humidity_2m,apparent_temperature,wind_speed_10m&wind_speed_unit=ms", latitude, longitude)

	// Выполняем запрос
	resp, err := http.Get(url)

	if err != nil {
		fmt.Println(err)
		return
	}

	// Закрываем канал
	defer resp.Body.Close()

	// Если сервер вернул ошибку - выходим
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Сервер вернул ответ:", resp.Status)
		return
	}

	// Читаем то, что вернул сервер

	body, err := io.ReadAll(resp.Body)

	// Если ошибка при чтении - выход
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

	fmt.Println("Погода для:", city.Name, city.Country)
	fmt.Println("Температура:", data.Current.Temperature, data.CurrentUnits.TemperatureUnit)
	fmt.Println("Ощущается как:", data.Current.ApparentTemp, data.CurrentUnits.ApparentTempSign)
	fmt.Println("Влажность:", data.Current.Humidity, data.CurrentUnits.HumiditySign)
	fmt.Println("Скорость ветра:", data.Current.WindSpeed, data.CurrentUnits.WindSpeedSign)
}
