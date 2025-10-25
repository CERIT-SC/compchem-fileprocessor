# Argo workflows

This directory contains all argo workflows related stuff like WorkflowTemplates and example Workflows.

### Access localhost from within the cluster

Below is a short guide on how to be able to connect to a local instance of the compchem repo running on port 5000 from within a kubernetes Pod. This is done because workflows run as Pods and they need to communicate with the repository. This is only done for the purposes of development, with a repo running directly on the dev machine.


Create a service, then endpoint to connect to docker
```json
apiVersion: v1
kind: Service
metadata:
  name: host-service
  namespace: argo
spec:
  clusterIP: None  # Headless service
  ports:
  - port: 5000
    targetPort: 5000

---
apiVersion: v1
kind: Endpoints
metadata:
  name: host-service
  namespace: argo
subsets:
- addresses:
  - ip: 172.17.0.1  # Your host IP as seen from the container
  ports:
  - port: 5000
```

On the machine use socat

sudo socat TCP-LISTEN:5000,bind=$(ip -4 addr show docker0 | grep -Po 'inet \K[\d.]+'),fork TCP:localhost:5000 &

Then from within a pod localhost:5000 is available on: host-service.argo.svc.cluster.local:5000

must use header -H "Host: localhost" in requests
