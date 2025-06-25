package pricingservice

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/stat"
)

// ---------- weight constants (positive = good, negative = penalty) ----------
const (
	wPPR  = 0.16
	wWIN  = 0.05
	wPOD  = 0.05
	wPTF  = 0.03
	wDNF  = -0.03
	wSHR  = 0.04
	wDEL  = 0.03
	wCH3Y = 0.03

	wREC  = 0.11
	wGAIN = 0.07
	wVOL  = -0.02
	wCLUT = 0.03
	wFAST = 0.03
	wCONS = 0.04 // raw (0-1) or ConsZ

	wTSTR = 0.03
	wMOM  = 0.04
	wCEIL = 0.03
	wREL  = 0.03
	wENG  = 0.02
	wBUD  = 0.02

	wDNA   = 0.04
	wDNAV  = -0.02
	wPOP   = 0.02
	wAGE   = -0.01
	wLEAD  = 0.01
	wCHPCT = 0.02
)

// DriverStyle represents a driver's racing style classification
type F1DriverStyleV2 string

const (
	F1AggressiveV2 F1DriverStyleV2 = "Aggressive"
	F1SmoothV2     F1DriverStyleV2 = "Smooth"
	F1DefensiveV2  F1DriverStyleV2 = "Defensive"
	F1OvertakerV2  F1DriverStyleV2 = "Overtaker"
	F1AllRounderV2 F1DriverStyleV2 = "All-Rounder"
	F1RookieV2     F1DriverStyleV2 = "Rookie"
)

type F1PopularityLevelV2 string

const (
	F1HighPopularityV2   F1PopularityLevelV2 = "High"
	F1MediumPopularityV2 F1PopularityLevelV2 = "Medium"
	F1LowPopularityV2    F1PopularityLevelV2 = "Low"
)

// TEAM DATA STRUCTURES (NEW)
//

// TeamSeasonHistory holds historical data for a team's season
type F1TeamSeasonHistoryV2 struct {
	Year       int     // Season year
	Position   int     // Championship position
	Points     float64 // Total points scored
	Wins       int     // Season wins
	Podiums    int     // Season podiums
	TotalRaces int
}

// TeamData holds directly observable information about a team
type F1TeamDataV2 struct {
	Name                      string                  // Team name
	PowerUnit                 string                  // Engine supplier (e.g., "Mercedes", "Ferrari")
	SeasonPosition            int                     // Current championship position
	SeasonPoints              float64                 // Current season points
	BudgetTier                string                  // "Top", "Upper-Mid", "Lower-Mid", "Backmarker" (publicly known)
	SeasonHistory             []F1TeamSeasonHistoryV2 // Previous seasons data
	Wins                      int                     // Current season wins
	Podiums                   int                     // Current season podiums
	DNFs                      int                     // Current season DNFs
	RecentUpgrades            bool                    // Whether team has introduced upgrades in last ~3 races
	RecentRacePositions       []float64               // Last 5-6 races average finish positions
	RecentQualifyingPositions []float64               // Last 5-6 races average qualifying positions
	TotalRaces                int
	CurrentRace               int
}

//
// BASIC DATA STRUCTURES (USER-PROVIDED)
//

// BasicDriverData contains only the minimal info a user needs to provide
type F1BasicDriverDataV2 struct {
	// Basic information
	Name string
	Team string
	Age  int

	// Career statistics (publicly available)
	ChampionshipWins int
	CareerPodiums    int
	CareerStarts     int

	// Season performance data (publicly available)
	Seasons []F1BasicSeasonStatsV2

	// Driver style classification (one-time classification)
	PrimaryStyle   F1DriverStyleV2
	SecondaryStyle F1DriverStyleV2

	// Track specialties (optional, can be empty)
	Specialties []string // Track types they excel at (e.g. "Wet", "Street")
	Weaknesses  []string // Track types they struggle with

	// Rookie info
	IsRookie bool

	// User-specified popularity (instead of nationality-based calculation)
	MarketPopularity F1PopularityLevelV2

	// Team data (new field)
	TeamData F1TeamDataV2

	IsTeamLeader       bool // Designated #1 driver
	CurrentRaceNumber  int  // Current race of the season
	TotalRacesInSeason int  // Total races in current season

	PreviousTeam         string
	RacesWithCurrentTeam int
}

type F1RaceResultV2 struct {
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
type F1BasicSeasonStatsV2 struct {
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
	RecentRaces    []F1RaceResultV2
}

//
// COMPLETE DATA STRUCTURES (CALCULATED)
//

// CompleteDriver combines user input with calculated attributes
type F1CompleteDriverV2 struct {
	// User-provided data
	BasicData      F1BasicDriverDataV2
	SpecialtiesMap map[string]bool
	WeaknessesMap  map[string]bool
	Abilities      map[string]float64

	RECz, GAINz, VOLz                                       float64 // already had
	ClutchZ, FastLapZ, ConsZ                                float64
	GainRaw, VolRaw                                         float64
	ClutchRaw, FastLapRaw, RecRaw                           float64
	CeilingZ, MomentumZ, ReliabZ, TeamStrengthZ             float64
	TeamStrengthRaw, TeamReliabRaw, MomentumRaw, CeilingRaw float64
	EngineTierRaw, EngineTierZ, BudgetTierRaw, BudgetTierZ  float64
	DNAvarZ                                                 float64
	Rows                                                    int     // classified rows (damping)
	ConsRaw                                                 float64 // CONS (0-1)

	// ------------- 3-YEAR AGGREGATES --------------
	PPR3yRaw, PPR3yZ     float64
	WIN3yRaw, WIN3yZ     float64
	POD3yRaw, POD3yZ     float64
	PTFIN3yRaw, PTFIN3yZ float64
	DNF3yRaw, DNF3yZ     float64
	SHARE3yRaw, SHARE3yZ float64
	DELTA3yRaw, DELTA3yZ float64
	CHAMP3yRaw, CHAMP3yZ float64

	// -------------- DNA LAYER ---------------------
	PerformanceRatio float64 // DNA_Core
	Consistency      float64 // DNA_Var

	ChampPctRaw, ChampPctZ float64

	RawScore           float64
	Strength           float64
	NormalizedStrength float64
	ScaledStrength     float64
	Price              float64 // last published price; set = 0 on first run
}

// TeamSeasonStats holds team performance statistics
type F1TeamSeasonStatsV2 struct {
	Position        int
	Points          float64
	WinCount        int     // Number of race wins
	PodiumCount     int     // Number of podiums
	PointsPercent   float64 // Percentage of total points available
	AvgGridPosition float64 // Average qualifying position
}

// DriverPrice represents a driver with their calculated price
type F1DriverPriceV2 struct {
	Driver             F1CompleteDriverV2
	Price              float64
	ComponentBreakdown map[string]float64
}

//
// DRIVER ATTRIBUTE CALCULATION FUNCTIONS
//

type F1QuantumPricingModelV2 struct{}

// ============================================================
//  UTILITIES
// ============================================================

func clamp(x, lo, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

func meanStd(vals []float64) (mu, sd float64) {
	n := float64(len(vals))
	for _, v := range vals {
		mu += v
	}
	mu /= n
	for _, v := range vals {
		sd += math.Pow(v-mu, 2)
	}
	sd = math.Sqrt(sd / n)
	return
}

func weightedMean(vals, wts []float64) float64 {
	var sum, wSum float64
	for i, v := range vals {
		if !math.IsNaN(v) && wts[i] > 0 {
			sum += v * wts[i]
			wSum += wts[i]
		}
	}
	if wSum == 0 {
		return math.NaN()
	}
	return sum / wSum
}

// ============================================================
// 1. PER‑SEASON RATIO HELPERS  (methods on F1BasicSeasonStatsV2)
// ============================================================

func (s *F1BasicSeasonStatsV2) PPR() float64 {
	if s.Races == 0 {
		return 0
	}
	return s.Points / float64(s.Races)
}
func (s *F1BasicSeasonStatsV2) WinRate() float64 {
	if s.Races == 0 {
		return 0
	}
	return float64(s.Wins) / float64(s.Races)
}
func (s *F1BasicSeasonStatsV2) PodRate() float64 {
	if s.Races == 0 {
		return 0
	}
	return float64(s.Podiums) / float64(s.Races)
}
func (s *F1BasicSeasonStatsV2) PTFRate() float64 {
	if s.Races == 0 {
		return 0
	}
	return float64(s.PointFinishes) / float64(s.Races)
}
func (s *F1BasicSeasonStatsV2) DNFRate() float64 {
	if s.Races == 0 {
		return 0
	}
	return float64(s.DNFs) / float64(s.Races)
}
func (s *F1BasicSeasonStatsV2) TeamShare() float64 {
	if s.TeamPoints == 0 {
		return 0.5
	}
	return s.Points / s.TeamPoints
}
func (s *F1BasicSeasonStatsV2) MateDelta() float64 { return s.Points - s.TeammatePoints }
func (s *F1BasicSeasonStatsV2) ChampPct(grid int) float64 {
	if grid == 0 {
		return 0
	}
	return 1 - float64(s.TeamPosition-1)/float64(grid-1)
}

// ============================================================
// 2. THREE‑YEAR ROLL‑UP + Z‑SCORE
// ============================================================

type seasonAgg struct{ PPR, WIN, POD, PTF, DNF, SHARE, DELTA, CHAMP float64 }

func seasonWeight(i int) float64 {
	switch i {
	case 0:
		return 0.60
	case 1:
		return 0.36
	case 2:
		return 0.216
	}
	return 0
}

func (d *F1CompleteDriverV2) compute3y(grid int) seasonAgg {
	out := seasonAgg{}
	var sumW float64
	seasons := d.BasicData.Seasons
	sort.Slice(seasons, func(i, j int) bool { return seasons[i].Year > seasons[j].Year })
	for idx, s := range seasons {
		if idx > 2 {
			break
		}
		w := seasonWeight(idx)
		sumW += w
		out.PPR += w * s.PPR()
		out.WIN += w * s.WinRate()
		out.POD += w * s.PodRate()
		out.PTF += w * s.PTFRate()
		out.DNF += w * s.DNFRate()
		out.SHARE += w * s.TeamShare()
		out.DELTA += w * s.MateDelta()
		out.CHAMP += w * s.ChampPct(grid)
	}
	if sumW == 0 {
		return out
	}
	out.PPR /= sumW
	out.WIN /= sumW
	out.POD /= sumW
	out.PTF /= sumW
	out.DNF /= sumW
	out.SHARE /= sumW
	out.DELTA /= sumW
	out.CHAMP /= sumW
	return out
}

func (d *F1CompleteDriverV2) store3yRaw(grid int) {
	a := d.compute3y(grid)
	d.PPR3yRaw, d.WIN3yRaw, d.POD3yRaw = a.PPR, a.WIN, a.POD
	d.PTFIN3yRaw, d.DNF3yRaw = a.PTF, a.DNF
	d.SHARE3yRaw, d.DELTA3yRaw, d.CHAMP3yRaw = a.SHARE, a.DELTA, a.CHAMP
}

func zBatch(drvs []*F1CompleteDriverV2, get func(*F1CompleteDriverV2) float64, set func(*F1CompleteDriverV2, float64)) {
	arr := make([]float64, len(drvs))
	for i, d := range drvs {
		arr[i] = get(d)
	}
	mu, sd := meanStd(arr)
	for i, d := range drvs {
		z := 0.0
		if sd > 0 {
			z = (arr[i] - mu) / sd
		}
		set(d, clamp(z, -3, 3))
	}
}

// ============================================================
// 3. LIVE‑WINDOW METRICS  (methods on season)
// ============================================================

func (r *F1RaceResultV2) gain() int { return r.StartPosition - r.FinishPosition }

func (s *F1BasicSeasonStatsV2) window() []F1RaceResultV2 {
	sort.Slice(s.RecentRaces, func(i, j int) bool { return s.RecentRaces[i].RaceNumber > s.RecentRaces[j].RaceNumber })
	out := make([]F1RaceResultV2, 0, 5)
	for _, rr := range s.RecentRaces {
		if rr.Classified {
			out = append(out, rr)
			if len(out) == 5 {
				break
			}
		}
	}
	return out
}

func (s *F1BasicSeasonStatsV2) GainRaw() float64 {
	w := s.window()
	if len(w) == 0 {
		return 0
	}
	sum := 0
	for _, rr := range w {
		sum += rr.gain()
	}
	return float64(sum) / float64(len(w))
}

func (s *F1BasicSeasonStatsV2) VolRaw() float64 {
	w := s.window()
	if len(w) < 2 {
		return 0
	}
	gains := make([]float64, len(w))
	for i, rr := range w {
		gains[i] = float64(rr.gain())
	}
	_, sd := meanStd(gains)
	return sd
}

func (s *F1BasicSeasonStatsV2) ClutchRaw() float64 {
	w := s.window()
	if len(w) == 0 {
		return 0
	}
	top := 0
	for _, rr := range w {
		if rr.FinishPosition <= 5 {
			top++
		}
	}
	return float64(top) / float64(len(w))
}

func (s *F1BasicSeasonStatsV2) FastRaw() float64 {
	w := s.window()
	if len(w) == 0 {
		return 0
	}
	fl := 0
	for _, rr := range w {
		if rr.FastestLap {
			fl++
		}
	}
	return float64(fl) / float64(len(w))
}

func ComputeConsZ(drvs []*F1CompleteDriverV2) {
	vals := make([]float64, len(drvs))
	for i, d := range drvs {
		vals[i] = d.ConsRaw
	} // 0-1 range
	mu, sd := meanStd(vals)
	for i, d := range drvs {
		if sd == 0 {
			d.ConsZ = 0
			continue
		}
		d.ConsZ = clamp((vals[i]-mu)/sd, -3, 3) // no rows damping
	}
}

// 0.35-alpha EWMA of the last ≤5 classified races
func (s *F1BasicSeasonStatsV2) RecRaw() float64 {
	win := s.lastClassified(5)
	if len(win) == 0 {
		return 0
	}

	alpha := 0.35
	ewma := alpha * win[0].PointsScored
	mult := 1.0
	for i := 1; i < len(win); i++ {
		mult *= (1 - alpha)
		ewma += mult * alpha * win[i].PointsScored
	}
	return ewma
}

func (s *F1BasicSeasonStatsV2) Rows() int { return len(s.window()) }
func (s *F1BasicSeasonStatsV2) ConsRaw() float64 {
	if s.Races == 0 {
		return 1
	}
	return 1 - float64(s.DNFs)/float64(s.Races)
}

func (s *F1BasicSeasonStatsV2) lastClassified(max int) []F1RaceResultV2 {
	sort.Slice(s.RecentRaces, func(i, j int) bool {
		return s.RecentRaces[i].RaceNumber > s.RecentRaces[j].RaceNumber
	})
	out := make([]F1RaceResultV2, 0, max)
	for _, rr := range s.RecentRaces {
		if rr.Classified {
			out = append(out, rr)
			if len(out) == max {
				break
			}
		}
	}
	return out
}

// driver‑level attachment
func (d *F1CompleteDriverV2) attachLiveRaw() {
	if len(d.BasicData.Seasons) == 0 {
		return
	}
	sort.Slice(d.BasicData.Seasons, func(i, j int) bool {
		return d.BasicData.Seasons[i].Year > d.BasicData.Seasons[j].Year
	})

	latest := &d.BasicData.Seasons[len(d.BasicData.Seasons)-1]

	d.GainRaw, d.VolRaw = latest.GainRaw(), latest.VolRaw()
	d.RecRaw = latest.RecRaw()
	d.ClutchRaw, d.FastLapRaw = latest.ClutchRaw(), latest.FastRaw()
	d.Rows = latest.Rows()
	d.ConsRaw = latest.ConsRaw()
}

func dampedZ(val, mu, sd float64, rows int, clampIt bool) float64 {
	if sd == 0 {
		return 0
	}
	z := (val - mu) / sd
	if clampIt {
		z = clamp(z, -3, 3)
	}
	z *= float64(rows) / 5.0 // early‑season damping
	return z
}

// 1. RECz (clamped)
func ComputeRecZ(drvs []*F1CompleteDriverV2) {
	vals := make([]float64, len(drvs))
	rows := make([]int, len(drvs))
	for i, d := range drvs {
		vals[i], rows[i] = d.RecRaw, d.Rows
	}
	mu, sd := meanStd(vals)
	for i, d := range drvs {
		d.RECz = dampedZ(vals[i], mu, sd, rows[i], true)
	}
}

// 2. GAINz  –‑ **no clamp** per latest spec
func ComputeGainZ(drvs []*F1CompleteDriverV2) {
	vals, rows := make([]float64, len(drvs)), make([]int, len(drvs))
	for i, d := range drvs {
		vals[i], rows[i] = d.GainRaw, d.Rows
	}
	mu, sd := meanStd(vals)
	for i, d := range drvs {
		d.GAINz = dampedZ(vals[i], mu, sd, rows[i], false) // no clamp
	}
}

// 3. VOLz  (clamped)
func ComputeVolZ(drvs []*F1CompleteDriverV2) {
	vals, rows := make([]float64, len(drvs)), make([]int, len(drvs))
	for i, d := range drvs {
		vals[i], rows[i] = d.VolRaw, d.Rows
	}
	mu, sd := meanStd(vals)
	for i, d := range drvs {
		d.VOLz = dampedZ(vals[i], mu, sd, rows[i], true)
	}
}

// 4. ClutchZ  (clamped)
func ComputeClutchZ(drvs []*F1CompleteDriverV2) {
	vals, rows := make([]float64, len(drvs)), make([]int, len(drvs))
	for i, d := range drvs {
		vals[i], rows[i] = d.ClutchRaw, d.Rows
	}
	mu, sd := meanStd(vals)
	for i, d := range drvs {
		d.ClutchZ = dampedZ(vals[i], mu, sd, rows[i], true)
	}
}

// 5. FastLapZ  (clamped)
func ComputeFastLapZ(drvs []*F1CompleteDriverV2) {
	vals, rows := make([]float64, len(drvs)), make([]int, len(drvs))
	for i, d := range drvs {
		vals[i], rows[i] = d.FastLapRaw, d.Rows
	}
	mu, sd := meanStd(vals)
	for i, d := range drvs {
		d.FastLapZ = dampedZ(vals[i], mu, sd, rows[i], true)
	}
}

// ============================================================
// 4. TEAM SNAPSHOT & HISTORY  (methods on F1TeamDataV2)
// ============================================================

// ---------- snapshot ratios ---------------------------------

// Strength = team points ÷ total constructor points
func (t *F1TeamDataV2) Strength(total float64) float64 {
	if total == 0 {
		return 0
	}
	return t.SeasonPoints / total
}

// Reliability = 1 − DNF rate (current season)
func (t *F1TeamDataV2) Reliability() float64 {
	if t.CurrentRace == 0 {
		return 1
	}
	return 1 - float64(t.DNFs)/float64(t.CurrentRace)
}

// categorical maps (latent horsepower & budget muscle)
func mapEngineTier(pu string) float64 {
	switch strings.ToLower(pu) {
	case "mercedes":
		return 1.00
	case "ferrari":
		return 0.90
	case "honda", "rbpt":
		return 0.85
	case "renault", "alpine":
		return 0.80
	default:
		return 0.60
	}
}

func mapBudgetTier(bt string) float64 {
	switch strings.ToLower(bt) {
	case "top":
		return 1.00
	case "upper-mid":
		return 0.75
	case "lower-mid":
		return 0.50
	case "backmarker":
		return 0.25
	default:
		return 0.50
	}
}

// ---------- two‑year history ratios --------------------------

// helper: project current‑season points until mid‑season
func projectSeasonPts(t *F1TeamDataV2) float64 {
	if t.CurrentRace == 0 {
		return 0
	}
	half := t.TotalRaces / 2
	if t.CurrentRace < half && half > 0 {
		return t.SeasonPoints * float64(t.TotalRaces) / float64(t.CurrentRace)
	}
	return t.SeasonPoints
}

// Momentum = weighted pts trend (0.60,0.36,0.216)
func (t *F1TeamDataV2) Momentum() float64 {
	rows := []struct{ pts, w float64 }{{projectSeasonPts(t), 0.60}}
	if len(t.SeasonHistory) > 0 {
		rows = append(rows, struct{ pts, w float64 }{t.SeasonHistory[0].Points, 0.36})
	}
	if len(t.SeasonHistory) > 1 {
		rows = append(rows, struct{ pts, w float64 }{t.SeasonHistory[1].Points, 0.216})
	}
	var sum, wSum float64
	for _, r := range rows {
		sum += r.pts * r.w
		wSum += r.w
	}
	if wSum == 0 {
		return 0
	}
	return sum / wSum
}

// Ceiling = (wins Y0+Y‑1+Y‑2) ÷ (races Y0+Y‑1+Y‑2)
func (t *F1TeamDataV2) Ceiling(gridMean float64) float64 {
	wins := t.Wins
	races := t.CurrentRace
	if len(t.SeasonHistory) > 0 {
		wins += t.SeasonHistory[0].Wins
		races += t.SeasonHistory[0].TotalRaces
	}
	if len(t.SeasonHistory) > 1 {
		wins += t.SeasonHistory[1].Wins
		races += t.SeasonHistory[1].TotalRaces
	}
	if races == 0 {
		return gridMean
	}
	return float64(wins) / float64(races)
}

// GridMeanCeil = weighted mean of ceilings across constructors
func GridMeanCeil(teams []*F1TeamDataV2) float64 {
	var vals, wts []float64
	for _, t := range teams {
		wins := t.Wins
		races := t.CurrentRace
		for i := 0; i < len(t.SeasonHistory) && i < 2; i++ {
			wins += t.SeasonHistory[i].Wins
			races += t.SeasonHistory[i].TotalRaces
		}
		if races == 0 {
			continue
		}
		vals = append(vals, float64(wins)/float64(races))
		wts = append(wts, float64(races))
	}
	gm := weightedMean(vals, wts)
	if math.IsNaN(gm) {
		return 0.05
	}
	return gm
}

// ============================================================
// 5. ABILITY‑VECTOR HELPERS  (DNA_Core & DNA_Var)
// ============================================================

// ---- ability key list (12) --------------------------------
var abilityKeys = []string{
	"WetWeather", "TireManagement", "BrakingStability", "TechnicalCorners",
	"RaceStart", "QualifyingPace", "SetupAdaptability", "OvertakingSkill",
	"RaceConsistency", "ERSManagement", "FuelSaving", "SafetyCarRestart",
}

// ---- slot weights feeding DNA_Core ------------------------
var abilityW = map[string]float64{
	"QualifyingPace": 0.18, "RaceStart": 0.14, "OvertakingSkill": 0.14,
	"RaceConsistency": 0.12, "TireManagement": 0.11, "TechnicalCorners": 0.08,
	"ERSManagement": 0.07, "SetupAdaptability": 0.06, "FuelSaving": 0.04,
	"BrakingStability": 0.03, "WetWeather": 0.02, "SafetyCarRestart": 0.01,
}

// ---- style bump table (primary full, secondary ×0.5) ------
var styleBumps = map[string]map[string]float64{
	"Aggressive":        {"OvertakingSkill": 0.15, "RaceStart": 0.15, "ERSManagement": 0.15},
	"Smooth":            {"TireManagement": 0.15, "FuelSaving": 0.15, "RaceConsistency": 0.15},
	"Quali-Ace":         {"QualifyingPace": 0.20},
	"Tyre-Whisperer":    {"TireManagement": 0.20},
	"Rain-Master":       {"WetWeather": 0.25},
	"Engineer’s Driver": {"SetupAdaptability": 0.15, "TechnicalCorners": 0.15},
	"Defensive":         {"BrakingStability": 0.15, "RaceConsistency": 0.15, "SafetyCarRestart": 0.10},
}

func styleVector(p, s F1DriverStyleV2) map[string]float64 {
	out := map[string]float64{}
	for k, v := range styleBumps[string(p)] {
		out[k] += v
	}
	for k, v := range styleBumps[string(s)] {
		out[k] += v * 0.5
	}
	return out
}

// ---- build grid‑wide mean & σ for raw ability numbers -------
func BuildAbilityMeanStd(drvs []*F1CompleteDriverV2) (map[string]float64, map[string]float64) {
	mu, sd := map[string]float64{}, map[string]float64{}
	for _, key := range abilityKeys {
		mu[key] = 0
		sd[key] = 0
	}
	// mean
	for _, d := range drvs {
		for _, k := range abilityKeys {
			mu[k] += d.Abilities[k]
		}
	}
	for _, k := range abilityKeys {
		mu[k] /= float64(len(drvs))
	}
	// std
	for _, d := range drvs {
		for _, k := range abilityKeys {
			sd[k] += math.Pow(d.Abilities[k]-mu[k], 2)
		}
	}
	for _, k := range abilityKeys {
		sd[k] = math.Sqrt(sd[k] / float64(len(drvs)))
	}
	return mu, sd
}

// ---- driver‑level DNA calculation --------------------------
func (d *F1CompleteDriverV2) setDNA(mu, sd map[string]float64) {
	styleVec := styleVector(d.BasicData.PrimaryStyle, d.BasicData.SecondaryStyle)
	V := make([]float64, len(abilityKeys))
	var core float64
	for i, key := range abilityKeys {
		// slot z (±2 clamp)
		z := 0.0
		if sd[key] > 0 {
			z = clamp((d.Abilities[key]-mu[key])/sd[key], -2, 2)
		}
		// style bump
		z += styleVec[key]
		// tag bumps
		if d.SpecialtiesMap != nil && d.SpecialtiesMap[key] {
			z += 0.10
		}
		if d.WeaknessesMap != nil && d.WeaknessesMap[key] {
			z -= 0.10
		}
		// upgrade synergy
		if d.BasicData.TeamData.RecentUpgrades && (key == "SetupAdaptability" || key == "TechnicalCorners" || key == "RaceConsistency") {
			z += 0.05
		}
		z = clamp(z, -2, 2)
		V[i] = z
		core += abilityW[key] * z
	}
	// DNA_Core -> PerformanceRatio
	d.PerformanceRatio = core

	// DNA_Var -> Consistency (σ of V slots)
	var meanV float64
	for _, v := range V {
		meanV += v
	}
	meanV /= float64(len(V))
	var varSum float64
	for _, v := range V {
		varSum += math.Pow(v-meanV, 2)
	}
	d.Consistency = math.Sqrt(varSum / float64(len(V)))
}

// AttachChampPctRaw fills ChampPctRaw for every driver and returns leaderPts.
func AttachChampPctRaw(drvs []*F1CompleteDriverV2) float64 {
	var leaderPts float64
	// first pass: find leader points
	for _, d := range drvs {
		if len(d.BasicData.Seasons) == 0 {
			continue
		}
		pts := d.BasicData.Seasons[len(d.BasicData.Seasons)-1].Points
		if pts > leaderPts {
			leaderPts = pts
		}
	}
	if leaderPts == 0 { // preseason: everyone raw 0
		for _, d := range drvs {
			d.ChampPctRaw = 0
		}
		return 0
	}
	// second pass: ratio for each driver
	for _, d := range drvs {
		pts := d.BasicData.Seasons[len(d.BasicData.Seasons)-1].Points
		d.ChampPctRaw = pts / leaderPts
	}
	return leaderPts
}

// ComputeChampPctZ – converts ChampPctRaw to Z‑score (±3 clamp, no damping)
func ComputeChampPctZ(drvs []*F1CompleteDriverV2) {
	vals := make([]float64, len(drvs))
	for i, d := range drvs {
		vals[i] = d.ChampPctRaw
	}
	mu, sd := meanStd(vals)
	for i, d := range drvs {
		z := 0.0
		if sd > 0 {
			z = (vals[i] - mu) / sd
		}
		d.ChampPctZ = clamp(z, -3, 3)
	}
}

// BuildCompleteDriver creates a single complete‑driver shell from basic data.
func (m *F1QuantumPricingModelV2) NewDriver(b F1BasicDriverDataV2) *F1CompleteDriverV2 {
	// 1. specialties / weaknesses to maps for O(1) lookup
	spMap := map[string]bool{}
	for _, tag := range b.Specialties {
		spMap[tag] = true
	}
	wkMap := map[string]bool{}
	for _, tag := range b.Weaknesses {
		wkMap[tag] = true
	}

	// 2. make sure Abilities has ALL 12 keys (fill 0 if missing)
	abil := make(map[string]float64, len(abilityKeys))
	for _, k := range abilityKeys {
		abil[k] = 0
	}
	// if external data provided, copy over
	if bAbil, ok := any(b).(interface{ GetAbilities() map[string]float64 }); ok {
		for k, v := range bAbil.GetAbilities() {
			abil[k] = v
		}
	}

	return &F1CompleteDriverV2{
		BasicData:      b,
		SpecialtiesMap: spMap,
		WeaknessesMap:  wkMap,
		Abilities:      abil,
	}
}

// BuildCompleteDriverSlice converts a slice in one go.
func (m *F1QuantumPricingModelV2) NewDriverSet(basics []F1BasicDriverDataV2) []*F1CompleteDriverV2 {
	out := make([]*F1CompleteDriverV2, len(basics))
	for i, b := range basics {
		out[i] = m.NewDriver(b)
	}
	return out
}

// B) if you only have drivers -----------------------------------------------------------
func (model *F1QuantumPricingModelV2) BuildTeamMapFromDrivers(drvs []*F1CompleteDriverV2) map[string]*F1TeamDataV2 {
	m := make(map[string]*F1TeamDataV2)
	for _, d := range drvs {
		key := strings.ToLower(d.BasicData.TeamData.Name)
		// first occurrence wins so every driver for the same team shares the same ptr
		if _, exists := m[key]; !exists {
			m[key] = &d.BasicData.TeamData
		}
	}
	return m
}

func (model *F1QuantumPricingModelV2) PopulateDriverStats(drvs []*F1CompleteDriverV2, teams map[string]*F1TeamDataV2) {
	// ---------- 1. LIVE WINDOW RAW  -------------------------
	for _, d := range drvs {
		d.attachLiveRaw()
	}
	ComputeRecZ(drvs)
	ComputeGainZ(drvs)
	ComputeVolZ(drvs)
	ComputeClutchZ(drvs)
	ComputeFastLapZ(drvs)
	ComputeConsZ(drvs)

	// ---------- 2. 3‑YEAR SEASON ROLL‑UPS -------------------
	gridSize := len(teams)
	for _, d := range drvs {
		d.store3yRaw(gridSize)
	}
	// zBatch helper defined earlier – run for each metric
	zBatch(drvs, func(x *F1CompleteDriverV2) float64 { return x.PPR3yRaw }, func(x *F1CompleteDriverV2, z float64) { x.PPR3yZ = z })
	zBatch(drvs, func(x *F1CompleteDriverV2) float64 { return x.WIN3yRaw }, func(x *F1CompleteDriverV2, z float64) { x.WIN3yZ = z })
	zBatch(drvs, func(x *F1CompleteDriverV2) float64 { return x.POD3yRaw }, func(x *F1CompleteDriverV2, z float64) { x.POD3yZ = z })
	zBatch(drvs, func(x *F1CompleteDriverV2) float64 { return x.PTFIN3yRaw }, func(x *F1CompleteDriverV2, z float64) { x.PTFIN3yZ = z })
	zBatch(drvs, func(x *F1CompleteDriverV2) float64 { return x.DNF3yRaw }, func(x *F1CompleteDriverV2, z float64) { x.DNF3yZ = z })
	zBatch(drvs, func(x *F1CompleteDriverV2) float64 { return x.SHARE3yRaw }, func(x *F1CompleteDriverV2, z float64) { x.SHARE3yZ = z })
	zBatch(drvs, func(x *F1CompleteDriverV2) float64 { return x.DELTA3yRaw }, func(x *F1CompleteDriverV2, z float64) { x.DELTA3yZ = z })
	zBatch(drvs, func(x *F1CompleteDriverV2) float64 { return x.CHAMP3yRaw }, func(x *F1CompleteDriverV2, z float64) { x.CHAMP3yZ = z })

	// ---------- 3. TEAM SNAPSHOT + HISTORY ------------------
	// First compute grid totals & mean ceiling
	var totalPts float64
	for _, t := range teams {
		totalPts += t.SeasonPoints
	}
	gridMean := GridMeanCeil(values(teams))

	// calculate raw snapshot/history and prepare slices for Z
	var st, mom, ceil, rel []float64
	for _, d := range drvs {
		team := teams[strings.ToLower(d.BasicData.TeamData.Name)]
		d.TeamStrengthRaw = team.Strength(totalPts)
		d.TeamReliabRaw = team.Reliability()
		d.MomentumRaw = team.Momentum()
		d.CeilingRaw = team.Ceiling(gridMean)
		st = append(st, d.TeamStrengthRaw)
		rel = append(rel, d.TeamReliabRaw)
		mom = append(mom, d.MomentumRaw)
		ceil = append(ceil, d.CeilingRaw)
	}
	// Z-score these four slices
	mu, sd := meanStd(st)
	for i, d := range drvs {
		d.TeamStrengthZ = clamp((st[i]-mu)/sd, -3, 3)
	}
	mu, sd = meanStd(rel)
	for i, d := range drvs {
		d.ReliabZ = clamp((rel[i]-mu)/sd, -3, 3)
	}
	mu, sd = meanStd(mom)
	for i, d := range drvs {
		d.MomentumZ = clamp((mom[i]-mu)/sd, -3, 3)
	}
	mu, sd = meanStd(ceil)
	for i, d := range drvs {
		d.CeilingZ = clamp((ceil[i]-mu)/sd, -3, 3)
	}

	// engine & budget tiers (categorical) – store raw Z vs mean
	var eng, bud []float64
	for _, d := range drvs {
		team := teams[strings.ToLower(d.BasicData.TeamData.Name)]
		d.EngineTierRaw = mapEngineTier(strings.ToLower(team.PowerUnit))
		d.BudgetTierRaw = mapBudgetTier(strings.ToLower(team.BudgetTier))
		eng = append(eng, d.EngineTierRaw)
		bud = append(bud, d.BudgetTierRaw)
	}
	mu, sd = meanStd(eng)
	for i, d := range drvs {
		d.EngineTierZ = (eng[i] - mu) / sd
	}
	mu, sd = meanStd(bud)
	for i, d := range drvs {
		d.BudgetTierZ = (bud[i] - mu) / sd
	}

	// ---------- 4. DNA --------------------------------------
	muA, sdA := BuildAbilityMeanStd(drvs)
	for _, d := range drvs {
		d.setDNA(muA, sdA)
	}
	// Z of DNA_Var (Consistency field)
	dnaVarSlice := make([]float64, len(drvs))
	for i, d := range drvs {
		dnaVarSlice[i] = d.Consistency
	}
	mu, sd = meanStd(dnaVarSlice)
	for i, d := range drvs {
		d.DNAvarZ = (dnaVarSlice[i] - mu) / sd
	}

	// ---------- 5. Continuous champ % -----------------------
	AttachChampPctRaw(drvs)
	ComputeChampPctZ(drvs)
}

// helper to convert map → slice
func values(m map[string]*F1TeamDataV2) []*F1TeamDataV2 {
	out := make([]*F1TeamDataV2, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}

func rawScore(d *F1CompleteDriverV2) float64 {
	return 0.15 + // constant bias
		wPPR*d.PPR3yZ + wWIN*d.WIN3yZ + wPOD*d.POD3yZ + wPTF*d.PTFIN3yZ +
		wDNF*d.DNF3yZ + wSHR*d.SHARE3yZ + wDEL*d.DELTA3yZ + wCH3Y*d.CHAMP3yZ +

		wREC*d.RECz + wGAIN*d.GAINz + wVOL*d.VOLz + wCLUT*d.ClutchZ +
		wFAST*d.FastLapZ + wCONS*(d.ConsRaw) + // raw 0-1

		wTSTR*d.TeamStrengthZ + wMOM*d.MomentumZ + wCEIL*d.CeilingZ +
		wREL*d.ReliabZ + wENG*d.EngineTierZ + wBUD*d.BudgetTierZ +

		wDNA*d.PerformanceRatio + wDNAV*d.DNAvarZ +
		wPOP*0 + wAGE*0 + wLEAD*0 + wCHPCT*d.ChampPctZ // pop/age/lead not in struct yet
}

func logistic(x float64) float64 { return 1 / (1 + math.Exp(-x)) }

// Economic knobs are baked-in constants.
func solveBand(drvs []*F1CompleteDriverV2, cap float64, roster int) (pMin, pMax float64) {

	const (
		tau  = 0.90 // target spend as % of cap
		mMin = 0.40 // floor as % of avg slot
		mMax = 1.35 // ceiling as % of avg slot
	)

	slot := cap / float64(roster) // 25 for 50/2
	target := tau * cap           // 45 for tau=0.9

	// --- Strength stats ----------------------------------------------------
	var sMin, sMax, sumS float64
	sMin = math.MaxFloat64
	for _, d := range drvs {
		s := d.Strength
		sumS += s
		if s < sMin {
			sMin = s
		}
		if s > sMax {
			sMax = s
		}
	}
	n := float64(len(drvs))

	// --- Start with theoretical floor / ceiling ----------------------------
	pMin = mMin * slot // 10
	pMax = mMax * slot // 33.8

	// --- Can the mean-spend be met with this band? -------------------------
	// Σ price = n·pMin + (pMax-pMin) Σ S_norm
	// S_norm = (S-Smin)/(Smax-Smin); if Smax==Smin → all S_norm = 0
	var sumNorm float64
	if sMax > sMin+1e-9 {
		sumNorm = (sumS - n*sMin) / (sMax - sMin)
	} else {
		sumNorm = 0 // perfectly flat grid
	}
	spend := n*pMin + (pMax-pMin)*sumNorm

	// --- If spend too HIGH → push pMin UP  (keep premium feel) -------------
	if spend < target {
		// a  + b·Smin = pMin  ;  Σ price = target
		// => pMin' = (target - b·ΣS) / n     with  b fixed (pMax-pMin)/(Smax-Smin)
		// In flat grid b = 0, so pMin' = target/n
		if sMax > sMin+1e-9 {
			b := (pMax - pMin) / (sMax - sMin)
			pMin = (target - b*sumS) / n
			if pMin < mMin*slot {
				pMin = mMin * slot
			}
			// keep same slope -> recompute pMax
			pMax = pMin + b*(sMax-sMin)
		} else { // flat grid, spread by design
			pMin = target / n // everyone costs the same
			if pMin < mMin*slot {
				pMin = mMin * slot
			}
			pMax = mMax * slot // still show a ceiling for UI tiers
		}
	}

	return
}

// price charm: nearest 0.1 M then replace .0 ⇒ .9 and .6 ⇒ .4
func charm(x float64) float64 {
	return math.Ceil(x*2) / 2
}

func (model *F1QuantumPricingModelV2) PriceDrivers(drvs []*F1CompleteDriverV2, cap float64, roster int) {
	// 1) RAW + Strength
	score := make([]float64, 0, len(drvs))
	for _, d := range drvs {
		d.RawScore = rawScore(d)
		d.Strength = logistic(d.RawScore)
		score = append(score, d.Strength)
	}

	mean := stat.Mean(score, nil)
	stdDev := stat.StdDev(score, nil)

	min := floats.Min(score)
	max := floats.Max(score)

	for _, d := range drvs {
		d.NormalizedStrength = (d.Strength - mean) / stdDev
		d.ScaledStrength = (d.Strength - min) / (max - min)
	}

	// 2) dynamic band
	pMin, pMax := solveBand(drvs, cap, roster)

	// 3) base & elastic price
	for _, d := range drvs {
		base := pMin + (pMax-pMin)*d.ScaledStrength // linear interpolation
		base = charm(base)                          // psychological rounding

		// elasticity – steeper if unreliable or volatile
		elast := 0.45 + 0.25*(1-d.ConsRaw) +
			0.10*math.Max(0, d.DNAvarZ) +
			0.10*math.Max(0, d.VOLz)

		if d.Price == 0 { // first call ever
			d.Price = base
		} else {
			d.Price += elast * (base - d.Price) // move toward base
		}
		fmt.Println(d.BasicData.Name)
		fmt.Println(d.BasicData.TeamData.Name)
		fmt.Println(d.Price)
	}
}
