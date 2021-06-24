package internal

import (
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

const tableName = "department_daily_report"

type StorageSQL struct {
	db *sql.DB
}

func NewStorageSQL(db *sql.DB) *StorageSQL {
	return &StorageSQL{db: db}
}

func (s *StorageSQL) Save(reports ...DepartmentDailyReportPerAge) error {
	sqlStr := "INSERT INTO " + tableName + "(department_number, day, age_category, tests_total, tests_positives, population)" +
		" VALUES (?, ?, ?, ?, ?, ?) " +
		"ON DUPLICATE KEY UPDATE tests_total=VALUES(tests_total), tests_positives=VALUES(tests_positives), population=VALUES(population)"
	stmt, err := s.db.Prepare(sqlStr)
	if err != nil {
		return err
	}

	for _, report := range reports {
		if _, err := stmt.Exec(report.DepartmentNumber, report.Day, report.AgeCategory, report.TestsTotal, report.TestsPositives, report.Population); err != nil {
			return err
		}
	}
	return nil
}

func (s *StorageSQL) Departments() ([]string, error) {
	departments := []string{}

	rows, err := s.db.Query("SELECT DISTINCT department_number FROM " + tableName)
	if err != nil {
		return departments, errors.Wrap(err, "cannot retrieve departments")
	}

	defer rows.Close()
	for rows.Next() {
		department := ""
		err := rows.Scan(&department)
		if err != nil {
			return departments, errors.Wrap(err, "cannot scan department")
		}
		departments = append(departments, department)
	}

	return departments, nil
}

func (s *StorageSQL) AgeCategories() ([]int, error) {
	ageCategories := []int{}

	rows, err := s.db.Query("SELECT DISTINCT age_category FROM " + tableName)
	if err != nil {
		return ageCategories, errors.Wrap(err, "cannot retrieve age categories")
	}

	defer rows.Close()
	for rows.Next() {
		ageCategory := 0
		err := rows.Scan(&ageCategory)
		if err != nil {
			return ageCategories, errors.Wrap(err, "cannot scan age category")
		}

		ageCategories = append(ageCategories, ageCategory)
	}

	return ageCategories, nil
}

func (s *StorageSQL) DaysLimit() (time.Time, time.Time, error) {
	var from, to time.Time
	rows, err := s.db.Query("SELECT min(day), max(day) FROM " + tableName)
	if err != nil {
		return from, to, errors.Wrap(err, "cannot retrieve days limit")
	}
	defer rows.Close()

	for rows.Next() {
		var fromStr, toStr string
		err := rows.Scan(&fromStr, &toStr)
		if err != nil {
			return from, to, errors.Wrap(err, "cannot scan days limit")
		}

		from, err = time.Parse("2006-01-02", fromStr)
		if err != nil {
			return from, to, errors.Wrap(err, "cannot parse from for days limit")
		}

		to, err = time.Parse("2006-01-02", toStr)
		if err != nil {
			return from, to, errors.Wrap(err, "cannot parse to for days limit")
		}
	}
	return from, to, nil
}

func (s *StorageSQL) Reports(from, to time.Time, departments ...string) ([]DepartmentDailyReportPerAge, error) {
	reports := []DepartmentDailyReportPerAge{}

	query := "SELECT department_number, day, age_category, tests_total, tests_positives, population FROM " + tableName +
		" WHERE day >= ? AND day <= ? "
	args := []interface{}{from, to}
	if len(departments) > 0 {
		query += "AND department_number IN ("
		for _, department := range departments {
			query += "?,"
			args = append(args, department)
		}
		query = query[0 : len(query)-1]
		query += ")"
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return reports, errors.Wrap(err, "cannot query reports")
	}
	defer rows.Close()
	for rows.Next() {
		report := DepartmentDailyReportPerAge{}
		var dayStr string
		err := rows.Scan(&report.DepartmentNumber, &dayStr, &report.AgeCategory, &report.TestsTotal, &report.TestsPositives, &report.Population)
		if err != nil {
			return reports, errors.Wrap(err, "cannot scan daily department report")
		}
		day, err := time.Parse("2006-01-02", dayStr)
		if err != nil {
			return reports, errors.Wrapf(err, "invalid day data (%s)", dayStr)
		}
		report.Day = day

		reports = append(reports, report)
	}

	return reports, nil
}

func (s *StorageSQL) NationalReports(from, to time.Time) ([]NationalDailyReport, error) {
	reports := []NationalDailyReport{}

	query := "SELECT day, sum(tests_total), sum(tests_positives) FROM " + tableName +
		" WHERE day >= ? AND day <= ? " +
		"GROUP BY day"

	rows, err := s.db.Query(query, from, to)
	if err != nil {
		return reports, errors.Wrap(err, "cannot query national reports")
	}
	defer rows.Close()
	for rows.Next() {
		report := NationalDailyReport{}
		var dayStr string
		err := rows.Scan(&dayStr, &report.TestsTotal, &report.TestsPositives)
		if err != nil {
			return reports, errors.Wrap(err, "cannot scan daily national report")
		}
		day, err := time.Parse("2006-01-02", dayStr)
		if err != nil {
			return reports, errors.Wrapf(err, "invalid day data (%s)", dayStr)
		}
		report.Day = day
		if report.TestsTotal != 0 {
			report.Ratio = float64(report.TestsPositives) / float64(report.TestsTotal)
		}

		reports = append(reports, report)
	}

	return reports, nil
}

func (s *StorageSQL) DepartmentResume(department string) (DepartmentResume, error) {
	resume := DepartmentResume{
		DayWithMostTests:     time.Time{},
		DayWithMostPositives: time.Time{},
		DayWithHighestRatio:  time.Time{},
	}
	maxTests := 0
	maxPositives := 0
	maxRatio := 0.0

	query := "SELECT day, sum(tests_total), sum(tests_positives) FROM " + tableName +
		" WHERE department_number = ?" +
		" GROUP BY day"

	rows, err := s.db.Query(query, department)
	if err != nil {
		return resume, errors.Wrap(err, "cannot query department resume")
	}
	defer rows.Close()
	for rows.Next() {
		var dayStr string
		var testsTotal, testsPositives int
		err := rows.Scan(&dayStr, &testsTotal, &testsPositives)
		if err != nil {
			return resume, errors.Wrap(err, "cannot scan department resume")
		}
		day, err := time.Parse("2006-01-02", dayStr)
		if err != nil {
			return resume, errors.Wrapf(err, "invalid day data (%s)", dayStr)
		}

		if testsTotal <= 10 {
			continue
		}

		ratio := 0.0
		if testsTotal > 0 {
			ratio = float64(testsPositives) / float64(testsTotal)
		}

		if testsTotal > maxTests {
			resume.DayWithMostTests = day
			maxTests = testsTotal
		}

		if testsPositives > maxPositives {
			resume.DayWithMostPositives = day
			maxPositives = testsPositives
		}

		if ratio > maxRatio {
			resume.DayWithHighestRatio = day
			maxRatio = ratio
		}
	}

	return resume, nil
}

func (s *StorageSQL) DailyTop5(day time.Time) (DailyTop5, error) {
	top5 := DailyTop5{
		AgeCategories: map[int]*DailyTop5PerAgeCategory{},
	}
	reports, err := s.Reports(day, day)
	if err != nil {
		return top5, err
	}

	for _, report := range reports {
		if report.TestsTotal <= 10 {
			continue
		}
		ratio := float64(report.TestsPositives) / float64(report.TestsTotal)

		if _, ok := top5.AgeCategories[report.AgeCategory]; !ok {
			top5.AgeCategories[report.AgeCategory] = &DailyTop5PerAgeCategory{Top5: []DepartmentDailyReportPerAge{report}}
			continue
		}

		isInserted := false
		for i, topReport := range top5.AgeCategories[report.AgeCategory].Top5 {
			topRatio := float64(topReport.TestsPositives) / float64(topReport.TestsTotal)
			if ratio < topRatio {
				continue
			}

			top5.AgeCategories[report.AgeCategory].Top5 = append(
				top5.AgeCategories[report.AgeCategory].Top5[:i],
				append([]DepartmentDailyReportPerAge{report}, top5.AgeCategories[report.AgeCategory].Top5[i:]...)...,
			)
			isInserted = true
			break
		}
		if !isInserted {
			top5.AgeCategories[report.AgeCategory].Top5 = append(top5.AgeCategories[report.AgeCategory].Top5, report)
		}

		if len(top5.AgeCategories[report.AgeCategory].Top5) > 5 {
			top5.AgeCategories[report.AgeCategory].Top5 = top5.AgeCategories[report.AgeCategory].Top5[:5]
		}
	}

	return top5, nil
}
