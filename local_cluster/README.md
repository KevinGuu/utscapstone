# Create a Kind cluster for local testing

## Create Kind cluster 
1. Create Kind cluster with config file.
`kind create cluster --config kind-cluster.yaml`

2. Check if cluster has been created successfully.
`kubectl cluster-info --context kind-kind`

## Install Calico
1. Install required CRD
`kubectl create -f https://projectcalico.docs.tigera.io/manifests/tigera-operator.yaml`
2. Intall Calico
`kubectl create -f calico.yaml`

## Install Kube Dashboard
1. Deploy the dashboard to the cluster
`kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v2.5.0/aio/deploy/recommended.yaml`
2. Create an admin user RBAC to authenticate into the dashboard
`kubectl apply -f rbac-dashboard.yaml`
3. Get the authentication token
`kubectl -n kubernetes-dashboard describe secret $(kubectl -n kubernetes-dashboard get secret | grep admin-user-token | awk '{print $1}')`
4. Dashboard will be available at this URl, use the token to login
http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/#/login

