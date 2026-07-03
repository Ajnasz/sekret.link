# Create a new entry with a disallowed Content-Type
POST {{api_host}}/api/
Content-Type: text/html
```
<script>alert(1)</script>
```

HTTP 200
[Captures]
entry_uuid: header "x-entry-uuid"
entry_key: header "x-entry-key"


# The disallowed Content-Type must be coerced, not echoed back verbatim
GET {{api_host}}/api/{{entry_uuid}}/{{entry_key}}

HTTP 200
[Asserts]
header "Content-Type" == "application/octet-stream"
body == "<script>alert(1)</script>\n"
