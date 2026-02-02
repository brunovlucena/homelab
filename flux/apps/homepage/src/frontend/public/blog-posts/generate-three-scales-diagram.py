#!/usr/bin/env python3
"""
Generate a visual diagram for the Three Scales of Understanding framework.

Creates a professional diagram showing Scale 1, Scale 2, and Scale 3 with their
characteristics, examples, limits, and relationships.

Requirements:
    pip install matplotlib numpy

Usage:
    python3 generate-three-scales-diagram.py

Output:
    - three-scales-framework.png (300 DPI PNG, in current directory)
    - three-scales-framework.svg (SVG version, scalable)
    - Also saves to: ../../../../storage/homepage-blog/images/graphs/

The diagram shows:
    - Scale 1: Formal/Symbolic Computation (blue)
    - Scale 2: Approximate/Heuristic Computation (purple)
    - Scale 3: Embodied/Contextual Computation (orange)
    - Relationships between scales
    - Integration of all scales in real systems
"""

import matplotlib.pyplot as plt
import matplotlib.patches as mpatches
from matplotlib.patches import FancyBboxPatch, FancyArrowPatch, ConnectionPatch
import numpy as np

# Set up the figure with a clean, professional style
plt.rcParams['font.family'] = 'sans-serif'
plt.rcParams['font.sans-serif'] = ['Arial', 'DejaVu Sans', 'Helvetica']
plt.rcParams['font.size'] = 10
plt.rcParams['axes.linewidth'] = 0

fig, ax = plt.subplots(figsize=(14, 10))
ax.set_xlim(0, 10)
ax.set_ylim(0, 12)
ax.axis('off')

# Color scheme - professional and accessible
colors = {
    'scale1': '#2E86AB',  # Blue for formal/symbolic
    'scale2': '#A23B72',  # Purple for heuristic
    'scale3': '#F18F01',  # Orange for embodied
    'background': '#F5F5F5',
    'text': '#2C3E50',
    'arrow': '#7F8C8D'
}

# Title
title = ax.text(5, 11.5, 'Three Scales of Understanding', 
                ha='center', va='center', fontsize=20, fontweight='bold',
                color=colors['text'])

# Scale 1: Formal/Symbolic Computation
scale1_box = FancyBboxPatch((0.5, 8.5), 2.8, 2.2,
                           boxstyle="round,pad=0.1", 
                           facecolor=colors['scale1'], 
                           edgecolor='white', linewidth=2,
                           alpha=0.9)
ax.add_patch(scale1_box)

scale1_title = ax.text(1.9, 10.3, 'SCALE 1\nFormal/Symbolic', 
                       ha='center', va='center', fontsize=12, 
                       fontweight='bold', color='white')

scale1_content = [
    '• Discrete symbols, explicit rules',
    '• Logical inference',
    '',
    'Examples:',
    '  Theorem provers (Coq, Isabelle)',
    '  Expert systems',
    '',
    'Limits:',
    '  Gödel\'s incompleteness',
    '  Turing\'s halting problem',
    '',
    'Characteristics:',
    '  + Verifiable',
    '  + Interpretable',
    '  - Brittle'
]

y_pos = 9.8
for line in scale1_content:
    ax.text(1.9, y_pos, line, ha='center', va='top', 
           fontsize=9, color='white')
    y_pos -= 0.18

# Scale 2: Approximate/Heuristic Computation
scale2_box = FancyBboxPatch((3.6, 8.5), 2.8, 2.2,
                           boxstyle="round,pad=0.1", 
                           facecolor=colors['scale2'], 
                           edgecolor='white', linewidth=2,
                           alpha=0.9)
ax.add_patch(scale2_box)

scale2_title = ax.text(5.0, 10.3, 'SCALE 2\nApproximate/Heuristic', 
                       ha='center', va='center', fontsize=12, 
                       fontweight='bold', color='white')

scale2_content = [
    '• Continuous optimization',
    '• Statistical learning',
    '',
    'Examples:',
    '  Neural networks',
    '  LLMs, Deep learning',
    '',
    'Limits:',
    '  Sample complexity',
    '  Generalization bounds',
    '  Bias-variance tradeoff',
    '',
    'Characteristics:',
    '  + Handles ambiguity',
    '  + Scales well',
    '  - Opaque'
]

y_pos = 9.8
for line in scale2_content:
    ax.text(5.0, y_pos, line, ha='center', va='top', 
           fontsize=9, color='white')
    y_pos -= 0.18

# Scale 3: Embodied/Contextual Computation
scale3_box = FancyBboxPatch((6.7, 8.5), 2.8, 2.2,
                           boxstyle="round,pad=0.1", 
                           facecolor=colors['scale3'], 
                           edgecolor='white', linewidth=2,
                           alpha=0.9)
ax.add_patch(scale3_box)

scale3_title = ax.text(8.1, 10.3, 'SCALE 3\nEmbodied/Contextual', 
                       ha='center', va='center', fontsize=12, 
                       fontweight='bold', color='white')

scale3_content = [
    '• Physical/social interaction',
    '• Real-world grounding',
    '',
    'Examples:',
    '  Human cognition',
    '  Embodied robotics',
    '',
    'Limits:',
    '  Evolutionary constraints',
    '  Sensorimotor experience',
    '',
    'Characteristics:',
    '  + Robust',
    '  + Adaptive',
    '  - Slow, resource-intensive'
]

y_pos = 9.8
for line in scale3_content:
    ax.text(8.1, y_pos, line, ha='center', va='top', 
           fontsize=9, color='white')
    y_pos -= 0.18

# Arrows showing relationships
arrow1 = FancyArrowPatch((3.3, 9.6), (3.6, 9.6),
                        arrowstyle='->', mutation_scale=20,
                        color=colors['arrow'], linewidth=2)
ax.add_patch(arrow1)
ax.text(3.45, 9.8, 'different\nconstraints', ha='center', va='bottom',
       fontsize=8, color=colors['text'], style='italic')

arrow2 = FancyArrowPatch((6.4, 9.6), (6.7, 9.6),
                        arrowstyle='->', mutation_scale=20,
                        color=colors['arrow'], linewidth=2)
ax.add_patch(arrow2)
ax.text(6.55, 9.8, 'different\nconstraints', ha='center', va='bottom',
       fontsize=8, color=colors['text'], style='italic')

# Bottom note about real systems
note_box = FancyBboxPatch((1.0, 5.5), 8.0, 1.5,
                         boxstyle="round,pad=0.15", 
                         facecolor='#ECF0F1', 
                         edgecolor=colors['text'], linewidth=1.5,
                         alpha=0.8)
ax.add_patch(note_box)

note_text = (
    'Note: Real systems (like the Kubernetes case study) operate across\n'
    'multiple scales simultaneously, combining formal guarantees (Scale 1),\n'
    'heuristic discovery (Scale 2), and contextual judgment (Scale 3).'
)
ax.text(5.0, 6.5, note_text, ha='center', va='center',
       fontsize=10, color=colors['text'], style='italic',
       bbox=dict(boxstyle='round,pad=0.5', facecolor='white', alpha=0.7))

# Add a connecting diagram showing overlap
overlap_y = 3.5
overlap_width = 0.8
overlap_height = 1.0

# Overlapping circles to show integration
circle1 = plt.Circle((3.5, overlap_y), 0.4, color=colors['scale1'], 
                    alpha=0.6, zorder=1)
circle2 = plt.Circle((5.0, overlap_y), 0.4, color=colors['scale2'], 
                    alpha=0.6, zorder=2)
circle3 = plt.Circle((6.5, overlap_y), 0.4, color=colors['scale3'], 
                    alpha=0.6, zorder=3)
ax.add_patch(circle1)
ax.add_patch(circle2)
ax.add_patch(circle3)

ax.text(3.5, overlap_y, '1', ha='center', va='center',
       fontsize=12, fontweight='bold', color='white', zorder=4)
ax.text(5.0, overlap_y, '2', ha='center', va='center',
       fontsize=12, fontweight='bold', color='white', zorder=5)
ax.text(6.5, overlap_y, '3', ha='center', va='center',
       fontsize=12, fontweight='bold', color='white', zorder=6)

ax.text(5.0, overlap_y - 0.8, 'Real systems integrate all scales', 
       ha='center', va='top', fontsize=10, color=colors['text'],
       fontweight='bold')

# Add constraint labels
constraint_y = 1.5
ax.text(1.9, constraint_y, 'Formal Logic\nConstraints', ha='center', va='center',
       fontsize=9, color=colors['scale1'], style='italic')
ax.text(5.0, constraint_y, 'Statistical\nLearning Bounds', ha='center', va='center',
       fontsize=9, color=colors['scale2'], style='italic')
ax.text(8.1, constraint_y, 'Evolutionary\nConstraints', ha='center', va='center',
       fontsize=9, color=colors['scale3'], style='italic')

# Save the figure
import os

# Determine output directory - use /output if available (Kubernetes job), otherwise use relative path
output_dir = '/output' if os.path.exists('/output') else '../../../../storage/homepage-blog/images/graphs'
os.makedirs(output_dir, exist_ok=True)

output_path = os.path.join(output_dir, 'three-scales-framework.png')
plt.tight_layout()
plt.savefig(output_path, dpi=300, bbox_inches='tight', 
           facecolor='white', edgecolor='none')
print(f"Diagram saved to: {output_path}")

# Also save as SVG for scalability
output_path_svg = os.path.join(output_dir, 'three-scales-framework.svg')
plt.savefig(output_path_svg, format='svg', bbox_inches='tight',
           facecolor='white', edgecolor='none')
print(f"SVG version saved to: {output_path_svg}")

# Also save a copy in the current directory for easy access (if not already there)
if output_dir != '.':
    local_output = 'three-scales-framework.png'
    plt.savefig(local_output, dpi=300, bbox_inches='tight',
               facecolor='white', edgecolor='none')
    print(f"Local copy saved to: {local_output}")

plt.close()

