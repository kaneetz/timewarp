package timewarp

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"
)

// TimeKeeper manages the simulated time
type TimeKeeper struct {
	startRealTime time.Time
	startSimTime  time.Time
	multiplier    float64
	mutex         sync.Mutex
}

// New initializes a new TimeKeeper instance
func New(startDate, startTime, timeZone string, multiplier float64) (*TimeKeeper, error) {
	location, err := time.LoadLocation(timeZone)
	if err != nil {
		return nil, err
	}

	startSimTime, err := time.ParseInLocation("2006-01-02 15:04", startDate+" "+startTime, location)
	if err != nil {
		return nil, err
	}

	startRealTime := time.Now()

	return &TimeKeeper{
		startRealTime: startRealTime,
		startSimTime:  startSimTime,
		multiplier:    multiplier,
	}, nil
}

// Now returns the current simulated time
func (tk *TimeKeeper) Now() time.Time {
	tk.mutex.Lock()
	defer tk.mutex.Unlock()

	elapsedReal := time.Since(tk.startRealTime)
	elapsedSim := time.Duration(float64(elapsedReal) * tk.multiplier)

	return tk.startSimTime.Add(elapsedSim)
}

// Duration calculates the simulated duration between two timestamps
func (tk *TimeKeeper) Duration(from, to time.Time) time.Duration {
	return time.Duration(float64(to.Sub(from)) * tk.multiplier)
}

// SetMultiplier updates the time speed dynamically
func (tk *TimeKeeper) SetMultiplier(multiplier float64) {
	tk.mutex.Lock()
	defer tk.mutex.Unlock()
	tk.multiplier = multiplier
}

// Reset restarts the simulation with the initial settings
func (tk *TimeKeeper) Reset() {
	tk.mutex.Lock()
	defer tk.mutex.Unlock()
	tk.startRealTime = time.Now()
}

// Synchronize fetches time from a remote API
func (tk *TimeKeeper) Synchronize(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var data struct {
		SimulatedTime string `json:"simulated_time"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	simTime, err := time.Parse(time.RFC3339, data.SimulatedTime)
	if err != nil {
		return err
	}

	tk.mutex.Lock()
	defer tk.mutex.Unlock()
	tk.startSimTime = simTime
	tk.startRealTime = time.Now()

	return nil
}
