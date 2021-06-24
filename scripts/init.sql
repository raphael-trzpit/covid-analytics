CREATE TABLE IF NOT EXISTS department_daily_report (
	department_number VARCHAR(255) NOT NULL,
	day DATE NOT NULL,
	age_category INT NOT NULL,
	tests_total INT NOT NULL,
	tests_positives INT NOT NULL,
	population INT NOT NULL,
	PRIMARY KEY (department_number, day, age_category)
);

CREATE INDEX department_daily_report_department_number ON department_daily_report (department_number);
CREATE INDEX department_daily_report_day ON department_daily_report (day);
CREATE INDEX department_daily_report_age_category ON department_daily_report (age_category);
