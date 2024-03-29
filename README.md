# vault-benchmarks

This repo holds the code to deploy, run and evaluate [vault-benchmark](https://github.com/hashicorp/vault-benchmark/) on AKS, GKE and [Constellation](https://github.com/edgelesssys/constellation).
It also holds the raw data of various runs of vault-benchmark against AKS, GKE and Constellation.

# Benchmark Setup

We run a Vault cluster with n nodes, each scheduled on a separate VM.
One primary and n-1 read-replica.
Vault only [scales vertically](https://developer.hashicorp.com/vault/tutorials/operations/performance-tuning#performance-standbys), if not using the enterprise edition.
A separate VM (n+1 th) runs the load generator (vault-benchmark).
In many scenarios Vault's performance is [I/O bound](https://developer.hashicorp.com/vault/tutorials/operations/performance-tuning#a-note-about-cpu-scaling).

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

To calculate basic statistics from the data in this repo and regenerate the boxplots run: `cd vegeta-parser && go run .`.

Results 2 replicas, 100 runs, 1300 workers, 1 job, 10 seconds per run.:
```
========== Results AKS ==========
Mean:   mean: 1.729962, variance: 0.194084
P99:    mean: 8.216240, variance: 4.842087
Max:    mean: 8.811985, variance: 4.267199
Min:    mean: 0.009774, variance: 0.000169
========== Results GKE ==========
Mean:   mean: 1.733809, variance: 0.196227
P99:    mean: 7.974444, variance: 5.930670
Max:    mean: 8.613526, variance: 4.831148
Min:    mean: 0.008509, variance: 0.000121
========== Results C11n ==========
Mean:   mean: 1.847684, variance: 0.198169
P99:    mean: 7.518989, variance: 4.830565
Max:    mean: 8.161877, variance: 3.909015
Min:    mean: 0.011946, variance: 0.000225
========== AKS vs C11n ==========
Mean:   +6.371342 % (lower is better)
P99:    -9.273208 % (lower is better)
Max:    -7.965180 % (lower is better)
Min:    +18.178964 % (lower is better)
========== GKE vs C11n ==========
Mean:   +6.163156 % (lower is better)
P99:    -6.057404 % (lower is better)
Max:    -5.533651 % (lower is better)
Min:    +28.767820 % (lower is better)
```

Results 5 replicas, 100 runs, 1300 workers, 1 job, 10 seconds per run.:
```
========== Results AKS ==========
Mean:   mean: 1.632200, variance: 0.002057
P99:    mean: 5.480679, variance: 2.263700
Max:    mean: 6.651001, variance: 2.808401
Min:    mean: 0.011415, variance: 0.000133
========== Results GKE ==========
Mean:   mean: 1.656435, variance: 0.003615
P99:    mean: 6.030807, variance: 3.955051
Max:    mean: 7.164843, variance: 3.300004
Min:    mean: 0.010233, variance: 0.000111
========== Results C11n ==========
Mean:   mean: 1.651549, variance: 0.001610
P99:    mean: 5.780422, variance: 3.016106
Max:    mean: 6.942997, variance: 3.075796
Min:    mean: 0.013774, variance: 0.000228
========== AKS vs C11n ==========
Mean:   +1.171577 %
P99:    +5.185495 %
Max:    +4.205618 %
Min:    +17.128781 %
========== GKE vs C11n ==========
Mean:   -0.295851 %
P99:    -4.331603 %
Max:    -3.195248 %
Min:    +25.710886 %
```

Please see the [Constellation documentation](https://docs.edgeless.systems/constellation/overview/performance/application) for an explanation of the results.
