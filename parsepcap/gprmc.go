package parsepcap

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// GPS contains the latitude and longitude of a GPS point
type GPS struct {
	lat float32
	lon float32
}

// GetTimeStamp returns the timestamp of the GPRMC packet
func (gps GPS) String() string {
	return fmt.Sprintf("%f, %f", gps.lat, gps.lon)
}

// GPRMC contains the decoded information of the GPRMC message
type GPRMC struct {
	Bytes   []byte
	message []string
	// Timestamp         time.Time
	// ReceiveStatus     string
	// Latitude          float32
	// Longitude         float32
	// Speed             float32
	// TrackMadeGood     float32
	// MagneticVariation float32
}

// GetTimeStamp returns the timestamp of the GPRMC packet
func (gprmc GPRMC) GetTimeStamp() time.Time {
	if len(gprmc.message) == 0 {
		gprmc.setMessage()
	}
	_time := gprmc.message[1]
	_date := gprmc.message[9]

	dd := _date[0:2]
	mo := _date[2:4]
	yy := _date[4:6]

	hh := _time[0:2]
	mm := _time[2:4]
	ss := _time[4:6]

	t, _ := time.Parse(
		time.RFC3339,
		fmt.Sprintf("20%s-%s-%sT%s:%s:%s+09:00", yy, mo, dd, hh, mm, ss))
	return t
}

// GetMessage returns the decoded GPRMC message
func (gprmc GPRMC) GetMessage() []string {
	if len(gprmc.message) == 0 {
		gprmc.setMessage()
	}
	return gprmc.message
}

// GetLocation returns the latitude and longitude
func (gprmc GPRMC) GetLocation() GPS {
	if len(gprmc.message) == 0 {
		gprmc.setMessage()
	}
	latDeg, _ := strconv.ParseFloat(gprmc.message[3][0:2], 32)
	latMin, _ := strconv.ParseFloat(gprmc.message[3][2:], 32)
	lat := latDeg + latMin/60

	lonDeg, _ := strconv.ParseFloat(gprmc.message[5][0:3], 32)
	lonMin, _ := strconv.ParseFloat(gprmc.message[5][3:], 32)
	lon := lonDeg + lonMin/60

	return GPS{float32(lat), float32(lon)}
}

func (gprmc *GPRMC) setMessage() {
	message := string(gprmc.Bytes)
	start := strings.LastIndex(message, "$GPRMC")
	end := start + 72
	(*gprmc).message = strings.Split(message[start:end], ",")
}
