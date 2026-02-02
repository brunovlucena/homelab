# Intelligent Camera Requirements: McDonald's POS/Edge Agent

**Comprehensive requirements analysis for AI-powered cameras for McDonald's restaurant automation**

---

## Executive Summary

**Purpose**: Deploy intelligent cameras with edge AI processing for McDonald's restaurant operations  
**Use Cases**: Kitchen monitoring, queue detection, drive-thru optimization, customer analytics, safety compliance  
**Budget**: $200-500 per camera (target: best price/performance from China)  
**Total Cameras per Location**: 7-10 cameras  
**Total Cost per Location**: $1,400-$5,000

---

## Camera Types & Use Cases

### 1. Kitchen Cameras (2-3 cameras)

**Location**: Grill area, fryer & assembly station  
**Purpose**: Food preparation monitoring, quality control, safety compliance

**Requirements**:
- **Resolution**: 1080p (1920×1080) minimum, 4K preferred
- **FPS**: 30 FPS minimum
- **AI Capabilities**:
  - Food tracking (item identification, quantity)
  - Prep time estimation
  - Station monitoring (equipment status)
  - Hygiene check (PPE detection, hand washing)
  - Temperature estimation (food safety)
  - Portion verification
- **Field of View**: Wide angle (90-120°)
- **Mounting**: Ceiling mount, IP67 rated
- **Lighting**: Low-light capable (IR or color night vision)
- **Edge Processing**: On-device AI inference (reduce latency)

**Recommended Models**:
- **Dahua IPC-HDBW5842R-ASE**: 8MP 4K, AI people/food detection, $135-138
- **Hikvision DS-2CD2T47G1-L**: 4MP ColorVu, AI detection, $65-75

---

### 2. Counter Camera (1 camera)

**Location**: Front counter, order point  
**Purpose**: Customer service, queue management, transaction monitoring

**Requirements**:
- **Resolution**: 1080p minimum
- **FPS**: 30 FPS
- **AI Capabilities**:
  - Queue counting (number of customers)
  - Wait time prediction
  - Peak detection
  - Customer sentiment analysis
  - Service quality monitoring
- **Field of View**: Wide angle (90-110°)
- **Mounting**: Wall/ceiling mount
- **Lighting**: Good indoor lighting, color accurate

**Recommended Models**:
- **Dahua IPC-HDBW3441R-ZAS**: 4MP dome, WizSense AI, $108-117
- **Hikvision DS-2CD2T47G1-L**: 4MP ColorVu, $65-75

---

### 3. Drive-Thru Cameras (2 cameras)

**Location**: Order point (menu board), pickup window  
**Purpose**: Vehicle detection, license plate recognition, service time tracking

**Requirements**:
- **Order Point Camera**:
  - **Resolution**: 4K (3840×2160) preferred
  - **FPS**: 60 FPS (for fast-moving vehicles)
  - **AI Capabilities**:
    - Vehicle detection
    - License plate recognition (LPR)
    - Order association (vehicle → order)
    - Service time tracking
- **Pickup Window Camera**:
  - **Resolution**: 1080p minimum
  - **FPS**: 30 FPS
  - **AI Capabilities**:
    - Vehicle identification
    - Order verification
    - Service completion detection
- **Field of View**: Medium angle (60-90°)
- **Mounting**: Pole/wall mount, weatherproof (IP67)
- **Lighting**: Day/night capable, IR for night

**Recommended Models**:
- **Dahua IPC-HFW5842H-ZHE**: 8MP 4K bullet, WizMind AI, LPR, $250-259
- **Hikvision DS-2CD2T47G1-L**: 4MP ColorVu, $65-75 (pickup window)

---

### 4. Dining Area Camera (1 camera)

**Location**: Dining area, seating area  
**Purpose**: Customer analytics, cleanliness monitoring, occupancy tracking

**Requirements**:
- **Resolution**: 1080p
- **FPS**: 30 FPS
- **AI Capabilities**:
  - Occupancy counting
  - Cleanliness score (spill detection, table status)
  - Trash monitoring
  - Customer flow analysis
- **Field of View**: Wide angle (110-130°)
- **Mounting**: Ceiling mount
- **Lighting**: Good indoor lighting

**Recommended Models**:
- **Dahua IPC-HDBW3441R-ZAS**: 4MP dome, WizSense AI, $108-117
- **Hikvision DS-2CD2T47G1-L**: 4MP ColorVu, $65-75

---

### 5. Entrance Camera (1 camera)

**Location**: Main entrance  
**Purpose**: Customer counting, entry/exit tracking, security

**Requirements**:
- **Resolution**: 1080p
- **FPS**: 30 FPS
- **AI Capabilities**:
  - People counting (entry/exit)
  - Customer flow analysis
  - Peak hour detection
- **Field of View**: Medium angle (70-90°)
- **Mounting**: Wall/ceiling mount
- **Lighting**: Day/night capable

**Recommended Models**:
- **Dahua IPC-HDBW3441R-ZAS**: 4MP dome, WizSense AI, $108-117
- **Hikvision DS-2CD2T47G1-L**: 4MP ColorVu, $65-75

---

## Technical Specifications Summary

### Minimum Requirements (All Cameras)

| Specification | Minimum | Recommended | Premium |
|---------------|---------|-------------|---------|
| **Resolution** | 1080p (2MP) | 4MP | 8MP (4K) |
| **FPS** | 25 FPS | 30 FPS | 60 FPS (drive-thru) |
| **Compression** | H.264 | H.265 | H.265+ |
| **IP Rating** | IP65 | IP67 | IP67+ |
| **Night Vision** | IR (30m) | IR (50m) | ColorVu/Starlight |
| **AI Processing** | Cloud | Edge (on-device) | Edge + Cloud hybrid |
| **PoE** | Yes | Yes (PoE+) | Yes (PoE++) |
| **Storage** | Cloud | Edge (SD card) | Edge + Cloud |
| **Wide Dynamic Range** | Basic | WDR | True WDR |

### AI Capabilities Required

| Capability | Kitchen | Counter | Drive-Thru | Dining | Entrance |
|------------|---------|---------|------------|--------|----------|
| **Object Detection** | ✅ | ✅ | ✅ | ✅ | ✅ |
| **People Counting** | ❌ | ✅ | ✅ | ✅ | ✅ |
| **Vehicle Detection** | ❌ | ❌ | ✅ | ❌ | ❌ |
| **License Plate Recognition** | ❌ | ❌ | ✅ | ❌ | ❌ |
| **Food Detection** | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Queue Detection** | ❌ | ✅ | ✅ | ❌ | ❌ |
| **Sentiment Analysis** | ❌ | ✅ | ❌ | ✅ | ❌ |
| **Safety Compliance** | ✅ | ❌ | ❌ | ❌ | ❌ |

---

## Chinese Manufacturer Recommendations

### 1. Dahua Technology (大华)

**Strengths**:
- Leading Chinese security camera manufacturer
- Strong AI capabilities (WizSense, WizMind)
- Good price/performance ratio
- Wide product range
- Good support in China

**Recommended Models**:

| Model | Resolution | AI Features | Price (USD) | Use Case |
|-------|------------|-------------|-------------|----------|
| **IPC-HDBW3441R-ZAS** | 4MP | WizSense AI, people counting | $108-117 | Counter, Dining, Entrance |
| **IPC-HDBW5842R-ASE** | 8MP 4K | WizMind AI, food detection | $135-138 | Kitchen |
| **IPC-HFW5842H-ZHE** | 8MP 4K | WizMind AI, LPR, vehicle detection | $250-259 | Drive-Thru |
| **IPC-HFW3441T-AS** | 4MP | WizSense AI, basic detection | $45-55 | Budget option |

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
| **DS-2CD2T87G2-L** | 8MP 4K | ColorVu, DeepinMind AI | $120-140 | Kitchen, Drive-Thru |
| **DS-2CD2347G1-LU** | 4MP | AI, people counting | $80-90 | Counter, Entrance |

**Where to Buy**:
- Alibaba.com (Hikvision official store)
- Hikvision China website
- Authorized distributors
- Price: 40-60% lower than international prices

---

### 3. Chinese AI Camera Startups

**Megvii (旷视科技) - Face++**
- **Focus**: AI computer vision, facial recognition
- **Products**: AI cameras with advanced analytics
- **Price**: $150-300 per camera
- **Best for**: Advanced AI features, custom solutions

**SenseTime (商汤科技)**
- **Focus**: AI vision, object detection
- **Products**: AI cameras with SenseTime SDK
- **Price**: $200-400 per camera
- **Best for**: Custom AI models, restaurant-specific solutions

**Where to Buy**:
- Direct from manufacturers
- Alibaba.com (authorized resellers)
- Custom OEM orders (minimum quantity required)

---

## Recommended Camera Configuration per Location

### Budget Configuration ($1,400-2,000 per location)

| Camera | Model | Quantity | Unit Price | Total |
|--------|-------|----------|------------|-------|
| Kitchen | Dahua IPC-HDBW3441R-ZAS | 2 | $110 | $220 |
| Counter | Dahua IPC-HDBW3441R-ZAS | 1 | $110 | $110 |
| Drive-Thru (Order) | Dahua IPC-HFW3441T-AS | 1 | $50 | $50 |
| Drive-Thru (Window) | Dahua IPC-HDBW3441R-ZAS | 1 | $110 | $110 |
| Dining | Dahua IPC-HDBW3441R-ZAS | 1 | $110 | $110 |
| Entrance | Dahua IPC-HDBW3441R-ZAS | 1 | $110 | $110 |
| **Total** | | **7** | | **$1,400** |

**Features**: Basic AI detection, 4MP resolution, 30 FPS

---

### Recommended Configuration ($2,500-3,500 per location)

| Camera | Model | Quantity | Unit Price | Total |
|--------|-------|----------|------------|-------|
| Kitchen | Dahua IPC-HDBW5842R-ASE | 2 | $135 | $270 |
| Counter | Dahua IPC-HDBW3441R-ZAS | 1 | $110 | $110 |
| Drive-Thru (Order) | Dahua IPC-HFW5842H-ZHE | 1 | $255 | $255 |
| Drive-Thru (Window) | Dahua IPC-HDBW3441R-ZAS | 1 | $110 | $110 |
| Dining | Dahua IPC-HDBW3441R-ZAS | 1 | $110 | $110 |
| Entrance | Dahua IPC-HDBW3441R-ZAS | 1 | $110 | $110 |
| **Total** | | **7** | | **$2,500** |

**Features**: Enhanced AI (food detection, LPR), 4K for drive-thru, 4MP for others

---

### Premium Configuration ($4,000-5,000 per location)

| Camera | Model | Quantity | Unit Price | Total |
|--------|-------|----------|------------|-------|
| Kitchen | Hikvision DS-2CD2T87G2-L | 2 | $130 | $260 |
| Counter | Hikvision DS-2CD2T47G1-L | 1 | $70 | $70 |
| Drive-Thru (Order) | Dahua IPC-HFW5842H-ZHE | 1 | $255 | $255 |
| Drive-Thru (Window) | Hikvision DS-2CD2T47G1-L | 1 | $70 | $70 |
| Dining | Hikvision DS-2CD2T47G1-L | 1 | $70 | $70 |
| Entrance | Hikvision DS-2CD2T47G1-L | 1 | $70 | $70 |
| **Total** | | **7** | | **$4,000** |

**Features**: 4K where needed, advanced AI, ColorVu night vision, DeepinMind AI

---

## Integration Requirements

### Network Requirements

- **PoE Switches**: 8-port PoE+ switch per location ($50-100)
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
4. Negotiate bulk pricing (10+ units)
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
2. Request quote for restaurant AI cameras
3. Specify requirements (AI features, quantity)
4. Negotiate pricing (MOQ usually 50+ units)
5. Arrange shipping (DDP - Delivered Duty Paid)

**Expected Savings**: 40-60% vs. international prices

---

### Option 3: Taobao (For Small Quantities)

**Advantages**:
- Good for small orders (1-10 units)
- Competitive prices
- Fast shipping within China

**Disadvantages**:
- Language barrier (Chinese only)
- Less support
- Quality varies by seller

**Process**:
1. Use Taobao agent (like Superbuy, Pandabuy)
2. Search for camera models
3. Order through agent
4. Agent handles shipping

**Expected Savings**: 20-40% vs. international prices

---

## Cost Breakdown

### Per Location (Recommended Configuration)

| Item | Quantity | Unit Price | Total |
|------|----------|------------|-------|
| **Cameras** | 7 | $110-255 | $2,500 |
| **PoE Switch** | 1 | $80 | $80 |
| **Cables (Cat6)** | 200m | $0.50/m | $100 |
| **Mounting Hardware** | 7 sets | $10 | $70 |
| **Installation** | 1 | $500 | $500 |
| **Total** | | | **$3,250** |

### Bulk Pricing (10+ Locations)

- **Camera Discount**: 15-20% off unit price
- **Total per Location**: $2,800-3,000
- **Savings**: $250-450 per location

---

## Implementation Timeline

### Phase 1: Research & Testing (Week 1-2)
- [ ] Order 2-3 sample cameras (different models)
- [ ] Test AI capabilities (food detection, queue counting)
- [ ] Test integration with agent-pos-edge
- [ ] Evaluate performance (latency, accuracy)
- [ ] Select final models

### Phase 2: Procurement (Week 3-4)
- [ ] Negotiate bulk pricing (10+ locations)
- [ ] Place order (Alibaba or direct)
- [ ] Arrange shipping (DDP preferred)
- [ ] Customs clearance (if needed)

### Phase 3: Installation (Week 5-6)
- [ ] Install cameras at pilot location
- [ ] Configure network (PoE, VLAN)
- [ ] Integrate with agent-pos-edge
- [ ] Test AI features
- [ ] Train staff

### Phase 4: Rollout (Week 7+)
- [ ] Install at remaining locations
- [ ] Monitor performance
- [ ] Optimize AI models
- [ ] Scale to all locations

---

## Recommended Action Plan

### Immediate (This Week)
1. **Order 2-3 sample cameras** from Alibaba.com
   - Dahua IPC-HDBW3441R-ZAS (1 unit) - $110
   - Dahua IPC-HDBW5842R-ASE (1 unit) - $135
   - Dahua IPC-HFW5842H-ZHE (1 unit) - $255
   - **Total**: $500 (for testing)

2. **Test AI capabilities**:
   - Food detection (kitchen)
   - Queue counting (counter)
   - Vehicle detection (drive-thru)
   - Integration with agent-pos-edge

### Short-term (Next Month)
1. **Select final models** based on testing
2. **Negotiate bulk pricing** (10+ locations)
3. **Place order** for pilot location (7 cameras)
4. **Install and test** at one location

### Long-term (Next Quarter)
1. **Roll out to all locations** (if pilot successful)
2. **Optimize AI models** (custom training)
3. **Scale infrastructure** (command center)

---

## Key Success Factors

1. **Edge AI Processing**: On-device inference reduces latency and bandwidth
2. **Standard Protocols**: RTSP/ONVIF for easy integration
3. **Price/Performance**: Chinese cameras offer best value
4. **Scalability**: Standardized configuration across locations
5. **Reliability**: IP67 rating for harsh environments

---

## Conclusion

**Recommended Configuration**:
- **7 cameras per location**: $2,500 (Recommended config)
- **Total per location**: $3,250 (including installation)
- **Bulk discount**: $2,800 per location (10+ locations)

**Best Value**: Dahua IPC-HDBW3441R-ZAS ($110) for most locations, Dahua IPC-HFW5842H-ZHE ($255) for drive-thru

**Next Steps**: Order samples, test integration, negotiate bulk pricing, deploy pilot location.

---

**Ready to order?** Start with 2-3 sample cameras from Alibaba.com for testing! 📹🤖
