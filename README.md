# eLibrary Backend

## By: Reza Hadi Fairuztama

An implementation of the backend for the eLibrary project

## Installation

The easiest way to deploy the API from the ground up is to use:
```cmd
docker-compose up
```

## Usage

### /auth/login:

body - 
```json
{
    "email": "username@example.co.id",
    "password": "P4sswordꦏꦤ꧀"
}
```
response - 200 OK
```json
{
    "token": "a.b.c",
    "scheme": "Bearer",
    "expires_at": "2010-07-28T12:54:27+09:00"
}
```
response - 400 Bad Request; 401 Unauthorized Request; 422 Validation Failed
```json
{
    "error_type": "Bad Request",
    "message": "error-message"
}
```

### /auth/google:

body - 
```json
{
    "token": "aa.bb.cc",
}
```
response - 200 OK
```json
{
    "token": "a.b.c",
    "scheme": "Bearer",
    "expires_at": "2010-07-28T12:54:27+09:00"
}
```
response - 400 Bad Request; 401 Unauthorized Request; 422 Validation Failed
```json
{
    "error_type": "Bad Request",
    "message": "error-message"
}
```

### /auth/register:

body - 
```json
{
    "email": "username@example.co.id",
    "name": "Joko",
    "password": "P4sswordꦏꦤ꧀"
}
```
response - 200 OK
```json
{
    "new_id": "username@example.co.id"
}
```
response - 400 Bad Request
```json
{
    "error_type": "Bad Request",
    "message": "error-message"
}
```

### /auth/register/google:

body - 
```json
{
    "token": "aa.bb.cc",
}
```
response - 200 OK
```json
{
    "new_id": "username@example.co.id"
}
```
response - 400 Bad Request; 422 Validation Failed
```json
{
    "error_type": "Bad Request",
    "message": "error-message"
}
```

### /books?criteria=new:

header -
```
Authorization: Bearer ...
```
response - 200 OK
```json
{
    "data": [
        {
            "id": "01234567-89ab-cdef-0123-456789abcdef",
            "title": "Chäos;HEĀd",
            "author": "Hayashi Naotaka",
            "cover_url": "https://s2.vndb.org/cv/59/49759.jpg",
            "readers": 0,
            "rating": 3.5,
            "is_favorite": false
        },
    ]
}
```
response - 422 Validation Failed
```json
{
    "error_type": "Validation Failed",
    "message": "error-message"
}
```

### /books?criteria=popular:

header -
```
Authorization: Bearer ...
```
response - 200 OK
```json
{
    "data": [
        {
            "id": "01234567-89ab-cdef-0123-456789abcdef",
            "title": "Chäos;HEĀd",
            "author": "Hayashi Naotaka",
            "cover_url": "https://s2.vndb.org/cv/59/49759.jpg",
            "readers": 0,
            "rating": 3.5,
            "is_favorite": false
        },
    ]
}
```
response - 422 Validation Failed
```json
{
    "error_type": "Validation Failed",
    "message": "error-message"
}
```

### /books?criteria=newHomepage&page=0:

header -
```
Authorization: Bearer ...
```
response - 200 OK
```json
{
    "data": [
        {
            "id": "01234567-89ab-cdef-0123-456789abcdef",
            "title": "Chäos;HEĀd",
            "author": "Hayashi Naotaka",
            "cover_url": "https://s2.vndb.org/cv/59/49759.jpg",
            "readers": 0,
            "rating": 3.5,
            "is_favorite": false
        },
    ]
}
```
response - 422 Validation Failed
```json
{
    "error_type": "Validation Failed",
    "message": "error-message"
}
```

### /books?criteria=popularHomepage&page=0:

header -
```
Authorization: Bearer ...
```
response - 200 OK
```json
{
    "data": [
        {
            "id": "01234567-89ab-cdef-0123-456789abcdef",
            "title": "Chäos;HEĀd",
            "author": "Hayashi Naotaka",
            "cover_url": "https://s2.vndb.org/cv/59/49759.jpg",
            "readers": 0,
            "rating": 3.5,
            "is_favorite": false
        },
    ]
}
```
response - 422 Validation Failed
```json
{
    "error_type": "Validation Failed",
    "message": "error-message"
}
```

### /books?criteria=search&query=...:

header -
```
Authorization: Bearer ...
```
query -
```
query= text
```
response - 200 OK
```json
{
    "data": [
        {
            "id": "01234567-89ab-cdef-0123-456789abcdef",
            "title": "Chäos;HEĀd",
            "author": "Hayashi Naotaka",
            "cover_url": "https://s2.vndb.org/cv/59/49759.jpg",
            "readers": 0,
            "rating": 3.5,
            "is_favorite": false
        },
    ]
}
```
response - 400 Bad Request; 422 Validation Failed
```json
{
    "error_type": "Bad Request",
    "message": "error-message"
}
```

## Modules used:

[![Go Reference](https://pkg.go.dev/badge/github.com/go-chi/chi/v5@v5.0.7.svg)](https://pkg.go.dev/github.com/go-chi/chi/v5@v5.0.7)
``chi``, for REST handling

[![Go Reference](https://pkg.go.dev/badge/github.com/go-chi/render@v1.0.1.svg)](https://pkg.go.dev/github.com/go-chi/render@v1.0.1)
``chi/render``, for parsing HTTP request and response bodies

[![Go Reference](https://pkg.go.dev/badge/github.com/go-chi/jwtauth@v1.2.0.svg)](https://pkg.go.dev/github.com/go-chi/jwtauth@v1.2.0)
``chi/jwtauth`` module, for authentication via JWT tokens

[![Go reference](https://pkg.go.dev/badge/golang.org/x/crypto@v0.0.0-20220722155217-630584e8d5aa.svg)](https://pkg.go.dev/golang.org/x/crypto@v0.0.0-20220722155217-630584e8d5aa)
``crypto``, for password hashing via ``bcrypt``

[![Go reference](https://pkg.go.dev/badge/github.com/joho/godotenv@v1.4.0.svg)](https://pkg.go.dev/github.com/joho/godotenv@v1.4.0)
``godotenv``, for enviroment variable loading

[![Go reference](https://pkg.go.dev/badge/gopkg.in/gomail.v2@v2.0.0-20160411212932-81ebce5c23df.svg)](https://pkg.go.dev/gopkg.in/gomail.v2@v2.0.0-20160411212932-81ebce5c23df)
``gomail``, for sending account activation email

[![Go reference](https://pkg.go.dev/badge/google.golang.org/api@v0.93.0.svg)](https://pkg.go.dev/google.golang.org/api@v0.93.0)
``google/apo``, for validating google sign-in's token 

[![Go reference](https://pkg.go.dev/badge/github.com/google/uuid@v1.3.0.svg)](https://pkg.go.dev/github.com/google/uuid@v1.3.0)
``google/uuid``, for session id's uuid generator

[![Go Reference](https://pkg.go.dev/badge/github.com/lib/pq@v1.10.6.svg)](https://pkg.go.dev/github.com//lib/pq@v1.10.6)
``pq``, for connecting to PostgreSQL server 

[![Go Reference](https://pkg.go.dev/badge/github.com/sethvargo/go-envconfig@v0.8.2.svg)](https://pkg.go.dev/github.com/sethvargo/go-envconfig@v0.8.2)
``sethvargo/go-envconfig``, for structured env vars

[![Go reference](https://pkg.go.dev/badge/github.com/simukti/sqldb-logger@v0.0.0-20220521163925-faf2f2be0eb6.svg)](https://pkg.go.dev/github.com/simukti/sqldb-logger@v0.0.0-20220521163925-faf2f2be0eb6)
``sqldb-logger``, for SQL logging

[![Go reference](https://pkg.go.dev/badge/github.com/rs/zerolog@v1.26.1.svg)](https://pkg.go.dev/github.com/rs/zerolog@v1.26.1)
``zerolog``, for JSON structured logger

### Testing

[![Go reference](https://pkg.go.dev/badge/github.com/stretchr/testify@v1.8.0.svg)](https://pkg.go.dev/github.com/stretchr/testify@v1.8.0)
``testify``, for cleaner unit testing assertions

[![Go reference](https://pkg.go.dev/badge/github.com/DATA-DOG/go-sqlmock@v1.5.0.svg)](https://pkg.go.dev/github.com/DATA-DOG/go-sqlmock@v1.5.0)
``go-sqlmock``, for mocking an ``*sql.SQL`` object in unit testing
