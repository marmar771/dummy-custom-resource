# dummy-operator
A simple dummy kubernetes operator

## Description
This operator creates a new pod for each API object, when object is deleted, the pod is also deleted. This operator also logs the Pods status in the yaml status.podStatus value.
In addition the operator copies the value from spec.message to status.specEcho and logs the name, namespace and spec.message.

## Getting Started
You’ll need a Kubernetes cluster to run against. You can use MINIKUBE.

### Running on the cluster
1. Pull the image from DockerHub:

```sh
docker pull marmar771/dummy-operator:v1.0.0
```

2. Run the docker image. 
If you want to check the name of the previously pulled image, please run this command:

```sh
docker image list --all
```
To run the image, please exec this command:

```sh
docker run <image-name>
```

3. Apply the CRD:

```sh
kubectl apply -f config/crd/bases/cache.interview.com_dummies.yaml
```

4. Create API object:

```sh
kubectl create -f config/samples/cache_v1alpha1_dummy.yaml
```

4. Check if pod is created by exec this command:

```sh
kubectl get pods
```

5. Check if the status.specEcho and status.podStatus are modified. When pod's status will be running, the status.podStatus should be "Running".

```sh
kubectl get dummy.cache.interview.com/dummy -o yaml
```

5. Delete the API object and check if the pod will get terminated:

```sh
kubectl delete -f config/samples/cache_v1alpha1_dummy.yaml
kubectl get pods
```

## License

Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

