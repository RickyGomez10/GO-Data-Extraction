package domain

type FileRepository interface {
	ReadData(string) []string
	UnmarshalToCaseDetails([]string) []CaseDetails
}
