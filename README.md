# Covid analytics

A go API importing covid data from datagouv and serving analytics handler.

## Dependencies

 - Go >= 1.16
 - A myslq database
 - Connected to internet (to gather data from datagouv)
 
You can use the docker-compose file to setup a local mysql database. It will automatically setup the database schema.
Just do `make run-dev` to start the database and run the API.

## Configuration

The main executable is cmd/api. It reads these environment variables:

Name | Description
-----|------------
DATAGOUV_URL | The url where we import the data from
DB_DSN | The mysql DSN to connect to the database (see [https://github.com/go-sql-driver/mysql#dsn-data-source-name](https://github.com/go-sql-driver/mysql#dsn-data-source-name))
HTTP_PORT | The HTTP port the server will be listening to

## Routes

### /import
Call this route to trigger an import. It will override the database.
This call takes some time!

It returns either an error response or a result: OK.

### /departments

This route returns the list of department codes for which we have data.

It returns an array of string

### /age_categories

This route returns the list of age categories for which we have data.

It returns an array of int

### /days

This route returns the date range for which we have data.

It returns:

```json
{
    "from": "2020-05-13T00:00:00Z",
    "to": "2021-06-20T00:00:00Z"
}
```

### /data

This route returns the data we have for a given date range.

QueryParams:
 - from: the start day for the date range filter (mandatory)
 - to: the end day for the date range filter (mandatory)
 - departments: an array of departments (mandatory)

You can use the same day for from and to to query data for a single day.

ie: /data?from=2021-06-01&to=2021-06-01&departments=01&departments=02 will query all data for the 1st june 2021 for the departments 01 and 02.

It returns:

```json
[
    {
        "department_number": "01",
        "day": "2021-06-01T00:00:00Z",
        "age_category": 0,
        "tests_total": 3181,
        "tests_positives": 72,
        "population": 656955
    },
   ...
]
```

### /national

This route returns data at a national level for a given date range.

QueryParams:
 - from: the start day for the date range filter (mandatory)
 - to: the end day for the date range filter (mandatory)

ie: /national?from=2021-06-01&to=2021-06-02

It returns:

```json
[
    {
        "day": "2021-06-01T00:00:00Z",
        "tests_total": 692716,
        "tests_positives": 15864,
        "ratio": 0.02290116007137124
    },
   ...
]
```


### /department

This route returns a summary for a department.

QueryParams:
 - department: the department code (mandatory)

It returns:
 
 ```json
{
    "day_with_most_tests": "2020-12-18T00:00:00Z",
    "day_with_most_positives": "2020-11-02T00:00:00Z",
    "day_with_highest_ratio": "2020-11-01T00:00:00Z"
}
```

### /daily_top5

This route returns the top5 department data per age category, for a given day.

QueryParams:
 - day: the day (mandatory)
 
 It returns:
 
 ```json
{
    "age_categories": {
        "0": {
            "top5": [
                {
                    "department_number": "90",
                    "day": "2021-01-06T00:00:00Z",
                    "age_category": 0,
                    "tests_total": 738,
                    "tests_positives": 97,
                    "population": 140145
                },
                ...
           ]
        },
        "9": ...
    }
 }
```
