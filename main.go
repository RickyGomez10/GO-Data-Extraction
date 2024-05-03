package main

import (
	"openLaw-dataextraction2/application"
	"openLaw-dataextraction2/infrastructure"
)

const fileName = ""
const outputFileName = ""

func main() {

	fileRepositoryFactory := infrastructure.FileRepositoryFactory{}
	fileRepository := fileRepositoryFactory.GetFileRepository(fileName)
	fileProcessorService := application.FileProcessorService{FileRepository: fileRepository}
	fileProcessorService.ProcessFile(fileName, outputFileName)

}
