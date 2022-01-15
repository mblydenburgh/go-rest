package domain

type Car struct {
	UserId         string `json:"userId"`
	ModelTypeAndId string `json:"modelTypeAndId"`
	Manufacturer   string `json:"manufacturer"`
	Model          string `json:"model"`
	Trim           string `json:"trim"`
	Year           int32  `json:"year"`
	VehicleType    string `json:"vehicleType"`
	Color          string `json:"color"`
	VIN            string `json:"vin"`
}

type SaveCarPayload struct {
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	Trim         string `json:"trim"`
	Year         int32  `json:"year"`
	VehicleType  string `json:"vehicleType"`
	Color        string `json:"color"`
	VIN          string `json:"vin"`
}
