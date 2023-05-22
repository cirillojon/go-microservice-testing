# build
docker build -t cirillojon/calculation-service .

# run
docker run -e MONGO_USER=username -e MONGO_PASSWORD=password -e MONGO_DB_NAME=database -p localPort:8080 cirillojon/calculation-service

# note
Remember to use the corresponding escape character for special characters in password 