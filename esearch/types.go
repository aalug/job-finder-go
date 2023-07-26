package esearch

type Job struct {
	ID           int32    `json:"id"`
	Title        string   `json:"title"`
	Industry     string   `json:"industry"`
	CompanyName  string   `json:"company_name"`
	Description  string   `json:"description"`
	Location     string   `json:"location"`
	SalaryMin    int32    `json:"salary_min"`
	SalaryMax    int32    `json:"salary_max"`
	Requirements string   `json:"requirements"`
	JobSkills    []string `json:"job_skills"`
}

type contextKey struct {
	Key int
}

var JobKey contextKey = contextKey{Key: 1}
var ClientKey contextKey = contextKey{Key: 2}
