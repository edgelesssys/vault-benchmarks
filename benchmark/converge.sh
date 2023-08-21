#!/usr/bin/env bash

if [[ -z "$1" || ! "$1" =~ ^[0-9]+$ ]]; then
    echo "Usage: $0 <iterations> <csp>"
    exit 1
fi

if [[ ! "$2" =~ ^(gke|aks|c11n)$ ]]; then
    echo "Invalid csp, got $2 expected <gke|aks|c11n>"
    exit 1
fi


function converge() {
    datadir=./data/1300w/5replicas/$2
    mkdir -p $datadir
    KUBECONFIG=./kubecfg-$2.yaml

    for i in $(seq 1 $1); do
        echo "Running iteration $i / $1"
        RESULTS_FILE=$datadir/results-$(date +%Y%m%d%H%M%S).json
        kubectl delete --ignore-not-found=true job vault-benchmark -n vault && \
        kubectl apply -f ./benchmark/transit-$2.yaml && \
        kubectl wait --for=condition=complete job/vault-benchmark -n vault && \
        PODNAME=$(kubectl get pods --selector=job-name=vault-benchmark -n vault --output=jsonpath='{.items[*].metadata.name}') && \
        kubectl logs $PODNAME -n vault > $RESULTS_FILE
        sleep 2
    done
}

converge $1 $2
