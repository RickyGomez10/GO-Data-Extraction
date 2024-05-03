package application

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"openLaw-dataextraction2/domain"
	"sort"
)

type FileProcessorService struct {
	FileRepository domain.FileRepository
}

func (fps FileProcessorService) ProcessFile(inputFileNameWithExtension string, outputFileNameWithExtension string) {
	data := fps.FileRepository.ReadData(inputFileNameWithExtension)
	caseDetails := fps.FileRepository.UnmarshalToCaseDetails(data)
	jsonBytes := fps.unmarshalJSON(caseDetails)
	fps.createFile(outputFileNameWithExtension, jsonBytes)
}

func (FileProcessorService) unmarshalJSON(caseDetails []domain.CaseDetails) []byte {
	jsonBytes, err := json.MarshalIndent(caseDetails, "", " ")
	if err != nil {
		log.Fatalf("Error while unmarshaling case details %v", err)
		return nil
	}
	return jsonBytes
}

func (FileProcessorService) createFile(fileName string, dataToInsertInFile []byte) {
	ioutil.WriteFile(fileName, dataToInsertInFile, 0644)
}

func (FileProcessorService) sortByAddress(details []domain.CaseDetails) {
	sort.Slice(details, func(i, j int) bool {
		addressI := ""
		for _, party := range details[i].Parties {
			if party.Kind == "Defendant" {
				addressI = party.Address
				break
			}
		}

		addressJ := ""
		for _, party := range details[j].Parties {
			if party.Kind == "Defendant" {
				addressJ = party.Address
				break
			}
		}

		return addressI < addressJ
	})
}

func (FileProcessorService) sortByCaseType(details []domain.CaseDetails) {
	sort.Slice(details, func(i, j int) bool {
		return details[i].CaseType < details[j].CaseType
	})
}

func (fps FileProcessorService) sortByDefendantName(details []domain.CaseDetails) {
	sort.Slice(details, func(i, j int) bool {
		defNameI := ""
		for _, party := range details[i].Parties {
			if party.Kind == "Defendant" {
				defNameI = party.Name
				break
			}
		}

		defNameJ := ""
		for _, party := range details[j].Parties {
			if party.Kind == "Defendant" {
				defNameJ = party.Name
				break
			}
		}

		return defNameI < defNameJ
	})
}
