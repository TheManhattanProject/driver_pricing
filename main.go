package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	pricingservice "podiumpe.com/driver_pricing/pricing_service"
)

func main() {
	fmt.Println("\n=== Sport ===")
	fmt.Println("1. Formula 1")
	fmt.Println("2. MotoGP")
	fmt.Println("3. Formula E")
	fmt.Println("4. Exit")
	fmt.Print("Enter your choice (1-4): ")

	choice := GetUserChoice()
	if choice == 4 {
		fmt.Println("Exiting...")
		return
	}

	if choice < 1 || choice > 3 {
		fmt.Println("Invalid choice. Please enter a number between 1 and 3.")
		return
	}

	if choice == 1 {
		fmt.Println("Formula 1 selected")
	} else if choice == 2 {
		fmt.Println("MotoGP selected")
	} else if choice == 3 {
		fmt.Println("Formula E selected")
	}

	driverDataPath := GetInput("Driver Data Json file path: ")

	if choice == 1 {
		fmt.Println("\n=== Versions ===")
		fmt.Println("1. Version 1")
		fmt.Println("2. Version 2")

		versionNo := GetUserChoice()

		if versionNo == 1 {
			drivers, err := readFormula1DriversFromJSON(driverDataPath)
			if err != nil {
				fmt.Println("Error reading Formula 1 drivers from JSON:", err)
				return
			}

			totalNumberOfRacesStr := GetInput("Total Number of Races in a Season: ")
			totalNumberOfRaces, err := strconv.Atoi(totalNumberOfRacesStr)
			if err != nil {
				fmt.Println("Error reading Total Number of Races:", err)
				return
			}

			lastRoundStr := GetInput("Last Round: ")
			lastRound, err := strconv.Atoi(lastRoundStr)
			if err != nil {
				fmt.Println("Error reading Last Round:", err)
				return
			}

			totalPointsSeasonStr := GetInput("Total Points in a Season: ")
			totalPointsSeason, err := strconv.Atoi(totalPointsSeasonStr)
			if err != nil {
				fmt.Println("Error reading Total Points in a Season:", err)
				return
			}

			pricingModel := pricingservice.NewF1QuantumPricingModel(totalNumberOfRaces, lastRound, totalPointsSeason)

			driverPrices := pricingModel.ProcessAllDrivers(drivers)
			pricingModel.PrintDriverAttributesTable(driverPrices)
			pricingModel.PrintDriverAbilitiesTable(driverPrices)
			pricingModel.PrintDriverPrices(driverPrices)
			return
		}

		if versionNo == 2 {
			drivers, err := readFormula1DriversV2FromJSON(driverDataPath)
			if err != nil {
				fmt.Println("Error reading Formula 1 drivers from JSON:", err)
				return
			}

			pricingModel := &pricingservice.F1QuantumPricingModelV2{}

			driversSet := pricingModel.NewDriverSet(drivers)
			teams := pricingModel.BuildTeamMapFromDrivers(driversSet)
			pricingModel.PopulateDriverStats(driversSet, teams)
			pricingModel.PriceDrivers(driversSet, 50, 2)
			return
		}

	}

}

func readFormula1DriversFromJSON(filePath string) ([]pricingservice.F1BasicDriverData, error) {
	// Read the file
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading JSON file: %v", err)
	}

	// Unmarshal the JSON into driver structs
	var drivers []pricingservice.F1BasicDriverData
	err = json.Unmarshal(jsonData, &drivers)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON data: %v", err)
	}

	fmt.Printf("Successfully read %d drivers from JSON file\n", len(drivers))
	return drivers, nil
}

func readFormula1DriversV2FromJSON(filePath string) ([]pricingservice.F1BasicDriverDataV2, error) {
	// Read the file
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading JSON file: %v", err)
	}

	// Unmarshal the JSON into driver structs
	var drivers []pricingservice.F1BasicDriverDataV2
	err = json.Unmarshal(jsonData, &drivers)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON data: %v", err)
	}

	fmt.Printf("Successfully read %d drivers from JSON file\n", len(drivers))
	return drivers, nil
}

func GetUserChoice() int {
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil {
		return 0
	}
	return choice
}

func GetInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}
