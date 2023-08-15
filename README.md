# vault-benchmarks

This repo holds the code to deploy, run and evaluate (vault-benchmark)[https://github.com/hashicorp/vault-benchmark/] on AKS, GKE and (Constellation)[https://github.com/edgelesssys/constellation].
It also holds the raw data of 100 runs of vault-benchmark against AKS, GKE and Constellation.

# Benchmark Setup

We run a Vault cluster with two nodes, each scheduled on a separate VM.
One primary and one read-replica.
Vault only (scales vertically)[https://developer.hashicorp.com/vault/tutorials/operations/performance-tuning#performance-standbys], if not using the enterprise edition.
A third VM runs the load generator (vault-benchmark).
In many scenarios Vault's performance is (I/O bound)[https://developer.hashicorp.com/vault/tutorials/operations/performance-tuning#a-note-about-cpu-scaling].

# Deploy
- Start your target cluster:
    - `az aks create -g <resourcegroup> -n <name> --node-count 3 -s Standard_DC4as_v5`
    - `gcloud container clusters create <name> --zone europe-west3-b --node-locations europe-west3-b --machine-type n2d-standard-4 --num-nodes 3`
    - `constellation create -c 3 -w 3 --debug -y && constellation init`
- Fetch the credentials and put them into separate files in the root of the repo:
    - `touch kubecfg-gke.yaml && export KUBECONFIG=./kubecfg-gke.yaml && gcloud container clusters get-credentials <name> --region europe-west3-b`
    - `touch kubecfg-aks.yaml && export KUBECONFIG=./kubecfg-aks.yaml && az aks get-credentials -g <resourcegroup> --name <name>`
    - `mv constellation-admin.yaml kubecfg-c11n.yaml`
- Run the setup script: `source ./deploy/setup.sh`

# Benchmarks

- Replace `replaceme_vault_token` in `transit-<target>.yaml` with the root token you have saved in `$VAULT_KEYS` after running `setup.sh`.
- Replace `replaceme_node_name` in `transit-<target>.yaml` with the hostname of a node you want the load generator to run on.
- Run, e.g. 5 iterations on Constellation: `./benchmark/converge.sh 5 c11n`

# Evaluation

To calculate some basic statistics from the data in this repo and regenerate the boxplots cd into `vegeta-parser` and run: `go run .`.

You will be presented with results like these:
```
========== Results AKS ==========
Mean: mean: 1.729962, variance: 0.194084
P99: mean: 8.216240, variance: 4.842087
Max: mean: 8.811985, variance: 4.267199
Min: mean: 0.009774, variance: 0.000169
========== Results GKE ==========
Mean: mean: 1.733809, variance: 0.196227
P99: mean: 7.974444, variance: 5.930670
Max: mean: 8.613526, variance: 4.831148
Min: mean: 0.008509, variance: 0.000121
========== Results C11n ==========
Mean: mean: 1.847684, variance: 0.198169
P99: mean: 7.518989, variance: 4.830565
Max: mean: 8.161877, variance: 3.909015
Min: mean: 0.011946, variance: 0.000225
========== AKS vs C11n ==========
Mean: +6.371342 %
P99: -9.273208 %
Max: -7.965180 %
Min: +18.178964 %
========== GKE vs C11n ==========
Mean: +6.163156 %
P99: -6.057404 %
Max: -5.533651 %
Min: +28.767820 %
```

The above results are from 100 runs, 1300 workers, 1 job, 10 seconds per run.
