# build
docker build -t cirillojon/calculation-service .

# run
docker run -e MONGO_USER=jonathancirillo -e MONGO_PASSWORD=password%5E%5E -e MONGO_DB_NAME=Calculator -p 8081:8080 cirillojon/calculation-service

# push
docker push cirillojon/calculation-service

# kubernetes secret
kubectl create secret generic mongodb-secret --from-literal=mongo_user=jonathancirillo --from-literal=mongo_password=password%5E%5E --from-literal=mongo_db_name=Calculator

# kubernetes deployment
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml

# check kubernetes status
kubectl get deployments
kubectl get pods
kubectl get services

# kubernetes port forwarding
kubectl port-forward service/calculation-service 8081:8080
