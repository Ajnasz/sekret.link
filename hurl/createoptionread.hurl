# Create a new entry
POST {{api_host}}/api/
content-type: application/json
{
    "name": "John Doe",
    "email": "john.do@acheron.space"
}

HTTP 200
[Captures]
entry_uuid: header "x-entry-uuid"
entry_key: header "x-entry-key"
entry_expire: header "x-entry-expire"
entry_delete_key: header "x-entry-delete-key"


# Check options response
OPTIONS {{api_host}}/api/{{entry_uuid}}/{{entry_key}}
Origin: https://acheron.space

HTTP 200
[Asserts]
header "Access-Control-Allow-Origin" == "https://acheron.space"
header "Access-Control-Allow-Methods" == "POST, GET, OPTIONS, DELETE"
header "Access-Control-Allow-Headers" == "Accept, Content-Type, Content-Length, Accept-Encoding, x-entry-uuid, x-entry-key, x-entry-delete-key, x-entry-expire"

# Retrieve the entry
GET {{api_host}}/api/{{entry_uuid}}/{{entry_key}}

HTTP 200
[Asserts]
{
    "name": "John Doe",
    "email": "john.do@acheron.space"
}

