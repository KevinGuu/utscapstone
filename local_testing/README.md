# Create a Kind cluster for local testing

## Create Kind cluster 
`kind create cluster --config kind-cluster.yaml`
`kubectl cluster-info --context kind-kind`

## Install Calico
`kubectl create -f tigera-operator.yaml`
`kubectl craete -f calico.yaml`

## Install Kube Dashboard
`kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v2.5.0/aio/deploy/recommended.yaml`
`kubectl -n kubernetes-dashboard describe secret $(kubectl -n kubernetes-dashboard get secret | grep admin-user-token | awk '{print $1}')`