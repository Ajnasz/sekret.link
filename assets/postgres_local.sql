docker exec -i postgres psql -U postgres -c "CREATE USER secret_link WITH PASSWORD 'tOliZrYVVz3Dtjl97oo3YaaDsVVCVzzCtfgaciB6lXDAfz5tf0gIWJr4luKR'"
docker exec -i postgres psql -U postgres -c "CREATE DATABASE secret_link WITH ENCODING='UTF-8' OWNER secret_link"
