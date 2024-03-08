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


# Delete an entry
DELETE {{api_host}}/api/{{entry_uuid}}/{{entry_key}}/{{entry_delete_key}}

HTTP 202


# Get an entry
GET {{api_host}}/api/{{entry_uuid}}/{{entry_key}}

HTTP 404
