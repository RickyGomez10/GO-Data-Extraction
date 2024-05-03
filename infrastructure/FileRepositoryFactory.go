package infrastructure

import (
	"bufio"
	"github.com/tealeg/xlsx"
	"log"
	"openLaw-dataextraction2/domain"
	"os"
	"path"
	"regexp"
	"strings"
)

type FileRepositoryFactory struct{}

func (fpf FileRepositoryFactory) GetFileRepository(fileNameWithExtension string) domain.FileRepository {
	extension := path.Ext(fileNameWithExtension)
	var fileRepository domain.FileRepository
	switch extension {
	case ".xlsx":
		fileRepository = fpf.verifyXlsxToUse(fileNameWithExtension)
		break
	case ".txt":
		fileRepository = fpf.verifyTextFileToUse(fileNameWithExtension)
	default:
		fileRepository = OtherTextFile{}
		break
	}
	return fileRepository
}

func (FileRepositoryFactory) verifyXlsxToUse(fileNameWithExtension string) domain.FileRepository {
	xlFile, err := xlsx.OpenFile(fileNameWithExtension)
	if err != nil {
		log.Fatalf("Error opening Excel file: %v", err)
	}
	for _, sheet := range xlFile.Sheets {

		for _, cell := range sheet.Rows[0].Cells {
			text := cell.String()
			if text == "CaseNbr" ||
				text == "FileDate" ||
				text == "Judge" ||
				text == "Style" ||
				text == "PartyNameFMLS" ||
				text == "AddressLine1" ||
				text == "AddressLine2" ||
				text == "AddressCity" ||
				text == "AddressState" ||
				text == "AddressZip" ||
				text == "ArrestDate" ||
				text == "OffenseDate" ||
				text == "OffenseCode" ||
				text == "OffenseCodeDescription" {
				return ExcelFileService2{}
			}

		}

	}

	return ExcelFileService{}
}

func (FileRepositoryFactory) verifyTextFileToUse(fileNameWithExtension string) domain.FileRepository {
	regexpPatternCircuitCounty := `^(0?[1-9]|1[0-2])/(0?[1-9]|[12]\d|3[01])/\d{4}$`
	regexPatternCivil := `^.{2}$`
	var splitRow []string
	var fileRows []string
	fileData, err := os.Open(fileNameWithExtension)
	if err != nil {
		log.Fatalf("Error opening File: %v", err)
	}

	scanner := bufio.NewScanner(fileData)

	for scanner.Scan() {
		if !strings.Contains(scanner.Text(), "EOF") {
			fileRow := strings.ReplaceAll(scanner.Text(), ",", ";")
			fileRows = append(fileRows, fileRow)
		}
	}

	for _, fileRow := range fileRows[:1] {
		unescaped := strings.ReplaceAll(fileRow, "\"", "")
		splitRow = strings.Split(unescaped, ";")

		matchCircuitCountyFileFormat, _ := regexp.MatchString(regexpPatternCircuitCounty, splitRow[2])
		matchCivilFileFormat, _ := regexp.MatchString(regexPatternCivil, splitRow[2])

		if matchCircuitCountyFileFormat {
			return TextFileCircuit{}
		} else if matchCivilFileFormat {
			return TextFileCivil{}
		} else {
			return TextFileTraffic{}
		}

	}

	return TextFileCircuit{}
}
