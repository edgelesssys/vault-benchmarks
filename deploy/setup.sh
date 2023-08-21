#!/usr/bin/env bash

function install() {
  if helm list -q -n vault | grep -q "vault"; then
    echo "Vault release already exists. Aborting..."
    return 1
  fi

  echo "Install fresh vault certificate? (y/n)"
  read confirm
  if [[ "$confirm" = "y" ]]; then
      kubectl delete --ignore-not-found=true csr vault-csr
      ./deploy/create_cert.sh
  fi

  echo "Have you updated the node names in vault-overrides.yaml? (y/n)"
  read confirm
  if [[ ! "$confirm" = "y" ]]; then
    echo "Please update the values under server.affinity first."
    return 1
  fi

  if [[ ! -z $(helm repo list | grep -q "https://helm.releases.hashicorp.com") ]]; then
    helm repo add hashicorp https://helm.releases.hashicorp.com
  fi

  helm install vault hashicorp/vault --namespace vault --create-namespace -f ./deploy/vault-overrides.yaml
}

function initialize() {
  for i in $(seq 0 $(($1 - 1))); do
    kubectl wait --for=condition=Ready --timeout=2m -n vault pod/vault-$i
    if [[ ! $? = 0 ]]; then
      echo "Vault not ready. Aborting..."
      return 1
    fi
  done


  export VAULT_KEYS=$(kubectl exec --stdin=true --tty=true -n vault vault-0 -- vault operator init -format json)
  if [[ -z "$VAULT_KEYS" ]]; then
    echo "Vault is already initialized. Aborting..."
    return 1
  fi

  export VAULT_KEY_1=$(echo $VAULT_KEYS | yq eval ".unseal_keys_b64[0]" -) && export VAULT_KEY_2=$(echo $VAULT_KEYS | yq eval ".unseal_keys_b64[1]" -) && export VAULT_KEY_3=$(echo $VAULT_KEYS | yq eval ".unseal_keys_b64[2]" -)

  for i in $(seq 0 $(($1 - 1))); do
    echo -n "Waiting for vault-0 to be ready"
    until kubectl exec --stdin=true --tty=true -n vault vault-$i -- sh -c "vault operator unseal $VAULT_KEY_1 && vault operator unseal $VAULT_KEY_2 && vault operator unseal $VAULT_KEY_3"; do
      echo -n "."
      sleep 0.5
    done
    echo ""
  done

}

echo "How many replicas did you configure?"
read replicas
if [[ ! "$replicas" =~ ^[0-9]+$ ]]; then
  echo "Please enter a number."
  return 1
fi

install
# # Wait for pods to be spawned
sleep 5
initialize $replicas
