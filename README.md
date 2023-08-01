# Go job search

### Built in Go 1.20

### The app uses:
- Postgres
- Docker
- [Gin](https://github.com/gin-gonic/gin)
- [Elasticsearch](https://github.com/elastic/go-elasticsearch)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [sqlc](https://github.com/kyleconroy/sqlc)
- [testify](https://github.com/stretchr/testify)
- [PASETO Security Tokens](github.com/o1egl/paseto)
- [Viper](https://github.com/spf13/viper)
- [gin-swagger](https://github.com/swaggo/gin-swagger) for generating docs

<hr>

## Getting started
1. Clone the repository
2. Go to the project's root directory
3. Rename `app.env.sample` to `app.env` and replace the values
4. Run in your terminal:
     - `docker-compose up` to run the database container
     - `make migrate_up` - to run migrations
     - `make runserver` - to run HTTP server
5. Now everything should be ready and server running on `SERVER_ADDRESS` specified in `app.env`
<hr>

## Testing
1. Run the postgres container (`docker-compose up`)
2. Run in your terminal:
    - `make test` to run all tests
   or
    - `make test_coverage p={PATH}` - to get the coverage in the HTML format - where `{PATH}` is the path to the target directory for which you want to generate test coverage. The `{PATH}` should be replaced with the actual path you want to use. For example `./api`
   or
    - use standard `go test` commands (e.g. `go test -v ./api`)
<hr>

## API endpoints
This API provides a set of endpoints for managing:
- users
- employers
- jobs

After running the server, the Swagger documentation is available at http://localhost:8080/swagger/index.html. 
You can find there detailed information about the API endpoints, including their parameters, 
request and response formats, and examples. You can use the Swagger UI to test the API 
endpoints and see their responses in real-time.

### The base path for all endpoints is `/api/v1`
so for example `/api/v1/users/login`


Here is a summary of the available endpoints and their functionality:


### Users

+ `POST /users`: This endpoint creates a new user. The request body must contain the user details 
in JSON format. On success, the response has a `201 Created` status code and returns the created 
user in JSON format. If the request body is invalid, a `400 Bad Request` status code is returned. 
If a user with the given email already exists, a `403 Forbidden` status code is returned. In case 
of any other error, a `500 Internal Server Error` status code is returned.

+ `POST /users/login`: This endpoint logs in a user. The request body must contain the user credentials
(email, password) in JSON format. On success, the response has a `200 OK` status code and returns 
an access token and the authenticated user in JSON format. If the request body is invalid, a 
`400 Bad Request` status code is returned. If the password is incorrect, a `401 Unauthorized` 
status code is returned. If a user with the given email does not exist, a `404 Not Found` status 
code is returned. In case of any other error, a `500 Internal Server Error` status code is returned.

+ `GET /users`: This endpoint retrieves the details of the logged-in user. On success, the response 
has a `200 OK` status code and returns the user details in JSON format. In case of any error, a 
`500 Internal Server Error` status code is returned.

+ `PATCH /users`: This endpoint updates the details of the logged-in user. The request body must 
contain the updated user details in JSON format. On success, the response has a `200 OK` status code 
and returns the updated user in JSON format. If the request body is invalid, a `400 Bad Request` 
status code is returned. In case of any other error, a `500 Internal Server Error` status code is returned.

+ `PATCH /users/password`: This endpoint updates the password of the logged-in user. The request body 
must contain the old and new password in JSON format. On success, the response has a `200 OK` status 
code and returns a success message. If the request body is invalid, a `400 Bad Request` status code 
is returned. If the old password is incorrect, a `401 Unauthorized` status code is returned. In case 
of any other error, a `500 Internal Server Error` status code is returned.

+ `DELETE /users`: This endpoint deletes the logged-in user. On success, the response has a 
`204 No Content` status code. In case of any error, a `500 Internal Server Error` status code is returned.


### Employers

+ `POST /employers`: This endpoint creates a new employer. The request body
must contain the employer and company details in JSON format. 
On success, the response has a `201 Created` status code and returns 
the created employer in JSON format. If the request body is invalid, 
a `400` status code is returned. If a company with the given name or an 
employer with the given email already exists, a `403 Forbidden` status 
code is returned. In case of any other error, a 500 Internal Error status code is returned.

+ `POST /employers/login`: This endpoint logs in an employer. The request body 
must contain the employer credentials (email, password) in JSON format. On success, 
the response has a `200 OK` status code and returns an access token and the authenticated employer 
in JSON format. If the request body is invalid, a `400 Bad Request` status code is returned. 
If the password is incorrect, a `401 Unauthorized` status code is returned. 
If an employer with the given email or a company with the given id does not 
exist, a `404 Not Found` status code is returned. In case of any other error, 
a `500 Internal Server Error` status code is returned.

+ `GET /employers`: This endpoint retrieves the details of the 
authenticated employer. The response is in JSON format and has a `200 OK` 
status code on success. In case of an internal error, a `500 Internal Server Error` 
status code is returned.

+ `PATCH /employers`: This endpoint updates the details of the 
authenticated employer. The request body must contain the updated 
employer details in JSON format. On success, the response has a `200 OK` 
status code and returns the updated employer in JSON format. In case of 
any error, a `500 Internal Server Error` status code is returned.

+ `PATCH /employers/password`: This endpoint updates the password of the logged-in 
employer. The request body must contain the old and new password in JSON format.
On success, the response has a `200 OK` status code and returns a success message. 
If the request body is invalid, a `400 Bad Request` status code is returned. 
If the old password is incorrect, a `401 Unauthorized` status code is returned. 
In case of any other error, a `500 Internal Server Error` status code is returned.

+ `DELETE /employers`: This endpoint deletes the logged-in employer. 
On success, the response has a `204 No Content` status code. In case of 
any error, a `500 Internal Server Error` status code is returned.


### Jobs

+ `POST /jobs`: This endpoint creates a new job. The request body must contain 
the job details in JSON format. On success, the response has a `201 Created` status 
code and returns the created job in JSON format. If the request body is invalid, 
a `400 Bad Request` status code is returned. In case of any other error, a 
`500 Internal Server Error `status code is returned.

+ `GET /jobs/search`: This endpoint searches for jobs with elasticsearch. 
The request must contain the `page`, `page_size`, and `search` parameters in the 
query. On success, the response has a `200 OK` status code and returns an array 
of jobs that match the search query in JSON format. If the query is invalid, a 
`400 Bad Request` status code is returned. In case of any other error, a `500 Internal Server Error` status code is returned.

+ `GET /jobs`: This endpoint filters and lists jobs based on the provided query 
parameters. The `page` and `page_size` query parameters are required and specify
the page number and page size, respectively. The `title`, `industry`, `job_location`, 
`salary_min`, and `salary_max` query parameters are optional and can be used to 
filter the jobs by title, industry, location, and salary range, respectively. 
On success, the response has a `200 OK` status code and returns a list of jobs 
in JSON format. If the query is invalid, a `400` status code is returned. 
In case of any other error, a `500 Internal Server Error` status code is returned.

+ `GET /jobs/company`: This endpoint lists jobs by company name, id, or part 
of the name. The `page` and `page_size` query parameters are required and specify 
the page number and page size, respectively. The `id`, `name`, and `name_contains` 
query parameters are optional (one of them has to be provided) and can be used 
to filter the jobs by company id, exact company name, or part of the company name, 
respectively. Only one of these three parameters is allowed in a single request. 
On success, the response has a `200 OK` status code and returns a list of jobs 
in JSON format. If the query is invalid, a `400 Bad Request` status code is returned. 
In case of any other error, a `500 Internal Server Error` status code is returned.

+ `GET /jobs/{id}`: This endpoint retrieves the details of the job with the given id. 
The id path parameter is required and specifies the id of the job to retrieve. On success, 
the response has a `200 OK` status code and returns the job details in JSON format. If the 
request query is invalid, a `400 Bad Request` code is returned. If the job with the given id is 
not found, a `404 Not Found` status code is returned. In case of any other error, a 
`500 Internal Server Error` status code is returned.

+ `PATCH /jobs/{id}`: This endpoint updates the job with the given id. The id path parameter is required 
and specifies the id of the job to update. The request body must contain the updated job details 
in JSON format. On success, the response has a `200 OK` status code and returns the updated job in 
JSON format. In case of any error, a `500 Internal Server Error` status code is returned.

+ `DELETE /jobs/{id}`: This endpoint deletes the job with the given id. The id path parameter is 
required and specifies the id of the job to delete. On success, the response has a `204 No Content` 
status code. In case the job is not found, returns `404 Not Found`, in case of any other error, a `500 Internal Server Error` status code is returned.

