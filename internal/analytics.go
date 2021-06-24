package internal

import "time"

type DepartmentDailyReportPerAge struct {
	DepartmentNumber string    `json:"department_number"`
	Day              time.Time `json:"day"`
	AgeCategory      int       `json:"age_category"`
	TestsTotal       int       `json:"tests_total"`
	TestsPositives   int       `json:"tests_positives"`
	Population       int       `json:"population"`
}

type NationalDailyReport struct {
	Day            time.Time `json:"day"`
	TestsTotal     int       `json:"tests_total"`
	TestsPositives int       `json:"tests_positives"`
	Ratio          float64   `json:"ratio"`
}

type DepartmentResume struct {
	DayWithMostTests     time.Time `json:"day_with_most_tests"`
	DayWithMostPositives time.Time `json:"day_with_most_positives"`
	DayWithHighestRatio  time.Time `json:"day_with_highest_ratio"`
}

type DailyTop5 struct {
	AgeCategories map[int]*DailyTop5PerAgeCategory `json:"age_categories"`
}

type DailyTop5PerAgeCategory struct {
	Top5 []DepartmentDailyReportPerAge `json:"top5"`
}
