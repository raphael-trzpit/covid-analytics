package internal

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type Importer interface {
	Get() ([]DepartmentDailyReportPerAge, error)
}

type Storage interface {
	Save(reports ...DepartmentDailyReportPerAge) error
	Departments() ([]string, error)
	AgeCategories() ([]int, error)
	Reports(from time.Time, to time.Time, departments ...string) ([]DepartmentDailyReportPerAge, error)
	DaysLimit() (time.Time, time.Time, error)
	NationalReports(from, to time.Time) ([]NationalDailyReport, error)
	DepartmentResume(department string) (DepartmentResume, error)
	DailyTop5(day time.Time) (DailyTop5, error)
}

type ImportHandler struct {
	importer Importer
	storage  Storage
}

func NewImportHandler(importer Importer, storage Storage) *ImportHandler {
	return &ImportHandler{importer: importer, storage: storage}
}

func (h *ImportHandler) ImportHandler(w http.ResponseWriter, r *http.Request) {
	reports, err := h.importer.Get()
	if err != nil {
		writeJSONError(w, err)
		return
	}

	err = h.storage.Save(reports...)
	if err != nil {
		writeJSONError(w, err)
		return
	}

	writeJSONResponse(w, http.StatusOK, struct {
		Result string `json:"result"`
	}{
		Result: "OK",
	})
}

type AnalyticsHandler struct {
	storage Storage
}

func NewAnalyticsHandler(storage Storage) *AnalyticsHandler {
	return &AnalyticsHandler{storage: storage}
}

func (h *AnalyticsHandler) DepartmentListHandler(w http.ResponseWriter, r *http.Request) {
	departments, err := h.storage.Departments()
	if err != nil {
		writeJSONError(w, err)
		return
	}

	writeJSONResponse(w, http.StatusOK, departments)
}

func (h *AnalyticsHandler) AgeCategoryListHandler(w http.ResponseWriter, r *http.Request) {
	ageCategories, err := h.storage.AgeCategories()
	if err != nil {
		writeJSONError(w, err)
		return
	}

	writeJSONResponse(w, http.StatusOK, ageCategories)
}

func (h *AnalyticsHandler) DaysLimitHandler(w http.ResponseWriter, r *http.Request) {
	from, to, err := h.storage.DaysLimit()
	if err != nil {
		writeJSONError(w, err)
	}

	writeJSONResponse(w, http.StatusOK, struct {
		From time.Time `json:"from"`
		To   time.Time `json:"to"`
	}{
		From: from,
		To:   to,
	})
}

func (h *AnalyticsHandler) DataHandler(w http.ResponseWriter, r *http.Request) {
	fromParam, ok := r.URL.Query()["from"]
	if !ok || len(fromParam) != 1 {
		writeJSONResponse(w, http.StatusBadRequest, ErrorResponse{Error: "query param from is mandatory"})
		return
	}

	from, err := time.Parse("2006-01-02", fromParam[0])
	if err != nil {
		writeJSONResponse(w, http.StatusBadRequest, ErrorResponse{Error: errors.Wrap(err, "from must be a date").Error()})
		return
	}

	toParam, ok := r.URL.Query()["to"]
	if !ok || len(toParam) != 1 {
		writeJSONResponse(w, http.StatusBadRequest, ErrorResponse{Error: "query param to is mandatory"})
		return
	}

	to, err := time.Parse("2006-01-02", toParam[0])
	if err != nil {
		writeJSONResponse(w, http.StatusBadRequest, ErrorResponse{Error: errors.Wrap(err, "to must be a date").Error()})
		return
	}

	departments, ok := r.URL.Query()["departments"]

	if !ok || len(departments) == 0 {
		writeJSONResponse(w, http.StatusBadRequest, ErrorResponse{Error: "query param departments is mandatory"})
		return
	}

	data, err := h.storage.Reports(from, to, departments...)
	if err != nil {
		writeJSONError(w, err)
		return
	}

	writeJSONResponse(w, http.StatusOK, data)
}

func (h *AnalyticsHandler) NationalReportsHandler(w http.ResponseWriter, r *http.Request) {
	fromParam, ok := r.URL.Query()["from"]
	if !ok || len(fromParam) != 1 {
		writeJSONResponse(w, http.StatusBadRequest, ErrorResponse{Error: "query param from is mandatory"})
		return
	}

	from, err := time.Parse("2006-01-02", fromParam[0])
	if err != nil {
		writeJSONResponse(w, http.StatusBadRequest, ErrorResponse{Error: errors.Wrap(err, "from must be a date").Error()})
		return
	}

	toParam, ok := r.URL.Query()["to"]
	if !ok || len(toParam) != 1 {
		writeJSONResponse(w, http.StatusBadRequest, ErrorResponse{Error: "query param to is mandatory"})
		return
	}

	to, err := time.Parse("2006-01-02", toParam[0])
	if err != nil {
		writeJSONResponse(w, http.StatusBadRequest, ErrorResponse{Error: errors.Wrap(err, "to must be a date").Error()})
		return
	}

	reports, err := h.storage.NationalReports(from, to)
	if err != nil {
		writeJSONError(w, err)
		return
	}

	writeJSONResponse(w, http.StatusOK, reports)
}

func (h *AnalyticsHandler) DepartmentResumeHandler(w http.ResponseWriter, r *http.Request) {
	department, ok := r.URL.Query()["department"]

	if !ok || len(department) != 1 {
		writeJSONResponse(w, http.StatusBadRequest, ErrorResponse{Error: "query param department is mandatory"})
		return
	}

	data, err := h.storage.DepartmentResume(department[0])
	if err != nil {
		writeJSONError(w, err)
		return
	}

	writeJSONResponse(w, http.StatusOK, data)
}

func (h *AnalyticsHandler) DailyTop5Handler(w http.ResponseWriter, r *http.Request) {
	dayParam, ok := r.URL.Query()["day"]
	if !ok || len(dayParam) != 1 {
		writeJSONResponse(w, http.StatusBadRequest, ErrorResponse{Error: "query param day is mandatory"})
		return
	}

	day, err := time.Parse("2006-01-02", dayParam[0])
	if err != nil {
		writeJSONResponse(w, http.StatusBadRequest, ErrorResponse{Error: errors.Wrap(err, "day must be a date").Error()})
		return
	}

	top5, err := h.storage.DailyTop5(day)
	if err != nil {
		writeJSONError(w, err)
		return
	}

	writeJSONResponse(w, http.StatusOK, top5)
}

func writeJSONResponse(w http.ResponseWriter, status int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

func writeJSONError(w http.ResponseWriter, err error) {
	writeJSONResponse(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
}
