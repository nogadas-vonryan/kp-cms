package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type CaseFields struct {
	Complainants     []string `json:"complainants"`
	Respondents      []string `json:"respondents"`
	Status           string   `json:"status"`
	CaseStage        string   `json:"case_stage"`
	DocketNumber     string   `json:"docket_number"`
	CourtBranch      string   `json:"court_branch"`
	AssignedAttorney string   `json:"assigned_attorney"`
	IncidentDate     string   `json:"incident_date"`
	FilingDate       string   `json:"filing_date"`
	NextHearingDate  string   `json:"next_hearing_date"`
	PriorityLevel    string   `json:"priority_level"`
	Tags             []string `json:"tags"`
	Remarks          string   `json:"remarks"`
}

type LegalCase struct {
	Title  string     `json:"title"`
	Fields CaseFields `json:"fields"`
}

var (
	firstNames = []string{"Juan", "Maria", "Ricardo", "Elena", "Antonio", "Liza", "Jose", "Carmela"}
	lastNames  = []string{"Dela Cruz", "Santos", "Reyes", "Gonzales", "Bautista", "Garcia", "Aquino"}
	stages     = []string{"Preliminary Investigation", "Arraignment", "Pre-Trial", "Trial", "Judgment", "Appeal"}
	branches   = []string{"RTC Branch 14", "RTC Branch 257", "MTC Branch 77", "RTC Branch 128"}
	attorneys  = []string{"Atty. Clara Vincent", "Atty. Mark Rover", "Atty. Sofia Zobel", "Atty. Jun Sabado"}
	caseTypes  = []string{"Land Dispute", "Estafa", "Libel", "Physical Injuries", "Theft"}
)

// GenerateCases creates a slice of N unique-ish cases
func GenerateCases(n int) []LegalCase {
	rand.Seed(time.Now().UnixNano())
	cases := make([]LegalCase, n)

	for i := 0; i < n; i++ {
		cases[i] = LegalCase{
			Title: fmt.Sprintf("%s - %s vs %s #%d",
				caseTypes[rand.Intn(len(caseTypes))],
				lastNames[rand.Intn(len(lastNames))],
				lastNames[rand.Intn(len(lastNames))],
				i+1000),
			Fields: CaseFields{
				Complainants:     generatePeople(1, 3),
				Respondents:      generatePeople(1, 2),
				Status:           "criminal",
				CaseStage:        stages[rand.Intn(len(stages))],
				DocketNumber:     fmt.Sprintf("I.S. NO. XV-01-INV-%d%c-%05d", rand.Intn(26)+20, 'A'+rand.Intn(5), i),
				CourtBranch:      fmt.Sprintf("%s - Parañaque City", branches[rand.Intn(len(branches))]),
				AssignedAttorney: attorneys[rand.Intn(len(attorneys))],
				IncidentDate:     time.Now().AddDate(0, 0, -rand.Intn(365)).Format("2006-01-02"),
				FilingDate:       time.Now().AddDate(0, 0, -rand.Intn(30)).Format("2006-01-02"),
				NextHearingDate:  time.Now().AddDate(0, 0, rand.Intn(60)).Format(time.RFC3339),
				PriorityLevel:    []string{"Low", "Medium", "High", "Urgent"}[rand.Intn(4)],
				Tags:             []string{"Legal", "Property", "Criminal Case"},
				Remarks:          generateRandomText(rand.Intn(3)), // 0: short, 1: medium, 2: long
			},
		}
	}
	return cases
}

func generatePeople(min, max int) []string {
	count := rand.Intn(max-min+1) + min
	people := make([]string, count)
	for i := 0; i < count; i++ {
		people[i] = fmt.Sprintf("%s %s", firstNames[rand.Intn(len(firstNames))], lastNames[rand.Intn(len(lastNames))])
	}
	return people
}

func generateRandomText(mode int) string {
	short := []string{"Case pending.", "Awaiting documents.", "Client notified."}
	medium := []string{
		"Initial counter-affidavits filed; awaiting rejoinder from complainants.",
		"Mediation scheduled for next month. Settlement is being explored.",
	}
	long := []string{
		"The respondent has requested an extension to file a counter-affidavit due to medical reasons. The court has granted 15 days. Next step is to verify the medical certificate provided by the local health office.",
		"Lengthy dispute over technical descriptions in the land title. Surveyors have been called to testify in the next hearing to clarify the encroachment boundaries in Brgy. San Antonio.",
	}

	switch mode {
	case 1:
		return medium[rand.Intn(len(medium))]
	case 2:
		return long[rand.Intn(len(long))]
	default:
		return short[rand.Intn(len(short))]
	}
}

const (
	TargetURL    = "http://localhost:8080/documents" // Change to local server URL
	TotalRecords = 0
)

func main() {
	fmt.Printf("🚀 Starting mock data generation for %d records...\n", TotalRecords)

	if TotalRecords == 0 {
		log.Println("Total record count is set to 0")
		return
	}

	cases := GenerateCases(TotalRecords)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	successCount := 0
	errorCount := 0

	for i, c := range cases {
		// Convert struct to JSON
		jsonData, err := json.Marshal(c)
		if err != nil {
			fmt.Printf("❌ Error marshaling case %d: %v\n", i, err)
			continue
		}

		// Create POST request
		req, err := http.NewRequest("POST", TargetURL, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("❌ Error creating request for case %d: %v\n", i, err)
			continue
		}

		req.SetBasicAuth("admin", "Amz2rPA7")
		req.Header.Set("Content-Type", "application/json")

		// Execute request
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("⚠️ Request %d failed: %v\n", i, err)
			errorCount++
			continue
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			successCount++
		} else {
			fmt.Printf("⚠️ Case %d received status: %d\n", i, resp.StatusCode)
			errorCount++
		}
		resp.Body.Close()

		// Optional: throttle requests slightly to avoid overwhelming a local dev server
		if i%100 == 0 && i != 0 {
			fmt.Printf("✅ Processed %d records...\n", i)
			time.Sleep(100 * time.Millisecond)
		}
	}

	fmt.Printf("\n--- Seeding Complete ---\n")
	fmt.Printf("Total Success: %d\n", successCount)
	fmt.Printf("Total Failures: %d\n", errorCount)
}
