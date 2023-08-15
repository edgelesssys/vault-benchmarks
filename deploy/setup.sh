#!/usr/bin/env bash

if helm list -q -n vault | grep -q "vault"; then
  echo "Vault release already exists. Aborting..."
  exit 0
fi

echo "Install fresh vault certificate? (y/n)"
read confirm
if [[ "$confirm" = "y" ]]; then
    kubectl delete --ignore-not-found=true csr vault-csr
    ./create_cert.sh
fi

echo "Have you updated the node names in vault-overrides.yaml? (y/n)"
read confirm
if [[ ! "$confirm" = "y" ]]; then
  echo "Please update the values under server.affinity first."
fi

if [[ ! -z $(helm repo list | grep -q "https://helm.releases.hashicorp.com") ]]; then
  helm repo add hashicorp https://helm.releases.hashicorp.com
fi

helm install vault hashicorp/vault --namespace vault --create-namespace -f ./deploy/vault-overrides.yaml

kubectl wait --for=condition=Ready --timeout=2m -n vault pod/vault-0
if [[ ! $? = 0 ]]; then
  echo "Vault not ready. Aborting..."
  exit 0
fi
kubectl wait --for=condition=Ready --timeout=2m -n vault pod/vault-1
if [[ ! $? = 0 ]]; then
  echo "Vault not ready. Aborting..."
  exit 0
fi

export VAULT_KEYS=$(kubectl exec --stdin=true --tty=true -n vault vault-0 -- vault operator init -format json)
if [[ -z "$VAULT_KEYS" ]]; then
  echo "Vault is already initialized. Aborting..."
  exit 0
fi

export VAULT_KEY_1=$(echo $VAULT_KEYS | yq eval ".unseal_keys_b64[0]" -) && export VAULT_KEY_2=$(echo $VAULT_KEYS | yq eval ".unseal_keys_b64[1]" -) && export VAULT_KEY_3=$(echo $VAULT_KEYS | yq eval ".unseal_keys_b64[2]" -)

echo -n "Waiting for vault-0 to be ready"
until kubectl exec --stdin=true --tty=true -n vault vault-0 -- sh -c "vault operator unseal $VAULT_KEY_1 && vault operator unseal $VAULT_KEY_2 && vault operator unseal $VAULT_KEY_3"; do
  echo -n "."
  sleep 0.5
done
echo ""

echo -n "Waiting for vault-1 to be ready"
until kubectl exec --stdin=true --tty=true -n vault vault-1 -- sh -c "vault operator unseal $VAULT_KEY_1 && vault operator unseal $VAULT_KEY_2 && vault operator unseal $VAULT_KEY_3"; do
  echo -n "."
  sleep 0.5
done
echo ""

kubectl wait --for=condition=Ready -n vault pod/vault-0 && kubectl wait --for=condition=Ready -n vault pod/vault-1
