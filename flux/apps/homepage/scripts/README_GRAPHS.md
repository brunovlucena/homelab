# Blog Post Graph Generation

This directory contains scripts to generate visualizations for blog posts from data extracted from markdown files.

## 📊 Generated Graphs

The script `generate_blog_graphs.py` creates 5 professional graphs for the **understanding-vs-knowledge** blog post:

### 1. Before/After Comparison (`01_before_after_comparison.png`)
Three-panel comparison showing:
- **Traffic Max/Min Ratio**: 8.8M:1 → 154:1 (57,275x improvement!)
- **Top 3 Pod Concentration**: 88% → 75.2% (12.8% reduction)
- **Pod Distribution CV**: 30% → 13.5% (55% reduction)

### 2. Pod Traffic Distribution (`02_pod_traffic_distribution.png`)
Horizontal bar chart showing per-pod network traffic (bytes/sec) for all 12 active pods, color-coded by parser:
- Blue: Parser 0197ad6c
- Red: Parser c42d2e6c  
- Green: Parser e0a711bd

### 3. Node Distribution (`03_node_distribution.png`)
Dual visualization (bar + pie) showing pod distribution across 3 Kubernetes nodes:
- 9:7:7 distribution (Max/Min: 1.29:1)
- CV: 13.5% (Excellent)

### 4. Prediction Accuracy (`04_prediction_accuracy.png`)
Comparison of conservative vs optimistic predictions vs actual measured results:
- Shows AI-assisted predictions were accurate
- Actual results fell between conservative and optimistic targets

### 5. Improvement Timeline (`05_improvement_timeline.png`)
Visual timeline of the complete journey (3 hours total):
- Discovery (30 min) → 8.8M:1 imbalance found
- Analysis (15 min) → 6 root causes identified
- Design (15 min) → 6 fixes designed
- Implementation (30 min) → v1.13.11 deployed
- Deployment (20 min) → Operator running
- Validation (90 min) → 57,000x improvement!

## 🚀 Usage

### Generate All Graphs

```bash
# From the scripts directory
python3 generate_blog_graphs.py

# Or from project root using make
make generate-graphs
```

### Install Dependencies

```bash
pip install -r requirements-graphs.txt
```

Required packages:
- `matplotlib>=3.7.0` - Graph generation
- `seaborn>=0.12.0` - Statistical visualizations
- `numpy>=1.24.0` - Numerical operations

## 📁 Output Location

Graphs are saved to:
```
public/blog-posts/graphs/
├── 01_before_after_comparison.png
├── 02_pod_traffic_distribution.png
├── 03_node_distribution.png
├── 04_prediction_accuracy.png
└── 05_improvement_timeline.png
```

## 🎨 Customization

### Graph Style
The script uses `seaborn`'s `darkgrid` theme with:
- Figure size: 12x8 or larger for multi-panel plots
- DPI: 300 (high resolution)
- Font sizes: 10-14pt for labels, 14-16pt for titles

### Colors
- Success/After: `#27ae60` (Green)
- Problem/Before: `#e74c3c` (Red)
- Prediction/Target: `#3498db` (Blue), `#e67e22` (Orange)
- Parser-specific: Blue, Red, Green for the 3 parsers

### Modifying Data

Edit the data directly in `generate_blog_graphs.py`:

```python
# Example: Update pod traffic data
pods = ['z4llt', 'rv9sj', ...]  # Pod names
traffic = [164589, 147622, ...]  # Traffic in bytes/sec
```

## 📝 Adding to Blog Post

Reference graphs in your markdown:

```markdown
![Before/After Comparison](./graphs/01_before_after_comparison.png)

**Figure 1**: Dramatic improvement across all metrics after applying fixes.
```

Or use HTML for more control:

```html
<img src="./graphs/01_before_after_comparison.png" 
     alt="Before/After Comparison" 
     style="max-width: 100%; height: auto;">
```

## 🔄 Automation

The graphs can be regenerated automatically:

1. **On data updates**: Run script when markdown data changes
2. **Via CI/CD**: Include in build pipeline
3. **Via Makefile**: `make generate-graphs` target

## 🐛 Troubleshooting

### Font Warnings
```
UserWarning: Glyph 9989 missing from font
```
**Solution**: Ignore or install additional fonts. Graphs still render correctly.

### Import Errors
```
ModuleNotFoundError: No module named 'matplotlib'
```
**Solution**: Install requirements: `pip install -r requirements-graphs.txt`

### Permission Errors
**Solution**: Ensure scripts have execute permission: `chmod +x generate_blog_graphs.py`

## 📊 Data Sources

All data extracted from:
```
public/blog-posts/understanding-vs-knowledge.md
```

Specifically:
- **Section**: "Post-Fix Measurements (2025-12-18T08:00 UTC - FINAL)"
- **Traffic data**: Per-pod network metrics table
- **Node distribution**: Pod placement matrix
- **Predictions**: Predicted vs. Actual Comparison table

## 🎯 Future Enhancements

Potential additions:
- [ ] Animated GIFs showing progression over time
- [ ] Interactive HTML charts (Plotly/Bokeh)
- [ ] CPU usage distribution graphs
- [ ] Latency improvement visualizations
- [ ] Coefficient of Variation calculations with confidence intervals
- [ ] Heat maps for pod-to-node placement

## 📖 Related Documentation

- [Blog Post: Understanding vs Knowledge](../public/blog-posts/understanding-vs-knowledge.md)
- [Knative Lambda Operator](../../knative-lambda-operator/)
- [Demo-Notifi Load Testing](../../../../demos/demo-notifi/)
