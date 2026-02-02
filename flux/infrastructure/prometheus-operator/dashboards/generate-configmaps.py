#!/usr/bin/env python3
"""
Generate Grafana dashboard ConfigMaps from JSON files.

This script reads dashboard JSON files and generates Kubernetes ConfigMaps
that Grafana's sidecar can auto-discover and provision.

Usage:
    python3 generate-configmaps.py

The script will:
1. Find all *.json files in the current directory
2. Generate corresponding ConfigMap YAML files
3. Update kustomization.yaml to include them
"""

import json
import os
import sys
from pathlib import Path
import yaml

DASHBOARDS_DIR = Path(__file__).parent
KUSTOMIZATION_FILE = DASHBOARDS_DIR / "kustomization.yaml"
NAMESPACE = "prometheus"

def sanitize_name(name: str) -> str:
    """Convert dashboard name to valid Kubernetes resource name."""
    # Remove .json extension and replace invalid chars
    name = name.replace(".json", "").replace("_", "-").lower()
    # Ensure it starts with a letter
    if name and not name[0].isalpha():
        name = "dashboard-" + name
    return name

def json_to_configmap(json_file: Path) -> dict:
    """Convert a dashboard JSON file to a ConfigMap."""
    with open(json_file, 'r') as f:
        dashboard_data = json.load(f)
    
    # Extract dashboard metadata
    dashboard = dashboard_data.get("dashboard", dashboard_data)
    dashboard_title = dashboard.get("title", json_file.stem)
    dashboard_uid = dashboard.get("uid", sanitize_name(json_file.stem))
    
    # Generate ConfigMap name
    cm_name = sanitize_name(json_file.stem)
    
    # Read the JSON content as string (preserve formatting)
    with open(json_file, 'r') as f:
        json_content = f.read()
    
    configmap = {
        "apiVersion": "v1",
        "kind": "ConfigMap",
        "metadata": {
            "name": f"{cm_name}-dashboard",
            "namespace": NAMESPACE,
            "labels": {
                "grafana_dashboard": "1",
                "app.kubernetes.io/name": "grafana",
                "app.kubernetes.io/component": "dashboard",
                "dashboard.kubernetes.io/name": dashboard_title,
                "dashboard.kubernetes.io/uid": dashboard_uid
            }
        },
        "data": {
            f"{json_file.name}": json_content
        }
    }
    
    return configmap

def generate_configmaps():
    """Generate ConfigMap YAML files for all JSON dashboards."""
    json_files = list(DASHBOARDS_DIR.glob("*.json"))
    
    if not json_files:
        print("❌ No JSON dashboard files found!")
        return []
    
    configmap_files = []
    
    for json_file in json_files:
        print(f"📊 Processing {json_file.name}...")
        
        try:
            configmap = json_to_configmap(json_file)
            cm_name = configmap["metadata"]["name"]
            yaml_file = DASHBOARDS_DIR / f"{cm_name}.yaml"
            
            # Write ConfigMap YAML
            with open(yaml_file, 'w') as f:
                yaml.dump(configmap, f, default_flow_style=False, sort_keys=False, allow_unicode=True)
            
            configmap_files.append(yaml_file.name)
            print(f"✅ Generated {yaml_file.name}")
            
        except Exception as e:
            print(f"❌ Error processing {json_file.name}: {e}")
            continue
    
    return configmap_files

def update_kustomization(configmap_files: list):
    """Update kustomization.yaml to include generated ConfigMaps."""
    if not configmap_files:
        return
    
    # Read existing kustomization or create new
    if KUSTOMIZATION_FILE.exists():
        with open(KUSTOMIZATION_FILE, 'r') as f:
            kustomization = yaml.safe_load(f) or {}
    else:
        kustomization = {
            "apiVersion": "kustomize.config.k8s.io/v1beta1",
            "kind": "Kustomization",
            "metadata": {
                "name": "grafana-dashboards",
                "namespace": NAMESPACE
            }
        }
    
    # Ensure resources list exists
    if "resources" not in kustomization:
        kustomization["resources"] = []
    
    # Add ConfigMap files (avoid duplicates)
    existing_resources = set(kustomization["resources"])
    for cm_file in configmap_files:
        if cm_file not in existing_resources:
            kustomization["resources"].append(cm_file)
            existing_resources.add(cm_file)
    
    # Sort resources
    kustomization["resources"].sort()
    
    # Write back
    with open(KUSTOMIZATION_FILE, 'w') as f:
        yaml.dump(kustomization, f, default_flow_style=False, sort_keys=False)
    
    print(f"✅ Updated {KUSTOMIZATION_FILE.name}")

def main():
    print("🚀 Generating Grafana dashboard ConfigMaps...\n")
    
    # Generate ConfigMaps
    configmap_files = generate_configmaps()
    
    if not configmap_files:
        print("\n❌ No ConfigMaps generated!")
        sys.exit(1)
    
    # Update kustomization
    update_kustomization(configmap_files)
    
    print(f"\n✅ Successfully generated {len(configmap_files)} ConfigMap(s)!")
    print("\n📝 Next steps:")
    print("   1. Review the generated ConfigMap YAML files")
    print("   2. Commit and push to your Git repository")
    print("   3. Flux will sync them to Kubernetes")
    print("   4. Grafana sidecar will auto-discover and provision the dashboards")

if __name__ == "__main__":
    main()
