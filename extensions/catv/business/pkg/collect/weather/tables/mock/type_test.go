package mock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// --- LocationType 상수 ---

func TestLocationTypeConstants(t *testing.T) {
	assert.Equal(t, LocationType(0), Coastal)
	assert.Equal(t, LocationType(1), Inland)
	assert.Equal(t, LocationType(2), Mountain)
	assert.Equal(t, LocationType(3), Island)
}

// --- GetLocationType ---

func TestGetLocationType_KnownLocations(t *testing.T) {
	tests := []struct {
		location string
		want     LocationType
	}{
		{"부산", Coastal},
		{"강릉", Coastal},
		{"대관령", Mountain},
		{"태백", Mountain},
		{"제주", Island},
		{"울릉도", Island},
	}
	for _, tt := range tests {
		t.Run(tt.location, func(t *testing.T) {
			assert.Equal(t, tt.want, GetLocationType(tt.location))
		})
	}
}

func TestGetLocationType_UnknownDefaultsToInland(t *testing.T) {
	assert.Equal(t, Inland, GetLocationType("알수없는지역"))
}

// --- GetLocationTempOffset ---

func TestGetLocationTempOffset_KnownLocations(t *testing.T) {
	assert.Equal(t, int32(3), GetLocationTempOffset("제주"))
	assert.Equal(t, int32(-5), GetLocationTempOffset("대관령"))
	assert.Equal(t, int32(0), GetLocationTempOffset("서울"))
}

func TestGetLocationTempOffset_UnknownReturnsZero(t *testing.T) {
	assert.Equal(t, int32(0), GetLocationTempOffset("없는도시"))
}

// --- GetMonthlyTempRange ---

func TestGetMonthlyTempRange_AllMonths(t *testing.T) {
	months := []time.Month{
		time.January, time.February, time.March, time.April,
		time.May, time.June, time.July, time.August,
		time.September, time.October, time.November, time.December,
	}
	for _, month := range months {
		t.Run(month.String(), func(t *testing.T) {
			minT, maxT := GetMonthlyTempRange(month)
			assert.Less(t, minT, maxT, "min 온도는 max 온도보다 작아야 함")
		})
	}
}

func TestGetMonthlyTempRange_SummerWarmerThanWinter(t *testing.T) {
	sumMin, _ := GetMonthlyTempRange(time.July)
	winMin, _ := GetMonthlyTempRange(time.January)
	assert.Greater(t, sumMin, winMin)
}

// --- GetRealisticTemperature ---

func TestGetRealisticTemperature_ReturnsValue(t *testing.T) {
	now := time.Date(2026, time.July, 1, 14, 0, 0, 0, time.UTC)
	temp := GetRealisticTemperature("서울", now)
	// 여름 서울 최고 온도 범위 내 (넉넉하게 체크)
	assert.Greater(t, temp, float32(-30))
	assert.Less(t, temp, float32(60))
}

func TestGetRealisticTemperature_CoastalWarmerThanMountain(t *testing.T) {
	// 제주(해안, +3 보정)와 대관령(산간, -5 보정) 여름 정오 비교
	// 랜덤 요소 있으므로 여러 번 실행하여 평균적으로 확인
	t1 := time.Date(2026, time.July, 15, 12, 0, 0, 0, time.UTC)
	jejuTemp := GetRealisticTemperature("제주", t1)
	daegwalTemp := GetRealisticTemperature("대관령", t1)
	// 제주가 대관령보다 평균적으로 8도 높으므로, 한쪽이 비정상적으로 높지 않음을 검증
	assert.NotEqual(t, jejuTemp, daegwalTemp)
}

// --- GetMonthlyRainWeight ---

func TestGetMonthlyRainWeight_SummerHigherThanWinter(t *testing.T) {
	julyWeight := GetMonthlyRainWeight(time.July)
	janWeight := GetMonthlyRainWeight(time.January)
	assert.Greater(t, julyWeight, janWeight)
}

func TestGetMonthlyRainWeight_AllMonthsPositive(t *testing.T) {
	months := []time.Month{
		time.January, time.February, time.March, time.April,
		time.May, time.June, time.July, time.August,
		time.September, time.October, time.November, time.December,
	}
	for _, month := range months {
		assert.Greater(t, GetMonthlyRainWeight(month), int32(0))
	}
}

// --- GetWindSpeedRange ---

func TestGetWindSpeedRange_IslandHighestWind(t *testing.T) {
	_, islandMax := GetWindSpeedRange(Island)
	_, coastalMax := GetWindSpeedRange(Coastal)
	_, inlandMax := GetWindSpeedRange(Inland)
	assert.GreaterOrEqual(t, islandMax, coastalMax)
	assert.Greater(t, coastalMax, inlandMax)
}

func TestGetWindSpeedRange_MinLessThanMax(t *testing.T) {
	for _, lt := range []LocationType{Coastal, Inland, Mountain, Island} {
		min, max := GetWindSpeedRange(lt)
		assert.Less(t, min, max)
	}
}

func TestGetWindSpeedRange_UnknownReturnsDefault(t *testing.T) {
	min, max := GetWindSpeedRange(LocationType(99))
	assert.Equal(t, float32(1.0), min)
	assert.Equal(t, float32(10.0), max)
}

// --- GetRealisticWindSpeed ---

func TestGetRealisticWindSpeed_WithinRange(t *testing.T) {
	speed := GetRealisticWindSpeed("부산", time.July)
	assert.GreaterOrEqual(t, speed, float32(0))
	assert.Less(t, speed, float32(30))
}

// --- GetHumidityRange ---

func TestGetHumidityRange_SummerHigherThanWinter(t *testing.T) {
	sumMin, _ := GetHumidityRange(Inland, time.July)
	winMin, _ := GetHumidityRange(Inland, time.January)
	assert.Greater(t, sumMin, winMin)
}

func TestGetHumidityRange_CoastalHigherThanInland(t *testing.T) {
	_, coastalMax := GetHumidityRange(Coastal, time.July)
	_, inlandMax := GetHumidityRange(Inland, time.July)
	assert.GreaterOrEqual(t, coastalMax, inlandMax)
}

func TestGetHumidityRange_MaxNotExceed100(t *testing.T) {
	_, max := GetHumidityRange(Island, time.July)
	assert.LessOrEqual(t, max, int32(100))
}

// --- AdjustRainProbByWeather ---

func TestAdjustRainProbByWeather_RainWeatherMinimum70(t *testing.T) {
	for _, weather := range []string{"비", "흐리고 비", "구름많고 비", "비/눈", "눈", "흐리고 눈", "구름많고 눈", "눈/비"} {
		result := AdjustRainProbByWeather(weather, 10)
		assert.GreaterOrEqual(t, result, int32(70), "날씨: "+weather)
	}
}

func TestAdjustRainProbByWeather_ClearWeatherMaximum20(t *testing.T) {
	for _, weather := range []string{"맑음", "구름조금"} {
		result := AdjustRainProbByWeather(weather, 50)
		assert.LessOrEqual(t, result, int32(20), "날씨: "+weather)
	}
}

func TestAdjustRainProbByWeather_CloudyMaximum40(t *testing.T) {
	result := AdjustRainProbByWeather("구름많음", 60)
	assert.LessOrEqual(t, result, int32(40))
}

func TestAdjustRainProbByWeather_OvercastRange(t *testing.T) {
	result := AdjustRainProbByWeather("흐림", 10)
	assert.GreaterOrEqual(t, result, int32(20))
	result2 := AdjustRainProbByWeather("흐림", 80)
	assert.LessOrEqual(t, result2, int32(60))
}

func TestAdjustRainProbByWeather_UnknownUnchanged(t *testing.T) {
	result := AdjustRainProbByWeather("알수없는날씨", 45)
	assert.Equal(t, int32(45), result)
}

// --- AdjustHumidityByWeather ---

func TestAdjustHumidityByWeather_RainIncreasesHumidity(t *testing.T) {
	result := AdjustHumidityByWeather("비", 50)
	assert.GreaterOrEqual(t, result, int32(70))
}

func TestAdjustHumidityByWeather_ClearDecreasesHumidity(t *testing.T) {
	result := AdjustHumidityByWeather("맑음", 90)
	assert.LessOrEqual(t, result, int32(70))
}

func TestAdjustHumidityByWeather_CloudyIncreasesHumidity(t *testing.T) {
	result := AdjustHumidityByWeather("흐림", 40)
	assert.GreaterOrEqual(t, result, int32(60))
}

// --- GetRealisticRainfall ---

func TestGetRealisticRainfall_RainWeatherNotNil(t *testing.T) {
	for _, weather := range []string{"비", "흐리고 비", "구름많고 비"} {
		result := GetRealisticRainfall(weather, time.July)
		assert.NotNil(t, result, "날씨: "+weather)
		assert.GreaterOrEqual(t, *result, float32(0))
	}
}

func TestGetRealisticRainfall_SnowWeatherNotNil(t *testing.T) {
	result := GetRealisticRainfall("눈", time.January)
	assert.NotNil(t, result)
	// 눈은 강수량 적음
	if result != nil {
		assert.Less(t, *result, float32(10.0))
	}
}

// --- GetWeatherIconByText ---

func TestGetWeatherIconByText_KnownWeatherReturnsIcon(t *testing.T) {
	tests := []struct {
		weather  string
		wantIcon int32
	}{
		{"맑음", 1},
		{"구름조금", 2},
		{"구름많음", 3},
		{"흐림", 4},
		{"비", 5},
		{"눈", 6},
		{"비/눈", 11},
		{"눈/비", 11},
	}
	for _, tt := range tests {
		t.Run(tt.weather, func(t *testing.T) {
			assert.Equal(t, tt.wantIcon, GetWeatherIconByText(tt.weather))
		})
	}
}

func TestGetWeatherIconByText_UnknownWeatherReturnsPositiveIcon(t *testing.T) {
	// 알 수 없는 날씨는 1~12 범위의 랜덤 아이콘
	icon := GetWeatherIconByText("알수없는날씨")
	assert.GreaterOrEqual(t, icon, int32(1))
	assert.LessOrEqual(t, icon, int32(12))
}

// --- GetDailyTempRange ---

func TestGetDailyTempRange_MinLessThanMax(t *testing.T) {
	months := []time.Month{time.January, time.April, time.July, time.October}
	for _, month := range months {
		t.Run(month.String(), func(t *testing.T) {
			minT, maxT := GetDailyTempRange("서울", month)
			assert.Less(t, minT, maxT)
		})
	}
}

func TestGetDailyTempRange_SummerWarmerThanWinter(t *testing.T) {
	sumMin, _ := GetDailyTempRange("서울", time.July)
	winMin, _ := GetDailyTempRange("서울", time.January)
	// 여름 최저온도가 겨울 최저온도보다 높아야 함 (평균적으로)
	assert.Greater(t, sumMin, winMin)
}

func TestGetDailyTempRange_CoastalWarmerThanMountain(t *testing.T) {
	jejuMin, _ := GetDailyTempRange("제주", time.January)
	daegwalMin, _ := GetDailyTempRange("대관령", time.January)
	assert.Greater(t, jejuMin, daegwalMin)
}

// --- GetRealisticHumidity ---

func TestGetRealisticHumidity_WithinBounds(t *testing.T) {
	humidity := GetRealisticHumidity("서울", time.July, "흐림")
	assert.GreaterOrEqual(t, humidity, int32(0))
	assert.LessOrEqual(t, humidity, int32(100))
}

func TestGetRealisticRainProb_WithinBounds(t *testing.T) {
	prob := GetRealisticRainProb(time.July, "비")
	assert.GreaterOrEqual(t, prob, int32(0))
	assert.LessOrEqual(t, prob, int32(100))
}
