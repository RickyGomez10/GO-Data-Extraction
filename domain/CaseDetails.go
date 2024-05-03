package domain

type IssueDateDetails struct {
	Raw   string `json:"raw"`
	Day   int    `json:"day"`
	Month int    `json:"month"`
	Year  int    `json:"year"`
}

type Party struct {
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	LastName    string `json:"last_name"`
	FirstName   string `json:"first_name"`
	MiddleName  string `json:"middle_name"`
	Address     string `json:"address"`
	AddressOnly string `json:"address_only"`
	Building    string `json:"building"`
	CompanyName string `json:"company_name"`
	City        string `json:"city"`
	State       string `json:"state"`
	Zip         string `json:"zip"`
	Apartment   string `json:"apartment"`
}

type CaseDetails struct {
	Status            string           `json:"status,omitempty"`
	County            string           `json:"county" json:"county,omitempty"`
	CaseStyle         string           `json:"case_style" json:"caseStyle,omitempty"`
	Kind              string           `json:"kind" json:"kind,omitempty"`
	CaseNumber        string           `json:"casenumber" json:"caseNumber,omitempty"`
	ChargeDescription string           `json:"charge_description" json:"chargeDescription,omitempty"`
	CaseType          string           `json:"casetype" json:"caseType,omitempty"`
	IssueDateDetails  IssueDateDetails `json:"issue_date_details" json:"issueDateDetails"`
	Parties           []Party          `json:"parties" json:"parties,omitempty"`
}
