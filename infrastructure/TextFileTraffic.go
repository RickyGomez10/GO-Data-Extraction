package infrastructure

import (
	"bufio"
	"log"
	"openLaw-dataextraction2/domain"
	"openLaw-dataextraction2/utils"
	"os"
	"strings"
	"time"
)

type TextFileTraffic struct {
}

func (tft TextFileTraffic) ReadData(FileNameWithExtension string) []string {
	var fileRows []string
	fileData, err := os.Open(FileNameWithExtension)
	if err != nil {
		log.Fatalf("Error opening File: %v", err)
	}

	scanner := bufio.NewScanner(fileData)

	for scanner.Scan() {
		if !strings.Contains(scanner.Text(), "EOF") {
			fileRow := strings.ReplaceAll(scanner.Text(), "\",\"", ";")
			fileRows = append(fileRows, fileRow)
		}
	}

	return fileRows

}

func (tft TextFileTraffic) UnmarshalToCaseDetails(semiColonFileRows []string) []domain.CaseDetails {
	var fileRows []string
	var caseDetails []domain.CaseDetails

	for _, semiColonFileRow := range semiColonFileRows {
		unescaped := strings.ReplaceAll(semiColonFileRow, "\"", "")
		fileRows = strings.Split(unescaped, ";")
		caseDetail := tft.mapFileData(fileRows)
		if caseDetail.CaseNumber != "" {
			caseDetails = append(caseDetails, caseDetail)
		}

	}

	tft.setCaseStyles(caseDetails)
	return caseDetails
}

func (TextFileTraffic) parseFileDate(date string) time.Time {

	if date == "" {
		return time.Time{}
	}

	layout := "1/2/2006"
	parsedDate, err := time.Parse(layout, date)

	if err != nil {
		log.Print(err)
		return time.Time{}
	}
	return parsedDate
}

func (tft TextFileTraffic) mapFileData(fileColumn []string) domain.CaseDetails {
	CaseNumber := fileColumn[0]
	FirstName := fileColumn[2]
	MiddleName := fileColumn[3]
	LastName := fileColumn[1]
	ZipCode := fileColumn[9]
	State := fileColumn[8]
	City := fileColumn[7]
	Address := fileColumn[6]
	Status := fileColumn[10]
	ChargeDescription := fileColumn[14]
	issueDate := fileColumn[12]
	parsedDate := tft.parseFileDate(issueDate)

	issueDetail := domain.IssueDateDetails{
		Raw:   issueDate,
		Day:   parsedDate.Day(),
		Month: int(parsedDate.Month()),
		Year:  parsedDate.Year(),
	}

	parties := []domain.Party{{
		Name: "State Of " + utils.GetStateName(State),
		Kind: "Plaintiff",
	}, {

		Name:        LastName + ", " + FirstName + " " + MiddleName,
		Kind:        "Defendant",
		LastName:    LastName,
		FirstName:   FirstName,
		MiddleName:  MiddleName,
		Address:     Address + " " + City + ", " + State + " " + ZipCode,
		AddressOnly: Address,
		City:        City,
		State:       State,
		Zip:         ZipCode,
	},
	}

	caseDetail := domain.CaseDetails{
		Status:            Status,
		CaseNumber:        CaseNumber,
		ChargeDescription: ChargeDescription,
		CaseType:          "",
		IssueDateDetails:  issueDetail,
		Parties:           parties,
	}
	return caseDetail
}

func (TextFileTraffic) setCaseStyles(cases []domain.CaseDetails) {
	for index, lawCase := range cases {
		plaintiffs := ""
		defendants := ""
		for _, party := range lawCase.Parties {

			if strings.ToLower(party.Kind) == "plaintiff" {
				if strings.Contains(plaintiffs, party.Name) {
					continue
				}
				plaintiffs += party.Name + " and "
			}

			if strings.ToLower(party.Kind) == "defendant" {
				if strings.Contains(defendants, party.Name) {
					continue
				}
				defendants += party.Name + " and "
			}

		}

		if plaintiffs == "" || defendants == "" {
			continue
		}

		caseStyle := plaintiffs[:len(plaintiffs)-5] + " vs. " + defendants[:len(defendants)-5]
		cases[index].CaseStyle = caseStyle
	}

}

func (TextFileTraffic) mergeDuplicateCases(cases []domain.CaseDetails) []domain.CaseDetails {
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

func (TextFileTraffic) removeDuplicateParties(caseDetails []domain.CaseDetails) []domain.CaseDetails {
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
