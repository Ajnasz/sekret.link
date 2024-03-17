# Create a new entry
POST {{api_host}}/api/?maxReads=3
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


# Generate a new key for the entry
GET {{api_host}}/api/key/{{entry_uuid}}/{{entry_key}}

HTTP 200
[Captures]
entry_key2: header "x-entry-key"
entry_expire2: header "x-entry-expire"

GET {{api_host}}/api/key/{{entry_uuid}}/{{entry_key}}

HTTP 200
[Captures]
entry_key3: header "x-entry-key"
entry_expire3: header "x-entry-expire"

# Retrieve the entry with key 2
GET {{api_host}}/api/{{entry_uuid}}/{{entry_key2}}

HTTP 200
[Asserts]
{
    "name": "John Doe",
    "email": "john.do@acheron.space"
}

# Retrieve the entry with key 3
GET {{api_host}}/api/{{entry_uuid}}/{{entry_key3}}

HTTP 200
[Asserts]
{
    "name": "John Doe",
    "email": "john.do@acheron.space"
}

# # Should not be able to retrieve the entry again
# GET {{api_host}}/api/{{entry_uuid}}/{{entry_key2}}
# 
# HTTP 404
