# Generate TLS certificates using CFSSL

(With help from https://github.com/marcel-dempers/docker-development-youtube-series/blob/master/kubernetes/admissioncontrollers/introduction/tls/ssl_generate_self_signed.md)

```
1. cd into the cert working dir
cd utscapstone/tls_cert

2. start a debian docker container with a bash terminal that will auto clean up after the container exits, mount the parent dir(project dir)
docker run -it --rm -v $(dirname "${PWD}"):/work -w /work debian bash

apt-get update && apt-get install -y curl &&
curl -L https://github.com/cloudflare/cfssl/releases/download/v1.5.0/cfssl_1.5.0_linux_amd64 -o /usr/local/bin/cfssl && \
curl -L https://github.com/cloudflare/cfssl/releases/download/v1.5.0/cfssljson_1.5.0_linux_amd64 -o /usr/local/bin/cfssljson && \
chmod +x /usr/local/bin/cfssl && \
chmod +x /usr/local/bin/cfssljson

# generate CA in /tmp
cfssl gencert -initca ./tls_cert/ca-csr.json | cfssljson -bare /tmp/ca

# generate certificate in /tmp
cfssl gencert \
  -ca=/tmp/ca.pem \
  -ca-key=/tmp/ca-key.pem \
  -config=./tls_cert/ca-config.json \
  -hostname="capstone,capstone-webhook.default.svc.cluster.local,capstone-webhook.default.svc,localhost,127.0.0.1" \
  -profile=default \
  ./tls_cert/ca-csr.json | cfssljson -bare /tmp/capstone-webhook

# create a Kube secret storing the tls cert / key
cat <<EOF > ./kubernetes/manifests/webhook/secret-capstone-webhook-tls.yaml
apiVersion: v1
kind: Secret
metadata:
  name: capstone-webhook-tls
type: Opaque
data:
  tls.crt: $(cat /tmp/capstone-webhook.pem | base64 | tr -d '\n')
  tls.key: $(cat /tmp/capstone-webhook-key.pem | base64 | tr -d '\n') 
EOF

# generate CA Bundle + inject into template
ca_pem_b64="$(openssl base64 -A <"/tmp/ca.pem")"

sed -e 's@${CA_BUNDLE}@'"$ca_pem_b64"'@g' <"./kubernetes/manifests/webhook/webhook-template.yaml" \
    > ./kubernetes/manifests/webhook/webhook.yaml
```