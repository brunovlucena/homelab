#!/usr/bin/env python3
"""Generate traffic flow diagram - standalone version."""

import sys
import os

# Try to import, if fails, show helpful message
try:
    import matplotlib
    matplotlib.use('Agg')
    import matplotlib.pyplot as plt
    import matplotlib.patches as mpatches
    import numpy as np
    from pathlib import Path
    import seaborn as sns
except ImportError as e:
    print(f"❌ Missing dependency: {e}")
    print("\n📦 Install with:")
    print("   pip3 install --break-system-packages matplotlib seaborn numpy")
    print("\n   OR use a virtual environment:")
    print("   python3 -m venv venv")
    print("   source venv/bin/activate")
    print("   pip install matplotlib seaborn numpy")
    sys.exit(1)

# Set style
sns.set_theme(style="darkgrid")
plt.rcParams['figure.figsize'] = (16, 12)
plt.rcParams['font.size'] = 11

# Output directory
OUTPUT_DIR = Path(__file__).parent.parent / "src" / "frontend" / "public" / "blog-posts" / "graphs"
OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

def create_traffic_flow_diagram():
    """Create a flow diagram showing how traffic concentrates on pods."""
    fig, ax = plt.subplots(figsize=(16, 12))
    ax.set_xlim(0, 12)
    ax.set_ylim(0, 11)
    ax.axis('off')
    
    # Colors
    scheduler_color = '#3498db'  # Blue - WHO IS PULLING
    http_color = '#e67e22'       # Orange - ROUTING LAYER 1
    knative_color = '#9b59b6'   # Purple - ROUTING LAYER 2
    k8s_color = '#1abc9c'       # Teal - ROUTING LAYER 3
    pod_hot_color = '#e74c3c'    # Red (hot pods)
    pod_cold_color = '#95a5a6'   # Gray (cold pods)
    
    # Title
    ax.text(6, 10.5, 'Traffic Flow: Why Pods Get Uneven Load', 
            ha='center', va='top', fontsize=18, fontweight='bold')
    ax.text(6, 10.1, 'Step-by-step: How requests flow from Scheduler to Pods', 
            ha='center', va='top', fontsize=12, style='italic', color='#555')
    
    # ===== STEP 1: WHO IS PULLING =====
    scheduler_box = mpatches.FancyBboxPatch((4.5, 8.8), 3, 1,
                                           boxstyle="round,pad=0.15", 
                                           facecolor=scheduler_color, 
                                           edgecolor='black', linewidth=3)
    ax.add_patch(scheduler_box)
    ax.text(6, 9.5, 'WHO IS PULLING:', ha='center', va='center', 
            fontsize=11, fontweight='bold', color='white', style='italic')
    ax.text(6, 9.1, 'Scheduler Service\n(Sends CloudEvents)', ha='center', va='center', 
            fontsize=13, fontweight='bold', color='white')
    
    # Step number
    step1_circle = mpatches.Circle((3.8, 9.3), 0.25, facecolor='white', 
                                   edgecolor='black', linewidth=2)
    ax.add_patch(step1_circle)
    ax.text(3.8, 9.3, '1', ha='center', va='center', fontsize=12, fontweight='bold')
    
    # ===== STEP 2: ROUTING LAYER 1 - HTTP/2 =====
    http_box = mpatches.FancyBboxPatch((4, 7.2), 4, 1,
                                      boxstyle="round,pad=0.15",
                                      facecolor=http_color,
                                      edgecolor='black', linewidth=2)
    ax.add_patch(http_box)
    ax.text(6, 7.9, 'ROUTING LAYER 1: HTTP/2', ha='center', va='center', 
            fontsize=11, fontweight='bold', color='white', style='italic')
    ax.text(6, 7.5, 'Connection Pooling\n⚠️ Reuses same TCP connection\n→ Bypasses pod-level routing!', 
            ha='center', va='center', fontsize=12, fontweight='bold', color='white')
    
    step2_circle = mpatches.Circle((3.3, 7.7), 0.25, facecolor='white', 
                                   edgecolor='black', linewidth=2)
    ax.add_patch(step2_circle)
    ax.text(3.3, 7.7, '2', ha='center', va='center', fontsize=12, fontweight='bold')
    
    # Arrow from scheduler to HTTP/2
    ax.annotate('', xy=(6, 8.2), xytext=(6, 8.8),
                arrowprops=dict(arrowstyle='->', lw=3, color='black'))
    ax.text(6.5, 8.5, 'Sends\nEvents', ha='left', va='center', fontsize=9, 
            bbox=dict(boxstyle="round,pad=0.3", facecolor='white', alpha=0.8))
    
    # ===== STEP 3: ROUTING LAYER 2 - Knative =====
    knative_box = mpatches.FancyBboxPatch((4, 5.8), 4, 1,
                                         boxstyle="round,pad=0.15",
                                         facecolor=knative_color,
                                         edgecolor='black', linewidth=2)
    ax.add_patch(knative_box)
    ax.text(6, 6.5, 'ROUTING LAYER 2: Knative', ha='center', va='center', 
            fontsize=11, fontweight='bold', color='white', style='italic')
    ax.text(6, 6.1, 'Broker → Trigger\n⚠️ Routes at REVISION level\n→ Not pod-level!', 
            ha='center', va='center', fontsize=12, fontweight='bold', color='white')
    
    step3_circle = mpatches.Circle((3.3, 6.3), 0.25, facecolor='white', 
                                   edgecolor='black', linewidth=2)
    ax.add_patch(step3_circle)
    ax.text(3.3, 6.3, '3', ha='center', va='center', fontsize=12, fontweight='bold')
    
    # Arrow from HTTP/2 to Knative
    ax.annotate('', xy=(6, 6.6), xytext=(6, 7.2),
                arrowprops=dict(arrowstyle='->', lw=3, color='black'))
    
    # ===== STEP 4: ROUTING LAYER 3 - Kubernetes =====
    k8s_box = mpatches.FancyBboxPatch((4, 4.4), 4, 1,
                                      boxstyle="round,pad=0.15",
                                      facecolor=k8s_color,
                                      edgecolor='black', linewidth=2)
    ax.add_patch(k8s_box)
    ax.text(6, 5.1, 'ROUTING LAYER 3: Kubernetes Service', ha='center', va='center', 
            fontsize=11, fontweight='bold', color='white', style='italic')
    ax.text(6, 4.7, 'Service (ClusterIP)\n✅ Round-robin at SERVICE level\n❌ But connection already established!', 
            ha='center', va='center', fontsize=12, fontweight='bold', color='white')
    
    step4_circle = mpatches.Circle((3.3, 4.9), 0.25, facecolor='white', 
                                   edgecolor='black', linewidth=2)
    ax.add_patch(step4_circle)
    ax.text(3.3, 4.9, '4', ha='center', va='center', fontsize=12, fontweight='bold')
    
    # Arrow from Knative to K8s
    ax.annotate('', xy=(6, 5.2), xytext=(6, 5.8),
                arrowprops=dict(arrowstyle='->', lw=3, color='black'))
    
    # ===== STEP 5: PODS - THE PROBLEM =====
    # Hot pods (receiving most traffic) - 3 pods getting 88% of traffic
    hot_pod1 = mpatches.FancyBboxPatch((1, 2), 1.5, 1.5,
                                       boxstyle="round,pad=0.1",
                                       facecolor=pod_hot_color,
                                       edgecolor='black', linewidth=4)
    ax.add_patch(hot_pod1)
    ax.text(1.75, 2.75, 'Pod 1', ha='center', va='center',
            fontsize=11, fontweight='bold', color='white')
    ax.text(1.75, 2.4, '🔥🔥🔥', ha='center', va='center', fontsize=14)
    ax.text(1.75, 2.1, '30% traffic', ha='center', va='center',
            fontsize=9, fontweight='bold', color='white')
    
    hot_pod2 = mpatches.FancyBboxPatch((3.5, 2), 1.5, 1.5,
                                       boxstyle="round,pad=0.1",
                                       facecolor=pod_hot_color,
                                       edgecolor='black', linewidth=4)
    ax.add_patch(hot_pod2)
    ax.text(4.25, 2.75, 'Pod 2', ha='center', va='center',
            fontsize=11, fontweight='bold', color='white')
    ax.text(4.25, 2.4, '🔥🔥🔥', ha='center', va='center', fontsize=14)
    ax.text(4.25, 2.1, '30% traffic', ha='center', va='center',
            fontsize=9, fontweight='bold', color='white')
    
    hot_pod3 = mpatches.FancyBboxPatch((6, 2), 1.5, 1.5,
                                       boxstyle="round,pad=0.1",
                                       facecolor=pod_hot_color,
                                       edgecolor='black', linewidth=4)
    ax.add_patch(hot_pod3)
    ax.text(6.75, 2.75, 'Pod 3', ha='center', va='center',
            fontsize=11, fontweight='bold', color='white')
    ax.text(6.75, 2.4, '🔥🔥🔥', ha='center', va='center', fontsize=14)
    ax.text(6.75, 2.1, '28% traffic', ha='center', va='center',
            fontsize=9, fontweight='bold', color='white')
    
    # Cold pods (receiving little traffic) - 20 pods sharing 12%
    cold_pod_positions = [(0.3, 0.5), (2.2, 0.5), (4.7, 0.5), (7.2, 0.5),
                          (1.2, 0.2), (3.7, 0.2), (6.2, 0.2), (8.7, 0.2),
                          (0.7, 0.8), (2.7, 0.8), (5.2, 0.8), (7.7, 0.8),
                          (1.7, 1.1), (4.2, 1.1), (6.7, 1.1), (9.2, 1.1),
                          (0.5, 1.4), (3, 1.4), (5.5, 1.4), (8, 1.4)]
    
    for x, y in cold_pod_positions[:20]:
        pod = mpatches.FancyBboxPatch((x, y), 0.5, 0.5,
                                      boxstyle="round,pad=0.05",
                                      facecolor=pod_cold_color,
                                      edgecolor='gray', linewidth=1,
                                      alpha=0.5)
        ax.add_patch(pod)
    
    ax.text(9.5, 1.5, '20 pods\nsharing\n12%', ha='center', va='center',
            fontsize=9, fontweight='bold', color='#555',
            bbox=dict(boxstyle="round,pad=0.3", facecolor='#f0f0f0', alpha=0.8))
    
    # Thick arrows to hot pods (showing heavy traffic)
    for x_pos, label in [(1.75, 'Heavy'), (4.25, 'Heavy'), (6.75, 'Heavy')]:
        ax.annotate('', xy=(x_pos, 3.5), xytext=(6, 4.4),
                    arrowprops=dict(arrowstyle='->', lw=6, color=pod_hot_color, alpha=0.9))
    
    # Thin arrows to cold pods (showing minimal traffic)
    for x, y in cold_pod_positions[:5]:
        ax.annotate('', xy=(x+0.25, y+0.5), xytext=(6, 4.4),
                    arrowprops=dict(arrowstyle='->', lw=1, color=pod_cold_color, alpha=0.2))
    
    # Arrow from K8s to pods section
    ax.annotate('', xy=(6, 3.5), xytext=(6, 4.4),
                arrowprops=dict(arrowstyle='->', lw=3, color='black'))
    step5_circle = mpatches.Circle((3.3, 2.75), 0.25, facecolor='white', 
                                   edgecolor='black', linewidth=2)
    ax.add_patch(step5_circle)
    ax.text(3.3, 2.75, '5', ha='center', va='center', fontsize=12, fontweight='bold')
    
    # Legend
    legend_y = 0.3
    legend_items = [
        ('Scheduler (Who pulls)', scheduler_color),
        ('HTTP/2 (Layer 1)', http_color),
        ('Knative (Layer 2)', knative_color),
        ('K8s Service (Layer 3)', k8s_color),
        ('Hot Pods (88% traffic)', pod_hot_color),
        ('Cold Pods (12% traffic)', pod_cold_color),
    ]
    
    for i, (label, color) in enumerate(legend_items):
        x_pos = 0.5 + (i % 3) * 3.5
        y_pos = legend_y - (i // 3) * 0.4
        rect = mpatches.Rectangle((x_pos, y_pos), 0.3, 0.25, 
                                  facecolor=color, edgecolor='black', linewidth=1)
        ax.add_patch(rect)
        ax.text(x_pos + 0.4, y_pos + 0.125, label, ha='left', va='center', 
                fontsize=9, fontweight='bold')
    
    plt.tight_layout()
    output_file = OUTPUT_DIR / '06_traffic_flow_diagram.png'
    plt.savefig(output_file, dpi=300, bbox_inches='tight')
    print(f"✅ Created: {output_file}")
    print(f"📁 Full path: {output_file.absolute()}")
    plt.close()

if __name__ == "__main__":
    print("📊 Generating traffic flow diagram...\n")
    create_traffic_flow_diagram()
    print("\n✅ Done!")
