
# PostgreSQL Kubernetes Operator 

![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)
![Kubernetes](https://img.shields.io/badge/Kubernetes-Operator-326CE5?style=flat&logo=kubernetes)
![Docker](https://img.shields.io/badge/Docker-Container-2496ED?style=flat&logo=docker)

## 🚀 Overview
A custom Kubernetes Controller written in **Go** that automates the lifecycle management of PostgreSQL database clusters.

Unlike standard deployments, this Operator acts as a "Site Reliability Engineer in a box." It continuously watches the cluster state and enforces the desired configuration, offering **Self-Healing** capabilities that automatically recover deleted pods in under 5 seconds.

## ⚡ Key Features
* **Custom Resource Definition (CRD):** Extends the Kubernetes API with a new `SimpleDB` kind, allowing developers to request databases using native YAML.
* **Active Reconciliation:** Implements a control loop that constantly compares the "Desired State" (YAML) vs "Actual State" (Cluster).
* **Self-Healing Infrastructure:** Automatically detects if a database Pod is deleted or crashes and recreates it instantly.
* **Elastic Scaling:** Supports real-time horizontal scaling via `kubectl patch` or YAML updates.
* **Status Reporting:** Provides real-time feedback on the database health status directly in the terminal.

## 🏗 Architecture
The Operator follows the standard Kubernetes Controller pattern:

1.  **User** applies a `SimpleDB` YAML.
2.  **API Server** stores the request.
3.  **Controller (My Code)** detects the new event.
4.  **Reconciler Loop** creates and manages a Kubernetes Deployment for Postgres.
5.  **Observer** watches the Pods; if one dies, the Reconciler spins up a replacement.

## 🏃‍♂️ How to Run

### Prerequisites
* Go 1.22+
* Docker Desktop & Kind (Kubernetes in Docker)
* `kubectl`

### 1. Installation
Clone the repo and install the Custom Resource Definitions (CRDs) into your cluster:
```bash
make install

```

### 2. Start the Controller

Run the operator locally (it will connect to your current K8s context):

```bash
make run

```

### 3. Deploy a Database

Open a new terminal and apply the sample database configuration:

```bash
kubectl apply -f config/samples/database_v1_simpledb.yaml

```

### 4. Verify & Test Self-Healing

Check the status of your database:

```bash
kubectl get simpledbs
# Output: NAME              STATUS   AGE
#         simpledb-sample   Ready    2m

```

**🔥 The "Chaos Monkey" Test:**
Delete the pod manually to test resilience:

```bash
kubectl delete pod -l app=simpledb
kubectl get pods -w
# Result: You will see a new pod spin up instantly!

```

### 5. Test Scaling

Scale your database from 1 replica to 3 replicas on the fly:

```bash
kubectl patch simpledb simpledb-sample --type='merge' -p '{"spec":{"replicas":3}}'
kubectl get pods
# Result: 3 Pods running

```

## 🛠 Tech Stack

* **Language:** Go (Golang)
* **Framework:** Kubebuilder, Controller-Runtime
* **Orchestration:** Kubernetes (Kind)

---

*Built by [Hemraj](https://github.com/Hemraj8)*
