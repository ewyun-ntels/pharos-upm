package mock

import (
	"math/rand/v2"
	"strings"
	"time"
)

// LocationType 지역 유형
type LocationType int

const (
	Coastal  LocationType = iota // 해안
	Inland                       // 내륙
	Mountain                     // 산간
	Island                       // 섬
)

// 월별 전국 평균 기온 기준 (실제 기상청 데이터 기반)
var monthlyTempBase = map[time.Month]struct{ min, max int32 }{
	time.January:   {min: -6, max: 3},
	time.February:  {min: -3, max: 6},
	time.March:     {min: 2, max: 12},
	time.April:     {min: 8, max: 19},
	time.May:       {min: 14, max: 24},
	time.June:      {min: 19, max: 27},
	time.July:      {min: 23, max: 29},
	time.August:    {min: 23, max: 30},
	time.September: {min: 18, max: 26},
	time.October:   {min: 11, max: 20},
	time.November:  {min: 4, max: 12},
	time.December:  {min: -3, max: 5},
}

// 지역별 온도 보정값 (전국 평균 대비)
var locationTempOffset = map[string]int32{
	// 남부 해안 (온화)
	"제주": 3, "서귀포": 3, "부산": 2, "통영": 2, "여수": 2, "목포": 1, "포항": 1,
	// 산간 지역 (한랭)
	"대관령": -5, "태백": -4, "철원": -3, "정선": -3, "인제": -2, "홍천": -2,
	// 내륙 주요 도시
	"서울": 0, "인천": 0, "대전": 0, "광주": 1, "대구": 0, "울산": 1,
	// 섬 지역
	"백령도": -1, "울릉도": 1, "흑산도": 1, "완도": 2, "고흥": 1,
	// 기타 지역은 0으로 처리
}

// 지역 타입 매핑
var locationTypes = map[string]LocationType{
	// 해안
	"부산": Coastal, "속초": Coastal, "통영": Coastal, "포항": Coastal, "강릉": Coastal,
	"동해": Coastal, "삼척": Coastal, "울산": Coastal, "창원": Coastal, "거제": Coastal,
	"남해": Coastal, "여수": Coastal, "목포": Coastal, "인천": Coastal, "북강릉": Coastal,
	// 산간
	"대관령": Mountain, "태백": Mountain, "정선": Mountain, "철원": Mountain, "인제": Mountain,
	"홍천": Mountain, "추풍령": Mountain, "문경": Mountain, "영주": Mountain, "거창": Mountain,
	// 섬
	"제주": Island, "서귀포": Island, "고산": Island, "성산": Island, "울릉도": Island,
	"백령도": Island, "흑산도": Island, "완도": Island, "고흥": Island, "진도": Island,
	// 내륙 (기본값)
}

// 월별 강수 가능성 가중치 (장마철 6-8월 높음)
var monthlyRainWeight = map[time.Month]int32{
	time.January: 20, time.February: 20, time.March: 30,
	time.April: 40, time.May: 50, time.June: 70, // 장마 시작
	time.July: 80, time.August: 70, time.September: 60,
	time.October: 40, time.November: 30, time.December: 20,
}

// GetLocationTempOffset 지역별 온도 보정값 반환
func GetLocationTempOffset(location string) int32 {
	if offset, exists := locationTempOffset[location]; exists {
		return offset
	}
	return 0 // 기본값
}

// GetLocationType 지역 타입 반환
func GetLocationType(location string) LocationType {
	if locType, exists := locationTypes[location]; exists {
		return locType
	}
	return Inland // 기본값: 내륙
}

// GetMonthlyTempRange 월별 기준 온도 범위 반환
func GetMonthlyTempRange(month time.Month) (min, max int32) {
	tempRange := monthlyTempBase[month]
	return tempRange.min, tempRange.max
}

// GetRealisticTemperature 현실적인 온도 생성 (지역/월/시간 고려)
func GetRealisticTemperature(location string, t time.Time) float32 {
	// 1. 월별 기준 온도 범위
	tempMin, tempMax := GetMonthlyTempRange(t.Month())

	// 2. 지역별 보정
	offset := GetLocationTempOffset(location)

	// 3. 시간대별 일교차 반영 (06시 최저, 15시 최고)
	hour := t.Hour()
	var hourFactor float32
	if hour >= 0 && hour < 6 {
		hourFactor = -0.3 // 새벽: 최저 온도
	} else if hour >= 6 && hour < 12 {
		hourFactor = float32(hour-6) / 9.0 // 상승
	} else if hour >= 12 && hour < 15 {
		hourFactor = 0.7 + float32(hour-12)/10.0 // 최고 도달
	} else if hour >= 15 && hour < 18 {
		hourFactor = 0.5 // 하강 시작
	} else {
		hourFactor = 0.2 - float32(hour-18)/20.0 // 저녁
	}

	// 4. 기준 온도 계산
	baseTemp := float32(tempMin+offset) + float32(tempMax-tempMin)*0.5
	dailyRange := float32(tempMax - tempMin)

	// 5. 최종 온도 = 기준 + 일교차 + 랜덤 변동
	temperature := baseTemp + dailyRange*hourFactor + (rand.Float32()*4 - 2)

	return temperature
}

// GetWindSpeedRange 지역 타입별 풍속 범위 반환
func GetWindSpeedRange(locType LocationType) (min, max float32) {
	switch locType {
	case Coastal:
		return 2.0, 12.0 // 해안: 바람 강함
	case Island:
		return 3.0, 15.0 // 섬: 바람 매우 강함
	case Mountain:
		return 2.5, 14.0 // 산간: 바람 강함
	case Inland:
		return 0.5, 8.0 // 내륙: 바람 약함
	default:
		return 1.0, 10.0
	}
}

// GetRealisticWindSpeed 현실적인 풍속 생성
func GetRealisticWindSpeed(location string, month time.Month) float32 {
	locType := GetLocationType(location)
	windMin, windMax := GetWindSpeedRange(locType)

	// 겨울철(12-2월) 바람 강함
	if month == time.December || month == time.January || month == time.February {
		windMin += 1.0
		windMax += 2.0
	}

	return windMin + rand.Float32()*(windMax-windMin)
}

// GetHumidityRange 지역 타입 및 월별 습도 범위 반환
func GetHumidityRange(locType LocationType, month time.Month) (min, max int32) {
	baseMin, baseMax := int32(40), int32(80)

	// 여름철(6-8월) 습도 증가
	if month >= time.June && month <= time.August {
		baseMin, baseMax = 60, 90
	}

	// 겨울철(12-2월) 습도 감소
	if month == time.December || month == time.January || month == time.February {
		baseMin, baseMax = 30, 60
	}

	// 해안/섬 지역은 습도 높음
	if locType == Coastal || locType == Island {
		baseMin += 10
		baseMax += 10
		if baseMax > 100 {
			baseMax = 100
		}
	}

	return baseMin, baseMax
}

// GetRealisticHumidity 현실적인 습도 생성
func GetRealisticHumidity(location string, month time.Month, weatherText string) int32 {
	locType := GetLocationType(location)
	humidMin, humidMax := GetHumidityRange(locType, month)

	humidity := humidMin + rand.Int32N(humidMax-humidMin+1)

	// 날씨 텍스트에 따른 습도 조정
	humidity = AdjustHumidityByWeather(weatherText, humidity)

	return humidity
}

// GetMonthlyRainWeight 월별 강수 가중치 반환
func GetMonthlyRainWeight(month time.Month) int32 {
	return monthlyRainWeight[month]
}

// GetRealisticRainProb 현실적인 강수확률 생성
func GetRealisticRainProb(month time.Month, weatherText string) int32 {
	weight := GetMonthlyRainWeight(month)
	rainProb := min(
		// 가중치 + 변동폭
		rand.Int32N(weight+20), 100)

	// 날씨 텍스트에 따른 강수확률 조정
	rainProb = AdjustRainProbByWeather(weatherText, rainProb)

	return rainProb
}

// AdjustRainProbByWeather 날씨 텍스트에 따른 강수확률 조정
func AdjustRainProbByWeather(weatherText string, baseRainProb int32) int32 {
	switch weatherText {
	case "맑음", "구름조금":
		if baseRainProb > 20 {
			return 20 // 최대 20%
		}
		return baseRainProb
	case "구름많음":
		if baseRainProb > 40 {
			return 40 // 최대 40%
		}
		if baseRainProb < 10 {
			return 10
		}
		return baseRainProb
	case "흐림":
		if baseRainProb > 60 {
			return 60 // 최대 60%
		}
		if baseRainProb < 20 {
			return 20
		}
		return baseRainProb
	case "비", "흐리고 비", "구름많고 비", "비/눈":
		if baseRainProb < 70 {
			return 70 + rand.Int32N(31) // 최소 70%
		}
		return baseRainProb
	case "눈", "흐리고 눈", "구름많고 눈", "눈/비":
		if baseRainProb < 70 {
			return 70 + rand.Int32N(31) // 최소 70%
		}
		return baseRainProb
	default:
		return baseRainProb
	}
}

// AdjustHumidityByWeather 날씨 텍스트에 따른 습도 조정
func AdjustHumidityByWeather(weatherText string, baseHumidity int32) int32 {
	switch {
	case strings.Contains(weatherText, "비") || strings.Contains(weatherText, "눈"):
		// 강수시 습도 높음
		if baseHumidity < 70 {
			return 70 + rand.Int32N(31) // 70~100%
		}
		return baseHumidity
	case weatherText == "맑음" || weatherText == "구름조금":
		// 맑을 때 습도 낮음
		if baseHumidity > 70 {
			return 40 + rand.Int32N(31) // 40~70%
		}
		return baseHumidity
	case weatherText == "흐림":
		// 흐릴 때 습도 약간 높음
		if baseHumidity < 60 {
			return 60 + rand.Int32N(21) // 60~80%
		}
		return baseHumidity
	default:
		return baseHumidity
	}
}

// GetRealisticRainfall 현실적인 강수량 생성 (날씨 텍스트 기반)
func GetRealisticRainfall(weatherText string, month time.Month) *float32 {
	// 강수 관련 날씨가 아니면 nil 또는 0.0
	if !strings.Contains(weatherText, "비") && !strings.Contains(weatherText, "눈") {
		// 10% 확률로 미량 강수
		if rand.IntN(10) == 0 {
			val := rand.Float32() * 0.5 // 0~0.5mm
			return &val
		}
		return nil
	}

	// 강수 관련 날씨면 강수량 생성
	var rainfall float32

	// 장마철(6-8월) 강수량 많음
	if month >= time.June && month <= time.August {
		rainfall = rand.Float32() * 30.0 // 0~30mm
	} else {
		rainfall = rand.Float32() * 15.0 // 0~15mm
	}

	// 눈은 강수량 적음
	if strings.Contains(weatherText, "눈") && !strings.Contains(weatherText, "비") {
		rainfall = rand.Float32() * 5.0 // 0~5mm
	}

	return &rainfall
}

// GetWeatherIconByText 날씨 텍스트에 맞는 아이콘 번호 반환 (간단한 매핑)
func GetWeatherIconByText(weatherText string) int32 {
	iconMap := map[string]int32{
		"맑음":     1,
		"구름조금":   2,
		"구름많음":   3,
		"흐림":     4,
		"비":      5,
		"눈":      6,
		"흐리고 비":  7,
		"흐리고 눈":  8,
		"구름많고 비": 9,
		"구름많고 눈": 10,
		"비/눈":    11,
		"눈/비":    11,
		"흐린후갬":   4,
		"맑다가 흐림": 3,
	}

	if icon, exists := iconMap[weatherText]; exists {
		return icon
	}
	return rand.Int32N(12) + 1 // 기본값: 랜덤
}

// GetDailyTempRange 일별 최저/최고 온도 범위 생성
func GetDailyTempRange(location string, month time.Month) (min, max int32) {
	// 월별 기준 온도
	tempMin, tempMax := GetMonthlyTempRange(month)
	offset := GetLocationTempOffset(location)

	// 일교차 (계절별 차이)
	var dailyRange int32
	switch month {
	case time.March, time.April, time.May, time.September, time.October, time.November:
		dailyRange = 10 // 봄/가을: 일교차 큼
	case time.June, time.July, time.August:
		dailyRange = 8 // 여름: 일교차 작음
	case time.December, time.January, time.February:
		dailyRange = 7 // 겨울: 일교차 중간
	default:
		dailyRange = 9
	}

	minTemp := tempMin + offset - dailyRange/2
	maxTemp := tempMax + offset + dailyRange/2

	// 변동
	minTemp += rand.Int32N(5) - 2
	maxTemp += rand.Int32N(5) - 2

	// 최소/최대 온도가 역전되지 않도록
	if minTemp >= maxTemp {
		maxTemp = minTemp + 5
	}

	return minTemp, maxTemp
}
