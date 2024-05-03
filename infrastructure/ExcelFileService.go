package infrastructure

import (
	"github.com/tealeg/xlsx"
	"log"
	"openLaw-dataextraction2/domain"
	"strings"
	"time"
)

type ExcelFileService struct {
}

func (exs ExcelFileService) ReadData(FileNameWithExtension string) []string {

	xlFile, err := xlsx.OpenFile(FileNameWithExtension)
	if err != nil {
		log.Fatalf("Error opening Excel file: %v", err)
	}

	var data []string

	for _, sheet := range xlFile.Sheets {

		for rowIndex, row := range sheet.Rows {
			var rowData []string
			for i, cell := range row.Cells {
				text := cell.String()

				// Convert Excel serial date value to Go time.Time
				if i == 2 && rowIndex != 0 { // Assuming the date is in the third cell (index 2)
					serialDateValue, err := cell.Int()
					if err != nil {
						log.Fatalf("Error opening Excel file: %v", err)
					}

					// Calculate the number of days since the Excel epoch (December 30, 1899)
					epoch := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
					excelTime := epoch.Add(time.Duration(serialDateValue) * 24 * time.Hour)

					// Format the Excel time as a "mm/dd/yyyy" formatted string
					text = excelTime.Format("01/02/2006")
				}

				rowData = append(rowData, text)
			}
			combinedRowData := strings.Join(rowData, ";")

			data = append(data, combinedRowData)
		}
	}

	return data
}

func (exs ExcelFileService) UnmarshalToCaseDetails(semiColonFileRows []string) []domain.CaseDetails {
	var fileCells []string
	var caseDetails []domain.CaseDetails

	for _, fileRow := range semiColonFileRows {
		fileCells = strings.Split(fileRow, ";")
		caseDetail := exs.mapFileData(fileCells)
		caseDetails = append(caseDetails, caseDetail)

	}
	//Se debe hacer merge?
	semifinalCaseDetails := exs.mergeDuplicateCases(caseDetails)
	finalCaseDetails := exs.removeDuplicateParties(semifinalCaseDetails)
	return finalCaseDetails
}

func (ExcelFileService) parseFileDate(date string) time.Time {
	layout := "01/02/2006"
	parsedDate, err := time.Parse(layout, date)
	if err != nil {
		log.Print(err)
		return time.Now()
	}
	return parsedDate
}

func (exs ExcelFileService) mapFileData(fileCells []string) domain.CaseDetails {
	CaseNumber := fileCells[0]
	CaseType := fileCells[1]
	FilingDate := fileCells[2]
	//BPT := fileCells[3]
	PartyType := fileCells[4]
	NameLast := fileCells[5]
	NameFirst := fileCells[6]
	NameMid := fileCells[7]
	//Sfx := fileCells[8]
	Address1 := fileCells[9]
	Address2 := fileCells[10]
	City := fileCells[11]
	State := fileCells[12]
	ZipCode := fileCells[13]

	caseFilingDate := exs.parseFileDate(FilingDate)

	Party := domain.Party{
		Name:        NameLast + ", " + NameFirst + " " + NameMid,
		Kind:        PartyType,
		LastName:    NameLast,
		FirstName:   NameFirst,
		MiddleName:  NameMid,
		Address:     Address1 + Address2 + City + ", " + State + " " + ZipCode,
		AddressOnly: Address1 + Address2,
		Building:    "",
		CompanyName: "",
		City:        City,
		State:       State,
		Zip:         ZipCode,
		Apartment:   Address2,
	}

	issueDateDetails := domain.IssueDateDetails{
		Raw:   FilingDate,
		Day:   caseFilingDate.Day(),
		Month: int(caseFilingDate.Month()),
		Year:  caseFilingDate.Year(),
	}

	caseDetail := domain.CaseDetails{
		County:            "",
		CaseStyle:         "",
		Kind:              "",
		CaseNumber:        CaseNumber,
		ChargeDescription: "",
		CaseType:          CaseType,
		IssueDateDetails:  issueDateDetails,
		Parties:           []domain.Party{Party},
	}
	return caseDetail
}

func (ExcelFileService) mergeDuplicateCases(cases []domain.CaseDetails) []domain.CaseDetails {
	combinedStructs := make(map[string]domain.CaseDetails)

	// Combine properties for duplicate IDs
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

func (ExcelFileService) removeDuplicateParties(caseDetails []domain.CaseDetails) []domain.CaseDetails {
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
