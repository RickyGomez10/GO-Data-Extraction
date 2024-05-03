package infrastructure

import (
	"bufio"
	"log"
	"openLaw-dataextraction2/domain"
	"os"
	"strings"
	"time"
)

type TextFileCivil struct {
}

func (tfc TextFileCivil) ReadData(FileNameWithExtension string) []string {
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

func (tfc TextFileCivil) UnmarshalToCaseDetails(semiColonFileRows []string) []domain.CaseDetails {
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
	SemifinalCaseDetails := tfc.mergeDuplicateCases(caseDetails)
	finalCaseDetails := tfc.removeDuplicateParties(SemifinalCaseDetails)
	return finalCaseDetails
}

func (TextFileCivil) parseFileDate(date string) time.Time {

	if date == "" {
		return time.Time{}
	}

	layout := "01/02/2006"
	parsedDate, err := time.Parse(layout, date)

	if err != nil {
		log.Print(err)
		return time.Time{}
	}
	return parsedDate
}

func (tfc TextFileCivil) mapFileData(fileColumn []string) domain.CaseDetails {

	Status := fileColumn[7]
	StatusLowerCase := strings.ToLower(Status)

	if strings.Contains(StatusLowerCase, "closed") {
		return domain.CaseDetails{}
	}
	var Parties []domain.Party
	CaseNumber := fileColumn[3]
	IssueDate := fileColumn[4]

	Address := fileColumn[9]
	Apartment := fileColumn[10]
	City := fileColumn[len(fileColumn)-6]
	State := fileColumn[len(fileColumn)-5]
	ZipCode := fileColumn[len(fileColumn)-4]
	CaseType := fileColumn[6]
	PartyName := fileColumn[8]
	Kind := fileColumn[len(fileColumn)-3]

	Party := domain.Party{
		Name:        PartyName,
		Kind:        Kind,
		Address:     Address + " " + Apartment + " " + City + " " + State + ", " + ZipCode,
		AddressOnly: Address,
		Building:    "",
		CompanyName: "",
		City:        City,
		State:       State,
		Zip:         ZipCode,
		Apartment:   Apartment,
	}

	Parties = append(Parties, Party)

	issueDateDetails := domain.IssueDateDetails{
		Raw:   IssueDate,
		Day:   tfc.parseFileDate(IssueDate).Day(),
		Month: int(tfc.parseFileDate(IssueDate).Month()),
		Year:  tfc.parseFileDate(IssueDate).Year(),
	}

	caseDetail := domain.CaseDetails{
		Status:           Status,
		CaseStyle:        "",
		Kind:             "",
		CaseNumber:       CaseNumber,
		CaseType:         CaseType,
		IssueDateDetails: issueDateDetails,
		Parties:          Parties,
	}
	return caseDetail
}

func (TextFileCivil) mergeDuplicateCases(cases []domain.CaseDetails) []domain.CaseDetails {
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

func (TextFileCivil) removeDuplicateParties(caseDetails []domain.CaseDetails) []domain.CaseDetails {
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
