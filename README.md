# PostgreSQL Kubernetes Operator (Level 4 Cloud Native Project)

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

## 🏗 Architecture
The Operator follows the standard Kubernetes Controller pattern:

1.  **User** applies a `SimpleDB` YAML.
2.  **API Server** stores the request.
3.  **Controller (My Code)** detects the new event.
4.  **Reconciler Loop** creates a Deployment and Service for Postgres.
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