docker exec -i postgres psql -U postgres -c "CREATE USER secret_link_test WITH PASSWORD 'Km61HJgJbBjNA0FdABpjDmQxEz008PHAQMA8TLpUbnlaKN7U8G1bQGHk0wsm'"
docker exec -i postgres psql -U postgres -c "CREATE DATABASE secret_link_test WITH ENCODING='UTF-8' OWNER secret_link_test"
