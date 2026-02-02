from diagrams import Diagram, Cluster, Edge
from diagrams.k8s.compute import Pod, Job, StatefulSet, Deployment
from diagrams.k8s.network import Service, Ingress
from diagrams.k8s.storage import PVC, StorageClass
from diagrams.k8s.clusterconfig import HPA
from diagrams.onprem.compute import Server
from diagrams.onprem.queue import RabbitMQ, Kafka
from diagrams.onprem.monitoring import Prometheus, Grafana
from diagrams.onprem.database import PostgreSQL
from diagrams.onprem.inmemory import Redis
from diagrams.onprem.network import Nginx, Traefik, Linkerd
from diagrams.onprem.gitops import Flux
from diagrams.onprem.container import Docker
from diagrams.onprem.storage import Ceph
from diagrams.programming.language import Go, NodeJS, Python
from diagrams.aws.storage import S3
from diagrams.aws.compute import ECR
from diagrams.generic.storage import Storage

graph_attr = {
    "fontsize": "18",
    "bgcolor": "white",
    "pad": "0.8",
    "splines": "spline",
    "ranksep": "1.5",
    "nodesep": "0.8"
}

with Diagram("Knative Lambda - Serverless Function Platform", 
             show=False, 
             direction="TB",
             filename="/Users/brunolucena/workspace/bruno/repos/homelab/flux/infrastructure/knative-lambda/docs/knative-lambda-architecture",
             graph_attr=graph_attr,
             outformat="png"):
    
    # ========== EDGE & INGRESS ==========
    with Cluster("Edge Layer"):
        nginx_edge = Nginx("Nginx\nReverse Proxy\nSSL Termination")
        traefik = Traefik("Traefik Ingress\nHTTP/gRPC Router")
    
    # ========== EXTERNAL STORAGE & REGISTRY ==========
    with Cluster("External Services (AWS/Self-Hosted)"):
        with Cluster("Source Code Storage"):
            s3_source = S3("S3 Source Bucket\nnotifi-uw2-dev-fusion-modules\nParser Code Storage")
            s3_tmp = S3("S3 Temp Bucket\nknative-lambda-dev-context-tmp\nBuild Context Cache")
        
        with Cluster("Container Registry"):
            ecr_prod = ECR("ECR Production\n339954290315.dkr.ecr.us-west-2\nLambda Images")
            harbor_local = Docker("Harbor Registry\nlocalhost:5001\nDev/Base Images")
    
    # ========== KNATIVE LAMBDA PLATFORM ==========
    with Cluster("Knative Lambda Platform (knative-lambda namespace)"):
        
        # ========== BUILDER SERVICE (CORE) ==========
        with Cluster("Builder Service (Knative Serving)"):
            builder_svc = Go("Lambda Builder\nKnative Service\nScale 0→10\nGo 1.24")
            metrics_pusher = Go("Metrics Pusher\nSidecar\nPrometheus Remote Write")
            
            with Cluster("Builder Components"):
                event_handler = Service("CloudEvent Handler\nbuild.start/job.start/service.delete")
                build_mgr = Service("Build Context Mgr\nS3 Download\nRate Limiting")
                job_mgr = Service("Job Manager\nKaniko Orchestration\nIdempotency")
                service_mgr = Service("Service Manager\nKnative Service CRUD\nTrigger Creation")
                dlq_handler_svc = Service("DLQ Handler\nRetry Logic\nError Categorization")
        
        # ========== KANIKO BUILD SYSTEM ==========
        with Cluster("Kaniko Build Jobs (Dynamic)"):
            with Cluster("Build Job (Batch/v1)"):
                kaniko_job = Job("Kaniko Job\nSecure Container Build\nNo Docker Daemon")
                kaniko_container = Docker("Kaniko Executor\nv1.19.2\nMulti-Arch")
                sidecar_monitor = Go("Sidecar Monitor\nBuild Progress\nMetrics Collection")
            
            with Cluster("Build Configuration"):
                dockerfile_gen = Service("Dockerfile Generator\nPython/Node/Go Templates")
                npm_config = Service("NPM Config\nRegistry Mirror\nFetch Retries")
                base_images = Service("Base Images\nnode:22-alpine\npython:3.11-alpine\ngolang:1.25-alpine")
        
        # ========== EVENTING INFRASTRUCTURE ==========
        with Cluster("Knative Eventing (RabbitMQ Backend)"):
            with Cluster("RabbitMQ Cluster"):
                rabbitmq_main = RabbitMQ("RabbitMQ Node 1\nQuorum Queues\nCloudEvents")
                rabbitmq_node2 = RabbitMQ("RabbitMQ Node 2\nMirror")
                rabbitmq_node3 = RabbitMQ("RabbitMQ Node 3\nMirror")
            
            with Cluster("Knative Brokers"):
                broker_dev = Service("Lambda Broker Dev\nRabbitMQBroker\nEvent Routing")
                broker_config = Service("Broker Config\nQuorum Queue\nHA Configuration")
            
            with Cluster("Event Sources"):
                api_source = Service("APIServerSource\nWatch K8s Events")
                rabbitmq_source = Service("RabbitMQSource\nExternal Events")
        
        # ========== DYNAMICALLY CREATED LAMBDA FUNCTIONS ==========
        with Cluster("Lambda Functions (Knative Services)"):
            with Cluster("Parser Functions (Scale-to-Zero)"):
                lambda_parser1 = NodeJS("Parser Function 1\nNodeJS 22\nmin=0, max=50")
                lambda_parser2 = Python("Parser Function 2\nPython 3.11\nmin=0, max=50")
                lambda_parser3 = Go("Parser Function 3\nGo 1.25\nmin=0, max=50")
            
            with Cluster("Lambda Configuration"):
                lambda_trigger = Service("Knative Trigger\nEvent Filtering\nCloudEvents Routing")
                lambda_autoscale = HPA("Knative Autoscaler\nKPA (Knative Pod Autoscaler)\nConcurrency-Based")
                lambda_service = Service("Knative Service\nRevision Management\nTraffic Splitting")
        
        # ========== DEAD LETTER QUEUE ==========
        with Cluster("Dead Letter Queue (DLQ)"):
            dlq_exchange = RabbitMQ("DLQ Exchange\nknative-lambda-dlq-exchange")
            dlq_queue = RabbitMQ("DLQ Queue\nknative-lambda-dlq\n7d TTL, 50K msgs")
            dlq_handler_dep = Deployment("DLQ Handler\nRetry with Backoff\nError Analysis")
            dlq_cleanup = Job("DLQ Cleanup\nCronJob\n24h Interval")
        
        # ========== RATE LIMITING & RESILIENCE ==========
        with Cluster("Rate Limiting & Security"):
            with Cluster("Rate Limiters"):
                rate_limiter_build = Service("Build Context\n5 req/min, burst 2")
                rate_limiter_job = Service("K8s Job\n10 req/min, burst 3")
                rate_limiter_client = Service("Client\n5 req/min, burst 2")
                rate_limiter_s3 = Service("S3 Upload\n50 req/min, burst 10")
            
            with Cluster("Security"):
                rbac = Service("RBAC\nClusterRole\nServiceAccount")
                pod_security = Service("Pod Security\nrunAsNonRoot: true\nreadOnlyRootFS: true")
                tls_certs = Service("TLS Certs\nCert Manager\nLet's Encrypt")
    
    # ========== OBSERVABILITY STACK ==========
    with Cluster("Observability Stack (Self-Hosted)"):
        with Cluster("Metrics & Monitoring"):
            prometheus = Prometheus("Prometheus\nMetrics Scraping\nAlertManager")
            grafana = Grafana("Grafana\nDashboards\nVisualization")
        
        with Cluster("Logging & Tracing"):
            loki = Prometheus("Loki\nLog Aggregation\nJSON Logs")
            tempo = Prometheus("Tempo\nDistributed Tracing\nOpenTelemetry")
            alloy = Prometheus("Alloy\nOTel Collector\nMetrics/Traces")
        
        with Cluster("Metrics Types"):
            metrics_build = Service("Build Metrics\nbuild_duration_seconds\nbuild_success_rate")
            metrics_queue = Service("Queue Metrics\nqueue_depth\nevent_processing_time")
            metrics_lambda = Service("Lambda Metrics\ncold_start_duration\nrequest_rate")
    
    # ========== NOTIFI BACKEND INTEGRATION ==========
    with Cluster("Notifi Backend Services (notifi namespace)"):
        with Cluster("Core Services"):
            scheduler = Service("Scheduler\nFusion Execution\nCallback Handler")
            subscription_mgr = Service("Subscription Manager\nUser Subscriptions\ngRPC")
            storage_mgr = Service("Storage Manager\nEphemeral/Persistent\nS3 Proxy")
            fetch_proxy = Service("Fetch Proxy\nHTTP Proxy\nFusion APIs")
            blockchain_mgr = Service("Blockchain Manager\nEVM/Solana/Sui RPC\nMulti-Chain")
    
    # ========== GITOPS & DEPLOYMENT ==========
    with Cluster("GitOps & Platform Services"):
        flux = Flux("Flux CD\nGitOps Engine\nAuto-Reconcile")
        linkerd_mesh = Linkerd("Linkerd2\nmTLS\nService Mesh")
        cert_mgr = Service("Cert Manager\nLet's Encrypt")
        sealed_secrets = Service("Sealed Secrets\nEncrypted in Git")
    
    # ========== REDIS (OPTIONAL) ==========
    with Cluster("Caching & Session (Optional)"):
        redis_cluster = Redis("Redis Sentinel\nRate Limit State\nBuild Cache")
    
    # ========== STORAGE & VOLUMES ==========
    with Cluster("Persistent Storage"):
        local_storage = StorageClass("Local Path Provisioner\nNVMe/SSD")
        pvc_kaniko = PVC("Kaniko Cache PVC\n20GB\nBuild Artifacts")
        pvc_tmp = PVC("Temp Storage PVC\n10GB\nBuild Context")
    
    # ==================== CONNECTIONS ====================
    
    # === EDGE TO BUILDER ===
    nginx_edge >> traefik >> builder_svc
    
    # === BUILDER SERVICE INTERNAL ===
    builder_svc >> event_handler
    event_handler >> [build_mgr, job_mgr, service_mgr]
    builder_svc >> Edge(label="Sidecar", style="dashed") >> metrics_pusher
    
    # === RABBITMQ CLUSTER ===
    rabbitmq_main >> Edge(label="Mirror", style="dashed") >> [rabbitmq_node2, rabbitmq_node3]
    
    # === EVENT FLOW: RABBITMQ → BUILDER ===
    rabbitmq_main >> broker_dev
    broker_dev >> api_source >> event_handler
    rabbitmq_source >> broker_dev
    
    # === BUILD FLOW: S3 → BUILDER → KANIKO ===
    s3_source >> Edge(label="Download\nParser Code") >> build_mgr
    build_mgr >> s3_tmp  # Cache build context
    build_mgr >> job_mgr
    job_mgr >> Edge(label="Create Job") >> kaniko_job
    
    # === KANIKO BUILD PROCESS ===
    kaniko_job >> kaniko_container
    kaniko_container >> dockerfile_gen
    dockerfile_gen >> [npm_config, base_images]
    harbor_local >> Edge(label="Pull\nBase Images") >> kaniko_container
    kaniko_container >> Edge(label="Push\nBuilt Image") >> ecr_prod
    sidecar_monitor >> Edge(label="Monitor", style="dashed") >> kaniko_container
    sidecar_monitor >> broker_dev  # Send build.complete event
    
    # === LAMBDA SERVICE CREATION ===
    service_mgr >> Edge(label="Create") >> lambda_service
    lambda_service >> lambda_parser1
    lambda_service >> lambda_parser2
    lambda_service >> lambda_parser3
    service_mgr >> Edge(label="Create") >> lambda_trigger
    lambda_trigger >> broker_dev
    
    # === LAMBDA AUTOSCALING ===
    lambda_autoscale >> Edge(label="Scale", style="dashed") >> [lambda_parser1, lambda_parser2, lambda_parser3]
    
    # === LAMBDA EVENT ROUTING ===
    broker_dev >> lambda_trigger >> [lambda_parser1, lambda_parser2, lambda_parser3]
    
    # === LAMBDA → NOTIFI BACKEND ===
    lambda_parser1 >> scheduler  # Callback
    lambda_parser1 >> subscription_mgr
    lambda_parser1 >> storage_mgr
    lambda_parser1 >> fetch_proxy
    lambda_parser1 >> blockchain_mgr
    
    # === DEAD LETTER QUEUE ===
    broker_dev >> Edge(label="Failed\nEvents", color="red") >> dlq_exchange
    dlq_exchange >> dlq_queue
    dlq_queue >> dlq_handler_dep
    dlq_handler_dep >> Edge(label="Retry", color="orange") >> broker_dev
    dlq_handler_svc >> dlq_handler_dep
    dlq_cleanup >> Edge(label="Cleanup\n7d", style="dashed") >> dlq_queue
    
    # === RATE LIMITING ===
    [rate_limiter_build, rate_limiter_job, rate_limiter_client, rate_limiter_s3] >> Edge(style="dashed") >> build_mgr
    redis_cluster >> Edge(label="State", style="dashed") >> [rate_limiter_build, rate_limiter_job, rate_limiter_client]
    
    # === SECURITY ===
    rbac >> builder_svc
    pod_security >> [kaniko_job, builder_svc]
    cert_mgr >> Edge(label="TLS", style="dashed") >> [builder_svc, lambda_parser1]
    
    # === OBSERVABILITY: METRICS ===
    metrics_pusher >> Edge(label="Remote Write") >> prometheus
    [builder_svc, kaniko_job, lambda_parser1] >> Edge(label="Scrape", style="dashed") >> prometheus
    prometheus >> grafana
    
    # === OBSERVABILITY: LOGGING & TRACING ===
    [builder_svc, kaniko_job, lambda_parser1, dlq_handler_dep] >> Edge(label="Logs", style="dashed") >> alloy
    alloy >> loki
    alloy >> tempo
    [loki, tempo] >> grafana
    
    # === METRICS COLLECTION ===
    [metrics_build, metrics_queue, metrics_lambda] >> prometheus
    
    # === STORAGE ===
    [pvc_kaniko, pvc_tmp] >> local_storage
    kaniko_job >> pvc_kaniko
    build_mgr >> pvc_tmp
    
    # === GITOPS ===
    flux >> Edge(label="Reconcile", style="dashed") >> builder_svc
    flux >> Edge(label="Reconcile", style="dashed") >> broker_dev
    sealed_secrets >> Edge(label="Decrypt", style="dashed") >> builder_svc
    
    # === SERVICE MESH ===
    linkerd_mesh >> Edge(label="mTLS", style="dashed") >> [builder_svc, lambda_parser1]

print("\n✅ Knative Lambda Architecture Generated!")
print("\n🚀 KNATIVE LAMBDA - SERVERLESS FUNCTION PLATFORM:")
print("━" * 80)
print("1. 🏗️  DYNAMIC FUNCTION BUILDING")
print("   • Kaniko-based secure container builds (no Docker daemon)")
print("   • Source code from S3 (Python, Node.js, Go supported)")
print("   • Automatic Dockerfile generation from templates")
print("   • Multi-language dependency resolution (npm, pip, go mod)")
print("   • Build context caching in S3 temp bucket")
print("   • Sidecar monitor for build progress tracking")
print("")
print("2. ⚡ AUTO-SCALING & PERFORMANCE")
print("   • Scale-to-zero when idle (0 resource consumption)")
print("   • Rapid scale-up: 0→N in <30s (cold start <5s)")
print("   • Concurrency-based autoscaling (Knative KPA)")
print("   • Burst handling: max 50 replicas per function")
print("   • Resource optimization: configurable CPU/memory per function")
print("   • Builder Service: scale 0→10 based on event load")
print("")
print("3. 🔄 EVENT-DRIVEN ARCHITECTURE")
print("   • CloudEvents native (standards-based)")
print("   • RabbitMQ backend with Quorum Queues (HA)")
print("   • Knative Brokers for event routing & filtering")
print("   • Knative Triggers for function-specific event subscriptions")
print("   • Event types: build.start, job.start, service.delete, custom")
print("   • APIServerSource for Kubernetes event watching")
print("   • RabbitMQSource for external event integration")
print("")
print("4. 📊 FULL OBSERVABILITY")
print("   • Prometheus metrics: build_duration, success_rate, queue_depth")
print("   • OpenTelemetry tracing: end-to-end distributed tracing")
print("   • Structured JSON logging with context propagation")
print("   • Grafana dashboards: pre-built monitoring dashboards")
print("   • Metrics Pusher sidecar: Prometheus remote write")
print("   • Loki log aggregation: centralized logging")
print("   • Tempo distributed tracing: full trace visibility")
print("   • Alloy (OTel Collector): unified telemetry")
print("")
print("5. 🔒 ENTERPRISE SECURITY")
print("   • RBAC: ClusterRole with fine-grained permissions")
print("   • Pod Security: runAsNonRoot, readOnlyRootFilesystem")
print("   • TLS/mTLS: Linkerd2 service mesh with automatic mTLS")
print("   • Cert Manager: automated TLS certificate management")
print("   • Sealed Secrets: encrypted secrets in Git")
print("   • Rate limiting: multi-level (build context, K8s jobs, client, S3)")
print("   • Image scanning: Trivy vulnerability scanning (optional)")
print("   • EKS Pod Identity: AWS IAM role for service account")
print("")
print("6. 🎯 GITOPS & CI/CD")
print("   • Flux CD: automated GitOps deployments")
print("   • Helm-based: environment-specific configs (dev/staging/prod)")
print("   • Multi-environment support: isolated namespaces")
print("   • Automated rollbacks: on failure detection")
print("   • Progressive delivery: Flagger canary deployments (optional)")
print("   • Everything in Git: full infrastructure as code")
print("")
print("7. 💾 STORAGE & REGISTRY")
print("   • S3 Source Bucket: notifi-uw2-dev-fusion-modules (parser code)")
print("   • S3 Temp Bucket: knative-lambda-dev-context-tmp (build cache)")
print("   • ECR Production: 339954290315.dkr.ecr.us-west-2 (lambda images)")
print("   • Harbor Registry: localhost:5001 (dev/base images)")
print("   • Local Path Provisioner: NVMe/SSD persistent storage")
print("   • Kaniko cache PVC: 20GB build artifact cache")
print("")
print("8. 🔄 DEAD LETTER QUEUE (DLQ)")
print("   • RabbitMQ-based DLQ with exponential backoff")
print("   • 7-day message retention (604800000ms)")
print("   • Max 50,000 messages (drop-head overflow policy)")
print("   • 5 retry attempts with exponential backoff")
print("   • Error categorization: transient vs permanent failures")
print("   • Automated cleanup: 24h interval CronJob")
print("   • Alert threshold: 1h message age, 10 queue depth")
print("")
print("9. 🔗 NOTIFI BACKEND INTEGRATION")
print("   • Scheduler: Fusion execution callbacks (HTTP)")
print("   • Subscription Manager: user subscription lookups (gRPC)")
print("   • Storage Manager: ephemeral/persistent storage (S3 proxy)")
print("   • Fetch Proxy: HTTP proxy for Fusion APIs")
print("   • Blockchain Manager: EVM/Solana/Sui RPC (multi-chain)")
print("   • Encryption key: AES-256 for sensitive data")
print("")
print("━" * 80)
print("\n📈 KEY METRICS & SCALING:")
print("   • Builder Service: 0-10 replicas, 250m CPU, 256Mi RAM")
print("   • Lambda Functions: 0-50 replicas, 50m CPU, 64Mi RAM")
print("   • Kaniko Jobs: 500m-1000m CPU, 1Gi-2Gi RAM")
print("   • Cold Start: <5s (optimized base images)")
print("   • Build Time: ~30-90s (cached dependencies)")
print("   • Scale-to-Zero Grace Period: 30s")
print("   • Event Processing: 50 parallel RabbitMQ consumers")
print("")
print("🎯 USE CASES:")
print("   • ⚡ Webhook processors: scale from 0 when webhooks arrive")
print("   • 🔄 Data pipelines: transform/enrich data from multiple sources")
print("   • 🌐 API integrations: connect to 3rd-party services dynamically")
print("   • 📊 Background jobs: scheduled/triggered tasks with auto-scaling")
print("   • 🔨 Event handlers: process events from RabbitMQ/Kafka/custom sources")
print("")
print("✅ ADVANTAGES:")
print("   • Zero infrastructure management: upload code, get running function")
print("   • Cost optimization: pay only for active request processing time")
print("   • Auto-scaling: handle traffic spikes without manual intervention")
print("   • Developer-friendly: no container expertise required")
print("   • Enterprise-grade: full observability, security, and resilience")
print("   • GitOps-native: everything versioned and auditable")
print("   • Multi-language: Python, Node.js, Go, with extensible template system")
print("")
print("⚠️  LIMITATIONS:")
print("   • Cold start latency: ~5s first request (mitigated by warm pool)")
print("   • Build time: 30-90s per function (cached dependencies help)")
print("   • Resource constraints: max 50 replicas per function")
print("   • Stateless functions: no persistent state between invocations")
print("   • Network dependency: requires S3/ECR/RabbitMQ connectivity")
print("")
print("📂 Saved to: knative-lambda-architecture.png")
print("\n🚀 Deploy the platform:")
print("   cd flux/infrastructure/knative-lambda")
print("   make build-images-local  # Build builder/sidecar/metrics-pusher images")
print("   kubectl apply -k k8s/kustomize/studio/  # Deploy to studio cluster")
print("   flux reconcile helmrelease knative-lambda-studio -n knative-lambda-dev")
print("")
print("🧪 Test with CloudEvent:")
print('   curl -X POST http://knative-lambda-builder-dev.knative-lambda-dev.svc.cluster.local:8080/cloudevents \\')
print('     -H "Content-Type: application/json" \\')
print('     -H "Ce-Id: test-123" \\')
print('     -H "Ce-Source: test-cli" \\')
print('     -H "Ce-Type: network.notifi.lambda.parser.start" \\')
print('     -H "Ce-Specversion: 1.0" \\')
print('     -d \'{"third_party_id": "test", "parser_id": "parser-123", "s3_key": "parsers/test.js"}\'')
print("")

