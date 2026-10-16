package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const Version = "0.2.4"

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
type Daily struct {
	Time           []string  `json:"time"`
	TemperatureMax []float64 `json:"temperature_2m_max"`
	TemperatureMin []float64 `json:"temperature_2m_min"`
	PrecipProb     []int     `json:"precipitation_probability_max"`
	WeatherCode    []int     `json:"weather_code"`
	WindSpeedMax   []float64 `json:"wind_speed_10m_max"`
}

type Current struct {
	Time         string  `json:"time"`
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
	Latitude     float64      `json:"latitude"`
	Longitude    float64      `json:"longitude"`
	Current      Current      `json:"current"`
	CurrentUnits CurrentUnits `json:"current_units"`
	Daily        Daily        `json:"daily"`
	Error        bool         `json:"error"`
	Reason       string       `json:"reason"`
}

func weekdayRu(day time.Weekday) string {
	switch day {
	case time.Monday:
		return "Понедельник"
	case time.Tuesday:
		return "Вторник"
	case time.Wednesday:
		return "Среда"
	case time.Thursday:
		return "Четверг"
	case time.Friday:
		return "Пятница"
	case time.Saturday:
		return "Суббота"
	case time.Sunday:
		return "Воскресенье"
	}
	return " "
}

func weatherDescription(code int) string {
	switch code {
	case 0:
		return "Ясно"
	case 1:
		return "Преимущественно ясно"
	case 2:
		return "Переменная облачность"
	case 3:
		return "Пасмурно"
	case 45, 48:
		return "Туман"
	case 51:
		return "Лёгкая морось"
	case 53:
		return "Умеренная морось"
	case 55:
		return "Сильная морось"
	case 56, 57:
		return "Переохлаждённая морось"
	case 61, 63, 65:
		return "Дождь"
	case 66, 67:
		return "Ледяной дождь"
	case 71, 73, 75:
		return "Снегопад"
	case 77:
		return "Снежные зерна"
	case 80, 81, 82:
		return "Ливни"
	case 85, 86:
		return "Снежный ливень"
	case 95:
		return "Гроза"
	case 96:
		return "Гроза с небольшим градом"
	case 99:
		return "Сильная гроза с градом"
	default:
		return "N/A"
	}
}

func GetDefaultCity() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	ConfPath := filepath.Join(dir, "gweather", "config")
	FullConfigName, err := os.ReadFile(ConfPath)
	if err != nil {
		return "", err
	}

	CName := strings.TrimSpace(string(FullConfigName))
	if CName == "" {
		return "", fmt.Errorf("конфигурационный файл не содержит название города")
	}

	return CName, nil
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
	var cityName string

	// Проверка аргументов
	if len(os.Args) > 1 {
		if os.Args[1] == "-v" || os.Args[1] == "--version" {
			fmt.Printf("gweather v%s\n", Version)
			return
		}
		if os.Args[1] == "-h" || os.Args[1] == "--help" {
			fmt.Printf("gweather v%s — прогноз погоды", Version)
			fmt.Println()
			fmt.Println("Использование:")
			fmt.Println("  gweather             Погода для города из конфигурации")
			fmt.Println("  gweather <город>     Погода для указанного города")
			fmt.Println("  gweather -h          Показать справку")
			fmt.Println("  gweather --help      Показать справку")
			fmt.Println("  gweather -v          Показать версию")
			fmt.Println("  gweather --version   Показать версию")
			return
		}
		if strings.HasPrefix(os.Args[1], "-") {
			fmt.Printf("Неизвестный параметр: %s\n", os.Args[1])
			fmt.Println("Используйте gweather --help для справки")
			return
		}
	}

	if len(os.Args) < 2 {
		cName, er := GetDefaultCity()
		if er != nil {
			fmt.Println(er)
			return
		}
		cityName = cName
	} else {
		cityName = os.Args[1]
	}
	// Получаем наименование города из аргументов
	city, e := GetCity(cityName)
	if e != nil {
		fmt.Println(e)
		return
	}

	latitude := strconv.FormatFloat(city.Latitude, 'f', -1, 64)
	longitude := strconv.FormatFloat(city.Longitude, 'f', -1, 64)

	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%s&longitude=%s&current=temperature_2m,relative_humidity_2m,apparent_temperature,wind_speed_10m&wind_speed_unit=ms&daily=temperature_2m_max,temperature_2m_min,precipitation_probability_max,weather_code,wind_speed_10m_max&forecast_days=4&timezone=auto", latitude, longitude)

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

	if data.Error {
		fmt.Println("Ошибка Open-Meteo:", data.Reason)
		return
	}

	fmt.Println("")
	fmt.Println("Погода для    :", city.Name, city.Country, data.Latitude, data.Longitude)
	fmt.Println("Температура   :", data.Current.Temperature, data.CurrentUnits.TemperatureUnit)
	fmt.Println("Ощущается как :", data.Current.ApparentTemp, data.CurrentUnits.ApparentTempSign)
	fmt.Println("Влажность     :", data.Current.Humidity, data.CurrentUnits.HumiditySign)
	fmt.Println("Скорость ветра:", data.Current.WindSpeed, data.CurrentUnits.WindSpeedSign)

	fmt.Println("")
	fmt.Println("Прогноз на 3 дня")
	fmt.Println("")

	days := len(data.Daily.Time)
	if days < 2 {
		fmt.Println("Ошибка: недостаточно данных для прогноза")
		return
	}
	if len(data.Daily.TemperatureMax) != days {
		fmt.Println("Ошибка: неполные данные максимальной температуры")
		return
	}
	if len(data.Daily.TemperatureMin) != days {
		fmt.Println("Ошибка: неполные данные минимальной температуры")
		return
	}
	if len(data.Daily.PrecipProb) != days {
		fmt.Println("Ошибка: неполные данные вероятности осадков")
		return
	}
	if len(data.Daily.WeatherCode) != days {
		fmt.Println("Ошибка: неполные данные кода погоды")
		return
	}
	if len(data.Daily.WindSpeedMax) != days {
		fmt.Println("Ошибка: неполные данные максимальной скорости ветра")
		return
	}

	for i := 1; i < len(data.Daily.Time); i++ {
		date, err := time.Parse("2006-01-02", data.Daily.Time[i])
		if err != nil {
			fmt.Println("Ошибка разбора даты:", err)
			return
		}
		dayName := weekdayRu(date.Weekday())
		fmt.Printf("%s, %s\n", dayName, date.Format("02.01.2006"))
		fmt.Printf("Температура макс.   : %.1f °C\n", data.Daily.TemperatureMax[i])
		fmt.Printf("Температура мин.    : %.1f °C\n", data.Daily.TemperatureMin[i])
		fmt.Printf("Вероятность осадков : %d %%\n", data.Daily.PrecipProb[i])
		fmt.Printf("Погода              : %s\n", weatherDescription(data.Daily.WeatherCode[i]))
		fmt.Printf("Скорость ветра макс.: %.1f м/с\n", data.Daily.WindSpeedMax[i])
		fmt.Println("")
	}

}
