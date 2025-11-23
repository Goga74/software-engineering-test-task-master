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
