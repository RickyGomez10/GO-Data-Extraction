package infrastructure

import (
	"bufio"
	"fmt"
	"log"
	"openLaw-dataextraction2/domain"
	"os"
	"strings"
	"time"
)

type OtherTextFile struct {
}

func (otf OtherTextFile) ReadData(FileNameWithExtension string) []string {
	var fileRows []string
	fileData, err := os.Open(FileNameWithExtension)
	if err != nil {
		log.Fatalf("Error opening File: %v", err)
	}

	scanner := bufio.NewScanner(fileData)

	for scanner.Scan() {
		if !strings.Contains(scanner.Text(), "EOF") {
			fileRow := strings.ReplaceAll(scanner.Text(), "|", ";")
			fileRows = append(fileRows, fileRow)
		}

	}

	return fileRows

}

func (otf OtherTextFile) UnmarshalToCaseDetails(semiColonFileRows []string) []domain.CaseDetails {
	var fileRows []string
	var caseDetails []domain.CaseDetails
	//var caseDetails []domain.CaseDetails
	for index, semiColonFileRow := range semiColonFileRows {
		fileRows = strings.Split(semiColonFileRow, ";")
		fmt.Println(index, "-", fileRows)
		caseDetail := otf.mapFileData(fileRows)
		caseDetails = append(caseDetails, caseDetail)
	}

	semiFinalCaseDetails := otf.mergeDuplicateCases(caseDetails)
	finalCaseDetails := otf.removeDuplicateParties(semiFinalCaseDetails)
	//json, _ := json.MarshalIndent(caseDetails, "", " ")
	return finalCaseDetails
}

func (OtherTextFile) parseFileDate(date string) time.Time {
	layout := "2006-01-02"
	parsedDate, err := time.Parse(layout, date)
	if err != nil {
		log.Print(err)
		return time.Now()
	}
	return parsedDate
}

func (otf OtherTextFile) mapFileData(fileColumn []string) domain.CaseDetails {
	CaseNumber := fileColumn[5]
	Date := fileColumn[6]
	CaseType := fileColumn[8]
	Name1 := fileColumn[9]
	Name2 := fileColumn[20]
	caseDate := otf.parseFileDate(Date)
	Name1Split := strings.Split(Name1, ",")
	Name2Split := strings.Split(Name2, ",")

	Parties := otf.mapParties(Name1Split, Name2Split)

	issueDateDetails := domain.IssueDateDetails{
		Raw:   Date,
		Day:   caseDate.Day(),
		Month: int(caseDate.Month()),
		Year:  caseDate.Year(),
	}

	caseDetail := domain.CaseDetails{
		CaseNumber:       CaseNumber,
		CaseType:         CaseType,
		IssueDateDetails: issueDateDetails,
		Parties:          Parties,
	}
	return caseDetail
}

func (OtherTextFile) mapParties(name1 []string, name2 []string) []domain.Party {
	var party1 domain.Party
	var party2 domain.Party
	if len(name1) > 1 {
		party1 = domain.Party{
			Name:      name1[0] + ", " + name1[1],
			LastName:  name1[0],
			FirstName: name1[1],
		}
	}

	if len(name2) > 1 {
		party2 = domain.Party{
			Name:      name2[0] + ", " + name2[1],
			LastName:  name2[0],
			FirstName: name2[1],
		}
	}
	return []domain.Party{party1, party2}
}

func (OtherTextFile) mergeDuplicateCases(cases []domain.CaseDetails) []domain.CaseDetails {
	combinedStructs := make(map[string]domain.CaseDetails)

	for i, _ := range cases {
		existingItem, found := combinedStructs[cases[i].CaseNumber]
		if found {
			existingItem.Parties = append(existingItem.Parties, cases[i].Parties[0])
			combinedStructs[cases[i].CaseNumber] = existingItem
		} else {
			combinedStructs[cases[i].CaseNumber] = cases[i]
		}
	}

	combinedSlice := make([]domain.CaseDetails, 0, len(combinedStructs))
	for _, item := range combinedStructs {
		combinedSlice = append(combinedSlice, item)
	}

	return combinedSlice
}

func (OtherTextFile) removeDuplicateParties(caseDetails []domain.CaseDetails) []domain.CaseDetails {
	for i := range caseDetails {
		uniqueParties := make([]domain.Party, 0)
		seen := make(map[string]bool)

		for _, party := range caseDetails[i].Parties {
			if _, exists := seen[party.Name]; !exists {
				seen[party.Name] = true
				uniqueParties = append(uniqueParties, party)
			}
		}

		caseDetails[i].Parties = uniqueParties
	}

	return caseDetails
}
