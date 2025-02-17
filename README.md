# **GoMark - Configuration Engine**

## **1. Project Overview**
GoMark is a scalable backend configuration engine designed to manage HubSpot deal objects and Upso automated outreach cadences. It ensures consistency, reliability, and scalability by handling API integrations, preventing external changes, and optimizing performance using caching and queueing.

---

## **2. Running the Project Locally**

### **Prerequisites**
- Golang (1.20+)
- Docker & Docker Compose

### **Setup & Execution**

#### **Clone the Repository**
```sh
git clone https://github.com/f5aaff/gomark.git
cd gomark
```

#### **Start Services Using Docker Compose**
```sh
docker-compose up -d
```
This will start:
- PostgreSQL (Database)
- Redis (Caching layer)
- RabbitMQ (Message queue for async operations)
- Backend  GoLang service

#### **Run the Backend Manually (Without Docker)**
```sh
export DATABASE_URL="postgres://user:password@localhost:5432/gomark"
export REDIS_HOST="localhost:6379"
go run main.go
```

#### **API Endpoints**
| Method | Endpoint | Description |
|--------|---------|-------------|
| POST | `/hubspot/fields/add` | Add a new field to a HubSpot deal object |
| PUT | `/hubspot/fields/modify/{company_id}/{old_name}` | Modify an existing HubSpot deal field name |
| PUT | `/upso/cadence/modify/{company_id}` | Modify an Upso email cadence template and timing |

#### **Testing API Calls**
Use cURL or Postman:
```sh
curl -X POST http://localhost:8080/hubspot/fields/add -d '{"company_id":"123", "name":"Deal Stage", "type":"string", "value":"Negotiation"}' -H "Content-Type: application/json"
```

---

## **3. Design Reasoning**

### **Microservices Approach with Separation of Concerns**
- GoLang Microservices → Handles deal object changes for HubSpot & manages automated outreach email cadences through upso.
- Orchestration Layer → Ensures consistency across integrations.
- Queue Worker Service (RabbitMQ) → Prevents API rate limits and ensures async processing.

### **PostgreSQL as a Source of Truth**
- Maintains a single authoritative configuration.
- Prevents data inconsistency from external API modifications.
- Supports audit logs and rollback features.

### **Caching & Performance Optimizations**
- Redis caches frequently accessed configurations to reduce database queries.
- API request queueing (RabbitMQ) prevents long-running API calls from blocking requests.
- Rate limiting and retry mechanisms prevent external API failures from disrupting operations.

---

## **4. Deployment Strategy**

### **Containerized Deployment (Docker & Kubernetes)**
For a scalable cloud deployment, Docker & Kubernetes are used in conjunction:

#### **1. Build the Docker Image**
```sh
cd src
docker build -t gomark-backend .
```

push the image to your personal dockerhub if you prefer, however, this will work fine with a local image; provided you edit the image paths.

#### **2. Kubernetes Deployment YAML**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gomark-backend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: gomark-backend
  template:
    metadata:
      labels:
        app: gomark-backend
    spec:
      containers:
      - name: gomark-backend
        image: your-dockerhub-username/gomark-backend:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          value: "postgres://user:password@db-service:5432/gomark"
```
#### **3. Deploy to Kubernetes**
```sh
kubectl apply -f deployment.yaml
kubectl expose deployment gomark-backend --type=LoadBalancer --port=80 --target-port=8080
```

### **Cloud Load Balancing & Scaling**
- Use AWS ALB or GCP Load Balancer to distribute traffic across instances.
- Horizontal Scaling with Kubernetes `HorizontalPodAutoscaler`:
```sh
kubectl autoscale deployment gomark-backend --cpu-percent=50 --min=3 --max=10
```

## **Summary**
### **Why This Design Works for GoMark**
- Scalable & Fault-Tolerant → Supports multiple companies without downtime.
- Optimized Performance → Caching, queueing, and load balancing reduce latency.
- Cloud-Ready Deployment → Easily deployable via Docker & Kubernetes.

### **Future Enhancements**
- Add GraphQL Support for more flexible queries.
- Implement Role-Based Access Control (RBAC) for security.
