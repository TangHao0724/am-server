package service

import (
	"math"
)

type Userlocation struct {
	Lon float64 `json:"lon"`
	Lat float64 `json:"lat"`
}

type LocationResult struct {
	DistanceMeters float64 `json:"distance_meters"`
	BearingDegrees float64 `json:"bearing_degrees"`
	Direction      string  `json:"direction"`
}

// CalculateLocation 計算玩家到目標位置的距離與方向
func CalculateLocation(
	user Userlocation,
	targetLat float64,
	targetLon float64,
) LocationResult {
	// 轉換角度為弧度
	lat1 := toRadians(user.Lat)
	lat2 := toRadians(targetLat)
	lon1 := toRadians(user.Lon)
	lon2 := toRadians(targetLon)

	dLat := lat2 - lat1
	dLon := lon2 - lon1

	// Haversine 公式：計算距離
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	// 避免浮點誤差造成 sqrt 負數或 asin 超出範圍
	a = math.Max(0, math.Min(1, a))

	earthRadius := 6371000.0 // 地球平均半徑，公尺
	distance := 2 * earthRadius * math.Asin(math.Sqrt(a))

	// Bearing 公式：計算從玩家指向目標的方位角
	y := math.Sin(dLon) * math.Cos(lat2)
	x := math.Cos(lat1)*math.Sin(lat2) -
		math.Sin(lat1)*math.Cos(lat2)*math.Cos(dLon)

	bearing := math.Atan2(y, x)
	bearing = math.Mod(toDegrees(bearing)+360, 360)

	return LocationResult{
		DistanceMeters: distance,
		BearingDegrees: bearing,
		Direction:      bearingToDirection(bearing),
	}
}

func toRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func toDegrees(radians float64) float64 {
	return radians * 180 / math.Pi
}

// bearingToDirection 將方位角轉換成八方位
func bearingToDirection(bearing float64) string {
	directions := []string{
		"北", "東北", "東", "東南",
		"南", "西南", "西", "西北",
	}

	index := int(math.Floor((bearing+22.5)/45)) % 8
	return directions[index]
}
