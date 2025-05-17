# Non containerized workloads for Kubernetes
This project integrates Nomad HashiCorp with Kubernetes API to achieve native-like management of non-containerized workloads through Kubernetes. This approach addresses the issues of disk I/O performance degradation as a result of containerization, allowing to manage containerized microservices and non-containerized disk-heavy services together. 

## Project Motivation

**Context** 
When building *high-reliability high-load large-scale distributed systems*, components orchestration is always required. One of the most widely used orchestration systems that fits a variety of use cases is `Kubernetes`. System "components" managed by `Kubernetes` are often called **workloads**.

**Workload types**
 - **Stateless**
  - most microservices
   - some load balancers (e.g. Nginx by default)
 - **Stateful non-persistent**
   - some microservices (e.g. ones that cache some user data or enable background processing for heavy requests)
   - some caches (e.g. Redis by default)
   - some load balancers (e.g. Nginx when setup to cache data)
 - **Stateful persistent**
   - all databases
   - some caches (e.g. Redis when setup to persist the cache)
   - message brokers (e.g. Apache Kafka)
   - log aggregation systems

**Issue 1:**
**Stateful persistent workloads** heavily rely on **disk I/O operations**. However, these operations suffer intense performance degradation when executed in containerized environments (tests report up to **10x** slowdown).

> **Potential solution:**
> Deploy bare-metal stateful persistent workloads and manage them as non-containerized workloads in Kubernetes.

**Issue 2:**
Kubernetes does not natively support non-containerized workloads, nor does it have any plugins for such purposes

> **Potential solution:**
> Manage stateful persistent workloads outside of Kubernetes.

**Issue 3:**
Databases, message brokers, caches and other workloads have to be auto-scaled, auto-healed, updated, distributed, and controlled in the same way as all the other workloads. Implementing this functionality outside of Kubernetes is not feasible and is redundant.

> **Potential solution:**
> Use a Kubernetes alternative that supports both containerized and non-containerized environments, e.g. HashiCorp Nomad.

**Issue 4:**
While being feasible, this solution sacrifices the Kubernetes capabilities in exchange for support for non-containerized workloads. At the moment, Kubernetes is the superior choice among all alternatives for a general case. Thus, it is highly desired to avoid sacrificing Kubernetes.

> **Potential solution:**
> Combine Kubernetes with Nomad, using the first for containerized workloads and using the second for non-containerized workloads.

**Issue 5:**
This solution is acceptable, but inconvenient, since it now requires to setup two centralized orchestration systems. Having two sets of configurations, two implementations for every reliability mechanism and so on is undesirable.

> **Actual solution:**
> **Integrate** Nomad capabilities of orchestrating non-containerized workloads **into Kubernetes** using Kubernetes API. Build an abstraction level, so that every non-contsinerized workload appears as a native *"pseudo-container"* in Kubernetes.

---

## Project Requirements

**Goal:**
Allow seamless integration of any non-containerized workloads into Kubernetes without compromising any functionality.

**Objective 1:**
Implement a Nomad abstraction layer under Kubernetes API for managing stateful persistent non-containerized workloads as "pseudo-containers", enabling native integration through Kubernetes CRD.

**Objective 2:**
Simplify the process of configuring new workloads to match the Kubernetes native configurations as much as possible.

**Objective 3:**
Create a Linux package for the API translation (integration) system with simplified configuration

**Objective 4 (optional):**
Provide default setups for frequently used services:
- Databases
  - PostgreSQL (relational)
  - MongoDB (no sql)
  - MinIO (file storage)
  - Weaviate (vector storage)
  - Prometheus (metrics collection)
  - Loki (storage and indexing of logs)
- Message Brokers
  - Apache Kafka

**Deliverables:**
Git repository containing:
1. Implementation of integration of Nomad into Kubernetes API
2. Simple and flexible configuration file(s) for setting up non-containerized workloads with the expressive power of Kubernetes configuration
3. Linux package release encapsulating the entire system
4. Example configurations for frequently used services from Objective 4 (optional)
5. Clear, concise, and easily usable documentation with installation and usage specifications
