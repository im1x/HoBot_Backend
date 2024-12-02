package utility

import (
	"math"
	"sort"
)

func Median(ratings []int) float64 {
	n := len(ratings)
	if n == 0 {
		return 0 // Handle empty input gracefully
	}

	sort.Ints(ratings) // Sort the array

	// If odd, return the middle element
	if n%2 == 1 {
		return float64(ratings[n/2])
	}

	// If even, return the average of the two middle elements
	mid1, mid2 := ratings[(n/2)-1], ratings[n/2]
	return float64(mid1+mid2) / 2.0
}

// Arithmetic mean calculation
func Mean(ratings []int) float64 {
	n := len(ratings)
	if n == 0 {
		return 0 // Handle empty input gracefully
	}

	sum := 0
	for _, rating := range ratings {
		sum += rating
	}
	return float64(sum) / float64(n)
}

// Trimmed mean calculation
func TrimmedMean(ratings []int, trimPercent float64) float64 {
	n := len(ratings)
	if n == 0 {
		return 0
	}

	sort.Ints(ratings) // Sort the array

	trimCount := int(float64(n) * trimPercent) // Calculate how many values to trim

	start := trimCount
	end := n - trimCount

	sum := 0
	for i := start; i < end; i++ {
		sum += ratings[i]
	}

	return float64(sum) / float64(end-start) // Calculate mean of the trimmed array
}

// Bayesian mean calculation
func BayesianMean(ratings []int, globalMean float64, weight int) float64 {
	n := len(ratings)
	if n == 0 {
		return globalMean // If no ratings, return the global mean
	}

	sum := 0
	for _, rating := range ratings {
		sum += rating
	}

	return float64(sum+weight*int(globalMean)) / float64(n+weight)
}

// Filtering based on mode
func FilterByMode(ratings []int, tolerance int) float64 {
	modeCount := make(map[int]int)
	for _, rating := range ratings {
		modeCount[rating]++
	}

	// Find mode (most common rating)
	mode := 0
	maxCount := 0
	for rating, count := range modeCount {
		if count > maxCount {
			mode = rating
			maxCount = count
		}
	}

	// Filter ratings within the specified tolerance of the mode
	filteredRatings := []int{}
	for _, rating := range ratings {
		if int(math.Abs(float64(rating-mode))) <= tolerance {
			filteredRatings = append(filteredRatings, rating)
		}
	}

	return Mean(filteredRatings) // Calculate mean of filtered ratings
}

// Filtering based on standard deviations
func FilterByStandardDeviation(ratings []int, threshold float64) float64 {
	if len(ratings) == 0 {
		return 0
	}

	m := Mean(ratings)
	variance := 0.0
	for _, rating := range ratings {
		variance += math.Pow(float64(rating)-m, 2)
	}
	stdDev := math.Sqrt(variance / float64(len(ratings)))

	// Filter out ratings beyond the specified standard deviation threshold
	filteredRatings := []int{}
	for _, rating := range ratings {
		if math.Abs(float64(rating)-m) <= threshold*stdDev {
			filteredRatings = append(filteredRatings, rating)
		}
	}

	return Mean(filteredRatings) // Calculate mean of filtered ratings
}
