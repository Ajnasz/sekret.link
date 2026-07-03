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


# A JSON read must not leak the delete key to the reader
GET {{api_host}}/api/{{entry_uuid}}/{{entry_key}}
Accept: application/json

HTTP 200
[Asserts]
jsonpath "$.DeleteKey" not exists
