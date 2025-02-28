# Building a local inference proxy image

* NOTE: If you want to push the new image to MaaS docker account please reach out to MaaS team to get credentials.
* NOTE: This instruction assume you have a running docker and configured poetry. For more info about poetry configuration see main repo ```README.MD```

1. Make sure that ```requirements.txt``` file is up to date, or run:
```bash
poetry export -f requirements.txt --output requirements.txt --without-hashes
```

2. Build a proxy-dev image:
```bash
docker build --no-cache --platform linux/amd64 -t proxy-dev -f src/inference_engine/Dockerfile --build-arg http_proxy=$HTTP_PROXY --build-arg https_proxy=$HTTPS_PROXY --build-arg no_proxy=$NO_PROXY .
```

3. Tag the image & push it to MaaS docker hub:
```bash
docker tag proxy-dev:latest cnvrg/infaas-inference:proxy-dev
docker push cnvrg/infaas-inference:proxy-dev
```

4. Test it using helm install dev:
```bash
helm install infaas-inference-llama-3-1-8b ./infaas-inference -n idcs-system --set region=$regionName --set deployedModel="meta-llama/Meta-Llama-3.1-8B-Instruct" --set image.agent.pullPolicy=Always -f ./infaas-inference/values-dev.yaml
```

5. If installation works - tag is as ```proxy``` & push
```bash
docker tag proxy-dev:latest cnvrg/infaas-inference:proxy
docker push cnvrg/infaas-inference:proxy
```

* NOTE: ```cnvrg/infaas-inference:proxy``` is a production image. Push only if you sure the image is working properly.
