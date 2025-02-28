#################################### MANUAL DEPLOYMENT ###########################
###################################For Automation Use Jenkins ####################
# development Image
docker build -f "deployment/Dockerfile" --force-rm --progress=plain -t internal-placeholder.com/intelcloud/idcs-portal:dev01 ..

docker push internal-placeholder.com/intelcloud/idcs-portal:dev01
#REM Push to K9s
kubectl delete -f DeployDev.yaml -n reuse 
kubectl apply -f DeployDev.yaml -n reuse
