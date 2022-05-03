# Webhook deployment

## 1. Deploy `Secret` containing TLS cert/key, to be used by the webhook server
`kubectl apply -f secret-capstone-webhook-tls.yaml`

## 2. Deploy `service` fronting the webhook server
`kubectl apply -f service.yaml`

## 3. Deploy `RBAC components (SA, CR & CRB)` for the deployment
`kubectl apply -f rbac.yaml`

## 4. Deploy `Configmap` for the sidecar container
`kubectl apply -f configmap-sidecar.yaml`

## 5. Deploy `Deployment` (webhook server)
`kubectl apply -f deployment.yaml`

## 6. Deploy `MutatingWebhookConfiguration`
`kubectl apply -f webhook.yaml`