# Surveillance Camera Guide for AI Agents

**Complete guide for purchasing and integrating AI-powered surveillance cameras for your homelab AI agents**

---

## Executive Summary

**Purpose**: Deploy intelligent cameras with edge AI processing for surveillance and monitoring  
**Use Cases**: Security monitoring, intrusion detection, people counting, vehicle tracking, anomaly detection  
**Budget**: $65-500 per camera (target: best price/performance from China)  
**Recommended**: 4-8 cameras for comprehensive coverage  
**Total Cost**: $500-4,000 depending on configuration

---

## Camera Types & Use Cases

### 1. Perimeter Security Cameras (2-4 cameras)

**Location**: Exterior walls, entrances, property boundaries  
**Purpose**: Intrusion detection, motion tracking, vehicle monitoring

**Requirements**:
- **Resolution**: 1080p (1920×1080) minimum, 4K preferred for license plate recognition
- **FPS**: 30 FPS minimum, 60 FPS for high-speed areas
- **AI Capabilities**:
  - Person detection
  - Vehicle detection
  - License plate recognition (LPR)
  - Intrusion detection (perimeter breach)
  - Motion tracking
  - Face detection (optional)
- **Field of View**: Medium to wide angle (70-110°)
- **Mounting**: Wall/pole mount, weatherproof (IP67)
- **Lighting**: Day/night capable, IR (50m+), ColorVu/Starlight preferred
- **Edge Processing**: On-device AI inference (reduce latency)

**Recommended Models**:
- **Dahua IPC-HFW5842H-ZHE**: 8MP 4K bullet, WizMind AI, LPR, $250-259
- **Hikvision DS-2CD2T47G1-L**: 4MP ColorVu, AI detection, $65-75
- **Dahua IPC-HFW3441T-AS**: 4MP bullet, WizSense AI, $45-55 (budget)

---

### 2. Entrance/Exit Cameras (1-2 cameras)

**Location**: Main entrances, doorways  
**Purpose**: People counting, access control, visitor tracking

**Requirements**:
- **Resolution**: 1080p minimum, 4MP preferred
- **FPS**: 30 FPS
- **AI Capabilities**:
  - People counting (entry/exit)
  - Face detection/recognition (optional)
  - Visitor flow analysis
  - Peak hour detection
  - Unauthorized access alerts
- **Field of View**: Wide angle (90-110°)
- **Mounting**: Ceiling/wall mount
- **Lighting**: Good indoor/outdoor lighting, color accurate

**Recommended Models**:
- **Dahua IPC-HDBW3441R-ZAS**: 4MP dome, WizSense AI, people counting, $108-117
- **Hikvision DS-2CD2347G1-LU**: 4MP, AI, people counting, $80-90

---

### 3. Indoor Monitoring Cameras (2-3 cameras)

**Location**: Interior spaces, hallways, common areas  
**Purpose**: Activity monitoring, occupancy tracking, safety compliance

**Requirements**:
- **Resolution**: 1080p minimum
- **FPS**: 30 FPS
- **AI Capabilities**:
  - Occupancy counting
  - Activity detection
  - Anomaly detection (unusual behavior)
  - Object detection (left items, suspicious packages)
  - Safety compliance (PPE detection, restricted areas)
- **Field of View**: Wide angle (110-130°)
- **Mounting**: Ceiling mount
- **Lighting**: Good indoor lighting, low-light capable

**Recommended Models**:
- **Dahua IPC-HDBW3441R-ZAS**: 4MP dome, WizSense AI, $108-117
- **Hikvision DS-2CD2T47G1-L**: 4MP ColorVu, $65-75

---

### 4. High-Value Area Cameras (1-2 cameras)

**Location**: Server rooms, storage areas, sensitive zones  
**Purpose**: Enhanced security, detailed monitoring, evidence collection

**Requirements**:
- **Resolution**: 4K (3840×2160) preferred
- **FPS**: 30-60 FPS
- **AI Capabilities**:
  - High-resolution object detection
  - Face recognition (if needed)
  - Detailed activity logging
  - Tamper detection
- **Field of View**: Medium angle (60-90°) for detail
- **Mounting**: Ceiling/wall mount
- **Lighting**: Excellent lighting, low-light capable

**Recommended Models**:
- **Dahua IPC-HDBW5842R-ASE**: 8MP 4K, WizMind AI, $135-138
- **Hikvision DS-2CD2T87G2-L**: 8MP 4K, ColorVu, DeepinMind AI, $120-140

---

## Technical Specifications Summary

### Minimum Requirements (All Cameras)

| Specification | Minimum | Recommended | Premium |
|---------------|---------|-------------|---------|
| **Resolution** | 1080p (2MP) | 4MP | 8MP (4K) |
| **FPS** | 25 FPS | 30 FPS | 60 FPS (high-speed areas) |
| **Compression** | H.264 | H.265 | H.265+ |
| **IP Rating** | IP65 | IP67 | IP67+ |
| **Night Vision** | IR (30m) | IR (50m) | ColorVu/Starlight |
| **AI Processing** | Cloud | Edge (on-device) | Edge + Cloud hybrid |
| **PoE** | Yes | Yes (PoE+) | Yes (PoE++) |
| **Storage** | Cloud | Edge (SD card) | Edge + Cloud |
| **Wide Dynamic Range** | Basic | WDR | True WDR |

### AI Capabilities Required

| Capability | Perimeter | Entrance | Indoor | High-Value |
|------------|-----------|----------|--------|------------|
| **Object Detection** | ✅ | ✅ | ✅ | ✅ |
| **People Counting** | ❌ | ✅ | ✅ | ❌ |
| **Vehicle Detection** | ✅ | ❌ | ❌ | ❌ |
| **License Plate Recognition** | ✅ | ❌ | ❌ | ❌ |
| **Intrusion Detection** | ✅ | ✅ | ✅ | ✅ |
| **Face Detection** | Optional | Optional | ❌ | Optional |
| **Anomaly Detection** | ✅ | ✅ | ✅ | ✅ |
| **Motion Tracking** | ✅ | ✅ | ✅ | ✅ |

---

## Chinese Manufacturer Recommendations

### 1. Dahua Technology (大华)

**Strengths**:
- Leading Chinese security camera manufacturer
- Strong AI capabilities (WizSense, WizMind)
- Good price/performance ratio
- Wide product range
- Good support

**Recommended Models**:

| Model | Resolution | AI Features | Price (USD) | Use Case |
|-------|------------|-------------|-------------|----------|
| **IPC-HDBW3441R-ZAS** | 4MP | WizSense AI, people counting | $108-117 | Entrance, Indoor |
| **IPC-HDBW5842R-ASE** | 8MP 4K | WizMind AI, advanced detection | $135-138 | High-Value Areas |
| **IPC-HFW5842H-ZHE** | 8MP 4K | WizMind AI, LPR, vehicle detection | $250-259 | Perimeter, LPR |
| **IPC-HFW3441T-AS** | 4MP | WizSense AI, basic detection | $45-55 | Budget Perimeter |

**Where to Buy**:
- Alibaba.com (Dahua official store)
- Taobao (authorized resellers)
- Direct from Dahua China
- Price: 30-50% lower than international prices

---

### 2. Hikvision (海康威视)

**Strengths**:
- World's largest video surveillance manufacturer
- Excellent AI capabilities (DeepinMind)
- High quality, reliable
- Good for enterprise deployments

**Recommended Models**:

| Model | Resolution | AI Features | Price (USD) | Use Case |
|-------|------------|-------------|-------------|----------|
| **DS-2CD2T47G1-L** | 4MP | ColorVu, AI detection | $65-75 | All locations |
| **DS-2CD2T87G2-L** | 8MP 4K | ColorVu, DeepinMind AI | $120-140 | High-Value |
| **DS-2CD2347G1-LU** | 4MP | AI, people counting | $80-90 | Entrance |

**Where to Buy**:
- Alibaba.com (Hikvision official store)
- Hikvision China website
- Authorized distributors
- Price: 40-60% lower than international prices

---

## Recommended Camera Configurations

### Budget Configuration ($500-800)

| Camera | Model | Quantity | Unit Price | Total |
|--------|-------|----------|------------|-------|
| Perimeter | Dahua IPC-HFW3441T-AS | 2 | $50 | $100 |
| Entrance | Hikvision DS-2CD2T47G1-L | 1 | $70 | $70 |
| Indoor | Hikvision DS-2CD2T47G1-L | 2 | $70 | $140 |
| **Total** | | **5** | | **$500** |

**Features**: Basic AI detection, 4MP resolution, 30 FPS

---

### Recommended Configuration ($1,200-1,800)

| Camera | Model | Quantity | Unit Price | Total |
|--------|-------|----------|------------|-------|
| Perimeter (LPR) | Dahua IPC-HFW5842H-ZHE | 1 | $255 | $255 |
| Perimeter (Standard) | Dahua IPC-HDBW3441R-ZAS | 1 | $110 | $110 |
| Entrance | Dahua IPC-HDBW3441R-ZAS | 1 | $110 | $110 |
| Indoor | Dahua IPC-HDBW3441R-ZAS | 2 | $110 | $220 |
| High-Value | Dahua IPC-HDBW5842R-ASE | 1 | $135 | $135 |
| **Total** | | **6** | | **$1,200** |

**Features**: Enhanced AI (LPR, advanced detection), 4K for perimeter, 4MP for others

---

### Premium Configuration ($2,500-4,000)

| Camera | Model | Quantity | Unit Price | Total |
|--------|-------|----------|------------|-------|
| Perimeter (LPR) | Dahua IPC-HFW5842H-ZHE | 2 | $255 | $510 |
| Perimeter (Standard) | Hikvision DS-2CD2T87G2-L | 1 | $130 | $130 |
| Entrance | Hikvision DS-2CD2347G1-LU | 1 | $85 | $85 |
| Indoor | Hikvision DS-2CD2T47G1-L | 2 | $70 | $140 |
| High-Value | Hikvision DS-2CD2T87G2-L | 1 | $130 | $130 |
| **Total** | | **7** | | **$2,500** |

**Features**: 4K where needed, advanced AI, ColorVu night vision, DeepinMind AI

---

## Integration Requirements

### Network Requirements

- **PoE Switches**: 8-port PoE+ switch ($50-100)
- **Bandwidth**: 100Mbps minimum per camera (1Gbps switch recommended)
- **Network**: Isolated VLAN for cameras (security)
- **Storage**: Edge storage (SD card) + cloud backup

### Software Integration

- **RTSP/ONVIF**: Standard protocol support (all cameras)
- **API Access**: REST API for camera control
- **AI SDK**: Integration with agent-pos-edge system
- **CloudEvents**: Event streaming to command center

### Edge Processing

- **On-Device AI**: Reduce latency, reduce bandwidth
- **Model Support**: TensorFlow Lite, ONNX, or manufacturer SDK
- **Processing Power**: NPU (Neural Processing Unit) preferred
- **Latency**: <100ms for real-time detection

---

## Integration with AI Agents

### Architecture

```
┌─────────────────────────────────────────────────────┐
│           AI Surveillance Agent Architecture         │
├─────────────────────────────────────────────────────┤
│                                                      │
│  IP Cameras (Edge)                                   │
│  ├─ RTSP Stream → Video Processing                  │
│  ├─ On-Device AI → Object Detection                 │
│  └─ Events → CloudEvents → Agent                    │
│                                                      │
│  ↓ (mTLS via Linkerd)                              │
│                                                      │
│  AI Agent (Kubernetes)                              │
│  ├─ agent-surveillance receives events              │
│  ├─ Advanced analysis (if needed)                    │
│  ├─ Alert generation                                │
│  └─ Dashboard updates                               │
│                                                      │
│  ↓                                                   │
│                                                      │
│  Command Center                                      │
│  ├─ Real-time alerts                                │
│  ├─ Video playback                                  │
│  ├─ Analytics dashboard                             │
│  └─ Historical data                                 │
│                                                      │
└─────────────────────────────────────────────────────┘
```

### Example Integration Code

```python
# agent-surveillance handler.py
import cv2
import rtsp
from cloudevents.http import CloudEvent

class SurveillanceAgent:
    def __init__(self):
        self.cameras = {
            'perimeter-1': 'rtsp://camera1:554/stream',
            'entrance-1': 'rtsp://camera2:554/stream',
            'indoor-1': 'rtsp://camera3:554/stream',
        }
    
    def process_camera_stream(self, camera_id: str):
        """Process RTSP stream from camera"""
        stream = rtsp.Client(self.cameras[camera_id])
        
        while True:
            frame = stream.read()
            
            # On-device AI detection (if camera supports)
            # Or process in agent
            detections = self.detect_objects(frame)
            
            # Generate events for significant detections
            if detections:
                event = CloudEvent(
                    type='io.surveillance.detection',
                    data={
                        'camera_id': camera_id,
                        'detections': detections,
                        'timestamp': datetime.now().isoformat()
                    }
                )
                self.send_event(event)
```

---

## Purchasing Strategy

### Option 1: Alibaba.com (Recommended)

**Advantages**:
- Verified suppliers
- Trade assurance (payment protection)
- Bulk discounts available
- Easy communication (English support)
- Shipping to most countries

**Process**:
1. Search for "Dahua AI camera" or "Hikvision AI camera"
2. Filter by: AI features, resolution, price
3. Contact suppliers (request quotes, MOQ)
4. Negotiate bulk pricing (5+ units)
5. Order sample (1-2 units for testing)
6. Place bulk order

**Expected Savings**: 30-50% vs. international prices

---

### Option 2: Direct from Manufacturer

**Advantages**:
- Best prices (no middleman)
- Custom configurations
- OEM options available
- Direct support

**Process**:
1. Contact Dahua/Hikvision China sales
2. Request quote for surveillance AI cameras
3. Specify requirements (AI features, quantity)
4. Negotiate pricing (MOQ usually 10+ units)
5. Arrange shipping (DDP - Delivered Duty Paid)

**Expected Savings**: 40-60% vs. international prices

---

## Cost Breakdown

### Per Setup (Recommended Configuration)

| Item | Quantity | Unit Price | Total |
|------|----------|------------|-------|
| **Cameras** | 6 | $110-255 | $1,200 |
| **PoE Switch** | 1 | $80 | $80 |
| **Cables (Cat6)** | 200m | $0.50/m | $100 |
| **Mounting Hardware** | 6 sets | $10 | $60 |
| **Installation** | 1 | $300 | $300 |
| **Total** | | | **$1,740** |

---

## Implementation Timeline

### Phase 1: Research & Testing (Week 1-2)
- [ ] Order 2-3 sample cameras (different models)
- [ ] Test AI capabilities (person detection, LPR)
- [ ] Test integration with agent-pos-edge
- [ ] Evaluate performance (latency, accuracy)
- [ ] Select final models

### Phase 2: Procurement (Week 3-4)
- [ ] Negotiate bulk pricing (5+ cameras)
- [ ] Place order (Alibaba or direct)
- [ ] Arrange shipping (DDP preferred)
- [ ] Customs clearance (if needed)

### Phase 3: Installation (Week 5-6)
- [ ] Install cameras
- [ ] Configure network (PoE, VLAN)
- [ ] Integrate with agent-pos-edge
- [ ] Test AI features
- [ ] Configure alerts

### Phase 4: Optimization (Week 7+)
- [ ] Monitor performance
- [ ] Optimize AI models
- [ ] Fine-tune detection zones
- [ ] Scale to additional locations

---

## Key Success Factors

1. **Edge AI Processing**: On-device inference reduces latency and bandwidth
2. **Standard Protocols**: RTSP/ONVIF for easy integration
3. **Price/Performance**: Chinese cameras offer best value
4. **Scalability**: Standardized configuration across locations
5. **Reliability**: IP67 rating for harsh environments

---

## Quick Start Checklist

### This Week
- [ ] Order 2-3 sample cameras from Alibaba.com ($300-500)
- [ ] Test AI capabilities (person detection, LPR)
- [ ] Test integration with agent-pos-edge

### Next Week
- [ ] Select final models
- [ ] Negotiate bulk pricing (5+ cameras)
- [ ] Place order for full setup

### Next Month
- [ ] Install cameras
- [ ] Integrate with agent-pos-edge
- [ ] Configure alerts and dashboards
- [ ] Test and optimize

---

## Recommended Action Plan

### Immediate (This Week)
1. **Order 2-3 sample cameras** from Alibaba.com
   - Dahua IPC-HFW5842H-ZHE (1 unit) - $255 (LPR test)
   - Dahua IPC-HDBW3441R-ZAS (1 unit) - $110 (general test)
   - **Total**: $365 (for testing)

2. **Test AI capabilities**:
   - Person detection
   - Vehicle detection
   - License plate recognition
   - Integration with agent-pos-edge

### Short-term (Next Month)
1. **Select final models** based on testing
2. **Negotiate bulk pricing** (5+ cameras)
3. **Place order** for full setup (6 cameras, $1,200)
4. **Install and test** at your location

### Long-term (Next Quarter)
1. **Optimize AI models** (custom training if needed)
2. **Scale infrastructure** (command center)
3. **Add more cameras** as needed

---

## Conclusion

**Recommended Configuration**:
- **6 cameras**: $1,200 (Recommended config)
- **Total setup**: $1,740 (including installation)
- **Best Value**: Dahua IPC-HDBW3441R-ZAS ($110) for most locations, Dahua IPC-HFW5842H-ZHE ($255) for perimeter/LPR

**Next Steps**: Order samples, test integration, negotiate bulk pricing, deploy setup.

---

**Ready to order?** Start with 2-3 sample cameras from Alibaba.com for testing! 📹🤖
