package infrastructure

import (
	"github.com/tealeg/xlsx"
	"log"
	"openLaw-dataextraction2/domain"
	"strings"
	"time"
)

type ExcelFileService2 struct {
}

func (efs2 ExcelFileService2) ReadData(FileNameWithExtension string) []string {

	xlFile, err := xlsx.OpenFile(FileNameWithExtension)
	if err != nil {
		log.Fatalf("Error opening Excel file: %v", err)
	}

	var data []string

	for _, sheet := range xlFile.Sheets {

		for _, row := range sheet.Rows[1:] {
			var rowData []string
			for i, cell := range row.Cells {
				text := cell.String()

				// Convert Excel serial date value to Go time.Time
				if i == 1 || i == 10 && cell.Value != "NULL" { // Assuming the date is in the third cell (index 2)
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

func (efs2 ExcelFileService2) UnmarshalToCaseDetails(semiColonFileRows []string) []domain.CaseDetails {
	var fileCells []string
	var caseDetails []domain.CaseDetails

	for _, fileRow := range semiColonFileRows {
		fileCells = strings.Split(fileRow, ";")
		caseDetail := efs2.mapFileData(fileCells)
		caseDetails = append(caseDetails, caseDetail)

	}
	//Hacer merge
	semifinalCaseDetails := efs2.mergeDuplicateCases(caseDetails)
	finalCaseDetails := efs2.removeDuplicateParties(semifinalCaseDetails)
	return finalCaseDetails
}

func (ExcelFileService2) parseFileDate(date string) time.Time {
	layout := "01/02/2006"
	parsedDate, err := time.Parse(layout, date)
	if err != nil {
		log.Print(err)
		return time.Now()
	}
	return parsedDate
}

func (efs2 ExcelFileService2) mapFileData(fileCells []string) domain.CaseDetails {
	CaseNumber := checkForNUllString(fileCells[0])
	fileDate := checkForNUllString(fileCells[1])
	//Judge := fileCells[2]
	Style := checkForNUllString(fileCells[3])
	PartyNameFMLS := checkForNUllString(fileCells[4])
	AddressLine1 := checkForNUllString(fileCells[5])
	AddressLine2 := checkForNUllString(fileCells[6])
	AddressCity := checkForNUllString(fileCells[7])
	AddressState := checkForNUllString(fileCells[8])
	AddressZip := checkForNUllString(fileCells[9])
	//ArrestDate := fileCells[10]
	//OffenseCode := fileCells[11]
	OffenseCodeDescription := checkForNUllString(fileCells[13])

	partiesSplit := strings.Split(Style, "VS.")

	caseFilingDate := efs2.parseFileDate(fileDate)
	parties := efs2.mapParties(partiesSplit, PartyNameFMLS, AddressLine1, AddressLine2, AddressCity, AddressState, AddressZip)

	issueDateDetails := domain.IssueDateDetails{
		Raw:   checkForNUllString(fileDate),
		Day:   caseFilingDate.Day(),
		Month: int(caseFilingDate.Month()),
		Year:  caseFilingDate.Year(),
	}

	caseDetail := domain.CaseDetails{
		County:            "",
		CaseStyle:         checkForNUllString(Style),
		Kind:              "",
		CaseNumber:        checkForNUllString(CaseNumber),
		ChargeDescription: checkForNUllString(OffenseCodeDescription),
		CaseType:          "",
		IssueDateDetails:  issueDateDetails,
		Parties:           parties,
	}
	return caseDetail
}

func (ExcelFileService2) mergeDuplicateCases(cases []domain.CaseDetails) []domain.CaseDetails {
	combinedStructs := make(map[string]domain.CaseDetails)

	// Combine properties for duplicate IDs
	for i, _ := range cases {
		existingItem, found := combinedStructs[cases[i].CaseNumber]
		if found {
			// Combine properties (if needed)
			// For simplicity, let's assume there's a field called "Value" to combine
			// You can modify this based on your struct's fields
			existingItem.Parties = append(existingItem.Parties, cases[i].Parties[0])
			combinedStructs[cases[i].CaseNumber] = existingItem
		} else {
			// If not a duplicate, add to the map
			combinedStructs[cases[i].CaseNumber] = cases[i]
		}
	}

	combinedSlice := make([]domain.CaseDetails, 0, len(combinedStructs))
	for _, item := range combinedStructs {
		combinedSlice = append(combinedSlice, item)
	}

	return combinedSlice
}

func (ExcelFileService2) mapParties(parties []string, partyNameFMLS string, address1 string, address2 string, city string, state string, zipcode string) []domain.Party {

	var partiesSlice []domain.Party
	for _, party := range parties {

		partyLowerCase := strings.ToLower(party)
		if strings.Contains(partyLowerCase, "state") {

			StateParty := domain.Party{
				Name: checkForNUllString(party),
				Kind: "State",
			}

			partiesSlice = append(partiesSlice, StateParty)
			continue
		}

		DefendantNameSplit := strings.Split(partyNameFMLS, " ")
		var DefendantFirstName string
		var DefendantMidName string

		var DefendantLastName string

		if len(DefendantNameSplit) == 2 {
			DefendantFirstName = checkForNUllString(DefendantNameSplit[0])
			DefendantLastName = checkForNUllString(DefendantNameSplit[1])
		}

		if len(DefendantNameSplit) > 2 {
			DefendantFirstName = checkForNUllString(DefendantNameSplit[0])
			DefendantMidName = checkForNUllString(DefendantNameSplit[1])
			DefendantLastName = checkForNUllString(DefendantNameSplit[2])
		}

		HumanParty := domain.Party{
			Name:        checkForNUllString(DefendantLastName) + ", " + checkForNUllString(DefendantFirstName) + " " + checkForNUllString(DefendantMidName),
			Kind:        "Defendant",
			LastName:    checkForNUllString(DefendantLastName),
			FirstName:   checkForNUllString(DefendantFirstName),
			MiddleName:  checkForNUllString(DefendantMidName),
			Address:     checkForNUllString(DefendantFirstName) + " " + checkForNUllString(DefendantLastName) + " " + checkForNUllString(address1) + " " + checkForNUllString(address2) + " " + checkForNUllString(city) + ", " + checkForNUllString(state) + " " + checkForNUllString(zipcode),
			AddressOnly: checkForNUllString(address1) + " " + checkForNUllString(address2),
			Building:    "",
			CompanyName: "",
			City:        checkForNUllString(city),
			State:       checkForNUllString(state),
			Zip:         checkForNUllString(zipcode),
			Apartment:   checkForNUllString(address2),
		}

		partiesSlice = append(partiesSlice, HumanParty)
	}
	return partiesSlice
}

func checkForNUllString(value string) string {
	if strings.ToLower(value) == strings.ToLower("NULL") {
		return ""
	}
	return value
}

func (ExcelFileService2) removeDuplicateParties(caseDetails []domain.CaseDetails) []domain.CaseDetails {
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
