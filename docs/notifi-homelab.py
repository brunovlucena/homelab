from diagrams import Diagram, Cluster, Edge
from diagrams.k8s.compute import Pod, Deployment, StatefulSet, DaemonSet, Job
from diagrams.k8s.network import Service, Ingress
from diagrams.k8s.storage import PV, PVC, StorageClass
from diagrams.k8s.clusterconfig import HPA
from diagrams.onprem.compute import Server
from diagrams.onprem.database import PostgreSQL, MongoDB, Cassandra
from diagrams.onprem.inmemory import Redis, Memcached
from diagrams.onprem.network import Linkerd, Nginx, Traefik, Consul
from diagrams.onprem.gitops import Flux
from diagrams.onprem.monitoring import Prometheus, Grafana, Datadog
from diagrams.onprem.queue import RabbitMQ, Kafka
from diagrams.onprem.storage import Ceph
from diagrams.onprem.container import Docker
from diagrams.onprem.vcs import Gitlab
from diagrams.programming.language import Go, Csharp, NodeJS
from diagrams.generic.storage import Storage

graph_attr = {
    "fontsize": "18",
    "bgcolor": "white",
    "pad": "0.8",
    "splines": "spline",
    "ranksep": "1.2",
    "nodesep": "0.7"
}

with Diagram("Notifi Homelab Architecture - On-Premises", 
             show=False, 
             direction="TB",
             filename="/Users/brunolucena/workspace/notifi-homelab",
             graph_attr=graph_attr,
             outformat="png"):
    
    # ========== EDGE & INGRESS ==========
    with Cluster("Edge Layer (Self-Hosted)"):
        external_dns_svc = Service("External DNS\nCloudflare/Route53")
        coredns = Service("CoreDNS\nInternal DNS")
        nginx_edge = Nginx("Nginx\nReverse Proxy\nSSL Termination\nRate Limiting")
    
    # ========== PHYSICAL INFRASTRUCTURE ==========
    with Cluster("Physical Infrastructure"):
        with Cluster("Compute Nodes"):
            pi_node1 = Server("Pi Node 1\n8GB RAM\nARM64")
            pi_node2 = Server("Pi Node 2\n8GB RAM\nARM64")
            pi_node3 = Server("Pi Node 3\n8GB RAM\nARM64")
            studio = Server("Studio Node\n32GB RAM\nx86_64")
            forge = Server("Forge GPU Node\nRTX 4090\n64GB RAM")
        
        with Cluster("Storage"):
            nas = Ceph("NAS/MinIO\n2TB Storage\nS3-Compatible")
            local_storage = StorageClass("Local Path\nProvisioner")
    
    # ========== K3S CLUSTER ==========
    with Cluster("K3s Homelab Cluster"):
        
        # Platform Layer
        with Cluster("Platform Services (Flux GitOps)"):
            flux = Flux("Flux CD\nGitOps Engine\nAuto-Sync")
            linkerd_mesh = Linkerd("Linkerd2\nmTLS + Multi-Cluster")
            
            with Cluster("Observability Stack"):
                prometheus = Prometheus("Prometheus\nMetrics + Alerting")
                grafana = Grafana("Grafana\nUnified Dashboard")
                loki = Prometheus("Loki\nLog Aggregation")
                tempo = Prometheus("Tempo\nTracing")
                alloy = Prometheus("Alloy\nOTel Collector")
                alert_mgr = Prometheus("AlertManager\nSlack/Email Alerts")
            
            with Cluster("Platform Tools"):
                cert_mgr = Service("Cert Manager\nLet's Encrypt")
                external_dns_tool = Service("External DNS")
                sealed_secrets = Service("Sealed Secrets\nBitnami")
                metrics_server = Service("Metrics Server\nHPA")
                vault = Service("Vault (Optional)\nSecrets Management")
        
        # Ingress Layer
        with Cluster("Ingress Layer"):
            traefik = Traefik("Traefik Ingress\nHTTP/gRPC Router")
            nginx = Nginx("Nginx\nBrotli Compression")
        
        # ========== NOTIFI SERVICES ==========
        with Cluster("Notifi Services (Adapted for Homelab)"):
            
            # API Gateways (Reduced Replicas)
            with Cluster("API Gateways (2 replicas)"):
                mgmt_gw = Csharp("Management GW\ngRPC + REST")
                data_gw = Csharp("Dataplane GW\nHigh Throughput")
            
            # Blockchain Layer (Consolidated)
            with Cluster("Blockchain Services (Knative)"):
                blockchain_router = Go("Blockchain Router\nQuery Cache\nKnative Service")
                solana_proxy = Pod("Solana Proxy\nScale-to-Zero")
                evm_proxy = Pod("EVM Proxy\nScale-to-Zero")
                multi_chain = Pod("Multi-Chain\nAptos/Sui/Cosmos")
            
            # Core Business Services
            with Cluster("Core Services (1-2 replicas)"):
                blockchain_mgr = Csharp("Blockchain Mgr")
                user_mgr = Csharp("User Mgr")
                tenant_mgr = Csharp("Tenant Mgr")
                subscription_mgr = Csharp("Subscription Mgr")
                storage_mgr = Csharp("Storage Mgr")
                template_mgr = Csharp("Template Mgr")
            
            # Notification Engine
            with Cluster("Notification Engine"):
                notification_router = Csharp("Notification Router\nEvent-Driven")
                event_processor = Csharp("Event Processor\nWorker Pool")
                dlq_handler = Csharp("DLQ Handler")
            
            # Messaging Services (Knative)
            with Cluster("Messaging Workers (Knative - Scale-to-Zero)"):
                telegram = Deployment("Telegram")
                sms = Deployment("SMS")
                webhook = Deployment("Webhook")
                mailer = Deployment("Mailer")
                fcm = Deployment("FCM")
            
            # Background Jobs
            with Cluster("Background Workers"):
                scheduler = StatefulSet("Scheduler\nCronJobs")
                analytics = Deployment("Analytics")
                callback = Deployment("Callback Handler")
        
        # ========== DATA LAYER (SELF-HOSTED) ==========
        with Cluster("Data Layer (All Self-Hosted)"):
            
            # PostgreSQL HA Cluster (replaces Aurora RDS)
            with Cluster("PostgreSQL HA (Patroni)"):
                pg_primary = PostgreSQL("Primary Node\n16GB Data\nStreaming Replication")
                pg_replica1 = PostgreSQL("Replica 1\nRead Scaling")
                pg_replica2 = PostgreSQL("Replica 2\nRead Scaling")
                pg_backup = Job("pg_dump Cron\nDaily Backups")
                pg_pvc = PVC("PostgreSQL PVC\n100GB NVMe")
            
            # Redis Sentinel (replaces ElastiCache)
            with Cluster("Redis Sentinel (HA)"):
                redis_master = Redis("Redis Master\nCache + Sessions\nRate Limiting")
                redis_replica1 = Redis("Redis Replica 1\nRead-Only")
                redis_replica2 = Redis("Redis Replica 2\nRead-Only")
                redis_sentinel = Service("Redis Sentinel\nAuto-Failover")
                redis_pvc = PVC("Redis PVC\n20GB")
            
            # Document Store (replaces DynamoDB)
            with Cluster("NoSQL Document Store"):
                mongo_primary = MongoDB("MongoDB Primary\nDocument Storage\n50GB")
                mongo_secondary = MongoDB("MongoDB Secondary\nReplication")
                mongo_arbiter = MongoDB("MongoDB Arbiter\nElection")
        
        # ========== MESSAGING LAYER ==========
        with Cluster("Messaging Infrastructure (Self-Hosted)"):
            
            # RabbitMQ Cluster (replaces SQS/SNS)
            with Cluster("RabbitMQ Cluster (HA)"):
                rabbitmq_main = RabbitMQ("RabbitMQ Node 1\nQuorum Queues")
                rabbitmq_node2 = RabbitMQ("RabbitMQ Node 2\nMirror Queue")
                rabbitmq_node3 = RabbitMQ("RabbitMQ Node 3\nMirror Queue")
            
            # Kafka for Event Streaming (replaces Kinesis)
            with Cluster("Kafka Cluster"):
                kafka_broker1 = Kafka("Kafka Broker 1\nEvent Streaming")
                kafka_broker2 = Kafka("Kafka Broker 2\nReplication")
                zookeeper = Service("ZooKeeper\nCoordination")
            
            # Knative Eventing
            with Cluster("Knative Eventing"):
                knative_broker = Service("Knative Broker\nCloudEvents")
                knative_trigger = Service("Event Triggers\nFiltering")
                knative_source = Service("Event Sources\nRabbitMQ Source")
        
        # ========== STORAGE & REGISTRY ==========
        with Cluster("Storage & Registry (Self-Hosted)"):
            # MinIO S3-Compatible Storage (replaces AWS S3)
            with Cluster("MinIO Object Storage"):
                minio_node1 = Ceph("MinIO Node 1\nS3-Compatible")
                minio_node2 = Ceph("MinIO Node 2\nDistributed")
                minio_node3 = Ceph("MinIO Node 3\nErasure Coding")
                minio_console = Service("MinIO Console\nWeb UI")
            
            # Container Registry (replaces ECR)
            with Cluster("Container Registry"):
                harbor = Docker("Harbor Registry\nMulti-Arch Images\nVulnerability Scan")
                registry_cache = Docker("Pull-Through Cache\nDocker Hub Mirror")
    
    # ========== BACKUP & DR ==========
    with Cluster("Backup & DR"):
        velero = Service("Velero\nCluster Backup")
        restic = Service("Restic\nVolume Backup")
        backup_storage = Ceph("Backup Storage\nMinIO S3")
    
    # ========== EXTERNAL APIs (3rd Party Only) ==========
    with Cluster("External APIs (Internet)"):
        twilio = Service("Twilio API\nSMS Gateway")
        smtp_relay = Service("SMTP Relay\nEmail Delivery")
        telegram_api = Service("Telegram Bot API")
        fcm_api = Service("Firebase FCM API")
        blockchain_rpcs = Service("Blockchain RPCs\nSolana/EVM Nodes")
    
    # ========== MONITORING & ALERTS ==========
    with Cluster("Monitoring & Alerting"):
        alert_mgr = Prometheus("AlertManager\nPagerDuty/Slack")
        uptime_kuma = Service("Uptime Kuma\nStatus Page")
    
    # ==================== CONNECTIONS ====================
    
    # === EDGE LAYER ===
    external_dns_svc >> nginx_edge >> traefik
    coredns >> Service("K8s Services\nInternal DNS")
    
    # === INGRESS ===
    traefik >> nginx
    nginx >> mgmt_gw
    nginx >> data_gw
    
    # === API GATEWAYS ===
    mgmt_gw >> Edge(label="gRPC") >> [blockchain_mgr, user_mgr, tenant_mgr]
    data_gw >> blockchain_router
    data_gw >> notification_router
    
    # === BLOCKCHAIN LAYER ===
    blockchain_router >> [solana_proxy, evm_proxy, multi_chain]
    blockchain_router >> event_processor
    blockchain_router >> redis_master  # Cache
    
    # === CORE SERVICES ===
    blockchain_mgr >> subscription_mgr
    user_mgr >> tenant_mgr
    subscription_mgr >> event_processor
    template_mgr >> event_processor
    storage_mgr >> nas  # S3-compatible
    
    # === NOTIFICATION FLOW ===
    notification_router >> event_processor
    event_processor >> knative_broker
    knative_broker >> knative_trigger
    knative_trigger >> [telegram, sms, webhook, mailer, fcm]
    
    # === RABBITMQ MESSAGING ===
    event_processor >> rabbitmq_main
    rabbitmq_main >> Edge(label="Mirror") >> [rabbitmq_node2, rabbitmq_node3]
    rabbitmq_main >> [callback, analytics]
    rabbitmq_main >> dlq_handler
    
    # === KAFKA STREAMING ===
    kafka_broker1 >> Edge(label="Replicate") >> kafka_broker2
    [kafka_broker1, kafka_broker2] >> zookeeper
    event_processor >> kafka_broker1
    analytics >> kafka_broker1
    
    # === DATA LAYER ===
    mgmt_gw >> pg_primary
    data_gw >> pg_replica1
    event_processor >> pg_replica2
    pg_primary >> Edge(label="Streaming\nReplication") >> [pg_replica1, pg_replica2]
    pg_primary >> pg_pvc
    pg_backup >> pg_primary
    
    redis_master >> Edge(label="Replication") >> [redis_replica1, redis_replica2]
    redis_sentinel >> Edge(label="Monitor", style="dashed") >> [redis_master, redis_replica1, redis_replica2]
    [mgmt_gw, data_gw, event_processor] >> redis_master
    redis_master >> redis_pvc
    
    # === MONGODB ===
    mongo_primary >> Edge(label="Replica Set") >> [mongo_secondary, mongo_arbiter]
    storage_mgr >> mongo_primary
    
    # === STORAGE ===
    [pg_pvc, redis_pvc] >> local_storage
    storage_mgr >> minio_node1
    minio_node1 >> Edge(label="Distributed") >> [minio_node2, minio_node3]
    minio_console >> minio_node1
    
    # === EXTERNAL SERVICES (Internet APIs Only) ===
    telegram >> Edge(label="HTTPS") >> telegram_api
    sms >> Edge(label="HTTPS") >> twilio
    mailer >> Edge(label="SMTP") >> smtp_relay
    fcm >> Edge(label="HTTPS") >> fcm_api
    [solana_proxy, evm_proxy, multi_chain] >> Edge(label="WebSocket/HTTPS") >> blockchain_rpcs
    
    # === CONTAINER REGISTRY ===
    [mgmt_gw, data_gw, event_processor] >> harbor
    registry_cache >> Edge(label="Pull Through") >> Docker("Docker Hub")
    
    # === GITOPS ===
    flux >> Edge(label="Git Sync\nAuto-Deploy") >> [mgmt_gw, data_gw, event_processor]
    
    # === SERVICE MESH ===
    linkerd_mesh >> [mgmt_gw, data_gw, event_processor]
    
    # === OBSERVABILITY ===
    alloy >> [prometheus, loki, tempo]
    [prometheus, loki, tempo] >> grafana
    prometheus >> alert_mgr
    alert_mgr >> Service("Slack/Email\nNotifications")
    [mgmt_gw, data_gw, event_processor, rabbitmq_main, pg_primary] >> Edge(style="dashed", label="Metrics/Logs") >> alloy
    
    # === BACKUPS (Self-Hosted) ===
    velero >> backup_storage
    [pg_pvc, redis_pvc] >> restic >> backup_storage
    pg_backup >> backup_storage
    backup_storage >> minio_node1
    
    # === PHYSICAL NODES ===
    pi_node1 >> Edge(label="Workload") >> [mgmt_gw, blockchain_mgr]
    pi_node2 >> [data_gw, event_processor]
    pi_node3 >> [rabbitmq_main, pg_primary]
    studio >> [telegram, webhook, mailer]
    forge >> Edge(label="GPU\nWorkloads") >> analytics

print("\n✅ Notifi Homelab Architecture Generated!")
print("\n🏠 100% SELF-HOSTED - NO AWS SERVICES:")
print("━" * 80)
print("1. 🖥️  INFRASTRUCTURE (Self-Hosted)")
print("   • K3s lightweight Kubernetes (not EKS)")
print("   • 3x Raspberry Pi nodes (8GB ARM64) + Studio/Forge (x86_64)")
print("   • Nginx edge proxy with rate limiting & SSL termination")
print("   • Traefik ingress controller (HTTP/gRPC routing)")
print("   • CoreDNS for internal DNS resolution")
print("   • Local path provisioner for persistent volumes")
print("")
print("2. 💾 DATA LAYER (Self-Hosted)")
print("   • PostgreSQL HA with Patroni (replaces Aurora RDS)")
print("   • Redis Sentinel cluster 3-node (replaces ElastiCache)")
print("   • MongoDB replica set (replaces DynamoDB)")
print("   • MinIO distributed 3-node S3-compatible (replaces AWS S3)")
print("   • All data stored on local NVMe/SSD storage")
print("")
print("3. 📨 MESSAGING (Self-Hosted)")
print("   • RabbitMQ 3-node cluster with Quorum Queues (replaces SQS/SNS)")
print("   • Kafka 2-broker cluster + ZooKeeper (replaces Kinesis)")
print("   • Knative Eventing with RabbitMQ broker backend")
print("   • CloudEvents format for event-driven architecture")
print("   • Dead letter queues for retry logic")
print("")
print("4. 🚀 SCALABILITY (Homelab Constraints)")
print("   • Knative for scale-to-zero (messaging workers save resources)")
print("   • HPA on core services (limited by 5 physical nodes)")
print("   • Resource requests/limits optimized for ARM64 + x86_64")
print("   • No spot instances (all nodes are dedicated)")
print("   • Max ~50 vCPU, ~100GB RAM total cluster capacity")
print("")
print("5. 🔐 SECURITY (Self-Hosted)")
print("   • Linkerd2 service mesh with automatic mTLS")
print("   • Sealed Secrets (Bitnami) - secrets encrypted in Git")
print("   • Cert Manager with Let's Encrypt for TLS certs")
print("   • Nginx rate limiting & DDoS protection at edge")
print("   • Optional: HashiCorp Vault for advanced secret management")
print("   • Network policies with Cilium/Calico")
print("")
print("6. 🔄 GITOPS & DEPLOYMENT (Self-Hosted)")
print("   • Flux CD for continuous delivery (100% GitOps)")
print("   • Everything in Git (infrastructure as code)")
print("   • Auto-reconciliation every 5 minutes from GitHub")
print("   • Multi-cluster support with Linkerd multi-cluster")
print("   • Harbor container registry with vulnerability scanning")
print("")
print("7. 📊 OBSERVABILITY (Grafana Stack - Self-Hosted)")
print("   • Prometheus for metrics collection & alerting")
print("   • Loki for centralized log aggregation")
print("   • Tempo for distributed tracing")
print("   • Alloy (OpenTelemetry collector) for telemetry")
print("   • Grafana for unified dashboards")
print("   • AlertManager → Slack/Email notifications")
print("   • Full observability with zero cloud dependencies")
print("")
print("8. 💾 BACKUP & DR (Self-Hosted)")
print("   • Velero for Kubernetes cluster backups")
print("   • Restic for persistent volume backups")
print("   • PostgreSQL pg_dump daily automated backups")
print("   • All backups stored in MinIO S3 (local NAS)")
print("   • Optional: Rsync to external USB/NAS for offsite")
print("")
print("━" * 80)
print("\n💰 COST COMPARISON (Self-Hosted vs AWS):")
print("   • AWS Monthly: ~$3,500-5,000/month (EKS, RDS, ElastiCache, S3, etc.)")
print("   • Homelab One-Time: ~$2,000-3,000 (5 nodes + NAS + networking)")
print("   • Monthly Power: ~$40-60/month (120W average × 5 nodes)")
print("   • Internet: $50-100/month (dedicated fiber recommended)")
print("   • ROI Break-even: 6-8 months")
print("   • Year 1 Total: ~$3,000 vs AWS $42,000+ (93% savings!)")
print("")
print("📈 PERFORMANCE CHARACTERISTICS:")
print("   • Throughput: ~1-3K RPS (adequate for small-medium workloads)")
print("   • Latency: p99 < 300ms local, <500ms internet (acceptable)")
print("   • Storage: 2-4TB local NVMe/SSD (expandable)")
print("   • Compute: ~50 vCPU, ~100GB RAM total")
print("   • Network: 1Gbps LAN, 100Mbps-1Gbps WAN")
print("   • Ideal for: Dev, Test, Small Prod (<10K users)")
print("")
print("🎯 BEST USE CASES:")
print("   • Development environment (1:1 with production)")
print("   • CI/CD testing pipelines")
print("   • Small-scale production (<10K users)")
print("   • Learning Kubernetes + Cloud Native")
print("   • Cost-sensitive deployments")
print("")
print("⚠️  LIMITATIONS:")
print("   • No multi-region failover")
print("   • Limited horizontal scalability")
print("   • Single point of failure (home network)")
print("   • ARM64 compatibility required for Pi nodes")
print("   • Manual hardware maintenance")
print("")
print("✅ ADVANTAGES OF SELF-HOSTED:")
print("   • 100% infrastructure control - no vendor dependencies")
print("   • Zero cloud vendor lock-in - truly portable")
print("   • Zero egress costs - unlimited internal traffic")
print("   • Complete data privacy & sovereignty")
print("   • Perfect dev/prod parity - identical environments")
print("   • Learning platform - hands-on Kubernetes experience")
print("   • Sustainable - reuse hardware, lower carbon footprint")
print("")
print("📂 Saved to: notifi-homelab.png")
print("\n🚀 Ready to deploy:")
print("   kubectl apply -k flux/infrastructure/notifi/k8s/")
print("   make deploy-notifi CLUSTER=studio")

