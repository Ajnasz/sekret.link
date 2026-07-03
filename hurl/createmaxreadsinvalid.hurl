# maxReads above the SMALLINT range must be rejected at parse time,
# not surface as a 500 from a database INSERT overflow
POST {{api_host}}/api/?maxReads=40000
content-type: application/json
{
    "name": "John Doe",
    "email": "john.do@acheron.space"
}

HTTP 400
