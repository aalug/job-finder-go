package esearch

// === Types for the ES part of the Application ===

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

// === for the Context ===
type contextKey struct {
	Key int
}

var JobKey = contextKey{Key: 1}
var ClientKey = contextKey{Key: 2}

// === Queries and Searches ===

type GetResponse struct {
	Index   string `json:"_index"`
	ID      string `json:"_id"`
	Version int    `json:"_version"`
	Source  *Job   `json:"_source"`
}

type SearchResponse struct {
	Hits struct {
		Total struct {
			Value int64 `json:"value"`
		} `json:"total"`
		Hits []*struct {
			Source *Job   `json:"_source"`
			ID     string `json:"_id"`
		} `json:"hits"`
	} `json:"hits"`
}
