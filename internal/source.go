package internal

import (
	"encoding/csv"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

type DataGouvSource struct {
	url *url.URL
}

func NewDataGouvSource(url *url.URL) *DataGouvSource {
	return &DataGouvSource{url: url}
}

func (s *DataGouvSource) Get() ([]DepartmentDailyReportPerAge, error) {
	reports := []DepartmentDailyReportPerAge{}

	resp, err := http.Get(s.url.String())
	if err != nil {
		return reports, errors.Wrap(err, "unable to retrieve csv file from url")
	}
	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)
	reader.Comma = ';'
	data, err := reader.ReadAll()
	if err != nil {
		return reports, errors.Wrap(err, "unable to parse csv file from url")
	}
	for i, row := range data {
		if i == 0 {
			if !isCSVHeaderValid(row) {
				return nil, errors.New("invalid csv headers - this csv format is not supported")
			}
			continue
		}

		report, err := parseCSVRow(row)
		if err != nil {
			log.Println(errors.Wrap(err, "unable to parse csv row"))
			continue
		}

		reports = append(reports, report)
	}

	return reports, nil
}

var (
	expectedCSVHeaders = []string{"dep", "jour", "P", "T", "cl_age90", "pop"}
)

func isCSVHeaderValid(headers []string) bool {
	if len(headers) != len(expectedCSVHeaders) {
		return false
	}

	for i, expectedHeader := range expectedCSVHeaders {
		if headers[i] != expectedHeader {
			return false
		}
	}

	return true

}

func parseCSVRow(row []string) (DepartmentDailyReportPerAge, error) {
	report := DepartmentDailyReportPerAge{
		DepartmentNumber: row[0],
	}

	day, err := time.Parse("2006-01-02", row[1])
	if err != nil {
		return report, errors.Wrapf(err, "invalid day (%s)", row[1])
	}
	report.Day = day

	ageCategory, err := strconv.Atoi(row[4])
	if err != nil {
		return report, errors.Wrapf(err, "invalid age category (%s)", row[4])
	}
	report.AgeCategory = ageCategory

	testsTotal, err := strconv.Atoi(row[3])
	if err != nil {
		return report, errors.Wrapf(err, "invalid total tests (%s)", row[3])
	}
	report.TestsTotal = testsTotal

	testsPositives, err := strconv.Atoi(row[2])
	if err != nil {
		return report, errors.Wrapf(err, "invalid total tests (%s)", row[2])
	}
	report.TestsPositives = testsPositives

	population, err := strconv.Atoi(row[5])
	if err != nil {
		// Try to parse as float and round. Some data from department 978 contains float data :/
		populationFloat, err := strconv.ParseFloat(row[5], 32)
		if err != nil {
			return report, errors.Wrapf(err, "invalid population (%s)", row[5])
		}
		population = int(math.Round(populationFloat))
	}
	report.Population = population

	return report, nil
}
