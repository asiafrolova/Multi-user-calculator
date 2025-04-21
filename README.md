# Calculate Service

Service to parse and calculate simple arithmetic expressions.

## Prerequisites
- Installed Go version 1.23 or higher
- curl, Postman, or any similar app to work with HTTP API   

## Installation
1. Clone the repository to your machine `git clone https://github.com/asiafrolova/Final_task.git`
2. Navigate into the project's directory `cd Final_task/orkestrator_service`
3. Download all dependencies `go mod tidy`
4. Run app `go run ./cmd/main.go`
5. Repeat everything for the agent
   
**Fast start**
```
git clone https://github.com/asiafrolova/Multi-user-calculator.git
cd Multi-user-calculator/backend/orkestrator_service
go mod tidy
go run ./cmd/main.go
```

In a separate terminal!

```
cd Multi-user-calculator/backend/agent_service
go mod tidy
go run ./cmd/main.go

```
## How to use
**Possibilities**
- addition `+`, subtract `-`, multiplicity `*`, divide `/` operations
- any complex nested parentheses with `(` and `)`
- int and float numbers (I hope within the range -1e308..1e308) with `.` as decimal separator ()
- unary minus `-` (regular minus sign) for numbers and parentheses group's
**Examples**
*Registration*
```
curl --location 'localhost:8080/api/v1/register' \
--header 'Content-Type: application/json' \
--data '{
  "login": "<user name>",
  "password":"<user password>"
}'
```
-Successfully
`{
    "login":"<user name>"
    "id":"<user id>"
}`

*Login*
```
curl --location 'localhost:8080/api/v1/login' \
--header 'Content-Type: application/json' \
--data '{
  "login": "<user name>",
  "password":"<user password>"
}'
```
-Successfully
`{
    "access-token":"<token>"
}`

*POST expression*
```
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer <token>' \
--data '{
  "expression": "(3.14*2)"
}'
```
-Successfully
`{
    "id": "1"
}
`

*GET expression by id*
```
curl --location 'localhost:8080/api/v1/expressions/1' \
--header 'Authorization: Bearer <token>' 
```
-Successfully
`{
    "expression":
        {
            "id": "1",
            "status": "Completed",
            "result": 6.28
        }
}`

*GET list expressions*
```
curl --location 'localhost:8080/api/v1/expressions' \
--header 'Authorization: Bearer <token>' 
```
-Successfully
`{
    "expressions": [
        {
            "id": "1",
            "status": "Completed",
            "result": 6.28
        },
        {
            "id": "2",
            "status": "Todo",
            "result": 0
        }
    ]
}
`

*DELETE expression by id*
```
curl --location 'localhost:8080/api/v1/delete/expressions/1' \
--header 'Authorization: Bearer <token>' 
```
-Successfully
`{
    {
    "count": 1,
    "error": ""
}
}`

*DELETE all user's expressions*

```
curl --location 'localhost:8080/api/v1/delete/expressions' \
--header 'Authorization: Bearer <token>' 
```
-Successfully
`{
    {
    "count": 3,
    "error": ""
}
}`

*DELETE user account*

```
curl --location 'localhost:8080/api/v1/delete/user' \
--header 'Authorization: Bearer <token>' 
```

-Successfully
`{
    "id": 1
}`



**Examples errors**

*Name or password is incorrect*

curl --location 'localhost:8080/api/v1/login' \
--header 'Content-Type: application/json' \
--data '{
  "login": "<user name>",
  "password":"<uncorrect password>"
}'

*Invalid expression*

`curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer <token>' \
--data '{
  "expression": "((32*2)"
}'`

`curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer <token>' \
--data '{
  "expression": "3a-1"
}'`

`curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer <token>' \
--data '{
  "expression": "3++1"
}'`

*Failed status*

`curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer <token>' \
--data '{
  "expression": "1/0"
}'`


`curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer <token>' \
--data '{
  "expression": "3.0.0-1"
}'`

**Frontend**

Open "frontend" folder, open "index.html" file in any browser



   
