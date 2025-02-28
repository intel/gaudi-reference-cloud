#################################### MANUAL DEPLOYMENT ###########################
###################################For Automation Use Jenkins ####################
# development Image
docker build -f "deployment/Dockerfile" --no-cache --force-rm --progress=plain -t internal-placeholder.com/intelcloud/idcs-admin:dev01 ..

docker push internal-placeholder.com/intelcloud/idcs-admin:dev01
#REM Push to K9s
kubectl delete -f manualdeployment.yaml -n console 
kubectl apply -f manualdeployment.yaml -n console
