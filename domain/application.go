package domain

type Application struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	Image        string `json:"image"`
	Label        string `json:"label"`
	Path         string `json:"path"`
	Hostname     string `json:"hostname"`
	Port         int    `json:"port"`
	Lifetime     int    `json:"lifetime"`
	LeaseRenewed int    `json:"lease_renewed"`
}
