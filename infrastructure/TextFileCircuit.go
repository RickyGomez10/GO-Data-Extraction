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

type TextFileCircuit struct {
}

func (tfc TextFileCircuit) ReadData(FileNameWithExtension string) []string {
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

func (tfc TextFileCircuit) UnmarshalToCaseDetails(semiColonFileRows []string) []domain.CaseDetails {
	var fileRows []string
	var caseDetails []domain.CaseDetails
	//var caseDetails []domain.CaseDetails
	for _, semiColonFileRow := range semiColonFileRows {
		unescaped := strings.ReplaceAll(semiColonFileRow, "\"", "")
		fileRows = strings.Split(unescaped, ";")
		caseDetail := tfc.mapFileData(fileRows)
		if caseDetail.CaseNumber != "" {
			caseDetails = append(caseDetails, caseDetail)
		}

	}
	semiFinalCaseDetails := tfc.mergeDuplicateCases(caseDetails)
	finalCaseDetails := tfc.removeDuplicateParties(semiFinalCaseDetails)
	return finalCaseDetails
}

func (TextFileCircuit) parseFileDate(date string) time.Time {
	layout := "01/02/2006"

	if date == "" {
		return time.Time{}
	}
	parsedDate, err := time.Parse(layout, date)
	if err != nil {
		log.Print(err)
		return time.Time{}
	}
	return parsedDate
}

func (tfc TextFileCircuit) mapFileData(fileColumn []string) domain.CaseDetails {

	Status := fileColumn[1]
	StatusLowerCase := strings.ToLower(Status)

	if strings.Contains(StatusLowerCase, "closed") {
		return domain.CaseDetails{}
	}

	CaseNumber := fileColumn[0]
	IssueDate := fileColumn[5]
	FirstName := fileColumn[9]
	MiddleName := fileColumn[8]
	LastName := fileColumn[7]
	State := fileColumn[13]
	ZipCode := fileColumn[14]
	ChargeDescription := fileColumn[22]
	City := fileColumn[12]
	Address := fileColumn[10]
	Apartment := fileColumn[11]

	CaseStyle := "State of " + utils.GetStateName(State) + " vs. " + FirstName + " " + LastName

	parties := []domain.Party{
		{
			Name: "State of " + utils.GetStateName(State),
			Kind: "Plaintiff",
		},
		{
			Name:        LastName + ", " + FirstName + " " + MiddleName,
			LastName:    LastName,
			Kind:        "Defendant",
			FirstName:   FirstName,
			MiddleName:  MiddleName,
			Address:     Address + " " + Apartment + " " + City + ", " + State + " " + ZipCode,
			AddressOnly: Address,
			City:        City,
			State:       State,
			Zip:         ZipCode,
			Apartment:   Apartment,
		},
	}

	issueDateDetails := domain.IssueDateDetails{
		Raw:   IssueDate,
		Day:   tfc.parseFileDate(IssueDate).Day(),
		Month: int(tfc.parseFileDate(IssueDate).Month()),
		Year:  tfc.parseFileDate(IssueDate).Year(),
	}

	caseDetail := domain.CaseDetails{
		Status:            Status,
		CaseStyle:         CaseStyle,
		Kind:              "",
		CaseNumber:        CaseNumber, //
		ChargeDescription: ChargeDescription,
		CaseType:          "",
		IssueDateDetails:  issueDateDetails,
		Parties:           parties,
	}
	return caseDetail
}

func (TextFileCircuit) mapParties(name1 []string, name2 []string) []domain.Party {
	var party1 domain.Party
	var party2 domain.Party
	if len(name1) > 1 {
		party1 = domain.Party{
			Name:      name1[1] + name1[0],
			LastName:  name1[0],
			FirstName: name1[1],
		}
	}

	if len(name2) > 1 {
		party2 = domain.Party{
			Name:      name2[1] + name2[0],
			LastName:  name2[0],
			FirstName: name2[1],
		}
	}
	return []domain.Party{party1, party2}
}

func (TextFileCircuit) mergeDuplicateCases(cases []domain.CaseDetails) []domain.CaseDetails {
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

func (TextFileCircuit) removeDuplicateParties(caseDetails []domain.CaseDetails) []domain.CaseDetails {
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
