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
    "username": "username",
    "password": "password"
}
```
response -
```json
{
    "token": "a.b.c",
    "scheme": "Bearer",
    "expires_at": "28-07-2010T12:17:27+09"
}
```

### /auth/google:

body - 
```json
{
    "token": "aa.bb.cc",
}
```
response -
```json
{
    "token": "a.b.c",
    "scheme": "Bearer",
    "expires_at": "28-07-2010T12:17:27+09"
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

[![Go reference](https://pkg.go.dev/badge/google.golang.org/api@v0.93.0.svg)](https://pkg.go.dev/google.golang.org/api@v0.93.0)
``google/apo``, for validating google sign-in's token 

[![Go reference](https://pkg.go.dev/badge/github.com/google/uuid@v1.3.0.svg)](https://pkg.go.dev/github.com/google/uuid@v1.3.0)
``google/uuid``, for session id's uuid generator

[![Go Reference](https://pkg.go.dev/badge/github.com/lib/pq@v1.10.6.svg)](https://pkg.go.dev/github.com//lib/pq@v1.10.6)
``pq``, for connecting to PostgreSQL server 

[![Go reference](https://pkg.go.dev/badge/github.com/simukti/sqldb-logger@v0.0.0-20220521163925-faf2f2be0eb6.svg)](https://pkg.go.dev/github.com/simukti/sqldb-logger@v0.0.0-20220521163925-faf2f2be0eb6)
``sqldb-logger``, for SQL logging

[![Go reference](https://pkg.go.dev/badge/github.com/rs/zerolog@v1.26.1.svg)](https://pkg.go.dev/github.com/rs/zerolog@v1.26.1)
``zerolog``, for JSON structured logger

### Testing

[![Go reference](https://pkg.go.dev/badge/github.com/stretchr/testify@v1.8.0.svg)](https://pkg.go.dev/github.com/stretchr/testify@v1.8.0)
``testify``, for cleaner unit testing assertions
