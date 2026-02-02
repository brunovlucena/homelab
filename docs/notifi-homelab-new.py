#!/usr/bin/env python3
"""
Notifi Homelab Architecture Diagram - Refactored Version
========================================================

Senior Cloud Architect Improvements:
- Modular, maintainable code structure
- Configuration-driven approach
- Reusable component functions
- Clear separation of concerns
- Better documentation
- Color-coded traffic flows
- Improved layer abstraction
"""

from diagrams import Diagram, Cluster, Edge
from diagrams.k8s.compute import Pod, Deployment, StatefulSet, Job
from diagrams.k8s.network import Service
from diagrams.k8s.storage import PVC, StorageClass
from diagrams.onprem.compute import Server
from diagrams.onprem.database import PostgreSQL, MongoDB
from diagrams.onprem.inmemory import Redis
from diagrams.onprem.network import Linkerd, Nginx, Traefik
from diagrams.onprem.gitops import Flux
from diagrams.onprem.monitoring import Prometheus, Grafana
from diagrams.onprem.queue import RabbitMQ, Kafka
from diagrams.onprem.storage import Ceph
from diagrams.onprem.container import Docker
from diagrams.programming.language import Go, Csharp
from typing import Dict, List, Tuple

# ============================================================================
# CONFIGURATION
# ============================================================================

DIAGRAM_CONFIG = {
    "name": "Notifi Homelab Architecture - Refactored",
    "filename": "/Users/brunolucena/workspace/bruno/repos/homelab/docs/notifi-homelab-new",
    "direction": "TB",
    "graph_attr": {
        "fontsize": "18",
        "bgcolor": "white",
        "pad": "0.8",
        "splines": "ortho",  # Changed to ortho for cleaner lines
        "ranksep": "1.5",
        "nodesep": "0.8",
        "compound": "true",
    }
}

# Traffic flow colors for better visualization
TRAFFIC_COLORS = {
    "external": "#FF6B6B",      # Red - External traffic
    "internal": "#4ECDC4",      # Teal - Internal service traffic
    "data": "#95E1D3",          # Light teal - Data replication
    "control": "#FFA07A",       # Light coral - Control plane
    "observability": "#DDA15E", # Brown - Metrics/logs
    "storage": "#9B59B6",       # Purple - Storage operations
}

# Edge styles for different traffic patterns
EDGE_STYLES = {
    "sync": {"style": "solid"},
    "async": {"style": "dashed", "color": TRAFFIC_COLORS["internal"]},
    "replication": {"style": "dotted", "color": TRAFFIC_COLORS["data"]},
    "monitoring": {"style": "dashed", "color": TRAFFIC_COLORS["observability"]},
    "control": {"style": "dotted", "color": TRAFFIC_COLORS["control"]},
}

# ============================================================================
# COMPONENT FACTORIES
# ============================================================================

def create_edge_layer() -> Dict:
    """Create edge networking components"""
    with Cluster("Edge Layer (L7)"):
        nginx_edge = Nginx("Nginx Edge\nSSL Termination\nRate Limiting\nDDoS Protection")
        external_dns = Service("External DNS\nCloudflare DNS")
        coredns = Service("CoreDNS\nInternal DNS")
    
    return {
        "nginx": nginx_edge,
        "external_dns": external_dns,
        "coredns": coredns
    }

def create_physical_infrastructure() -> Dict:
    """Create physical hardware nodes"""
    with Cluster("Physical Infrastructure (Bare Metal)"):
        with Cluster("Compute Nodes (5 nodes)"):
            pi_nodes = [
                Server(f"Pi Node {i}\n8GB RAM\nARM64\nWorker") 
                for i in range(1, 4)
            ]
            studio = Server("Studio Node\n32GB RAM\nx86_64\nWorker + Control")
            forge = Server("Forge Node\nRTX 4090\n64GB RAM\nGPU Workloads")
        
        with Cluster("Storage Layer"):
            nas = Ceph("MinIO S3\n2TB NVMe\nDistributed Storage")
            local_pv = StorageClass("Local Path\nProvisioner\nPer-Node Storage")
    
    return {
        "pi_nodes": pi_nodes,
        "studio": studio,
        "forge": forge,
        "nas": nas,
        "local_pv": local_pv
    }

def create_platform_services() -> Dict:
    """Create platform-level services (GitOps, Service Mesh, Observability)"""
    with Cluster("Platform Services Layer"):
        # GitOps
        with Cluster("GitOps & Deployment"):
            flux = Flux("Flux CD\nGitOps Controller\nAuto-Reconcile 5m")
            linkerd = Linkerd("Linkerd2\nmTLS Service Mesh\nMulti-Cluster")
        
        # Observability Stack
        with Cluster("Observability Stack (LGTM)"):
            prometheus = Prometheus("Prometheus\nMetrics + Alerts\n30d Retention")
            grafana = Grafana("Grafana\nUnified Dashboard\nAlerting")
            loki = Prometheus("Loki\nLog Aggregation\n7d Retention")
            tempo = Prometheus("Tempo\nDistributed Tracing\n24h Retention")
            alloy = Prometheus("Alloy\nOTel Collector\nAgent on All Nodes")
        
        # Platform Tools
        with Cluster("Platform Tools"):
            cert_mgr = Service("Cert Manager\nLet's Encrypt\nAuto-Renewal")
            sealed_secrets = Service("Sealed Secrets\nEncrypted in Git\nBitnami")
            metrics_server = Service("Metrics Server\nHPA Support")
    
    return {
        "flux": flux,
        "linkerd": linkerd,
        "prometheus": prometheus,
        "grafana": grafana,
        "loki": loki,
        "tempo": tempo,
        "alloy": alloy,
        "cert_mgr": cert_mgr,
        "sealed_secrets": sealed_secrets,
        "metrics_server": metrics_server,
    }

def create_ingress_layer() -> Dict:
    """Create ingress controllers"""
    with Cluster("Ingress Layer (L7 Routing)"):
        traefik = Traefik("Traefik\nHTTP/gRPC Router\nMiddleware Chain")
        nginx = Nginx("Nginx\nBrotli Compression\nCaching")
    
    return {"traefik": traefik, "nginx": nginx}

def create_application_layer() -> Dict:
    """Create Notifi application services"""
    with Cluster("Notifi Application Layer"):
        # API Gateways
        with Cluster("API Gateway Tier"):
            mgmt_gw = Csharp("Management Gateway\ngRPC + REST\n2 replicas")
            data_gw = Csharp("Dataplane Gateway\nHigh Throughput\n2 replicas")
        
        # Blockchain Layer
        with Cluster("Blockchain Services (Knative Scale-to-Zero)"):
            blockchain_router = Go("Blockchain Router\nQuery Cache\nKnative")
            blockchain_workers = [
                Pod("Solana Proxy\nmin=0 max=10"),
                Pod("EVM Proxy\nmin=0 max=10"),
                Pod("Multi-Chain\nAptos/Sui/Cosmos"),
            ]
        
        # Core Business Services
        with Cluster("Core Business Services"):
            core_services = [
                Csharp("Blockchain Mgr"),
                Csharp("User Mgr"),
                Csharp("Tenant Mgr"),
                Csharp("Subscription Mgr"),
                Csharp("Storage Mgr"),
                Csharp("Template Mgr"),
            ]
        
        # Notification Engine
        with Cluster("Notification Engine"):
            notification_router = Csharp("Notification Router\nEvent-Driven")
            event_processor = Csharp("Event Processor\nWorker Pool")
            dlq_handler = Csharp("DLQ Handler\nRetry Logic")
        
        # Messaging Workers (Knative)
        with Cluster("Messaging Workers (Knative)"):
            messaging_workers = [
                Deployment(f"{name} Worker")
                for name in ["Telegram", "SMS", "Webhook", "Email", "FCM"]
            ]
        
        # Background Jobs
        with Cluster("Background Jobs"):
            scheduler = StatefulSet("Scheduler\nCron Jobs")
            analytics = Deployment("Analytics\nData Processing")
    
    return {
        "mgmt_gw": mgmt_gw,
        "data_gw": data_gw,
        "blockchain_router": blockchain_router,
        "blockchain_workers": blockchain_workers,
        "core_services": core_services,
        "notification_router": notification_router,
        "event_processor": event_processor,
        "dlq_handler": dlq_handler,
        "messaging_workers": messaging_workers,
        "scheduler": scheduler,
        "analytics": analytics,
    }

def create_data_layer() -> Dict:
    """Create data persistence layer"""
    with Cluster("Data Layer (Self-Hosted HA)"):
        # PostgreSQL HA
        with Cluster("PostgreSQL HA Cluster (Patroni)"):
            pg_primary = PostgreSQL("Primary\nStreaming Replication\n100GB")
            pg_replicas = [PostgreSQL(f"Replica {i}\nRead Scaling") for i in [1, 2]]
            pg_backup = Job("pg_dump\nDaily Backups\nto MinIO")
            pg_pvc = PVC("PostgreSQL PVC\n100GB NVMe")
        
        # Redis Sentinel
        with Cluster("Redis Sentinel Cluster"):
            redis_master = Redis("Master\nCache + Sessions\nRate Limiting")
            redis_replicas = [Redis(f"Replica {i}\nRead-Only") for i in [1, 2]]
            redis_sentinel = Service("Sentinel\nAuto-Failover\n3 nodes")
            redis_pvc = PVC("Redis PVC\n20GB")
        
        # MongoDB Replica Set
        with Cluster("MongoDB Replica Set"):
            mongo_primary = MongoDB("Primary\nDocument Store\n50GB")
            mongo_secondary = MongoDB("Secondary\nReplication")
            mongo_arbiter = MongoDB("Arbiter\nElection Only")
    
    return {
        "pg_primary": pg_primary,
        "pg_replicas": pg_replicas,
        "pg_backup": pg_backup,
        "pg_pvc": pg_pvc,
        "redis_master": redis_master,
        "redis_replicas": redis_replicas,
        "redis_sentinel": redis_sentinel,
        "redis_pvc": redis_pvc,
        "mongo_primary": mongo_primary,
        "mongo_secondary": mongo_secondary,
        "mongo_arbiter": mongo_arbiter,
    }

def create_messaging_layer() -> Dict:
    """Create messaging infrastructure"""
    with Cluster("Messaging Layer (Event-Driven)"):
        # RabbitMQ Cluster
        with Cluster("RabbitMQ HA Cluster"):
            rabbitmq_nodes = [
                RabbitMQ(f"RabbitMQ Node {i}\nQuorum Queues")
                for i in [1, 2, 3]
            ]
        
        # Kafka Cluster
        with Cluster("Kafka Event Streaming"):
            kafka_brokers = [Kafka(f"Kafka Broker {i}") for i in [1, 2]]
            zookeeper = Service("ZooKeeper\nCoordination")
        
        # Knative Eventing
        with Cluster("Knative Eventing"):
            knative_broker = Service("Knative Broker\nCloudEvents")
            knative_triggers = Service("Event Triggers\nFiltering + Routing")
    
    return {
        "rabbitmq_nodes": rabbitmq_nodes,
        "kafka_brokers": kafka_brokers,
        "zookeeper": zookeeper,
        "knative_broker": knative_broker,
        "knative_triggers": knative_triggers,
    }

def create_storage_registry() -> Dict:
    """Create storage and container registry"""
    with Cluster("Storage & Registry"):
        # MinIO S3
        with Cluster("MinIO S3-Compatible Storage"):
            minio_nodes = [
                Ceph(f"MinIO Node {i}\nDistributed\nErasure Coding")
                for i in [1, 2, 3]
            ]
            minio_console = Service("MinIO Console\nWeb UI")
        
        # Container Registry
        with Cluster("Container Registry (Harbor)"):
            harbor = Docker("Harbor Registry\nMulti-Arch Images\nVuln Scanning")
            registry_cache = Docker("Pull-Through Cache\nDocker Hub Mirror")
    
    return {
        "minio_nodes": minio_nodes,
        "minio_console": minio_console,
        "harbor": harbor,
        "registry_cache": registry_cache,
    }

def create_backup_dr() -> Dict:
    """Create backup and disaster recovery"""
    with Cluster("Backup & Disaster Recovery"):
        velero = Service("Velero\nCluster Backup\nScheduled")
        restic = Service("Restic\nVolume Backup\nIncremental")
        backup_storage = Ceph("Backup Target\nMinIO S3\nOffsite Sync")
    
    return {
        "velero": velero,
        "restic": restic,
        "backup_storage": backup_storage,
    }

def create_external_apis() -> Dict:
    """Create external API integrations"""
    with Cluster("External APIs (Internet)"):
        twilio = Service("Twilio API\nSMS Gateway")
        smtp = Service("SMTP Relay\nEmail Delivery")
        telegram = Service("Telegram Bot API")
        fcm = Service("Firebase FCM")
        blockchain_rpcs = Service("Blockchain RPCs\nPublic Nodes")
    
    return {
        "twilio": twilio,
        "smtp": smtp,
        "telegram": telegram,
        "fcm": fcm,
        "blockchain_rpcs": blockchain_rpcs,
    }

# ============================================================================
# CONNECTION BUILDERS
# ============================================================================

def connect_edge_to_ingress(edge: Dict, ingress: Dict):
    """Connect edge layer to ingress"""
    edge["nginx"] >> Edge(label="HTTPS", color=TRAFFIC_COLORS["external"]) >> ingress["traefik"]

def connect_ingress_to_apps(ingress: Dict, apps: Dict):
    """Connect ingress to application gateways"""
    ingress["traefik"] >> Edge(label="gRPC/HTTP", color=TRAFFIC_COLORS["internal"]) >> ingress["nginx"]
    ingress["nginx"] >> Edge(color=TRAFFIC_COLORS["internal"]) >> [apps["mgmt_gw"], apps["data_gw"]]

def connect_apps_to_data(apps: Dict, data: Dict):
    """Connect applications to data layer"""
    # PostgreSQL connections
    apps["mgmt_gw"] >> Edge(label="Write", color=TRAFFIC_COLORS["data"]) >> data["pg_primary"]
    apps["data_gw"] >> Edge(label="Read", color=TRAFFIC_COLORS["data"]) >> data["pg_replicas"][0]
    
    # Redis connections
    [apps["mgmt_gw"], apps["data_gw"], apps["event_processor"]] >> \
        Edge(label="Cache", color=TRAFFIC_COLORS["data"]) >> data["redis_master"]

def connect_messaging_flows(apps: Dict, messaging: Dict):
    """Connect messaging flows"""
    # RabbitMQ
    apps["event_processor"] >> Edge(label="Events", **EDGE_STYLES["async"]) >> messaging["rabbitmq_nodes"][0]
    
    # Kafka
    apps["analytics"] >> Edge(label="Stream", **EDGE_STYLES["async"]) >> messaging["kafka_brokers"][0]
    
    # Knative Eventing
    apps["event_processor"] >> messaging["knative_broker"]
    messaging["knative_broker"] >> messaging["knative_triggers"]
    messaging["knative_triggers"] >> Edge(label="CloudEvents") >> apps["messaging_workers"]

def connect_observability(platform: Dict, apps: Dict):
    """Connect observability flows"""
    # Metrics collection
    [apps["mgmt_gw"], apps["data_gw"], apps["event_processor"]] >> \
        Edge(**EDGE_STYLES["monitoring"]) >> platform["alloy"]
    
    platform["alloy"] >> Edge(**EDGE_STYLES["monitoring"]) >> [
        platform["prometheus"],
        platform["loki"],
        platform["tempo"]
    ]
    
    [platform["prometheus"], platform["loki"], platform["tempo"]] >> platform["grafana"]

def connect_gitops_flows(platform: Dict, apps: Dict):
    """Connect GitOps deployment flows"""
    platform["flux"] >> Edge(label="Reconcile", **EDGE_STYLES["control"]) >> [
        apps["mgmt_gw"],
        apps["data_gw"],
        apps["event_processor"]
    ]

def connect_data_replication(data: Dict):
    """Connect data replication flows"""
    # PostgreSQL replication
    data["pg_primary"] >> Edge(label="Streaming\nReplication", **EDGE_STYLES["replication"]) >> data["pg_replicas"]
    
    # Redis replication
    data["redis_master"] >> Edge(label="Replication", **EDGE_STYLES["replication"]) >> data["redis_replicas"]
    data["redis_sentinel"] >> Edge(label="Monitor", **EDGE_STYLES["monitoring"]) >> \
        [data["redis_master"]] + data["redis_replicas"]
    
    # MongoDB replication
    data["mongo_primary"] >> Edge(label="Replica Set", **EDGE_STYLES["replication"]) >> \
        [data["mongo_secondary"], data["mongo_arbiter"]]

def connect_storage_flows(apps: Dict, storage: Dict, data: Dict):
    """Connect storage flows"""
    # MinIO storage
    [apps["core_services"][4]] >> Edge(label="S3 API", color=TRAFFIC_COLORS["storage"]) >> storage["minio_nodes"][0]
    
    # MinIO distributed storage
    storage["minio_nodes"][0] >> Edge(label="Distributed", **EDGE_STYLES["replication"]) >> \
        storage["minio_nodes"][1:]

def connect_backup_flows(backup: Dict, data: Dict, storage: Dict):
    """Connect backup and DR flows"""
    # Velero backups
    backup["velero"] >> Edge(label="Cluster Backup") >> backup["backup_storage"]
    
    # PostgreSQL backups
    data["pg_backup"] >> Edge(label="Daily Backup") >> backup["backup_storage"]
    
    # PVC backups
    [data["pg_pvc"], data["redis_pvc"]] >> backup["restic"] >> backup["backup_storage"]
    
    # Backup to MinIO
    backup["backup_storage"] >> storage["minio_nodes"][0]

def connect_external_apis(apps: Dict, external: Dict):
    """Connect to external APIs"""
    # Messaging workers to external APIs
    apps["messaging_workers"][0] >> Edge(label="HTTPS", color=TRAFFIC_COLORS["external"]) >> external["telegram"]
    apps["messaging_workers"][1] >> Edge(label="HTTPS", color=TRAFFIC_COLORS["external"]) >> external["twilio"]
    apps["messaging_workers"][3] >> Edge(label="SMTP", color=TRAFFIC_COLORS["external"]) >> external["smtp"]
    apps["messaging_workers"][4] >> Edge(label="HTTPS", color=TRAFFIC_COLORS["external"]) >> external["fcm"]
    
    # Blockchain proxies to external RPCs
    apps["blockchain_workers"] >> Edge(label="WebSocket/HTTPS", color=TRAFFIC_COLORS["external"]) >> \
        external["blockchain_rpcs"]

# ============================================================================
# MAIN DIAGRAM GENERATION
# ============================================================================

def generate_diagram():
    """Generate the complete architecture diagram"""
    
    with Diagram(
        DIAGRAM_CONFIG["name"],
        show=False,
        direction=DIAGRAM_CONFIG["direction"],
        filename=DIAGRAM_CONFIG["filename"],
        graph_attr=DIAGRAM_CONFIG["graph_attr"],
        outformat="png"
    ):
        # Create all components
        edge = create_edge_layer()
        physical = create_physical_infrastructure()
        
        with Cluster("K3s Kubernetes Cluster (5 nodes)"):
            platform = create_platform_services()
            ingress = create_ingress_layer()
            apps = create_application_layer()
            data = create_data_layer()
            messaging = create_messaging_layer()
            storage = create_storage_registry()
        
        backup = create_backup_dr()
        external = create_external_apis()
        
        # Build all connections
        connect_edge_to_ingress(edge, ingress)
        connect_ingress_to_apps(ingress, apps)
        connect_apps_to_data(apps, data)
        connect_messaging_flows(apps, messaging)
        connect_observability(platform, apps)
        connect_gitops_flows(platform, apps)
        connect_data_replication(data)
        connect_storage_flows(apps, storage, data)
        connect_backup_flows(backup, data, storage)
        connect_external_apis(apps, external)

# ============================================================================
# EXECUTION
# ============================================================================

if __name__ == "__main__":
    generate_diagram()
    
    print("\n" + "="*80)
    print("✅ Notifi Homelab Architecture Generated (REFACTORED)")
    print("="*80)
    print("\n📊 IMPROVEMENTS IN REFACTORED VERSION:")
    print("━" * 80)
    print("1. 🏗️  CODE ARCHITECTURE")
    print("   • Modular component factories (separation of concerns)")
    print("   • Configuration-driven approach (DIAGRAM_CONFIG)")
    print("   • Reusable connection builders")
    print("   • Type hints for better IDE support")
    print("   • Clear function responsibilities")
    print("")
    print("2. 🎨 VISUAL IMPROVEMENTS")
    print("   • Color-coded traffic flows (external, internal, data, control)")
    print("   • Different edge styles (sync, async, replication, monitoring)")
    print("   • Ortho splines for cleaner diagram lines")
    print("   • Better cluster organization")
    print("   • Consistent labeling patterns")
    print("")
    print("3. 🔧 MAINTAINABILITY")
    print("   • Easy to add/remove components")
    print("   • Consistent naming conventions")
    print("   • Documented functions with docstrings")
    print("   • Centralized configuration")
    print("   • DRY principle applied throughout")
    print("")
    print("4. 🏛️  ARCHITECTURE CLARITY")
    print("   • Clear layer separation (Edge → Ingress → App → Data)")
    print("   • Explicit HA configurations (Patroni, Sentinel, Replica Sets)")
    print("   • Better traffic flow visualization")
    print("   • Resource specifications included")
    print("   • Backup/DR clearly separated")
    print("")
    print(f"📂 Saved to: {DIAGRAM_CONFIG['filename']}.png")
    print("\n" + "="*80)

