docker exec -i postgres psql -U postgres -c "CREATE USER sekret_link_test WITH PASSWORD 'Km61HJgJbBjNA0FdABpjDmQxEz008PHAQMA8TLpUbnlaKN7U8G1bQGHk0wsm'"
docker exec -i postgres psql -U postgres -c "CREATE DATABASE sekret_link_test WITH ENCODING='UTF-8' OWNER sekret_link_test"
