# Kubernetes Deployment Guide

This guide provides comprehensive instructions for deploying the Go web application with PostgreSQL on Kubernetes.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Architecture Overview](#architecture-overview)
- [Quick Start](#quick-start)
- [Detailed Deployment Steps](#detailed-deployment-steps)
- [Configuration](#configuration)
- [Testing the Deployment](#testing-the-deployment)
- [Scaling the Application](#scaling-the-application)
- [Updating the Application](#updating-the-application)
- [Monitoring and Logging](#monitoring-and-logging)
- [Troubleshooting](#troubleshooting)
- [Production Recommendations](#production-recommendations)
- [Cleanup](#cleanup)

## Prerequisites

Before deploying, ensure you have:

1. **Kubernetes Cluster** (one of the following):
   - Minikube (local development)
   - Docker Desktop with Kubernetes enabled
   - DigitalOcean Kubernetes (DOKS)
   - Amazon EKS
   - Google GKE
   - Azure AKS

2. **kubectl** - Kubernetes command-line tool
   ```bash
   # Check kubectl is installed
   kubectl version --client

   # Verify cluster access
   kubectl cluster-info
   ```

3. **Docker Image** - Application image built and available
   ```bash
   # Build the Docker image
   docker build -t software-engineering-app:latest .

   # For cloud deployments, push to a registry
   docker tag software-engineering-app:latest your-registry/software-engineering-app:latest
   docker push your-registry/software-engineering-app:latest
   ```

4. **Storage Class** - Ensure your cluster has a default storage class
   ```bash
   kubectl get storageclass
   ```

## Architecture Overview

### Components

- **Namespace**: `cruder-app` - Isolated environment for all resources
- **Application**: 2 replicas with rolling updates
- **Database**: PostgreSQL 16 StatefulSet with persistent storage
- **Service**: LoadBalancer for external access on port 80
- **Storage**: 1Gi PersistentVolumeClaim for database data

### Resource Limits

**Application Pod**:
- CPU: 100m request, 500m limit
- Memory: 128Mi request, 512Mi limit

**PostgreSQL Pod**:
- CPU: 100m request, 500m limit
- Memory: 256Mi request, 512Mi limit

## Quick Start

For a quick deployment with default settings:

```bash
# Clone the repository (if not already done)
cd software-engineering-test-task-master

# Apply all manifests
kubectl apply -f k8s/

# Wait for all pods to be ready
kubectl wait --for=condition=ready pod -l app=cruder -n cruder-app --timeout=300s

# Get the external IP address
kubectl get service app-service -n cruder-app
```

## Local Testing with Minikube

This section provides comprehensive guidance for testing Kubernetes manifests locally using Minikube.

### Minikube Installation

**Windows (PowerShell):**
```powershell
# Option 1: Via Chocolatey (recommended)
choco install minikube -y
choco install kubernetes-cli -y

# Option 2: Via winget
winget install Kubernetes.minikube
winget install Kubernetes.kubectl

# Verify installation
minikube version
kubectl version --client
```

### Start Minikube Cluster
```powershell
# Start Minikube with Docker driver
minikube start --driver=docker

# First run takes 2-5 minutes (downloads ~1GB)

# Verify cluster status
minikube status
kubectl get nodes
```

Expected output:
```
NAME       STATUS   ROLES           AGE   VERSION
minikube   Ready    control-plane   1m    v1.34.0
```

### Known Issues and Solutions

#### Issue 1: Storage Provisioner Error

**Symptom:** `storage-provisioner` pod shows `ErrImagePull`

**Impact:** Minimal - basic PV functionality still works

**Solution:** Can be safely ignored for testing

#### Issue 2: Binary Permission Denied

**Symptom:** `exec: "./main": stat ./main: permission denied`

**Root Cause:** Binary lacks execute permissions for non-root user

**Solution:** Already fixed in Dockerfile:
```dockerfile
RUN chmod +x /root/main
```

#### Issue 3: PVC Stuck in Pending

**Symptom:** PostgreSQL StatefulSet PVC cannot bind to PV

**Root Cause:** Storage provisioner may not function properly

**Solution:** Use external PostgreSQL for testing:
```powershell
# Get Minikube host IP
minikube ssh
ip route show | grep default
# Example: default via 192.168.49.1 dev eth0
exit

# Update ConfigMap with host IP
# Edit k8s/configmap.yaml:
DB_HOST: "192.168.49.1"  # Replace with your Minikube host IP

# Apply and restart
kubectl apply -f k8s/configmap.yaml
kubectl rollout restart deployment app-deployment -n cruder-app
```

### Deploy to Minikube

#### Step 1: Build and Load Image
```powershell
# Build Docker image
docker build -t software-engineering-app:latest .

# Load into Minikube
minikube image load software-engineering-app:latest

# Verify
minikube image ls | Select-String "software-engineering-app"
```

#### Step 2: Update Secrets for Local Testing
```powershell
# Delete default secrets
kubectl delete secret app-secrets -n cruder-app --ignore-not-found

# Create with local credentials
kubectl create secret generic app-secrets -n cruder-app `
  --from-literal=DB_USER=postgres `
  --from-literal=DB_PASSWORD=postgres `
  --from-literal=X_API_KEY=prod-api-key-secure-12345
```

#### Step 3: Deploy Application
```powershell
# Apply manifests (skip postgres-statefulset if using external DB)
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secret.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml

# Watch deployment progress
kubectl get pods -n cruder-app --watch
```

Wait for `1/1 Running` status:
```
NAME                              READY   STATUS    RESTARTS   AGE
app-deployment-xxxxxxxxxx-xxxxx   1/1     Running   0          45s
app-deployment-xxxxxxxxxx-xxxxx   1/1     Running   0          45s
```

Press `Ctrl+C` to stop watching.

#### Step 4: Enable LoadBalancer Access
```powershell
# Terminal 1: Start tunnel (keep running)
minikube tunnel
# May require administrator password

# Terminal 2: Get external IP
kubectl get service app-service -n cruder-app
```

Expected output:
```
NAME          TYPE           CLUSTER-IP     EXTERNAL-IP   PORT(S)        AGE
app-service   LoadBalancer   10.105.x.x     127.0.0.1     80:xxxxx/TCP   2m
```

#### Step 5: Test API Endpoints
```powershell
# Test with valid API key
curl.exe http://127.0.0.1/api/v1/users/ -H "X-API-Key: prod-api-key-secure-12345"

# Expected: JSON array with user data
# [{"id":1,"uuid":"...","username":"jdoe",...}, ...]

# Test without API key (should return 401)
curl.exe http://127.0.0.1/api/v1/users/

# Expected: {"error":"API key required"}

# Test with invalid API key (should return 403)
curl.exe http://127.0.0.1/api/v1/users/ -H "X-API-Key: wrong-key"

# Expected: {"error":"Invalid API key"}
```

### View Application Logs
```powershell
# All application pods
kubectl logs -l app=cruder -n cruder-app --tail=50 -f

# Specific pod
kubectl logs <pod-name> -n cruder-app --tail=100

# Describe pod (view events)
kubectl describe pod <pod-name> -n cruder-app
```

### Minikube Resource Management

**When Running:**
- RAM: 2-3GB
- CPU: 1-2 cores  
- Disk: ~3GB

**When Stopped:**
- RAM: 0GB
- CPU: 0%
- Disk: ~3GB (persisted)
```powershell
# Stop Minikube (preserves state, frees resources)
minikube stop

# Start again when needed
minikube start

# Delete cluster entirely (removes all data)
minikube delete
```

### Cleanup After Testing
```powershell
# Stop tunnel (Ctrl+C in tunnel terminal)

# Delete application resources
kubectl delete namespace cruder-app

# Stop Minikube
minikube stop

# Optional: Delete Minikube VM
minikube delete
```

### ConfigMap Note for Local Testing

The `k8s/configmap.yaml` includes a helpful comment:
```yaml
data:
  DB_HOST: "postgres-service"
  # for local debug - specify IP for host.minikube.internal here like DB_HOST: "192.168.49.1"
```

This IP (192.168.49.1) is:
- Safe to document - it's a local RFC1918 private IP
- Only works within Minikube environment
- Changes with each Minikube instance
- Helps developers understand local testing

### Production vs. Local Differences

| Aspect | Minikube (Local) | Production (Cloud K8s) |
|--------|------------------|------------------------|
| **PostgreSQL** | External Docker or local StatefulSet | Managed Database Service |
| **Storage** | hostPath provisioner | Cloud persistent volumes |
| **LoadBalancer** | minikube tunnel (127.0.0.1) | Cloud LB (public IP) |
| **Secrets** | kubectl create manual | CI/CD automation or Sealed Secrets |
| **Container Registry** | Local image load | Docker Hub, ECR, GCR, DOCR |
| **DNS** | Requires host IP workaround | Native service discovery |
| **Startup** | Manual (`minikube start`) | Always available |

### Troubleshooting Minikube

**Problem:** Minikube won't start
```powershell
# Check Docker is running
docker ps

# Delete and recreate
minikube delete
minikube start --driver=docker
```

**Problem:** Image not found in Minikube
```powershell
# Verify image loaded
minikube image ls | Select-String "software-engineering-app"

# Reload if missing
minikube image load software-engineering-app:latest
```

**Problem:** Pods crash with database connection error
```powershell
# Check logs
kubectl logs -l app=cruder -n cruder-app

# Verify ConfigMap has correct DB_HOST
kubectl get configmap app-config -n cruder-app -o yaml

# Verify secrets have correct credentials
kubectl get secret app-secrets -n cruder-app -o jsonpath='{.data.DB_PASSWORD}' | base64 -d
```

### Successfully Tested

? **Test Date:** 2025-11-23  
? **Minikube Version:** v1.37.0  
? **Kubernetes Version:** v1.34.0  
? **All Features Verified:**
- Deployment with 2 replicas
- LoadBalancer service  
- API authentication (401/403/200)
- Database connectivity
- Health checks (liveness/readiness)
- Resource limits

---

## Detailed Deployment Steps

### Step 1: Customize Secrets (IMPORTANT)

Before deploying, update the secrets with your actual credentials.

**Option A: Using kubectl create secret**

```bash
# Create app secrets
kubectl create secret generic app-secrets \
  --from-literal=DB_USER=your_db_user \
  --from-literal=DB_PASSWORD=your_secure_password \
  --from-literal=X_API_KEY=your-api-key-here \
  -n cruder-app --dry-run=client -o yaml > k8s/secret-generated.yaml

# Create postgres secrets
kubectl create secret generic postgres-secrets \
  --from-literal=POSTGRES_USER=your_db_user \
  --from-literal=POSTGRES_PASSWORD=your_secure_password \
  --from-literal=POSTGRES_DB=testdb \
  -n cruder-app --dry-run=client -o yaml >> k8s/secret-generated.yaml

# Apply the generated secrets
kubectl apply -f k8s/secret-generated.yaml
```

**Option B: Manual base64 encoding**

```bash
# Encode your values
echo -n 'your_db_user' | base64
echo -n 'your_secure_password' | base64
echo -n 'your-api-key' | base64

# Edit k8s/secret.yaml and replace the base64 values
nano k8s/secret.yaml
```

### Step 2: Create Namespace

```bash
kubectl apply -f k8s/namespace.yaml

# Verify namespace creation
kubectl get namespace cruder-app
```

### Step 3: Deploy PostgreSQL

```bash
# Apply PostgreSQL resources
kubectl apply -f k8s/postgres-pvc.yaml
kubectl apply -f k8s/secret.yaml  # If not done in Step 1
kubectl apply -f k8s/postgres-statefulset.yaml
kubectl apply -f k8s/postgres-service.yaml

# Check PostgreSQL status
kubectl get statefulset postgres -n cruder-app
kubectl get pods -l app=postgres -n cruder-app

# Wait for PostgreSQL to be ready
kubectl wait --for=condition=ready pod -l app=postgres -n cruder-app --timeout=120s
```

### Step 4: Deploy Application

```bash
# Apply application resources
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml

# Check application status
kubectl get deployment app-deployment -n cruder-app
kubectl get pods -l app=cruder -n cruder-app

# Wait for application to be ready
kubectl wait --for=condition=ready pod -l app=cruder -n cruder-app --timeout=180s
```

### Step 5: Verify Deployment

```bash
# Check all resources
kubectl get all -n cruder-app

# Check pod logs
kubectl logs -l app=cruder -n cruder-app --tail=50

# Check PostgreSQL logs
kubectl logs -l app=postgres -n cruder-app --tail=50
```

## Configuration

### Environment Variables

Edit `k8s/configmap.yaml` to change non-sensitive configuration:

```yaml
data:
  DB_HOST: "postgres-service"  # PostgreSQL service name
  DB_PORT: "5432"              # PostgreSQL port
  DB_NAME: "testdb"            # Database name
  GIN_MODE: "release"          # Gin framework mode
  SERVER_PORT: "8080"          # Application port
```

### Database Configuration

To use an external managed database instead of the StatefulSet:

1. Remove or don't apply:
   - `k8s/postgres-statefulset.yaml`
   - `k8s/postgres-service.yaml`
   - `k8s/postgres-pvc.yaml`

2. Update `k8s/configmap.yaml`:
   ```yaml
   DB_HOST: "your-managed-db-host.example.com"
   DB_PORT: "5432"
   DB_NAME: "production_database"
   ```

3. Update `k8s/secret.yaml` with managed database credentials

### Resource Limits

Edit `k8s/deployment.yaml` to adjust resource limits:

```yaml
resources:
  requests:
    cpu: 100m      # Minimum CPU
    memory: 128Mi  # Minimum Memory
  limits:
    cpu: 500m      # Maximum CPU
    memory: 512Mi  # Maximum Memory
```

## Testing the Deployment

### Get External IP Address

```bash
# Get the LoadBalancer external IP
kubectl get service app-service -n cruder-app

# Wait for EXTERNAL-IP (may take a few minutes)
kubectl get service app-service -n cruder-app --watch
```

### Test API Endpoints

```bash
# Set variables
EXTERNAL_IP=$(kubectl get service app-service -n cruder-app -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
API_KEY="prod-api-key-secure-12345"  # Use your actual API key

# Test health check
curl -H "X-API-Key: $API_KEY" http://$EXTERNAL_IP/api/v1/users/

# Test getting all users
curl -H "X-API-Key: $API_KEY" http://$EXTERNAL_IP/api/v1/users/

# Test creating a user (POST)
curl -X POST -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com"}' \
  http://$EXTERNAL_IP/api/v1/users/

# Test authentication (should fail without API key)
curl http://$EXTERNAL_IP/api/v1/users/
# Expected: {"error":"API key required"}
```

### Access PostgreSQL (for debugging)

```bash
# Port forward to PostgreSQL
kubectl port-forward -n cruder-app svc/postgres-service 5432:5432

# Connect using psql (in another terminal)
psql -h localhost -p 5432 -U postgres -d testdb
# Password: postgres123 (or your configured password)
```

## Scaling the Application

### Manual Scaling

```bash
# Scale to 3 replicas
kubectl scale deployment app-deployment --replicas=3 -n cruder-app

# Verify scaling
kubectl get pods -l app=cruder -n cruder-app

# Scale down to 1 replica
kubectl scale deployment app-deployment --replicas=1 -n cruder-app
```

### Horizontal Pod Autoscaling (HPA)

```bash
# Create HPA (requires metrics-server)
kubectl autoscale deployment app-deployment \
  --cpu-percent=70 \
  --min=2 \
  --max=10 \
  -n cruder-app

# Check HPA status
kubectl get hpa -n cruder-app

# Describe HPA
kubectl describe hpa app-deployment -n cruder-app
```

## Updating the Application

### Rolling Update

```bash
# Build new version
docker build -t software-engineering-app:v2 .

# Tag for registry
docker tag software-engineering-app:v2 your-registry/software-engineering-app:v2
docker push your-registry/software-engineering-app:v2

# Update deployment
kubectl set image deployment/app-deployment \
  cruder-app=your-registry/software-engineering-app:v2 \
  -n cruder-app

# Watch rollout status
kubectl rollout status deployment/app-deployment -n cruder-app

# Check rollout history
kubectl rollout history deployment/app-deployment -n cruder-app
```

### Rollback

```bash
# Rollback to previous version
kubectl rollout undo deployment/app-deployment -n cruder-app

# Rollback to specific revision
kubectl rollout undo deployment/app-deployment --to-revision=1 -n cruder-app
```

## Monitoring and Logging

### View Logs

```bash
# Application logs (all pods)
kubectl logs -l app=cruder -n cruder-app --tail=100 -f

# Specific pod logs
kubectl logs <pod-name> -n cruder-app --tail=100 -f

# PostgreSQL logs
kubectl logs -l app=postgres -n cruder-app --tail=100 -f

# Previous container logs (if pod crashed)
kubectl logs <pod-name> -n cruder-app --previous
```

### Describe Resources

```bash
# Describe deployment
kubectl describe deployment app-deployment -n cruder-app

# Describe pod (check events)
kubectl describe pod <pod-name> -n cruder-app

# Describe service
kubectl describe service app-service -n cruder-app
```

### Execute Commands in Pods

```bash
# Get a shell in application pod
kubectl exec -it <app-pod-name> -n cruder-app -- /bin/sh

# Get a shell in PostgreSQL pod
kubectl exec -it postgres-0 -n cruder-app -- /bin/bash

# Run psql in PostgreSQL pod
kubectl exec -it postgres-0 -n cruder-app -- psql -U postgres -d testdb
```

## Troubleshooting

### Pods Not Starting

```bash
# Check pod status
kubectl get pods -n cruder-app

# Describe pod to see events
kubectl describe pod <pod-name> -n cruder-app

# Common issues:
# - ImagePullBackOff: Docker image not found or auth issue
# - CrashLoopBackOff: Application crashing on startup
# - Pending: Insufficient resources or PVC not bound
```

### Database Connection Issues

```bash
# Verify PostgreSQL is running
kubectl get pods -l app=postgres -n cruder-app

# Check PostgreSQL logs
kubectl logs postgres-0 -n cruder-app

# Verify service
kubectl get service postgres-service -n cruder-app

# Test connectivity from app pod
kubectl exec -it <app-pod-name> -n cruder-app -- nc -zv postgres-service 5432
```

### LoadBalancer Not Getting External IP

```bash
# Check service status
kubectl get service app-service -n cruder-app

# Describe service
kubectl describe service app-service -n cruder-app

# For Minikube:
minikube tunnel  # Run in separate terminal

# For Docker Desktop:
# LoadBalancer will use localhost

# For cloud providers:
# Ensure your cluster has load balancer support enabled
```

### Secret/ConfigMap Not Applied

```bash
# Check if secrets exist
kubectl get secrets -n cruder-app

# Check if configmaps exist
kubectl get configmaps -n cruder-app

# Describe to see data keys
kubectl describe secret app-secrets -n cruder-app
kubectl describe configmap app-config -n cruder-app

# Restart pods to pick up changes
kubectl rollout restart deployment/app-deployment -n cruder-app
```

### Application Health Check Failures

```bash
# Check liveness/readiness probe configuration
kubectl describe pod <pod-name> -n cruder-app

# View application logs for errors
kubectl logs <pod-name> -n cruder-app

# Test endpoint manually
kubectl port-forward <pod-name> 8080:8080 -n cruder-app
# In another terminal:
curl -H "X-API-Key: prod-api-key-secure-12345" http://localhost:8080/api/v1/users/
```

## Production Recommendations

### Security

1. **Use Managed Secrets**:
   ```bash
   # Use external-secrets or sealed-secrets
   # Example with sealed-secrets:
   kubeseal --format=yaml < secret.yaml > sealed-secret.yaml
   ```

2. **Network Policies**:
   - Restrict pod-to-pod communication
   - Only allow app pods to access PostgreSQL

3. **RBAC**:
   - Create service accounts with minimal permissions
   - Use Pod Security Policies or Pod Security Standards

4. **TLS/SSL**:
   - Use Ingress controller with cert-manager for HTTPS
   - Enable SSL for PostgreSQL connections

### High Availability

1. **Use Managed Database**:
   - DigitalOcean Managed Databases
   - AWS RDS
   - Google Cloud SQL
   - Azure Database for PostgreSQL

2. **Multi-Zone Deployment**:
   ```yaml
   # Add to deployment.yaml
   affinity:
     podAntiAffinity:
       requiredDuringSchedulingIgnoredDuringExecution:
         - topologyKey: topology.kubernetes.io/zone
   ```

3. **Backup Strategy**:
   ```bash
   # Regular backups for PostgreSQL PVC
   # Use Velero or cloud provider snapshots
   ```

### Monitoring

1. **Install Prometheus & Grafana**:
   ```bash
   # Using Helm
   helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
   helm install prometheus prometheus-community/kube-prometheus-stack
   ```

2. **Application Metrics**:
   - Add Prometheus metrics endpoint to application
   - Create Grafana dashboards

3. **Logging**:
   - ELK Stack (Elasticsearch, Logstash, Kibana)
   - Loki + Grafana
   - Cloud provider logging (CloudWatch, Stackdriver)

### Performance

1. **Resource Optimization**:
   - Monitor actual resource usage
   - Adjust requests/limits based on metrics

2. **Caching**:
   - Add Redis for caching
   - Configure CDN for static assets

3. **Database Optimization**:
   - Enable connection pooling
   - Add read replicas for read-heavy workloads

## Cleanup

### Remove All Resources

```bash
# Delete everything in the namespace
kubectl delete namespace cruder-app

# Or delete individual resources
kubectl delete -f k8s/

# Verify deletion
kubectl get all -n cruder-app
```

### Delete PersistentVolume

```bash
# List PVs
kubectl get pv

# Delete specific PV if needed
kubectl delete pv <pv-name>

# Cloud providers: Delete volumes from cloud console if necessary
```

## Additional Resources

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [kubectl Cheat Sheet](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)
- [Kubernetes Best Practices](https://kubernetes.io/docs/concepts/configuration/overview/)
- [Production Checklist](https://learnk8s.io/production-best-practices)

## Support

For issues and questions:
- Check [Troubleshooting](#troubleshooting) section
- Review pod logs and events
- Consult Kubernetes documentation
- Seek help from your platform provider's support

---

**Last Updated**: 2025-11-23
**Version**: 1.0.0
