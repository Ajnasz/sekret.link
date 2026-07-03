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
[Asserts]
header "Cache-Control" == "no-store"
header "X-Content-Type-Options" == "nosniff"
header "Referrer-Policy" == "no-referrer"


# Baseline security headers must also be present on reads
GET {{api_host}}/api/{{entry_uuid}}/{{entry_key}}

HTTP 200
[Asserts]
header "Cache-Control" == "no-store"
header "X-Content-Type-Options" == "nosniff"
header "Referrer-Policy" == "no-referrer"
