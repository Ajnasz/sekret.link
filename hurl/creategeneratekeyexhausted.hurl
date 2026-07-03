# Create a one-time-read entry
POST {{api_host}}/api/?maxReads=1
content-type: application/json
{
    "name": "John Doe",
    "email": "john.do@acheron.space"
}

HTTP 200
[Captures]
entry_uuid: header "x-entry-uuid"
entry_key: header "x-entry-key"


# Consume the single allowed read
GET {{api_host}}/api/{{entry_uuid}}/{{entry_key}}

HTTP 200
[Asserts]
{
    "name": "John Doe",
    "email": "john.do@acheron.space"
}


# Minting a fresh access key for an exhausted entry key must fail,
# not silently grant a new read/expiry budget
GET {{api_host}}/api/key/{{entry_uuid}}/{{entry_key}}?expire=60s&maxReads=5

HTTP 500
