package pricingservice

import (
	"fmt"
	"log"
	"math"
	"slices"
	"sort"
	"strings"

	"gonum.org/v1/gonum/stat"
)

//
// CORE ENUMS AND CONSTANTS
//

// DriverStyle represents a driver's racing style classification
type F1DriverStyle string

const (
	F1Aggressive F1DriverStyle = "Aggressive"
	F1Smooth     F1DriverStyle = "Smooth"
	F1Defensive  F1DriverStyle = "Defensive"
	F1Overtaker  F1DriverStyle = "Overtaker"
	F1AllRounder F1DriverStyle = "All-Rounder"
	F1Rookie     F1DriverStyle = "Rookie"
)

// PopularityLevel represents user-specified driver popularity
type F1PopularityLevel string

const (
	F1HighPopularity   F1PopularityLevel = "High"
	F1MediumPopularity F1PopularityLevel = "Medium"
	F1LowPopularity    F1PopularityLevel = "Low"
)

type F1SeasonPhase string

const (
	F1EarlySeason F1SeasonPhase = "Early" // First 1/4 of the season
	F1MidSeason   F1SeasonPhase = "Mid"   // Middle 1/2 of the season
	F1LateSeason  F1SeasonPhase = "Late"  // Final 1/4 of the season
)

//
// TEAM DATA STRUCTURES (NEW)
//

// TeamSeasonHistory holds historical data for a team's season
type F1TeamSeasonHistory struct {
	Year     int     // Season year
	Position int     // Championship position
	Points   float64 // Total points scored
	Wins     int     // Season wins
	Podiums  int     // Season podiums
	Races    int     // Season podiums
}

// TeamData holds directly observable information about a team
type F1TeamData struct {
	Name                      string                // Team name
	PowerUnit                 string                // Engine supplier (e.g., "Mercedes", "Ferrari")
	SeasonPosition            int                   // Current championship position
	CurrentRace               int                   // Current championship position
	SeasonPoints              float64               // Current season points
	BudgetTier                string                // "Top", "Upper-Mid", "Lower-Mid", "Backmarker" (publicly known)
	SeasonHistory             []F1TeamSeasonHistory // Previous seasons data
	Wins                      int                   // Current season wins
	Year                      int                   // Current season
	Podiums                   int                   // Current season podiums
	DNFs                      int                   // Current season DNFs
	RecentUpgrades            bool                  // Whether team has introduced upgrades in last ~3 races
	RecentRacePositions       []float64             // Last 5-6 races average finish positions
	RecentQualifyingPositions []float64             // Last 5-6 races average qualifying positions
}

//
// BASIC DATA STRUCTURES (USER-PROVIDED)
//

// BasicDriverData contains only the minimal info a user needs to provide
type F1BasicDriverData struct {
	// Basic information
	Name string
	Team string
	Age  int

	// Career statistics (publicly available)
	ChampionshipWins int
	CareerPodiums    int
	CareerStarts     int

	// Season performance data (publicly available)
	Seasons []F1BasicSeasonStats

	// Driver style classification (one-time classification)
	PrimaryStyle   F1DriverStyle
	SecondaryStyle F1DriverStyle

	// Track specialties (optional, can be empty)
	Specialties []string // Track types they excel at (e.g. "Wet", "Street")
	Weaknesses  []string // Track types they struggle with

	// Rookie info
	IsRookie bool

	// User-specified popularity (instead of nationality-based calculation)
	MarketPopularity F1PopularityLevel

	// Team data (new field)
	TeamData F1TeamData

	IsTeamLeader       bool // Designated #1 driver
	CurrentRaceNumber  int  // Current race of the season
	TotalRacesInSeason int  // Total races in current season

	PreviousTeam         string
	RacesWithCurrentTeam int
}

type F1RaceResult struct {
	RaceName       string
	RaceNumber     int // Race number in season (1-24)
	FinishPosition int
	StartPosition  int // Grid position
	PointsScored   float64
	FastestLap     bool
	DNF            bool
	Classified     bool // Completed >90% race distance
}

// BasicSeasonStats holds the minimum publicly available data for a season
type F1BasicSeasonStats struct {
	Year           int
	Team           string
	Points         float64
	Wins           int
	Podiums        int
	Races          int
	PointFinishes  int     // Number of races finished in points
	DNFs           int     // Number of Did Not Finish results
	TeamPoints     float64 // Team's total points that season
	TeamPosition   int     // Team's championship position
	TeammatePoints float64 // Points scored by teammate
	RecentRaces    []F1RaceResult
}

// CompleteDriver combines user input with calculated attributes
type F1CompleteDriver struct {
	// User-provided data
	BasicData F1BasicDriverData

	// Maps for specialties/weaknesses
	SpecialtiesMap map[string]bool
	WeaknessesMap  map[string]bool

	// Calculated attributes
	PointsVsTeammate       float64
	OverperformanceFactor  float64
	CareerTeamChanges      int
	SocialMediaFollowers   float64
	MediaMentions          float64
	MerchandiseSales       float64
	HomeMarketSize         float64
	FanbaseSize            float64
	ThreeYearTrend         float64
	TeamStrength           float64
	NormalizedTeamStrength float64

	// Technical abilities
	Abilities map[string]float64

	// Final calculated attributes for pricing
	PerformanceRatio float64
	Consistency      float64
	MarketPopularity float64
}

// TeamSeasonStats holds team performance statistics
type F1TeamSeasonStats struct {
	Position        int
	Points          float64
	WinCount        int     // Number of race wins
	PodiumCount     int     // Number of podiums
	PointsPercent   float64 // Percentage of total points available
	AvgGridPosition float64 // Average qualifying position
}

// DriverPrice represents a driver with their calculated price
type F1DriverPrice struct {
	Driver             F1CompleteDriver
	Price              float64
	ComponentBreakdown map[string]float64
}

//
// DRIVER ATTRIBUTE CALCULATION FUNCTIONS
//

type F1QuantumPricingModel struct {
	TotalNumberOfRaces     int
	CurrentRace            int
	TotalPointsInTheSeason int
}

func NewF1QuantumPricingModel(totalNumberOfRaces int, currentRace int, totalPointsInTheSeason int) *F1QuantumPricingModel {
	return &F1QuantumPricingModel{
		TotalNumberOfRaces:     totalNumberOfRaces,
		CurrentRace:            currentRace,
		TotalPointsInTheSeason: totalPointsInTheSeason,
	}
}

// NewCompleteDriver creates a complete driver from basic data and calculates all attributes
func (m *F1QuantumPricingModel) NewCompleteDriver(basicData F1BasicDriverData) F1CompleteDriver {
	// Create complete driver with empty maps
	driver := F1CompleteDriver{
		BasicData:      basicData,
		SpecialtiesMap: make(map[string]bool),
		WeaknessesMap:  make(map[string]bool),
		Abilities:      make(map[string]float64),
	}

	// Convert specialty/weakness arrays to maps
	for _, specialty := range basicData.Specialties {
		driver.SpecialtiesMap[specialty] = true
	}

	for _, weakness := range basicData.Weaknesses {
		driver.WeaknessesMap[weakness] = true
	}

	// Calculate all derived metrics
	m.calculateAllDerivedMetrics(&driver)

	return driver
}

// calculateAllDerivedMetrics computes all derived metrics from basic data
func (m *F1QuantumPricingModel) calculateAllDerivedMetrics(driver *F1CompleteDriver) {
	// Calculate career team changes
	driver.CareerTeamChanges = m.calculateCareerTeamChanges(driver)

	// Calculate points vs teammate ratio
	driver.PointsVsTeammate = m.calculatePointsVsTeammate(driver)

	// Calculate team strength
	driver.TeamStrength = m.calculateCurrentTeamStrength(driver)

	// Calculate popularity metrics
	m.calculatePopularityMetrics(driver)

	// Calculate three-year trend
	driver.ThreeYearTrend = m.calculateThreeYearTrend(driver)

	// Calculate overperformance factor
	driver.OverperformanceFactor = m.calculateOverperformanceFactor(driver)

	// Calculate technical abilities
	driver.Abilities = m.deriveDriverAbilities(driver)

	// Calculate final attributes for pricing
	driver.PerformanceRatio = m.calculatePerformanceRatio(driver)
	driver.Consistency = m.calculateConsistency(driver)
	driver.MarketPopularity = m.calculateMarketPopularity(driver)
}

// calculateCareerTeamChanges counts how many different teams a driver has raced for
func (m *F1QuantumPricingModel) calculateCareerTeamChanges(driver *F1CompleteDriver) int {
	// Early exit if no seasons data
	if len(driver.BasicData.Seasons) == 0 {
		return 0
	}

	// Track unique teams
	uniqueTeams := make(map[string]bool)
	for _, season := range driver.BasicData.Seasons {
		if season.Team != "" {
			uniqueTeams[season.Team] = true
		}
	}

	// Return the count of unique teams minus 1 (team changes)
	return len(uniqueTeams) - 1
}

// calculatePointsVsTeammate calculates the percentage of team points scored by driver
func (m *F1QuantumPricingModel) calculatePointsVsTeammate(driver *F1CompleteDriver) float64 {
	// Early exit if no seasons data
	if len(driver.BasicData.Seasons) == 0 {
		return 0.5 // Default equal split
	}

	// Focus on most recent season with a teammate

	sort.Slice(driver.BasicData.Seasons, func(i, j int) bool {
		return driver.BasicData.Seasons[i].Year > driver.BasicData.Seasons[j].Year
	})

	var recentSeason *F1BasicSeasonStats = &driver.BasicData.Seasons[0]

	// Calculate points ratio
	totalPoints := recentSeason.TeamPoints
	if totalPoints == 0 {
		// If neither driver scored points, estimate based on career stats
		if driver.BasicData.CareerPodiums > 0 && driver.BasicData.CareerStarts > 50 {
			return 0.75 // Veteran with podiums likely better than teammates in non-scoring teams
		}

		if driver.BasicData.IsRookie {
			return 0.32 // Veteran with podiums likely better than teammates in non-scoring teams
		}

		return 0.5 // Default if no points scored
	}

	return recentSeason.Points / totalPoints
}

// calculatePopularityMetrics estimates social media, merchandise sales, etc.
func (m *F1QuantumPricingModel) calculatePopularityMetrics(driver *F1CompleteDriver) {
	// Base popularity from championships and podiums
	basePopularity := 0.50

	// Championship bonus
	championBonus := math.Min(float64(driver.BasicData.ChampionshipWins)*0.07, 0.30)

	// Podium bonus
	podiumBonus := math.Min(float64(driver.BasicData.CareerPodiums)*0.002, 0.20)

	// Experience factor
	experienceBonus := math.Min(float64(driver.BasicData.CareerStarts)/200.0*0.15, 0.15)

	// Combined base
	combinedBase := basePopularity + championBonus + podiumBonus + experienceBonus

	// Recent success boost
	recentSuccessBoost := 0.0
	if len(driver.BasicData.Seasons) > 0 && driver.BasicData.Seasons[0].Wins > 0 {
		recentSuccessBoost = math.Min(float64(driver.BasicData.Seasons[0].Wins)*0.02, 0.2)
	}

	// Social media bonus for younger drivers
	socialMediaAge := 0.0
	if driver.BasicData.Age < 25 {
		socialMediaAge = 0.15
	} else if driver.BasicData.Age < 30 {
		socialMediaAge = 0.10
	} else if driver.BasicData.Age < 35 {
		socialMediaAge = 0.05
	}

	// Estimate social media following
	driver.SocialMediaFollowers = math.Min(combinedBase+socialMediaAge+recentSuccessBoost, 0.95)

	// Estimate media mentions (influenced by performance and championships)
	driver.MediaMentions = math.Min(combinedBase+recentSuccessBoost*1.5, 0.95)

	// Estimate merchandise sales (influenced by fanbase loyalty and team)
	teamMerchandiseBonus := 0.0
	if len(driver.BasicData.Seasons) > 0 {
		teamPosition := driver.BasicData.Seasons[0].TeamPosition
		if teamPosition == 1 {
			teamMerchandiseBonus = 0.15 // Top team has strong merchandise sales
		} else if teamPosition <= 3 {
			teamMerchandiseBonus = 0.10 // Other top teams
		}
	}
	driver.MerchandiseSales = math.Min(combinedBase+teamMerchandiseBonus, 0.95)

	// Use user-provided market popularity
	switch driver.BasicData.MarketPopularity {
	case F1HighPopularity:
		driver.HomeMarketSize = 0.85
		driver.FanbaseSize = 0.85
	case F1MediumPopularity:
		driver.HomeMarketSize = 0.65
		driver.FanbaseSize = 0.65
	case F1LowPopularity:
		driver.HomeMarketSize = 0.45
		driver.FanbaseSize = 0.45
	default:
		// Default to medium if not specified
		driver.HomeMarketSize = 0.65
		driver.FanbaseSize = 0.65
	}
}

// calculateCurrentTeamStrength calculates the current season team strength using the provided team data
func (m *F1QuantumPricingModel) calculateCurrentTeamStrength(driver *F1CompleteDriver) float64 {
	// Use the team data provided directly
	teamData := driver.BasicData.TeamData

	// Create team statistics from actual data
	teamStats := F1TeamSeasonStats{
		Position:      teamData.SeasonPosition,
		Points:        teamData.SeasonPoints,
		WinCount:      teamData.Wins,
		PodiumCount:   teamData.Podiums,
		PointsPercent: 0.0, // Will be calculated if needed
	}

	// Calculate average qualifying position if data available
	if len(teamData.RecentQualifyingPositions) > 0 {
		sum := 0.0
		for _, pos := range teamData.RecentQualifyingPositions {
			sum += pos
		}
		teamStats.AvgGridPosition = sum / float64(len(teamData.RecentQualifyingPositions))
	}

	var qualifyingScore float64 = 0.0
	if teamStats.AvgGridPosition > 0 {
		qualifyingScore = (10 - teamStats.AvgGridPosition) / 10
	}

	// We need to estimate the total points in the season for the points percentage
	// This can be based on the expected points per team or a fixed value
	teamStats.PointsPercent = teamData.SeasonPoints / float64(m.TotalPointsInTheSeason)

	// Use budget tier to help determine relative strength
	budgetFactor := 0.0
	switch teamData.BudgetTier {
	case "Top":
		budgetFactor = 1
	case "Upper-Mid":
		budgetFactor = 0.25
	case "Lower-Mid":
		budgetFactor = -0.25
	case "Backmarker":
		budgetFactor = -0.5
	}

	// Get team season positions for all teams in the competition
	// For simplicity, we'll assume 10 teams total
	numberOfTeams := 10

	// Calculate normalized position score (1.0 for first, 0.1 for last)
	positionScore := float64(numberOfTeams-teamData.SeasonPosition+1) / float64(numberOfTeams)
	positionScore = math.Pow(positionScore, 0.7) // Non-linear scaling
	positionScore = (positionScore - 0.5) / 0.5

	// Calculate recent form from race positions if available
	recentFormScore := 0.0
	if len(teamData.RecentRacePositions) > 0 {
		sum := 0.0
		for _, pos := range teamData.RecentRacePositions {
			sum += pos
		}
		avgPosition := sum / float64(len(teamData.RecentRacePositions))
		// Scale to 0-1 (1 is best)
		recentFormScore = (10 - avgPosition) / 10.0
	}

	// Recent upgrades bonus
	upgradeBonus := 0.0
	if teamData.RecentUpgrades {
		upgradeBonus = 0.05
	}

	// Historical performance pattern - bonus for consistently good teams
	pointsTrend := 0.0
	if len(teamData.SeasonHistory) > 0 {
		sort.Slice(teamData.SeasonHistory, func(i, j int) bool {
			return teamData.SeasonHistory[i].Year < teamData.SeasonHistory[j].Year
		})

		var teamPointsPerRace []float64
		for _, season := range teamData.SeasonHistory {
			pointsPerRace := float64(season.Points) / float64(season.Races)
			teamPointsPerRace = append(teamPointsPerRace, pointsPerRace)
		}

		teamPointsPerRace = append(teamPointsPerRace, float64(teamData.SeasonPoints)/float64(teamData.CurrentRace))

		if len(teamPointsPerRace) >= 2 {
			var changes []float64
			var weights []float64

			// Calculate all season-to-season changes
			for i := 1; i < len(teamPointsPerRace); i++ {
				change := teamPointsPerRace[i] - teamPointsPerRace[i-1]
				changes = append(changes, change)
				// Recent changes get higher weights: [1, 2, 3, 4] for 5 seasons
				weights = append(weights, float64(i))
			}

			var weightedSum, totalWeight float64
			for i, change := range changes {
				weightedSum += change * weights[i]
				totalWeight += weights[i]
			}

			avgChange := weightedSum / totalWeight

			// Calculate standard deviation of changes for normalization
			var changeSum, changeSquaredSum float64
			for _, change := range changes {
				changeSum += change
				changeSquaredSum += change * change
			}

			changeMean := changeSum / float64(len(changes))
			changeVariance := (changeSquaredSum / float64(len(changes))) - (changeMean * changeMean)
			changeStdDev := math.Sqrt(math.Max(changeVariance, 0.1)) // Minimum 0.1 to avoid tiny values

			// Scale trend using standard deviation
			// avgChange ± 2σ = trend ± 1.0

			if changeStdDev > 0 {
				pointsTrend = avgChange / (2.0 * changeStdDev)

				// Clamp to [-1, 1]
				pointsTrend = math.Tanh(pointsTrend)
			}

		}

	}

	// Calculate combined strength score
	strengthScore := 0.0

	// Combine factors
	var raceSuccessScore float64 = 0
	if m.CurrentRace != 0 {
		raceSuccessScore = (float64(teamStats.WinCount)/float64(m.CurrentRace))*0.65 +
			(float64(teamStats.PodiumCount)/float64(m.CurrentRace))*0.35
	}

	strengthScore = (positionScore * 0.20) +
		(raceSuccessScore * 0.10) +
		(recentFormScore * 0.35) +
		(qualifyingScore * 0.10) + (budgetFactor * 0.15) + upgradeBonus +
		(pointsTrend * 0.1)

	return strengthScore
}

func (m *F1QuantumPricingModel) calculateSeasonPhase(driver *F1CompleteDriver) F1SeasonPhase {
	// Default to mid-season if data not available
	if driver.BasicData.CurrentRaceNumber <= 0 || driver.BasicData.TotalRacesInSeason <= 0 {
		return F1MidSeason
	}

	// Calculate percentage through season
	seasonProgress := float64(driver.BasicData.CurrentRaceNumber) / float64(driver.BasicData.TotalRacesInSeason)

	if seasonProgress < 0.25 {
		return F1EarlySeason
	} else if seasonProgress > 0.75 {
		return F1LateSeason
	} else {
		return F1MidSeason
	}
}

// calculateThreeYearTrend computes a driver's performance trajectory
func (m *F1QuantumPricingModel) calculateThreeYearTrend(driver *F1CompleteDriver) float64 {
	// Early exit if rookie
	if driver.BasicData.IsRookie {
		return 0
	}

	// Check for hiatus in career
	hasHiatus := false
	if len(driver.BasicData.Seasons) >= 2 {
		yearGap := driver.BasicData.Seasons[0].Year - driver.BasicData.Seasons[1].Year
		if yearGap > 1 {
			hasHiatus = true
		}
	}

	// Early exit if insufficient data
	if len(driver.BasicData.Seasons) < 2 {
		return 0.0 // Neutral value
	}

	// Get up to 4 seasons of data
	// Calculate normalized points for each season
	normalizedPoints := make([]float64, len(driver.BasicData.Seasons))

	// For the previous seasons, we need to estimate team strength using historical data
	for i, season := range driver.BasicData.Seasons {
		// Normalize points by team strength
		normalizedPoints[i] = season.Points / float64(season.Races)
	}

	n := float64(len(normalizedPoints))
	sumX := 0.0
	sumY := 0.0
	sumXY := 0.0
	sumX2 := 0.0

	for i, point := range normalizedPoints {
		x := float64(i)
		y := point
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Calculate slope using least squares method
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)

	// Special handling for returning drivers after hiatus
	if !hasHiatus && len(normalizedPoints) >= 2 {
		// Scale to our -1.0 to 1.0 range
		return math.Max(-1.0, math.Min(1.0, slope))
	}

	if slope > 0 {
		return math.Max(-1.0, math.Min(1.0, slope))
	}

	if slope < -1 {
		return -0.4
	}

	return math.Max(-1.0, math.Min(1.0, slope+0.2))
}

// calculateOverperformanceFactor estimates how much a driver exceeds car potential
func (m *F1QuantumPricingModel) calculateOverperformanceFactor(driver *F1CompleteDriver) float64 {
	// Base value related to points vs teammate
	baseValue := driver.PointsVsTeammate
	baseValue = (baseValue - 0.5) / 0.5

	// Adjust for car performance - overperforming in a weak car is more impressive
	driversWinRate := 0.0
	if driver.BasicData.TeamData.Wins != 0 {
		seasonWins := 0
		idx := slices.IndexFunc(driver.BasicData.Seasons, func(s F1BasicSeasonStats) bool {
			return s.Year == driver.BasicData.TeamData.Year
		})
		if idx != -1 {
			seasonWins = driver.BasicData.Seasons[idx].Wins
		}

		driversWinRate = float64(seasonWins) / float64(driver.BasicData.TeamData.Wins)
		driversWinRate = (driversWinRate - 0.5) / 0.5
	}

	driversPodiumRate := 0.0
	if driver.BasicData.TeamData.Podiums != 0 {
		seasonPodiums := 0
		idx := slices.IndexFunc(driver.BasicData.Seasons, func(s F1BasicSeasonStats) bool {
			return s.Year == driver.BasicData.TeamData.Year
		})
		if idx != -1 {
			seasonPodiums = driver.BasicData.Seasons[idx].Podiums
		}

		driversPodiumRate = float64(seasonPodiums) / float64(driver.BasicData.TeamData.Podiums)
		driversPodiumRate = (driversPodiumRate - 0.5) / 0.5
	}

	// Adjust for regularly outperforming in qualifying (if team data available)
	qualifyingAdjustment := 0.0
	avgTeamQualifying := 0.0
	if len(driver.BasicData.TeamData.RecentQualifyingPositions) > 0 {
		// Check if driver has beaten typical team qualifying positions

		for _, pos := range driver.BasicData.TeamData.RecentQualifyingPositions {
			avgTeamQualifying += pos
		}
		avgTeamQualifying /= float64(len(driver.BasicData.TeamData.RecentQualifyingPositions))

		// Get driver's average qualifying position if available
		driverQualifying := 0.0
		qualifyingCount := 0
		if len(driver.BasicData.Seasons) > 0 {
			sort.Slice(driver.BasicData.Seasons, func(i, j int) bool {
				return driver.BasicData.Seasons[i].Year > driver.BasicData.Seasons[j].Year
			})
			for _, race := range driver.BasicData.Seasons[0].RecentRaces {
				if race.StartPosition > 0 {
					driverQualifying += float64(race.StartPosition)
					qualifyingCount++
				}
			}
			if qualifyingCount > 0 {
				driverQualifying /= float64(qualifyingCount)

				// Calculate position difference (negative = driver is better)
				positionDifference := avgTeamQualifying - driverQualifying

				// Scale to -1 to +1 range based on reasonable F1 position differences
				// Assume max difference of ±10 positions (adjust as needed)
				maxPositionDiff := 10.0
				qualifyingAdjustment = math.Max(-1.0, math.Min(1.0, 0.5+positionDifference/maxPositionDiff))
			}
		}
	}

	// Calculate final overperformance factor
	overperformanceFactor := baseValue*0.35 + driversWinRate*0.3 + driversPodiumRate*0.2 + qualifyingAdjustment*0.15

	// Ensure within valid range
	return overperformanceFactor
}

//
// DRIVER TECHNICAL ABILITIES CALCULATION
//

// deriveDriverAbilities calculates technical abilities from driver's career statistics and classification
func (m *F1QuantumPricingModel) deriveDriverAbilities(driver *F1CompleteDriver) map[string]float64 {
	// Initialize attributes map with base values
	abilities := map[string]float64{
		"WetWeather":        0.7,
		"TireManagement":    0.7,
		"BrakingStability":  0.7,
		"TechnicalCorners":  0.7,
		"RaceStart":         0.7,
		"QualifyingPace":    0.7,
		"SetupAdaptability": 0.7,
		"OvertakingSkill":   0.7,
		"RaceConsistency":   0.7,
		"ERSManagement":     0.7,
		"FuelSaving":        0.7,
		"SafetyCarRestart":  0.7,
	}

	// --- BASE MODIFIERS FROM CAREER STATISTICS ---
	experienceFactor := math.Min(float64(driver.BasicData.CareerStarts)/150.0, 1.0)
	abilities["TireManagement"] += experienceFactor * 0.1
	abilities["TechnicalCorners"] += experienceFactor * 0.05

	championFactor := math.Min(float64(driver.BasicData.ChampionshipWins)*0.03, 0.15)
	for key := range abilities {
		abilities[key] += championFactor
	}

	podiumRatio := float64(driver.BasicData.CareerPodiums) / math.Max(float64(driver.BasicData.CareerStarts), 1.0)
	technicalBoost := math.Min(podiumRatio*3.0, 0.15)
	abilities["TechnicalCorners"] += technicalBoost
	abilities["BrakingStability"] += technicalBoost

	if len(driver.BasicData.Seasons) > 0 {
		recentFormBoost := math.Min(driver.BasicData.Seasons[0].Points/400.0, 1.0) * 0.1
		for key := range abilities {
			abilities[key] += recentFormBoost
		}
	}

	// --- PRIMARY DRIVER STYLE MODIFIERS ---
	switch driver.BasicData.PrimaryStyle {
	case F1Aggressive:
		abilities["BrakingStability"] += 0.10
		abilities["RaceStart"] += 0.15
		abilities["TireManagement"] -= 0.10
		abilities["OvertakingSkill"] += 0.10
		abilities["FuelSaving"] -= 0.05
	case F1Smooth:
		abilities["TireManagement"] += 0.15
		abilities["TechnicalCorners"] += 0.10
		abilities["WetWeather"] += 0.05
		abilities["RaceStart"] -= 0.05
		abilities["RaceConsistency"] += 0.10
		abilities["FuelSaving"] += 0.05
	case F1Defensive:
		abilities["BrakingStability"] += 0.10
		abilities["TechnicalCorners"] += 0.05
		abilities["RaceStart"] += 0.05
		abilities["RaceConsistency"] += 0.10
	case F1Overtaker:
		abilities["BrakingStability"] += 0.15
		abilities["RaceStart"] += 0.10
		abilities["TireManagement"] -= 0.05
		abilities["OvertakingSkill"] += 0.15
	case F1AllRounder:
		for key := range abilities {
			abilities[key] += 0.05
		}
	case F1Rookie:
		for key := range abilities {
			abilities[key] -= 0.05
		}
	}

	// --- SECONDARY DRIVER STYLE MODIFIERS ---
	if driver.BasicData.SecondaryStyle != "" && driver.BasicData.SecondaryStyle != driver.BasicData.PrimaryStyle {
		switch driver.BasicData.SecondaryStyle {
		case F1Aggressive:
			abilities["BrakingStability"] += 0.05
			abilities["RaceStart"] += 0.07
			abilities["TireManagement"] -= 0.05
			abilities["OvertakingSkill"] += 0.05
		case F1Smooth:
			abilities["TireManagement"] += 0.07
			abilities["TechnicalCorners"] += 0.05
			abilities["WetWeather"] += 0.02
			abilities["RaceConsistency"] += 0.05
		case F1Defensive:
			abilities["BrakingStability"] += 0.05
			abilities["TechnicalCorners"] += 0.02
			abilities["RaceConsistency"] += 0.05
		case F1Overtaker:
			abilities["BrakingStability"] += 0.07
			abilities["RaceStart"] += 0.05
			abilities["OvertakingSkill"] += 0.07
		case F1AllRounder:
			for key := range abilities {
				abilities[key] += 0.02
			}
		}
	}

	// --- SPECIALTIES MODIFIERS ---
	for specialty := range driver.SpecialtiesMap {
		switch specialty {
		case "Street", "TechnicalCorners":
			abilities["TechnicalCorners"] += 0.15
			abilities["BrakingStability"] += 0.10
		case "WetWeather":
			abilities["WetWeather"] += 0.20
			abilities["BrakingStability"] += 0.05
		case "TechnicalTracks":
			abilities["TechnicalCorners"] += 0.15
			abilities["TireManagement"] += 0.05
		case "FastCorners":
			abilities["TechnicalCorners"] += 0.10
			abilities["TireManagement"] += 0.05
		case "RaceStart":
			abilities["RaceStart"] += 0.15
		case "TireManagement":
			abilities["TireManagement"] += 0.15
		case "OvertakingSkill":
			abilities["OvertakingSkill"] += 0.15
			abilities["RaceStart"] += 0.05
		case "RaceConsistency":
			abilities["RaceConsistency"] += 0.15
		case "ERSManagement":
			abilities["ERSManagement"] += 0.10
		case "FuelSaving":
			abilities["FuelSaving"] += 0.10
		case "SafetyCarRestart":
			abilities["SafetyCarRestart"] += 0.10
		case "QualifyingPace": // or "Qualifying" if that’s what your input uses
			abilities["QualifyingPace"] += 0.15
			abilities["SetupAdaptability"] += 0.05
		case "SetupAdaptability":
			abilities["SetupAdaptability"] += 0.15
		default:
			log.Printf("Unknown specialty: %s", specialty)
		}
	}

	// --- WEAKNESSES MODIFIERS ---
	for weakness := range driver.WeaknessesMap {
		switch weakness {
		case "Street", "TechnicalCorners":
			abilities["TechnicalCorners"] -= 0.10
		case "WetWeather":
			abilities["WetWeather"] -= 0.15
		case "TechnicalTracks":
			abilities["TechnicalCorners"] -= 0.10
		case "FastCorners":
			abilities["TechnicalCorners"] -= 0.05
		case "RaceStart":
			abilities["RaceStart"] -= 0.10
		case "TireManagement":
			abilities["TireManagement"] -= 0.10
		case "OvertakingSkill":
			abilities["OvertakingSkill"] -= 0.10
		case "RaceConsistency":
			abilities["RaceConsistency"] -= 0.10
		case "ERSManagement":
			abilities["ERSManagement"] -= 0.10
		case "FuelSaving":
			abilities["FuelSaving"] -= 0.10
		case "SafetyCarRestart":
			abilities["SafetyCarRestart"] -= 0.10
		case "QualifyingPace": // or "Qualifying"
			abilities["QualifyingPace"] -= 0.10
			abilities["SetupAdaptability"] -= 0.05
		case "SetupAdaptability":
			abilities["SetupAdaptability"] -= 0.10
		default:
			log.Printf("Unknown weakness: %s", weakness)
		}
	}

	// --- ENSURE ALL VALUES STAY WITHIN 0.0-1.0 RANGE ---
	for key := range abilities {
		abilities[key] = math.Max(0.0, math.Min(abilities[key], 1.0))
		abilities[key] = math.Round(abilities[key]*100) / 100
	}

	return abilities
}

//
// FINAL ATTRIBUTE CALCULATIONS
//

// calculatePerformanceRatio measures how a driver performs relative to their car's potential
// Range: 0.70-0.98
func (m *F1QuantumPricingModel) calculatePerformanceRatio(driver *F1CompleteDriver) float64 {

	if len(driver.BasicData.Seasons) == 0 {
		return 0 // Default for drivers with no season data
	}

	sort.Slice(driver.BasicData.Seasons, func(i, j int) bool {
		return driver.BasicData.Seasons[i].Year > driver.BasicData.Seasons[j].Year
	})

	var totalGain, raceCount float64
	for _, race := range driver.BasicData.Seasons[0].RecentRaces {
		if race.StartPosition > 0 && race.FinishPosition > 0 && race.FinishPosition <= 20 {
			gain := float64(race.StartPosition - race.FinishPosition) // Positive = gained positions
			totalGain += gain
			raceCount++
		}
	}

	var gainScore float64
	if raceCount > 0 {
		avgGain := totalGain / raceCount
		// Normalize to [-1, 1]: +3 avg = +1, 0 avg = 0, -3 avg = -1
		gainScore = math.Max(-1.0, math.Min(1.0, avgGain/4.0))
	}

	var trendRaces []float64
	var trendScore float64

	for _, race := range driver.BasicData.Seasons[0].RecentRaces {
		if race.FinishPosition > 0 && race.FinishPosition <= 20 {
			trendRaces = append(trendRaces, float64(race.FinishPosition))
		}
	}
	if len(trendRaces) >= 3 { // Need minimum races for meaningful trend
		// Calculate linear regression slope (trend)
		n := float64(len(trendRaces))
		var sumX, sumY, sumXY, sumX2 float64

		for i, position := range trendRaces {
			x := float64(i + 1) // Race number
			y := position       // Final position
			sumX += x
			sumY += y
			sumXY += x * y
			sumX2 += x * x
		}

		// Linear regression slope: negative slope = improving positions (lower numbers)
		slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)

		// Normalize slope to [-1, 1]: -1.0 slope = +1, 0 slope = 0, +1.0 slope = -1
		// Negative slope (improving) is good, positive slope (declining) is bad
		trendScore = math.Max(-1.0, math.Min(1.0, -slope))
	} else {
		trendScore = 0.0 // Neutral if insufficient data
	}

	// Measures raw pace regardless of grid/finish position
	var fastestLaps, totalRaces float64
	for _, race := range driver.BasicData.Seasons[0].RecentRaces {
		totalRaces++
		if race.FastestLap {
			fastestLaps++
		}
	}

	var fastestLapScore float64
	if totalRaces > 0 {
		fastestLapRate := fastestLaps / totalRaces
		// Normalize: 20% rate = +1, 10% = 0, 0% = -1
		fastestLapScore = math.Max(-1.0, math.Min(1.0, (fastestLapRate-0.075)/0.075))
	}

	var dnfCount float64
	for _, race := range driver.BasicData.Seasons[0].RecentRaces {
		if race.DNF {
			dnfCount++
		}
	}

	var dnfScore float64
	if totalRaces > 0 {
		dnfRate := dnfCount / totalRaces
		// Normalize to [-1, 1]: 0% DNF rate = +1, 15% = 0, 30% = -1
		dnfScore = math.Max(-1.0, math.Min(1.0, (0.15-dnfRate)/0.15))
	}

	performanceRatio := gainScore*0.5 + dnfScore*0.1 + fastestLapScore*0.1 + trendScore*0.3

	// Scale to appropriate range (0.70-0.98)
	return performanceRatio
}

// calculateConsistency measures how reliable a driver's performance is
// Range: 0.65-0.95
func (m *F1QuantumPricingModel) calculateConsistency(driver *F1CompleteDriver) float64 {

	sort.Slice(driver.BasicData.Seasons, func(i, j int) bool {
		return driver.BasicData.Seasons[i].Year > driver.BasicData.Seasons[j].Year
	})

	if len(driver.BasicData.Seasons) == 0 {
		return 0.80 // Default for drivers with no season data
	}

	// COMPONENT 1: Qualifying Position Consistency (35%)
	var qualifyingPositions []float64
	for _, race := range driver.BasicData.Seasons[0].RecentRaces {
		if race.StartPosition > 0 {
			qualifyingPositions = append(qualifyingPositions, float64(race.StartPosition))
		}
	}

	var qualifyingConsistencyScore float64
	if len(qualifyingPositions) > 1 {
		stdDev := stat.StdDev(qualifyingPositions, nil)
		maxPos := slices.Max(qualifyingPositions)
		minPos := slices.Min(qualifyingPositions)
		positionRange := maxPos - minPos

		if positionRange > 0 {
			consistencyRatio := stdDev / positionRange
			qualifyingConsistencyScore = math.Max(-1.0, math.Min(1.0, 1.0-(2.0*consistencyRatio)))
		} else {
			qualifyingConsistencyScore = 1.0
		}
	}

	var finalPositions []float64
	for _, race := range driver.BasicData.Seasons[0].RecentRaces {
		if race.FinishPosition > 0 && race.FinishPosition <= 20 {
			finalPositions = append(finalPositions, float64(race.FinishPosition))
		}
	}

	var finalPositionConsistencyScore float64
	if len(finalPositions) > 1 {
		stdDev := stat.StdDev(finalPositions, nil)
		maxPos := slices.Max(finalPositions)
		minPos := slices.Min(finalPositions)
		positionRange := maxPos - minPos

		if positionRange > 0 {

			consistencyRatio := stdDev / positionRange
			finalPositionConsistencyScore = math.Max(-1.0, math.Min(1.0, 1.0-(2.0*consistencyRatio)))
		} else {
			// Same position every time = perfect consistency
			finalPositionConsistencyScore = 1.0
		}
	}

	var totalRaces, finishedRaces float64
	for _, race := range driver.BasicData.Seasons[0].RecentRaces {
		totalRaces++
		if !race.DNF {
			finishedRaces++
		}
	}
	var dnfConsistencyScore float64
	if finishedRaces != totalRaces {
		dnfConsistencyScore = -0.2 * (totalRaces - finishedRaces)
	}

	return qualifyingConsistencyScore*0.3 + finalPositionConsistencyScore*0.5 + dnfConsistencyScore*0.2
}

// calculateMarketPopularity measures a driver's popularity with fans
func (m *F1QuantumPricingModel) calculateMarketPopularity(driver *F1CompleteDriver) float64 {

	// COMPONENT 3: Career Longevity (15%)
	careerLongevityFactor := math.Min(float64(driver.BasicData.CareerStarts)/150.0, 1.0)

	// COMPONENT 4: User-specified popularity level (40%)
	userSpecifiedPopularity := 0.5 // Medium default
	switch driver.BasicData.MarketPopularity {
	case F1HighPopularity:
		userSpecifiedPopularity = 0.9
	case F1MediumPopularity:
		userSpecifiedPopularity = 0.6
	case F1LowPopularity:
		userSpecifiedPopularity = 0.3
	}

	// Combine all factors with weights
	rawPopularity := (careerLongevityFactor * 0.25) + (userSpecifiedPopularity * 0.75)

	// Scale to appropriate range (0.40-0.95)
	return rawPopularity
}

//
// DRIVER PRICING SYSTEM
//

// calculateDriverPrice computes the final fantasy price for a driver
func (m *F1QuantumPricingModel) calculateDriverPrice(driver F1CompleteDriver) F1DriverPrice {
	// Determine season phase
	seasonPhase := m.calculateSeasonPhase(&driver)

	// Base price from team strength
	basePrice := math.Round(15.0 + (driver.NormalizedTeamStrength * 5.0))

	fmt.Println(driver.BasicData.Name)
	fmt.Println(driver.BasicData.TeamData.Name)
	fmt.Println(basePrice)

	// Base price from team strength
	// teammatePremium := math.Max(-2, math.Min(2, driver.OverperformanceFactor*3.0))
	teammatePremium := math.Max(2, driver.OverperformanceFactor*2)

	// Performance premium
	performancePremium := driver.PerformanceRatio * 3.0

	// Consistency value
	consistencyValue := (driver.Consistency) * 3.0

	// Championship premium
	championPremium := math.Min(float64(driver.BasicData.ChampionshipWins)*0.5, 3)

	// Trend adjustment
	trendAdjustment := driver.ThreeYearTrend * 3.0

	// Popularity premium - with season phase adjustment
	popularityPremium := (driver.MarketPopularity - 0.5) * 2.0

	// Recent team upgrades adjustment
	upgradeAdjustment := 0.0
	if driver.BasicData.TeamData.RecentUpgrades {
		upgradeAdjustment = 1 // Small bonus for teams that recently upgraded
	}

	// Mid-season team change adjustment
	teamChangeAdjustment := 0.0
	if driver.BasicData.PreviousTeam != "" {
		// Calculate experience factor with new team
		racesWithTeam := driver.BasicData.RacesWithCurrentTeam
		// Apply discount for limited experience with current team
		if racesWithTeam < 5 {
			teamChangeAdjustment = -2.0 + (0.4 * float64(racesWithTeam)) // Starts at -2.0M, reduces with each race
		}
	}

	// Season phase adjustment
	seasonPhaseAdjustment := 0.0
	switch seasonPhase {
	case F1EarlySeason:
		seasonPhaseAdjustment = trendAdjustment * 0.5 // More impact from trend
	case F1LateSeason:
		// Late season: More influenced by current performance
		seasonPhaseAdjustment = performancePremium * 0.5 // More impact from recent form
	}

	// --- Ability-based premium (uses ALL derived abilities) ---
	totalAbility := 0.0
	for _, val := range driver.Abilities {
		totalAbility += val
	}
	abilityScore := totalAbility / float64(len(driver.Abilities))
	abilityPremium := (abilityScore - 0.5) * 5.0 // ±4M range

	// Calculate raw price with all adjustments
	rawPrice := basePrice + teammatePremium + performancePremium + championPremium + trendAdjustment + popularityPremium +
		consistencyValue + upgradeAdjustment + teamChangeAdjustment + seasonPhaseAdjustment + abilityPremium

	if driver.BasicData.IsRookie && len(driver.BasicData.Seasons) == 0 {
		rawPrice = basePrice
	}

	// Create price breakdown for analysis
	breakdown := map[string]float64{
		"Base (Team Strength)":    basePrice,
		"Performance Premium":     performancePremium,
		"Championship Premium":    championPremium,
		"Trend Adjustment":        trendAdjustment,
		"Popularity Premium":      popularityPremium,
		"Consistency Value":       consistencyValue,
		"Upgrade Adjustment":      upgradeAdjustment,
		"Team Change Adjustment":  teamChangeAdjustment,
		"Season Phase Adjustment": seasonPhaseAdjustment,
		"Ability Premium":         abilityPremium,
		"Raw Price":               rawPrice,
	}

	// Apply psychological price anchoring
	finalPrice := m.applyPsychologicalPricing(rawPrice, driver.TeamStrength)
	breakdown["Final Price"] = finalPrice

	// Create and return DriverPrice object
	driverPrice := F1DriverPrice{
		Driver:             driver,
		Price:              finalPrice,
		ComponentBreakdown: breakdown,
	}

	return driverPrice
}

// applyPsychologicalPricing applies strategic price anchoring
func (m *F1QuantumPricingModel) applyPsychologicalPricing(price float64, teamStrength float64) float64 {
	// Apply team-based price anchoring
	if teamStrength > 1.3 { // Top teams
		if price > 20 {
			return math.Floor(price) + 0.9
		} else {
			return math.Floor(price) + 0.5
		}
	} else if teamStrength > 1.0 { // Midfield teams
		if price > 15 {
			return math.Floor(price) + 0.5
		} else {
			return math.Floor(price)
		}
	} else { // Backmarker teams
		return math.Floor(price*2) / 2 // Round to nearest 0.5
	}
}

// ProcessAllDrivers calculates all attributes and prices for a set of drivers
func (m *F1QuantumPricingModel) ProcessAllDrivers(basicDriversData []F1BasicDriverData) []F1DriverPrice {
	// Convert basic drivers to complete drivers
	completeDrivers := make([]F1CompleteDriver, len(basicDriversData))

	teamStrengthList := make([]float64, len(basicDriversData))
	for i, basicData := range basicDriversData {
		completeDrivers[i] = m.NewCompleteDriver(basicData)
		teamStrengthList[i] = completeDrivers[i].TeamStrength
	}

	teamStrengthMean := stat.Mean(teamStrengthList, nil)
	teamStrengthStd := stat.StdDev(teamStrengthList, nil)

	for i := range completeDrivers {
		completeDrivers[i].NormalizedTeamStrength = (completeDrivers[i].TeamStrength - teamStrengthMean) / teamStrengthStd
	}

	// Calculate prices for all drivers
	driverPrices := make([]F1DriverPrice, len(completeDrivers))
	for i, driver := range completeDrivers {
		driverPrices[i] = m.calculateDriverPrice(driver)
	}

	return driverPrices
}

// ProcessSingleDriver calculates attributes and price for a single driver
func (m *F1QuantumPricingModel) ProcessSingleDriver(basicDriverData F1BasicDriverData) F1DriverPrice {
	// Convert basic driver to complete driver
	completeDriver := m.NewCompleteDriver(basicDriverData)

	// Calculate and return price
	return m.calculateDriverPrice(completeDriver)
}

//
// HELPER FUNCTIONS FOR PRINTING RESULTS
//

// PrintDriverAttributesTable prints a formatted table of driver attributes
func (m *F1QuantumPricingModel) PrintDriverAttributesTable(driverPrices []F1DriverPrice) {
	fmt.Println("\n=== DRIVER ATTRIBUTES ===")
	fmt.Println(strings.Repeat("-", 105))
	fmt.Printf("%-20s %-12s %-12s %-12s %-12s %-12s\n",
		"DRIVER", "TEAM", "PERFORMANCE", "CONSISTENCY", "POPULARITY", "TREND")
	fmt.Println(strings.Repeat("-", 105))

	// Sort drivers by team
	sort.Slice(driverPrices, func(i, j int) bool {
		if driverPrices[i].Driver.BasicData.Team != driverPrices[j].Driver.BasicData.Team {
			return driverPrices[i].Driver.BasicData.Team < driverPrices[j].Driver.BasicData.Team
		}
		return driverPrices[i].Driver.BasicData.Name < driverPrices[j].Driver.BasicData.Name
	})

	currentTeam := ""

	for _, dp := range driverPrices {
		driver := dp.Driver

		// Print team header when team changes
		if driver.BasicData.Team != currentTeam {
			fmt.Printf("\n%-105s\n", driver.BasicData.Team)
			fmt.Println(strings.Repeat("-", 105))
			currentTeam = driver.BasicData.Team
		}

		fmt.Printf("%-20s %-12s %-12.2f %-12.2f %-12.2f %-12.2f\n",
			driver.BasicData.Name,
			driver.BasicData.Team,
			driver.PerformanceRatio,
			driver.Consistency,
			driver.MarketPopularity,
			driver.ThreeYearTrend)
	}

	fmt.Println(strings.Repeat("-", 105))
}

// PrintDriverAbilitiesTable prints a formatted table of driver technical abilities
func (m *F1QuantumPricingModel) PrintDriverAbilitiesTable(driverPrices []F1DriverPrice) {
	fmt.Println("\n=== DRIVER TECHNICAL ABILITIES ===")
	fmt.Println(strings.Repeat("-", 95))
	fmt.Printf("%-20s %-12s %-12s %-12s %-12s %-12s\n",
		"DRIVER", "WET WEATHER", "TIRE MGMT", "BRAKING", "TECHNICAL", "RACE START")
	fmt.Println(strings.Repeat("-", 95))

	// Sort drivers by team
	sort.Slice(driverPrices, func(i, j int) bool {
		if driverPrices[i].Driver.BasicData.Team != driverPrices[j].Driver.BasicData.Team {
			return driverPrices[i].Driver.BasicData.Team < driverPrices[j].Driver.BasicData.Team
		}
		return driverPrices[i].Driver.BasicData.Name < driverPrices[j].Driver.BasicData.Name
	})

	currentTeam := ""

	for _, dp := range driverPrices {
		driver := dp.Driver

		// Print team header when team changes
		if driver.BasicData.Team != currentTeam {
			fmt.Printf("\n%-95s\n", driver.BasicData.Team)
			fmt.Println(strings.Repeat("-", 95))
			currentTeam = driver.BasicData.Team
		}

		fmt.Printf("%-20s %-12.2f %-12.2f %-12.2f %-12.2f %-12.2f\n",
			driver.BasicData.Name,
			driver.Abilities["WetWeather"],
			driver.Abilities["TireManagement"],
			driver.Abilities["BrakingStability"],
			driver.Abilities["TechnicalCorners"],
			driver.Abilities["RaceStart"])
	}

	fmt.Println(strings.Repeat("-", 95))
}

// PrintDriverPrices prints formatted driver prices
func (m *F1QuantumPricingModel) PrintDriverPrices(driverPrices []F1DriverPrice) {
	fmt.Println("\n=== F1 FANTASY DRIVER PRICES ===")
	fmt.Println(strings.Repeat("-", 75))
	fmt.Printf("%-20s %-15s %-12s %-12s %s\n",
		"DRIVER", "TEAM", "PRICE", "TEAM STRENGTH", "BUDGET TIER")
	fmt.Println(strings.Repeat("-", 75))

	// Group by team for better visual organization
	teamDrivers := make(map[string][]F1DriverPrice)
	for _, dp := range driverPrices {
		teamDrivers[dp.Driver.BasicData.Team] = append(teamDrivers[dp.Driver.BasicData.Team], dp)
	}

	// Get teams in order from highest to lowest performance
	teams := make([]string, 0, len(teamDrivers))
	for team := range teamDrivers {
		teams = append(teams, team)
	}

	// Sort teams by average team strength
	sort.Slice(teams, func(i, j int) bool {
		teamIStrength := 0.0
		countI := 0
		for _, dp := range teamDrivers[teams[i]] {
			teamIStrength += dp.Driver.TeamStrength
			countI++
		}
		teamIAvg := teamIStrength / float64(countI)

		teamJStrength := 0.0
		countJ := 0
		for _, dp := range teamDrivers[teams[j]] {
			teamJStrength += dp.Driver.TeamStrength
			countJ++
		}
		teamJAvg := teamJStrength / float64(countJ)

		return teamIAvg > teamJAvg
	})

	// Print by team with team headers
	for _, team := range teams {
		dps := teamDrivers[team]

		// Sort drivers within team by price
		sort.Slice(dps, func(i, j int) bool {
			return dps[i].Price > dps[j].Price
		})

		// Print team name as header
		fmt.Printf("\n%-75s\n", team)
		fmt.Println(strings.Repeat("-", 75))

		// Print team drivers
		for _, dp := range dps {
			fmt.Printf("%-20s %-15s $%-12.1fM %-12.2f %s\n",
				dp.Driver.BasicData.Name,
				dp.Driver.BasicData.Team,
				dp.Price,
				dp.Driver.TeamStrength,
				dp.Driver.BasicData.TeamData.BudgetTier)
		}
	}

	// Print total budget information
	totalDrivers := len(driverPrices)
	avgPrice := 0.0
	for _, dp := range driverPrices {
		avgPrice += dp.Price
	}
	avgPrice /= float64(totalDrivers)

	fmt.Println(strings.Repeat("-", 75))
	fmt.Printf("\nAverage Driver Price: $%.1fM\n", avgPrice)
	fmt.Printf("Recommended Team Budget: $%.1fM\n", avgPrice*5) // Assuming 5 drivers per team
}

// PrintPriceBreakdown prints detailed price component breakdown for a driver
func (m *F1QuantumPricingModel) PrintPriceBreakdown(driverPrice F1DriverPrice) {
	fmt.Printf("\n=== PRICE BREAKDOWN FOR %s ===\n", driverPrice.Driver.BasicData.Name)
	fmt.Println(strings.Repeat("-", 50))

	components := []string{
		"Base (Team Strength)",
		"Team Tier Premium",
		"Performance Premium",
		"Championship Premium",
		"Rookie Premium",
		"Trend Adjustment",
		"Popularity Premium",
		"Consistency Value",
		"Ability Premium",
		"Backmarker Discount",
		"Pay Driver Adjustment",
		"Skill Premium",
		"Recent Form Premium",
		"Upgrade Adjustment",
		"Team Leader Premium",
		"Team Change Adjustment",
		"Season Phase Adjustment",
		"Raw Price",
		"Final Price",
	}

	for _, comp := range components {
		val, ok := driverPrice.ComponentBreakdown[comp]
		if !ok || (val == 0.0 && comp != "Raw Price" && comp != "Final Price") {
			continue
		}
		if comp == "Raw Price" {
			fmt.Println(strings.Repeat("-", 50))
		}
		fmt.Printf("%-25s $%6.2fM\n", comp, val)
	}
}
