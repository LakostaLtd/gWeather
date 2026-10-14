package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const Version = "0.1.1"

// Структура JSON возвращаемая сервером
type Current struct {
	Temperature   float64 `json:"temperature_2m"`
	Humidity      int64   `json:"relative_humidity_2m"`
	Apparent_Temp float64 `json:"apparent_temperature"`
	WindSpeed     float64 `json:"wind_speed_10m"`
}

type Current_Units struct {
	Celsium            string `json:"temperature_2m"`
	HumiditySign       string `json:"relative_humidity_2m"`
	Apparent_Temp_Sign string `json:"apparent_temperature"`
	WindSpeedSign      string `json:"wind_speed_10m"`
}

type Response struct {
	Current       Current       `json:"current"`
	Current_Units Current_Units `json:"current_units"`
}

func main() {
	// Выполняем запрос
	resp, err := http.Get("https://api.open-meteo.com/v1/forecast?latitude=43.24&longitude=76.92&current=temperature_2m,relative_humidity_2m,apparent_temperature,wind_speed_10m")

	if err != nil {
		fmt.Println(err)
		return
	}

	// Если сервер вернул ошибку - выходим
	if resp.StatusCode != 200 {
		fmt.Println("Сервер вернул ответ ", resp.Status)
		return
	}

	// Закрыываем канал и читаем то, что вернул сервер
	defer resp.Body.Close()
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

	fmt.Println("Температура:", data.Current.Temperature, data.Current_Units.Celsium)
	fmt.Println("Ощущается как:", data.Current.Apparent_Temp, data.Current_Units.Apparent_Temp_Sign)
	fmt.Println("Влажность:", data.Current.Humidity, data.Current_Units.HumiditySign)
	fmt.Println("Скорость ветра:", data.Current.WindSpeed, data.Current_Units.WindSpeedSign)
}
