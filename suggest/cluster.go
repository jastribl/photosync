package suggest

import (
	"math"
	"sort"
	"time"
)

// Cluster is a group of photos taken at one stop: one restaurant, one
// sight, one stretch of a hike. Center is the median position of its
// photos, which is more robust to a single bad GPS fix than the mean.
type Cluster struct {
	Photos     []Photo
	CenterLat  float64
	CenterLong float64
	Start      time.Time
	End        time.Time
}

// ClusterPhotos groups a folder's photos into clusters, oldest first.
// Photos are walked in time order; a new cluster starts when a photo is
// more than timeGap after the previous photo OR more than distanceM from
// the current cluster's centroid.
//
// The time split is what makes this work in practice. Pure distance
// fails twice: on a hike, photos spread along a trail would merge into
// one giant cluster, and at dinner-then-bar-next-door, two venues 30 m
// apart would merge into one. The time gap separates both cases (trail
// photos are spread over hours; dinner and the bar are different time
// blocks). Photos without GPS are left out of clustering; callers run
// InheritGPS first so that only a folder with no GPS at all has any.
func ClusterPhotos(photos []Photo, timeGap time.Duration, distanceM float64) []Cluster {
	sorted := make([]Photo, 0, len(photos))
	for _, p := range photos {
		if p.HasGPS {
			sorted = append(sorted, p)
		}
	}
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Taken.Equal(sorted[j].Taken) {
			return sorted[i].Name < sorted[j].Name
		}
		return sorted[i].Taken.Before(sorted[j].Taken)
	})

	var clusters []Cluster
	for _, p := range sorted {
		if len(clusters) == 0 {
			clusters = append(clusters, Cluster{Photos: []Photo{p}, Start: p.Taken, End: p.Taken})
			continue
		}
		last := &clusters[len(clusters)-1]
		prev := last.Photos[len(last.Photos)-1]
		centerLat, centerLong := clusterCenter(last.Photos)
		if p.Taken.Sub(prev.Taken) > timeGap || haversineM(p.Lat, p.Long, centerLat, centerLong) > distanceM {
			clusters = append(clusters, Cluster{Photos: []Photo{p}, Start: p.Taken, End: p.Taken})
			continue
		}
		last.Photos = append(last.Photos, p)
		last.End = p.Taken
	}
	for i := range clusters {
		clusters[i].CenterLat, clusters[i].CenterLong = clusterCenter(clusters[i].Photos)
	}
	return clusters
}

func clusterCenter(photos []Photo) (lat, lng float64) {
	lats := make([]float64, len(photos))
	lngs := make([]float64, len(photos))
	for i, p := range photos {
		lats[i], lngs[i] = p.Lat, p.Long
	}
	return median(lats), median(lngs)
}

func median(values []float64) float64 {
	sorted := append([]float64{}, values...)
	sort.Float64s(sorted)
	n := len(sorted)
	if n == 0 {
		return 0
	}
	if n%2 == 1 {
		return sorted[n/2]
	}
	return (sorted[n/2-1] + sorted[n/2]) / 2
}

// haversineM returns the great-circle distance between two points in
// meters.
func haversineM(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusM = 6371000
	toRad := func(d float64) float64 { return d * math.Pi / 180 }
	dLat := toRad(lat2 - lat1)
	dLng := toRad(lng2 - lng1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*math.Sin(dLng/2)*math.Sin(dLng/2)
	return 2 * earthRadiusM * math.Asin(math.Sqrt(a))
}
