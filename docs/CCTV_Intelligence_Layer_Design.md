# FireLine CCTV Intelligence Layer -- Complete Technical Design

**Document Version:** 1.0  
**Created:** 2026-04-06  
**Classification:** Confidential -- Internal & Investor Use  
**Designed By:** Multi-agent expert panel (Backend Architect, AI/ML Engineer, Security Engineer, Frontend Developer, SRE)

---

## Table of Contents

1. [Part 1: Backend Architecture](#part-1-backend-architecture)
2. [Part 2: AI/ML Computer Vision Pipeline](#part-2-aiml-computer-vision-pipeline)
3. [Part 3: Security & Privacy Architecture](#part-3-security--privacy-architecture)
4. [Part 4: Frontend UX Design](#part-4-frontend-ux-design)
5. [Part 5: Infrastructure & Reliability](#part-5-infrastructure--reliability)

---

## Executive Summary

The CCTV Intelligence Layer transforms passive security cameras into an active data source feeding every existing FireLine module. Edge GPU devices (NVIDIA Jetson Orin NX) run computer vision inference locally at each restaurant, publishing structured events via NATS JetStream. The cloud never receives raw video -- only anonymized clips and detection events.

**Key Capabilities:**
- Kitchen station occupancy and bottleneck detection
- Health code compliance monitoring (handwashing, gloves, hair nets)
- Food waste bin monitoring and receiving verification
- Dine-in occupancy and queue length detection
- Slip/fall safety alerts
- Video evidence attached to existing anomaly system
- Cross-location visual benchmarking for portfolio owners

**Architecture Highlights:**
- All inference at the edge ($599 one-time vs $1+/hr cloud GPU)
- Face anonymization before any frame leaves the device
- 48-hour offline resilience with automatic cloud sync
- ~$61/month per location at scale (200+ locations)
- BIPA/GDPR/CCPA/PCI DSS compliant by design

---

# Part 1: Backend Architecture

# CCTV Intelligence Layer -- Technical Design Document

## FireLine Platform -- `internal/vision/`

---

## 1. Module Structure

The CCTV intelligence layer lives under `internal/vision/` (not `internal/cctv/`) to reflect that the module's value is in the structured data it produces from visual analysis, not in the camera hardware management itself. The naming also leaves room for non-camera visual inputs like uploaded photos for waste logging or receiving verification.

### Package Layout

```
internal/vision/
    types.go          -- Core domain types: Camera, Detection, Zone, Clip, etc.
    service.go        -- Service struct, constructor, camera CRUD, detection queries
    pipeline.go       -- Ingestion pipeline: frame dispatch, detection aggregation
    detectors.go      -- Detector interface and built-in detector implementations
    storage.go        -- S3 clip storage: upload, presign, lifecycle tagging
    events.go         -- NATS event publishing and subscription registration
    privacy.go        -- Privacy zone masking, PII scrubbing, retention enforcement
    config.go         -- Per-camera and per-zone configuration types
    edge.go           -- Edge agent protocol: registration, heartbeat, result ingestion
```

### Key Interfaces and Types

```go
package vision

import (
    "context"
    "encoding/json"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/opsnerve/fireline/internal/event"
)

// ─── Service ───────────────────────────────────────────────────────────────

// Service provides camera management, detection event storage, clip management,
// and orchestration of the vision processing pipeline.
type Service struct {
    pool    *pgxpool.Pool
    bus     *event.Bus
    store   ClipStore
    cfg     GlobalConfig
}

func New(pool *pgxpool.Pool, bus *event.Bus, store ClipStore, cfg GlobalConfig) *Service {
    return &Service{pool: pool, bus: bus, store: store, cfg: cfg}
}

// ─── Core Domain Types ─────────────────────────────────────────────────────

// Camera represents a registered IP camera at a location.
type Camera struct {
    CameraID    string          `json:"camera_id"`
    OrgID       string          `json:"org_id"`
    LocationID  string          `json:"location_id"`
    Name        string          `json:"name"`
    StreamURL   string          `json:"stream_url"`   // RTSP URL, encrypted at rest
    Orientation string          `json:"orientation"`   // front_door, kitchen, register, dining, drive_thru, receiving
    Status      string          `json:"status"`        // active, inactive, offline, maintenance
    Zones       []DetectionZone `json:"zones"`
    PrivacyMask json.RawMessage `json:"privacy_mask"`  // polygon coordinates for blacked-out regions
    Config      CameraConfig    `json:"config"`
    LastFrameAt *time.Time      `json:"last_frame_at"`
    CreatedAt   time.Time       `json:"created_at"`
    UpdatedAt   time.Time       `json:"updated_at"`
}

// CameraConfig holds per-camera processing parameters.
type CameraConfig struct {
    FPS              int     `json:"fps"`                // target frame extraction rate (1-5 for analytics, 15-30 for recording)
    ResolutionCap    string  `json:"resolution_cap"`     // max resolution for processing: "720p", "1080p"
    DetectionModels  []string `json:"detection_models"`  // which models to run: ["occupancy", "compliance", "activity"]
    Sensitivity      float64 `json:"sensitivity"`        // 0.0 (loose) to 1.0 (strict) -- affects confidence thresholds
    RetentionDays    int     `json:"retention_days"`     // clip retention before cold storage transition
    RecordContinuous bool    `json:"record_continuous"`  // whether to record continuously or event-only clips
}

// DetectionZone defines a named region of interest within a camera's field of view.
type DetectionZone struct {
    ZoneID      string      `json:"zone_id"`
    Name        string      `json:"name"`        // "register_area", "prep_station_1", "front_entrance"
    ZoneType    string      `json:"zone_type"`   // occupancy, compliance, activity, restricted
    Polygon     []Point     `json:"polygon"`     // normalized 0.0-1.0 coordinates
    Thresholds  ZoneThreshold `json:"thresholds"`
}

// Point is a 2D coordinate normalized to the frame dimensions.
type Point struct {
    X float64 `json:"x"`
    Y float64 `json:"y"`
}

// ZoneThreshold holds per-zone alerting thresholds.
type ZoneThreshold struct {
    MaxOccupancy     *int     `json:"max_occupancy,omitempty"`
    MinOccupancy     *int     `json:"min_occupancy,omitempty"`
    DwellTimeSecs    *int     `json:"dwell_time_secs,omitempty"`     // alert if someone stays longer than this
    EmptyAlertSecs   *int     `json:"empty_alert_secs,omitempty"`    // alert if zone is empty for this long
    ConfidenceFloor  float64  `json:"confidence_floor"`              // minimum model confidence to count as a detection
}

// ─── Detection Events ──────────────────────────────────────────────────────

// Detection represents a single structured observation from the vision pipeline.
type Detection struct {
    DetectionID   string          `json:"detection_id"`
    OrgID         string          `json:"org_id"`
    LocationID    string          `json:"location_id"`
    CameraID      string          `json:"camera_id"`
    ZoneID        *string         `json:"zone_id"`
    DetectionType string          `json:"detection_type"` // occupancy, compliance_violation, activity, anomaly
    Confidence    float64         `json:"confidence"`
    Payload       json.RawMessage `json:"payload"`        // type-specific structured data
    ClipID        *string         `json:"clip_id"`        // associated video clip if extracted
    FrameURL      *string         `json:"frame_url"`      // S3 URL of the key frame snapshot
    DetectedAt    time.Time       `json:"detected_at"`
    CreatedAt     time.Time       `json:"created_at"`
}

// OccupancyPayload is the structured data for occupancy-type detections.
type OccupancyPayload struct {
    PersonCount    int     `json:"person_count"`
    PreviousCount  int     `json:"previous_count"`
    ZoneCapacity   *int    `json:"zone_capacity,omitempty"`
    UtilizationPct float64 `json:"utilization_pct"`
}

// CompliancePayload is the structured data for compliance violations.
type CompliancePayload struct {
    ViolationType string  `json:"violation_type"` // no_gloves, no_hairnet, no_apron, wet_floor_no_sign, blocked_exit
    PersonIndex   int     `json:"person_index"`   // which detected person in the frame
    RuleID        string  `json:"rule_id"`        // reference to the compliance rule that was violated
    Severity      string  `json:"severity"`       // info, warning, critical
}

// ActivityPayload is the structured data for general activity tracking.
type ActivityPayload struct {
    ActivityType  string  `json:"activity_type"` // station_active, station_idle, receiving_delivery, waste_disposal
    DurationSecs  int     `json:"duration_secs"`
    ActorCount    int     `json:"actor_count"`
}

// ─── Video Clips ───────────────────────────────────────────────────────────

// Clip represents an extracted video segment stored in S3.
type Clip struct {
    ClipID       string    `json:"clip_id"`
    OrgID        string    `json:"org_id"`
    LocationID   string    `json:"location_id"`
    CameraID     string    `json:"camera_id"`
    StartTime    time.Time `json:"start_time"`
    EndTime      time.Time `json:"end_time"`
    DurationSecs int       `json:"duration_secs"`
    StorageTier  string    `json:"storage_tier"` // hot, warm, cold
    S3Key        string    `json:"s3_key"`
    SizeBytes    int64     `json:"size_bytes"`
    Trigger      string    `json:"trigger"`      // event, manual, continuous_segment
    CreatedAt    time.Time `json:"created_at"`
}

// ─── Detector Interface ────────────────────────────────────────────────────

// Detector is the interface for all vision analysis backends. Each detector
// receives a frame (or batch of frames) and returns zero or more raw detections.
// Detectors run either on the edge agent or in the cloud pipeline.
type Detector interface {
    // Name returns the detector identifier (e.g., "occupancy", "compliance").
    Name() string

    // Detect processes a frame and returns raw detection results.
    Detect(ctx context.Context, frame Frame) ([]RawDetection, error)

    // ModelVersion returns the current model version for schema tagging.
    ModelVersion() string
}

// Frame is a single video frame passed to detectors.
type Frame struct {
    CameraID   string
    LocationID string
    Timestamp  time.Time
    ImageData  []byte    // JPEG-encoded frame
    Width      int
    Height     int
    Zones      []DetectionZone
}

// RawDetection is the output of a single Detector before aggregation.
type RawDetection struct {
    DetectionType string
    Confidence    float64
    BoundingBox   *BoundingBox
    ZoneID        *string
    Payload       json.RawMessage
}

// BoundingBox represents a detected region in normalized coordinates.
type BoundingBox struct {
    X      float64 `json:"x"`
    Y      float64 `json:"y"`
    Width  float64 `json:"width"`
    Height float64 `json:"height"`
}

// ─── Clip Storage Interface ────────────────────────────────────────────────

// ClipStore abstracts S3-compatible blob storage for video clips and frame snapshots.
type ClipStore interface {
    // UploadClip stores a video clip and returns the S3 key.
    UploadClip(ctx context.Context, orgID string, clip *Clip, data []byte) (s3Key string, err error)

    // UploadFrame stores a JPEG frame snapshot and returns the S3 URL.
    UploadFrame(ctx context.Context, orgID, cameraID string, ts time.Time, jpeg []byte) (url string, err error)

    // PresignClip returns a time-limited presigned URL for clip download.
    PresignClip(ctx context.Context, s3Key string, expiry time.Duration) (url string, err error)

    // DeleteClip removes a clip from storage.
    DeleteClip(ctx context.Context, s3Key string) error

    // TransitionTier moves a clip between storage tiers (e.g., hot -> warm).
    TransitionTier(ctx context.Context, s3Key string, targetTier string) error
}

// ─── Edge Agent Protocol ───────────────────────────────────────────────────

// EdgeAgent represents a registered edge processing node at a location.
type EdgeAgent struct {
    AgentID      string    `json:"agent_id"`
    OrgID        string    `json:"org_id"`
    LocationID   string    `json:"location_id"`
    Hostname     string    `json:"hostname"`
    Version      string    `json:"version"`
    Cameras      []string  `json:"cameras"`       // camera IDs managed by this agent
    Status       string    `json:"status"`         // online, offline, degraded
    LastHeartbeat time.Time `json:"last_heartbeat"`
    CreatedAt    time.Time `json:"created_at"`
}

// EdgeDetectionBatch is the payload sent from an edge agent to the cloud API.
// The edge agent processes RTSP streams locally and sends structured results
// rather than raw video, minimizing bandwidth.
type EdgeDetectionBatch struct {
    AgentID    string          `json:"agent_id"`
    CameraID   string          `json:"camera_id"`
    BatchID    string          `json:"batch_id"`
    Detections []RawDetection  `json:"detections"`
    KeyFrame   []byte          `json:"key_frame,omitempty"` // JPEG of the most significant frame
    StartTime  time.Time       `json:"start_time"`
    EndTime    time.Time       `json:"end_time"`
}
```

### Design Rationale

**Why `internal/vision/` and not `internal/cctv/`**: The term "CCTV" implies closed-circuit television hardware. The module's responsibility is extracting structured intelligence from visual data. It should accept frames from edge agents, uploaded images, and potentially future sources like drone footage for multi-location audits. "Vision" captures this abstraction.

**Why a `Detector` interface**: Different detection tasks (occupancy counting, compliance checking, activity classification) require different ML models with different latency and accuracy profiles. The interface allows swapping implementations (edge-local ONNX models vs. cloud API calls to a hosted model) without changing the pipeline.

**Why `EdgeDetectionBatch` instead of streaming raw video**: A restaurant's uplink bandwidth is typically 10-50 Mbps shared with POS and guest WiFi. A single 1080p RTSP stream at 15fps consumes roughly 4 Mbps. Sending structured detection results (a few KB per batch) instead of raw frames reduces bandwidth by 99.9%. The edge agent handles frame extraction, model inference, and clip recording locally.

---

## 2. Ingestion Pipeline

### Architecture: Edge-Primary, Cloud-Supervised

```
  Restaurant Location                          Cloud (FireLine)
  ====================                         ================

  IP Camera 1 --RTSP--> +----------------+     +------------------+
  IP Camera 2 --RTSP--> | Edge Agent     |     | Vision Service   |
  IP Camera 3 --RTSP--> |  (Go binary)   |     |                  |
                         |                |     |                  |
                         | - FFmpeg RTSP  |     | - Receives       |
                         |   demux        | --> |   EdgeDetection  |
                         | - ONNX Runtime |HTTPS|   Batches        |
                         |   inference    |     | - Aggregates     |
                         | - Local clip   |     | - Publishes NATS |
                         |   recording    |     |   events         |
                         | - S3 upload    |     | - Stores to PG   |
                         |   (clips only) |     |                  |
                         +----------------+     +------------------+
                                |                        |
                                v                        v
                         +------------+          +---------------+
                         | S3 Bucket  |          | PostgreSQL    |
                         | (clips +   |          | (detections,  |
                         |  frames)   |          |  cameras,     |
                         +------------+          |  clips meta)  |
                                                 +---------------+
```

### Edge Agent Responsibilities

The edge agent is a standalone Go binary deployed on a small compute device (NVIDIA Jetson Orin Nano or equivalent) at each location. It runs independently and tolerates cloud connectivity loss.

1. **RTSP Demux**: Uses FFmpeg (called via `os/exec`) to pull RTSP streams from each camera and extract frames at the configured FPS.

2. **Privacy Masking**: Before any inference, applies privacy zone masks (black rectangles over defined polygon regions) to the raw frame. This ensures PII-sensitive areas never reach any model or storage.

3. **Local Inference**: Runs lightweight ONNX models via ONNX Runtime Go bindings. Models are pulled from S3 on agent startup and updated on heartbeat if a newer version is available.

4. **Detection Batching**: Accumulates detections over a configurable window (default: 5 seconds) and sends an `EdgeDetectionBatch` to the cloud API via HTTPS POST.

5. **Clip Recording**: When a detection exceeds the configured severity threshold, the agent extracts a clip from the RTSP stream buffer (maintained as a rolling 60-second ring buffer). The clip includes 10 seconds before and 20 seconds after the trigger event.

6. **Clip Upload**: Clips and key frames are uploaded directly to S3 using presigned URLs obtained from the cloud API. This avoids routing large binary data through the application server.

7. **Offline Resilience**: If the cloud API is unreachable, the agent queues detection batches to local disk (SQLite WAL-mode) and replays them when connectivity returns. Clips are retained locally up to configurable disk quota.

### Cloud Pipeline Flow

```go
// pipeline.go -- simplified flow

// IngestBatch is the cloud-side handler for edge agent detection batches.
// It validates, enriches, persists, and publishes events for each detection.
func (s *Service) IngestBatch(ctx context.Context, orgID string, batch EdgeDetectionBatch) error {
    tenantCtx := tenant.WithOrgID(ctx, orgID)

    // 1. Validate agent registration and camera ownership
    // 2. For each detection in the batch:
    //    a. Apply confidence threshold filtering (camera-specific sensitivity)
    //    b. Map to detection zones via bounding box -> polygon intersection
    //    c. Aggregate with recent detections (debounce duplicate events)
    //    d. Persist Detection row to PostgreSQL
    //    e. If key frame present, store frame URL
    //    f. Publish appropriate NATS event based on detection_type
    // 3. Update camera.last_frame_at heartbeat
    // 4. Evaluate cross-detection rules (e.g., occupancy + no-employee = unstaffed station)

    return database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
        for _, raw := range batch.Detections {
            detection := s.enrichDetection(raw, batch)
            if detection.Confidence < s.cameraConfidenceThreshold(batch.CameraID) {
                continue
            }
            if err := s.persistDetection(tenantCtx, tx, detection); err != nil {
                return err
            }
            s.publishDetectionEvent(ctx, detection)
        }
        return s.updateCameraHeartbeat(tenantCtx, tx, batch.CameraID, batch.EndTime)
    })
}
```

### Why Edge-Primary

| Consideration | Edge Processing | Cloud Processing |
|---|---|---|
| Bandwidth | Sends KB of structured data | Would send GB of raw video |
| Latency | Sub-second detection | 2-5 second round trip minimum |
| Privacy | Raw frames never leave premises | Raw frames traverse internet |
| Reliability | Works during internet outage | Dead during outage |
| Cost | One-time hardware cost | Per-frame API inference cost |
| Model updates | Pull new ONNX on heartbeat | Immediate, but same pull mechanism |

The cloud retains authority over camera configuration, detection thresholds, and model versions. The edge agent is a stateless executor that does what the cloud tells it to do.

---

## 3. Event Schema

All CCTV events follow the existing `event.Envelope` convention with `OrgID`, `LocationID`, `Source: "vision"`, and `SchemaVersion: 1`.

### Event Subjects and Payloads

#### `vision.occupancy.update`

Published when person count changes in any monitored zone. Debounced to a minimum 5-second interval per zone.

```go
// OccupancyUpdateEvent is published on vision.occupancy.update
type OccupancyUpdateEvent struct {
    CameraID       string  `json:"camera_id"`
    ZoneID         string  `json:"zone_id"`
    ZoneName       string  `json:"zone_name"`
    PersonCount    int     `json:"person_count"`
    PreviousCount  int     `json:"previous_count"`
    Delta          int     `json:"delta"`           // positive = people entered, negative = people left
    UtilizationPct float64 `json:"utilization_pct"` // 0-100 if zone has capacity defined
    FrameURL       string  `json:"frame_url"`
    DetectedAt     string  `json:"detected_at"`     // RFC3339
}
```

**Consumers**: Customer module (dine-in occupancy, queue length), Operations module (station staffing), Labor module (station occupancy validation), Multi-location (cross-location benchmarking).

#### `vision.compliance.violation`

Published when a compliance rule is violated with confidence above the camera's threshold.

```go
// ComplianceViolationEvent is published on vision.compliance.violation
type ComplianceViolationEvent struct {
    CameraID      string  `json:"camera_id"`
    ZoneID        string  `json:"zone_id"`
    ViolationType string  `json:"violation_type"` // no_gloves, no_hairnet, no_apron, blocked_exit, wet_floor_no_sign
    Severity      string  `json:"severity"`       // warning, critical
    Confidence    float64 `json:"confidence"`
    FrameURL      string  `json:"frame_url"`
    ClipID        string  `json:"clip_id"`
    RuleID        string  `json:"rule_id"`
    DetectedAt    string  `json:"detected_at"`
}
```

**Consumers**: Operations module (compliance dashboard), Intelligence module (creates anomaly with video evidence), Alerting module (push notification to shift manager).

#### `vision.anomaly.detected`

Published when the vision system detects activity patterns that match anomaly signatures (buddy punching, cash drawer tampering, unauthorized after-hours access).

```go
// VisionAnomalyEvent is published on vision.anomaly.detected
type VisionAnomalyEvent struct {
    CameraID     string          `json:"camera_id"`
    AnomalyType  string          `json:"anomaly_type"` // buddy_punch, unauthorized_access, cash_irregularity, theft_indicator
    Severity     string          `json:"severity"`
    Confidence   float64         `json:"confidence"`
    Description  string          `json:"description"`
    Evidence     json.RawMessage `json:"evidence"` // type-specific evidence payload
    FrameURL     string          `json:"frame_url"`
    ClipID       string          `json:"clip_id"`
    DetectedAt   string          `json:"detected_at"`
}
```

**Consumers**: Intelligence module (creates Anomaly record with `evidence` containing video reference), Labor module (buddy-punch flag on shifts).

#### `vision.activity.update`

Published for general station activity tracking: idle detection, active prep, delivery receiving.

```go
// ActivityUpdateEvent is published on vision.activity.update
type ActivityUpdateEvent struct {
    CameraID     string `json:"camera_id"`
    ZoneID       string `json:"zone_id"`
    ActivityType string `json:"activity_type"` // station_active, station_idle, receiving_delivery, waste_disposal, cleaning
    ActorCount   int    `json:"actor_count"`
    DurationSecs int    `json:"duration_secs"` // how long this activity state has been observed
    DetectedAt   string `json:"detected_at"`
}
```

**Consumers**: Operations module (kitchen bottleneck detection, ticket queue correlation), Inventory module (waste bin monitoring, receiving verification), Labor module (station activity for ELU calculation).

#### `vision.camera.status`

Published when camera connectivity status changes.

```go
// CameraStatusEvent is published on vision.camera.status
type CameraStatusEvent struct {
    CameraID     string `json:"camera_id"`
    PreviousStatus string `json:"previous_status"`
    CurrentStatus  string `json:"current_status"` // online, offline, degraded
    Reason         string `json:"reason"`          // heartbeat_timeout, agent_reconnect, manual
    DetectedAt     string `json:"detected_at"`
}
```

**Consumers**: Alerting module (notify manager of camera outage), Multi-location (system health dashboard).

### Event Registration

```go
// events.go

// RegisterHandlers subscribes the vision service to events from other modules
// that trigger vision-related actions.
func (s *Service) RegisterHandlers() {
    // When a shift clock-in occurs, start monitoring for buddy-punch patterns
    s.bus.Subscribe("labor.shift.clock_in", s.onShiftClockIn)

    // When inventory receiving starts, activate receiving zone monitoring
    s.bus.Subscribe("inventory.po.receiving", s.onPOReceiving)
}

// PublishOccupancyUpdate publishes an occupancy change event.
func (s *Service) PublishOccupancyUpdate(ctx context.Context, evt OccupancyUpdateEvent, orgID, locationID string) {
    s.bus.Publish(ctx, event.Envelope{
        EventType:     "vision.occupancy.update",
        OrgID:         orgID,
        LocationID:    locationID,
        Source:        "vision",
        SchemaVersion: 1,
        Payload:       evt,
    })
}
```

---

## 4. Database Schema

### Migration: `021_vision.sql`

```sql
-- Vision / CCTV intelligence layer
-- Follows existing RLS pattern: org_id on every table, policy per table.

-- ============================================================
-- EDGE AGENTS
-- ============================================================

CREATE TABLE vision_edge_agents (
    agent_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    location_id    UUID NOT NULL REFERENCES locations(location_id),
    hostname       TEXT NOT NULL,
    version        TEXT NOT NULL DEFAULT '0.0.0',
    status         TEXT NOT NULL DEFAULT 'offline'
                   CHECK (status IN ('online', 'offline', 'degraded')),
    last_heartbeat TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_vision_agents_org ON vision_edge_agents(org_id);
CREATE INDEX idx_vision_agents_location ON vision_edge_agents(location_id);

ALTER TABLE vision_edge_agents ENABLE ROW LEVEL SECURITY;
ALTER TABLE vision_edge_agents FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON vision_edge_agents
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON vision_edge_agents TO fireline_app;

-- ============================================================
-- CAMERAS
-- ============================================================

CREATE TABLE vision_cameras (
    camera_id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id        UUID NOT NULL REFERENCES organizations(org_id),
    location_id   UUID NOT NULL REFERENCES locations(location_id),
    agent_id      UUID REFERENCES vision_edge_agents(agent_id),
    name          TEXT NOT NULL,
    stream_url    TEXT NOT NULL,  -- encrypted at application layer before storage
    orientation   TEXT NOT NULL DEFAULT 'general'
                  CHECK (orientation IN (
                      'front_door', 'kitchen', 'register', 'dining',
                      'drive_thru', 'receiving', 'storage', 'general'
                  )),
    status        TEXT NOT NULL DEFAULT 'inactive'
                  CHECK (status IN ('active', 'inactive', 'offline', 'maintenance')),
    zones         JSONB NOT NULL DEFAULT '[]',         -- array of DetectionZone
    privacy_mask  JSONB NOT NULL DEFAULT '[]',         -- array of polygon coordinates
    config        JSONB NOT NULL DEFAULT '{
        "fps": 2,
        "resolution_cap": "720p",
        "detection_models": ["occupancy"],
        "sensitivity": 0.7,
        "retention_days": 30,
        "record_continuous": false
    }',
    last_frame_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_vision_cameras_org ON vision_cameras(org_id);
CREATE INDEX idx_vision_cameras_location ON vision_cameras(org_id, location_id);
CREATE INDEX idx_vision_cameras_status ON vision_cameras(status) WHERE status = 'active';
CREATE INDEX idx_vision_cameras_agent ON vision_cameras(agent_id);

ALTER TABLE vision_cameras ENABLE ROW LEVEL SECURITY;
ALTER TABLE vision_cameras FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON vision_cameras
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON vision_cameras TO fireline_app;

-- ============================================================
-- DETECTIONS (TimescaleDB hypertable for time-series queries)
-- ============================================================

CREATE TABLE vision_detections (
    detection_id   UUID NOT NULL DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    location_id    UUID NOT NULL REFERENCES locations(location_id),
    camera_id      UUID NOT NULL REFERENCES vision_cameras(camera_id),
    zone_id        TEXT,                                    -- references zone within camera JSONB
    detection_type TEXT NOT NULL
                   CHECK (detection_type IN (
                       'occupancy', 'compliance_violation', 'activity', 'anomaly'
                   )),
    confidence     NUMERIC(4,3) NOT NULL CHECK (confidence BETWEEN 0 AND 1),
    payload        JSONB NOT NULL,                          -- type-specific structured data
    clip_id        UUID,                                    -- FK to vision_clips if clip was extracted
    frame_url      TEXT,                                    -- S3 URL to key frame snapshot
    model_version  TEXT NOT NULL DEFAULT 'v1.0.0',
    detected_at    TIMESTAMPTZ NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- Composite PK for TimescaleDB hypertable
    PRIMARY KEY (detection_id, detected_at)
);

-- Convert to hypertable for efficient time-range queries and automatic partitioning.
-- Chunk interval of 1 day balances query performance with chunk management overhead.
SELECT create_hypertable('vision_detections', 'detected_at',
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

CREATE INDEX idx_vision_det_org_time ON vision_detections(org_id, detected_at DESC);
CREATE INDEX idx_vision_det_location_time ON vision_detections(location_id, detected_at DESC);
CREATE INDEX idx_vision_det_camera_time ON vision_detections(camera_id, detected_at DESC);
CREATE INDEX idx_vision_det_type ON vision_detections(detection_type, detected_at DESC);
CREATE INDEX idx_vision_det_zone ON vision_detections(zone_id, detected_at DESC)
    WHERE zone_id IS NOT NULL;
-- GIN index for querying into detection payloads (e.g., violation_type)
CREATE INDEX idx_vision_det_payload ON vision_detections USING gin(payload jsonb_path_ops);

ALTER TABLE vision_detections ENABLE ROW LEVEL SECURITY;
ALTER TABLE vision_detections FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON vision_detections
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT ON vision_detections TO fireline_app;
-- No UPDATE/DELETE: detections are append-only. Retention handled by TimescaleDB policies.

-- Automatic retention: drop chunks older than 90 days.
SELECT add_retention_policy('vision_detections', INTERVAL '90 days', if_not_exists => TRUE);

-- ============================================================
-- CONTINUOUS AGGREGATES for dashboard queries
-- ============================================================

-- Hourly occupancy summaries per zone
CREATE MATERIALIZED VIEW vision_occupancy_hourly
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', detected_at) AS bucket,
    org_id,
    location_id,
    camera_id,
    zone_id,
    AVG((payload->>'person_count')::INT)  AS avg_count,
    MAX((payload->>'person_count')::INT)  AS max_count,
    MIN((payload->>'person_count')::INT)  AS min_count,
    COUNT(*)                               AS sample_count
FROM vision_detections
WHERE detection_type = 'occupancy'
GROUP BY bucket, org_id, location_id, camera_id, zone_id
WITH NO DATA;

SELECT add_continuous_aggregate_policy('vision_occupancy_hourly',
    start_offset   => INTERVAL '3 hours',
    end_offset     => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour',
    if_not_exists => TRUE
);

-- Hourly compliance violation counts
CREATE MATERIALIZED VIEW vision_compliance_hourly
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', detected_at) AS bucket,
    org_id,
    location_id,
    camera_id,
    payload->>'violation_type' AS violation_type,
    COUNT(*)                    AS violation_count,
    AVG(confidence)             AS avg_confidence
FROM vision_detections
WHERE detection_type = 'compliance_violation'
GROUP BY bucket, org_id, location_id, camera_id, payload->>'violation_type'
WITH NO DATA;

SELECT add_continuous_aggregate_policy('vision_compliance_hourly',
    start_offset   => INTERVAL '3 hours',
    end_offset     => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour',
    if_not_exists => TRUE
);

-- ============================================================
-- VIDEO CLIPS
-- ============================================================

CREATE TABLE vision_clips (
    clip_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id        UUID NOT NULL REFERENCES organizations(org_id),
    location_id   UUID NOT NULL REFERENCES locations(location_id),
    camera_id     UUID NOT NULL REFERENCES vision_cameras(camera_id),
    start_time    TIMESTAMPTZ NOT NULL,
    end_time      TIMESTAMPTZ NOT NULL,
    duration_secs INT GENERATED ALWAYS AS
                  (EXTRACT(EPOCH FROM (end_time - start_time))::INT) STORED,
    storage_tier  TEXT NOT NULL DEFAULT 'hot'
                  CHECK (storage_tier IN ('hot', 'warm', 'cold')),
    s3_key        TEXT NOT NULL,
    size_bytes    BIGINT NOT NULL DEFAULT 0,
    trigger       TEXT NOT NULL DEFAULT 'event'
                  CHECK (trigger IN ('event', 'manual', 'continuous_segment')),
    metadata      JSONB NOT NULL DEFAULT '{}',  -- trigger details, associated detection IDs
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_vision_clips_org ON vision_clips(org_id);
CREATE INDEX idx_vision_clips_location ON vision_clips(org_id, location_id);
CREATE INDEX idx_vision_clips_camera_time ON vision_clips(camera_id, start_time DESC);
CREATE INDEX idx_vision_clips_tier ON vision_clips(storage_tier);
CREATE INDEX idx_vision_clips_trigger ON vision_clips(trigger);

ALTER TABLE vision_clips ENABLE ROW LEVEL SECURITY;
ALTER TABLE vision_clips FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON vision_clips
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE ON vision_clips TO fireline_app;

-- ============================================================
-- COMPLIANCE RULES (configurable per location)
-- ============================================================

CREATE TABLE vision_compliance_rules (
    rule_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id        UUID NOT NULL REFERENCES organizations(org_id),
    location_id   UUID REFERENCES locations(location_id), -- NULL = org-wide default
    name          TEXT NOT NULL,
    violation_type TEXT NOT NULL,
    description   TEXT,
    severity      TEXT NOT NULL DEFAULT 'warning'
                  CHECK (severity IN ('info', 'warning', 'critical')),
    zones         TEXT[] NOT NULL DEFAULT '{}',  -- which zone_types this rule applies to
    active        BOOLEAN NOT NULL DEFAULT true,
    config        JSONB NOT NULL DEFAULT '{}',   -- rule-specific params (e.g., min_glove_confidence)
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_vision_rules_org ON vision_compliance_rules(org_id);
CREATE INDEX idx_vision_rules_location ON vision_compliance_rules(location_id);
CREATE INDEX idx_vision_rules_active ON vision_compliance_rules(active) WHERE active = true;

ALTER TABLE vision_compliance_rules ENABLE ROW LEVEL SECURITY;
ALTER TABLE vision_compliance_rules FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON vision_compliance_rules
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON vision_compliance_rules TO fireline_app;

-- ============================================================
-- ADD vision-related anomaly types to the anomalies CHECK constraint
-- ============================================================

ALTER TABLE anomalies DROP CONSTRAINT IF EXISTS anomalies_type_check;
ALTER TABLE anomalies ADD CONSTRAINT anomalies_type_check
    CHECK (type IN (
        'void_pattern', 'cash_variance', 'clock_irregularity',
        'shrinkage', 'transaction_pattern',
        -- New vision-sourced types:
        'buddy_punch_visual', 'unauthorized_access', 'compliance_violation',
        'cash_irregularity_visual', 'theft_indicator', 'station_abandonment'
    ));

-- Add FK from detections to clips (deferred because both tables are in this migration)
ALTER TABLE vision_detections
    ADD CONSTRAINT fk_detection_clip
    FOREIGN KEY (clip_id) REFERENCES vision_clips(clip_id);
```

### Schema Design Rationale

**TimescaleDB hypertable for detections**: The `vision_detections` table will receive the highest write throughput of any table in the system. At 2 FPS across 8 cameras with 3 zones each, a location generates roughly 50 detection rows per second, or 4.3M rows per day per location. TimescaleDB handles this gracefully with automatic chunking, and the continuous aggregates provide fast dashboard queries without expensive real-time aggregation.

**Continuous aggregates for dashboards**: The hourly occupancy and compliance views are pre-materialized. When a manager opens the dashboard and asks "what was peak occupancy in the dining room today?", the query hits the continuous aggregate (a few hundred rows) instead of scanning millions of raw detections.

**Append-only detections**: No UPDATE or DELETE grants on `vision_detections`. This ensures auditability: a detection once recorded cannot be tampered with. Retention is handled exclusively by TimescaleDB's `add_retention_policy` which drops entire chunks.

**JSONB for zones and config**: Camera zones and configuration change infrequently but have complex nested structures. Storing them as JSONB avoids a proliferation of junction tables while still supporting efficient indexing via `jsonb_path_ops`.

**Separate compliance_rules table**: Compliance rules are organizational policy, not detection data. Keeping them separate allows managers to configure rules through the dashboard without touching camera or detection tables.

---

## 5. Integration Points

### Labor Module

| CCTV Capability | Integration | Event/Mechanism |
|---|---|---|
| **Station occupancy** | Subscribe to `vision.occupancy.update` where zone_type is a kitchen station. Cross-reference with shift assignments to determine if the assigned employee is actually at their station. Feed into ELU (Employee Labor Utilization) calculations in `internal/labor/elu.go`. | `vision.occupancy.update` -> `labor.elu.recalculate` |
| **Buddy-punch detection** | Subscribe to `vision.anomaly.detected` where anomaly_type is `buddy_punch`. When the existing clock_irregularity detection in the intelligence module flags sequential same-device clock-ins, the vision system corroborates by checking if the camera near the time clock captured two distinct individuals in rapid succession. Creates an anomaly with both POS evidence and video evidence merged into the `evidence` JSON field. | `labor.shift.clock_in` triggers vision monitoring; `vision.anomaly.detected` enriches existing anomalies |

```go
// In internal/labor/service.go or a new labor/vision_handlers.go

func (s *LaborService) onVisionOccupancyUpdate(ctx context.Context, env event.Envelope) error {
    // Parse OccupancyUpdateEvent
    // Look up which employee is assigned to this zone's station for the current shift
    // If zone person_count == 0 and an employee is assigned, flag as potential station abandonment
    // If zone person_count > 0 and no employee is assigned, flag as unauthorized station access
    // Update ELU active-station-time tracking
}
```

### Operations Module

| CCTV Capability | Integration |
|---|---|
| **Kitchen bottlenecks** | Subscribe to `vision.occupancy.update` for kitchen zones. Correlate person_count and activity_type with existing `KitchenCapacity` data from `internal/operations/kitchen.go`. When vision sees a station is unmanned but the KDS has queued tickets for that station, generate a kitchen bottleneck alert. |
| **Ticket queue monitoring** | Subscribe to `vision.activity.update` where activity_type is `station_active` or `station_idle`. The operations module already computes `TicketTimeEstimate`. Vision data provides ground-truth validation: if the model predicts a ticket should be done in 8 minutes but the station has been idle for 3 of those minutes, the estimate needs to be adjusted upward. |

### Inventory Module

| CCTV Capability | Integration |
|---|---|
| **Waste bin monitoring** | Subscribe to `vision.activity.update` where activity_type is `waste_disposal`. When the receiving camera zone detects waste disposal activity, create a pending waste log entry in the inventory system that a manager must confirm (preventing false entries). This cross-references with `internal/inventory/` waste logging. |
| **Receiving verification** | Subscribe to `inventory.po.receiving` events in the vision module. When a PO is being received, the vision system activates heightened monitoring on the receiving dock camera. It captures a clip of the delivery, counts visible boxes/pallets if the model supports it, and attaches the clip and frame to the PO record's evidence field. |

### Customer Module

| CCTV Capability | Integration |
|---|---|
| **Dine-in occupancy** | Subscribe to `vision.occupancy.update` for dining room zones. Feed real-time occupancy into the customer-facing wait time estimator. If occupancy exceeds 80% of zone capacity, trigger waitlist recommendations. |
| **Queue length** | Subscribe to `vision.occupancy.update` for front_door/register zones. The person count in the queue zone, combined with historical average service time, produces a dynamic "estimated wait time" published to the customer module. |

### Intelligence Module

The intelligence module is the primary consumer of `vision.anomaly.detected`. The integration is straightforward because the anomaly table already supports a JSONB `evidence` field.

```go
// In internal/intelligence/vision_handlers.go

func (s *Service) onVisionAnomaly(ctx context.Context, env event.Envelope) error {
    var evt VisionAnomalyEvent
    // ... unmarshal from env.Payload

    // Build evidence that combines the vision data with any existing POS/shift data
    evidence, _ := json.Marshal(map[string]any{
        "source":      "vision",
        "camera_id":   evt.CameraID,
        "frame_url":   evt.FrameURL,
        "clip_id":     evt.ClipID,
        "confidence":  evt.Confidence,
        "model_data":  evt.Evidence,
    })

    input := AnomalyInput{
        LocationID:  env.LocationID,
        Type:        mapVisionAnomalyType(evt.AnomalyType), // maps to the anomalies CHECK constraint
        Severity:    evt.Severity,
        Title:       evt.Description,
        Description: fmt.Sprintf("Vision system detected %s with %.0f%% confidence", evt.AnomalyType, evt.Confidence*100),
        Evidence:    json.RawMessage(evidence),
    }

    _, err := s.CreateAnomaly(ctx, env.OrgID, input)
    return err
}
```

### Multi-Location Module

| CCTV Capability | Integration |
|---|---|
| **Cross-location visual benchmarking** | The `vision_occupancy_hourly` continuous aggregate provides per-location, per-zone hourly occupancy data. The multi-location portfolio module queries this view to produce comparative dashboards: which locations have the highest peak utilization, which have the most idle kitchen time, which have the longest customer queue durations. This is a read-only integration: no events, just queries against the continuous aggregate. |

---

## 6. Video Storage

### S3 Key Structure

```
s3://{bucket}/
  {org_id}/
    {location_id}/
      clips/
        {YYYY}/{MM}/{DD}/
          {camera_id}_{timestamp}_{clip_id}.mp4
      frames/
        {YYYY}/{MM}/{DD}/
          {camera_id}_{timestamp}_{detection_id}.jpg
      models/
        {model_name}_{version}.onnx
```

### Tiered Storage Lifecycle

| Tier | Duration | Storage Class | Access Pattern | Monthly Cost (per TB) |
|---|---|---|---|---|
| **Hot** | Days 0-7 | S3 Standard | Frequent playback, investigation | ~$23 |
| **Warm** | Days 8-30 | S3 Infrequent Access | Occasional review, anomaly follow-up | ~$12.50 |
| **Cold** | Days 31-90 | S3 Glacier Instant Retrieval | Compliance archive, legal holds | ~$4 |
| **Delete** | Day 91+ | Deleted | N/A | $0 |

### S3 Lifecycle Rules

```json
{
  "Rules": [
    {
      "ID": "vision-clips-warm",
      "Filter": {"Prefix": "clips/"},
      "Status": "Enabled",
      "Transitions": [
        {
          "Days": 7,
          "StorageClass": "STANDARD_IA"
        }
      ]
    },
    {
      "ID": "vision-clips-cold",
      "Filter": {"Prefix": "clips/"},
      "Status": "Enabled",
      "Transitions": [
        {
          "Days": 30,
          "StorageClass": "GLACIER_IR"
        }
      ]
    },
    {
      "ID": "vision-clips-expire",
      "Filter": {"Prefix": "clips/"},
      "Status": "Enabled",
      "Expiration": {"Days": 90}
    },
    {
      "ID": "vision-frames-expire",
      "Filter": {"Prefix": "frames/"},
      "Status": "Enabled",
      "Expiration": {"Days": 30}
    }
  ]
}
```

### Clip Extraction from Continuous Streams

The edge agent maintains a rolling 60-second ring buffer of the RTSP stream using FFmpeg's segment muxer. When a detection triggers clip extraction:

1. **Pre-roll**: Copy the 10 seconds before the trigger event from the ring buffer.
2. **Post-roll**: Continue recording for 20 seconds after the trigger.
3. **Encode**: Transcode to H.264 baseline profile at CRF 28 for storage efficiency (typical clip: 30 seconds, 720p, ~2-4 MB).
4. **Upload**: Request a presigned PUT URL from the cloud API, upload directly to S3.
5. **Register**: POST clip metadata (start_time, end_time, s3_key, size_bytes) to the cloud API for database registration.

### Retention Overrides

- Clips linked to open or investigating anomalies are exempt from lifecycle transitions until the anomaly is resolved. The `storage.go` module checks the anomaly status before allowing tier transitions.
- Manual "legal hold" flag on a clip prevents all lifecycle transitions and deletion indefinitely.
- Per-camera `retention_days` in `CameraConfig` overrides the default lifecycle for that camera's clips (useful for high-security zones like cash offices).

### Storage Budget Estimation

For a 4-location restaurant group with 6 cameras per location, event-triggered recording (not continuous), at approximately 50 clips/day/location averaging 3 MB each:

- Daily: 50 clips x 4 locations x 3 MB = 600 MB/day
- Monthly: ~18 GB in hot tier
- After lifecycle: ~18 GB hot + ~54 GB warm + ~54 GB cold = ~126 GB total
- Monthly cost: (0.018 x $23) + (0.054 x $12.50) + (0.054 x $4) = ~$1.30/month

This is negligible. Even continuous recording at 720p/2fps would stay under $50/month.

---

## 7. API Endpoints

All endpoints are under `/api/v1/vision/` and require the standard `authMW` middleware. They follow the existing pattern in `internal/api/handlers.go`.

### Camera Management

```
POST   /api/v1/vision/cameras                    -- Register a new camera
GET    /api/v1/vision/cameras                    -- List cameras (filter: location_id, status)
GET    /api/v1/vision/cameras/{id}               -- Get camera details
PUT    /api/v1/vision/cameras/{id}               -- Update camera config, zones, privacy mask
DELETE /api/v1/vision/cameras/{id}               -- Deactivate camera (soft delete via status)
PUT    /api/v1/vision/cameras/{id}/zones         -- Update detection zones
PUT    /api/v1/vision/cameras/{id}/privacy       -- Update privacy mask polygons
GET    /api/v1/vision/cameras/{id}/status        -- Get camera health and last frame time
```

### Detection Events

```
GET    /api/v1/vision/detections                 -- Query detections (filter: camera_id, zone_id, type, time range)
GET    /api/v1/vision/detections/{id}            -- Get single detection with frame URL
GET    /api/v1/vision/occupancy                  -- Current occupancy snapshot for all zones at a location
GET    /api/v1/vision/occupancy/history          -- Hourly occupancy history (reads continuous aggregate)
GET    /api/v1/vision/compliance/violations       -- List compliance violations (filter: type, severity, time range)
GET    /api/v1/vision/compliance/summary          -- Compliance summary: violations per type per day
```

### Video Clips

```
GET    /api/v1/vision/clips                      -- List clips (filter: camera_id, trigger, time range)
GET    /api/v1/vision/clips/{id}                 -- Get clip metadata
GET    /api/v1/vision/clips/{id}/url             -- Get presigned download URL (expires in 15 minutes)
POST   /api/v1/vision/clips/extract              -- Request manual clip extraction from a camera's stream
PUT    /api/v1/vision/clips/{id}/hold            -- Set or remove legal hold
```

### Edge Agent

```
POST   /api/v1/vision/agents/register            -- Register a new edge agent
POST   /api/v1/vision/agents/{id}/heartbeat       -- Agent heartbeat (returns config updates, model URLs)
POST   /api/v1/vision/agents/{id}/detections       -- Ingest a detection batch from edge agent
POST   /api/v1/vision/agents/{id}/clips/presign   -- Request presigned S3 URL for clip upload
```

### Compliance Rules

```
GET    /api/v1/vision/compliance/rules            -- List compliance rules
POST   /api/v1/vision/compliance/rules            -- Create a compliance rule
PUT    /api/v1/vision/compliance/rules/{id}       -- Update a compliance rule
DELETE /api/v1/vision/compliance/rules/{id}       -- Deactivate a compliance rule
```

### Handler Registration

```go
// internal/api/vision_handler.go

type VisionHandler struct {
    svc *vision.Service
}

func NewVisionHandler(svc *vision.Service) *VisionHandler {
    return &VisionHandler{svc: svc}
}

func (h *VisionHandler) RegisterRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
    // Cameras
    mux.Handle("POST /api/v1/vision/cameras", authMW(http.HandlerFunc(h.CreateCamera)))
    mux.Handle("GET /api/v1/vision/cameras", authMW(http.HandlerFunc(h.ListCameras)))
    mux.Handle("GET /api/v1/vision/cameras/{id}", authMW(http.HandlerFunc(h.GetCamera)))
    mux.Handle("PUT /api/v1/vision/cameras/{id}", authMW(http.HandlerFunc(h.UpdateCamera)))
    mux.Handle("DELETE /api/v1/vision/cameras/{id}", authMW(http.HandlerFunc(h.DeleteCamera)))
    mux.Handle("PUT /api/v1/vision/cameras/{id}/zones", authMW(http.HandlerFunc(h.UpdateCameraZones)))
    mux.Handle("PUT /api/v1/vision/cameras/{id}/privacy", authMW(http.HandlerFunc(h.UpdatePrivacyMask)))
    mux.Handle("GET /api/v1/vision/cameras/{id}/status", authMW(http.HandlerFunc(h.GetCameraStatus)))

    // Detections
    mux.Handle("GET /api/v1/vision/detections", authMW(http.HandlerFunc(h.ListDetections)))
    mux.Handle("GET /api/v1/vision/detections/{id}", authMW(http.HandlerFunc(h.GetDetection)))
    mux.Handle("GET /api/v1/vision/occupancy", authMW(http.HandlerFunc(h.GetOccupancySnapshot)))
    mux.Handle("GET /api/v1/vision/occupancy/history", authMW(http.HandlerFunc(h.GetOccupancyHistory)))
    mux.Handle("GET /api/v1/vision/compliance/violations", authMW(http.HandlerFunc(h.ListViolations)))
    mux.Handle("GET /api/v1/vision/compliance/summary", authMW(http.HandlerFunc(h.GetComplianceSummary)))

    // Clips
    mux.Handle("GET /api/v1/vision/clips", authMW(http.HandlerFunc(h.ListClips)))
    mux.Handle("GET /api/v1/vision/clips/{id}", authMW(http.HandlerFunc(h.GetClip)))
    mux.Handle("GET /api/v1/vision/clips/{id}/url", authMW(http.HandlerFunc(h.GetClipURL)))
    mux.Handle("POST /api/v1/vision/clips/extract", authMW(http.HandlerFunc(h.ExtractClip)))
    mux.Handle("PUT /api/v1/vision/clips/{id}/hold", authMW(http.HandlerFunc(h.SetClipHold)))

    // Edge agents (authenticated via agent API key, not user JWT)
    mux.Handle("POST /api/v1/vision/agents/register", authMW(http.HandlerFunc(h.RegisterAgent)))
    mux.Handle("POST /api/v1/vision/agents/{id}/heartbeat", authMW(http.HandlerFunc(h.AgentHeartbeat)))
    mux.Handle("POST /api/v1/vision/agents/{id}/detections", authMW(http.HandlerFunc(h.IngestDetections)))
    mux.Handle("POST /api/v1/vision/agents/{id}/clips/presign", authMW(http.HandlerFunc(h.PresignClipUpload)))

    // Compliance rules
    mux.Handle("GET /api/v1/vision/compliance/rules", authMW(http.HandlerFunc(h.ListComplianceRules)))
    mux.Handle("POST /api/v1/vision/compliance/rules", authMW(http.HandlerFunc(h.CreateComplianceRule)))
    mux.Handle("PUT /api/v1/vision/compliance/rules/{id}", authMW(http.HandlerFunc(h.UpdateComplianceRule)))
    mux.Handle("DELETE /api/v1/vision/compliance/rules/{id}", authMW(http.HandlerFunc(h.DeleteComplianceRule)))
}
```

### Live View Proxy

Live camera viewing is NOT proxied through the FireLine server. The edge agent runs a lightweight WebSocket-to-MJPEG bridge on the local network. The dashboard connects directly to the edge agent's local IP (discovered via the camera status endpoint which returns the agent's LAN address). This keeps latency under 200ms and avoids routing video through the cloud.

For remote viewing (manager accessing from home), the edge agent establishes a WebRTC peer connection brokered through the FireLine API as a signaling server. The actual video stream flows peer-to-peer via STUN/TURN. The API endpoint for this:

```
POST   /api/v1/vision/cameras/{id}/live/offer     -- WebRTC SDP offer relay
POST   /api/v1/vision/cameras/{id}/live/answer    -- WebRTC SDP answer relay
POST   /api/v1/vision/cameras/{id}/live/ice       -- ICE candidate exchange
```

---

## 8. Configuration

### Global Config

```go
// config.go

// GlobalConfig holds system-wide vision configuration.
type GlobalConfig struct {
    // MaxCamerasPerLocation limits camera count per location (licensing/resource guard)
    MaxCamerasPerLocation int `json:"max_cameras_per_location"`

    // DefaultRetentionDays is the fallback clip retention if not set per-camera
    DefaultRetentionDays int `json:"default_retention_days"`

    // DetectionBatchIntervalSecs is how often edge agents should send detection batches
    DetectionBatchIntervalSecs int `json:"detection_batch_interval_secs"`

    // OccupancyDebounceIntervalSecs prevents event spam from rapid count changes
    OccupancyDebounceIntervalSecs int `json:"occupancy_debounce_interval_secs"`

    // S3 configuration
    S3Bucket    string `json:"s3_bucket"`
    S3Region    string `json:"s3_region"`
    S3Endpoint  string `json:"s3_endpoint"`  // for MinIO/non-AWS

    // Model registry
    ModelRegistryPrefix string `json:"model_registry_prefix"` // S3 prefix for ONNX models
}
```

### Per-Location Camera Setup Flow

1. **Physical install**: Staff mounts camera and connects to location network.
2. **Discovery**: Edge agent scans the local subnet for ONVIF-compatible cameras and reports discovered RTSP URLs to the cloud API.
3. **Registration**: Manager opens the web dashboard, sees discovered cameras, and assigns each one a name and orientation (front_door, kitchen, register, etc.).
4. **Zone definition**: The dashboard shows the camera's live view. Manager draws polygons to define detection zones (e.g., "register queue area", "prep station 1"). The polygons are stored as normalized coordinates in the camera's `zones` JSONB.
5. **Privacy masking**: Manager draws rectangles/polygons over areas that must never be processed or stored (e.g., employee break room visible through a window, customer payment terminals showing card numbers). These are applied before any inference on the edge.
6. **Model assignment**: Based on orientation, the system automatically selects detection models. Kitchen cameras get occupancy + compliance (glove/hairnet detection). Register cameras get occupancy + anomaly (cash handling). Front door cameras get occupancy only. The manager can override.
7. **Sensitivity tuning**: A slider from 0.0 (loose, more detections, more false positives) to 1.0 (strict, fewer detections, fewer false positives). This maps to confidence thresholds on the edge agent.

### Detection Zone Configuration Examples

```json
{
  "zones": [
    {
      "zone_id": "z-prep-1",
      "name": "Prep Station 1",
      "zone_type": "occupancy",
      "polygon": [
        {"x": 0.1, "y": 0.2},
        {"x": 0.4, "y": 0.2},
        {"x": 0.4, "y": 0.8},
        {"x": 0.1, "y": 0.8}
      ],
      "thresholds": {
        "max_occupancy": 3,
        "empty_alert_secs": 300,
        "confidence_floor": 0.6
      }
    },
    {
      "zone_id": "z-register-queue",
      "name": "Register Queue",
      "zone_type": "occupancy",
      "polygon": [
        {"x": 0.5, "y": 0.1},
        {"x": 0.9, "y": 0.1},
        {"x": 0.9, "y": 0.7},
        {"x": 0.5, "y": 0.7}
      ],
      "thresholds": {
        "max_occupancy": 8,
        "dwell_time_secs": 180,
        "confidence_floor": 0.65
      }
    }
  ],
  "privacy_mask": [
    [
      {"x": 0.0, "y": 0.85},
      {"x": 0.3, "y": 0.85},
      {"x": 0.3, "y": 1.0},
      {"x": 0.0, "y": 1.0}
    ]
  ]
}
```

### Privacy and Compliance Configuration

```go
// privacy.go

// PrivacyConfig governs data protection behavior for the vision module.
type PrivacyConfig struct {
    // FaceBlurEnabled blurs all detected faces before frame storage.
    // When true, edge agents run face detection and apply Gaussian blur
    // before uploading frames or clips to S3.
    FaceBlurEnabled bool `json:"face_blur_enabled"`

    // RetainRawFrames controls whether raw (pre-blur) frames are ever stored.
    // Must be false for GDPR/CCPA compliance unless explicit consent is obtained.
    RetainRawFrames bool `json:"retain_raw_frames"`

    // EmployeeNotificationRequired indicates whether employees must be notified
    // that CCTV monitoring is active. When true, the system generates a
    // compliance task during onboarding to collect acknowledgment.
    EmployeeNotificationRequired bool `json:"employee_notification_required"`

    // DataRetentionMaxDays is the hard maximum for any clip or frame retention,
    // overriding per-camera settings. Set by legal/compliance.
    DataRetentionMaxDays int `json:"data_retention_max_days"`

    // AllowedPurposes documents the legal basis for processing.
    // Informational only, but included in API responses for transparency.
    AllowedPurposes []string `json:"allowed_purposes"`
}
```

---

## Summary of Key Files to Create

| File Path | Purpose |
|---|---|
| `/internal/vision/types.go` | All domain types: Camera, Detection, Clip, Zone, Edge, Frame, BoundingBox |
| `/internal/vision/service.go` | Service struct, constructor, camera CRUD, detection queries |
| `/internal/vision/pipeline.go` | `IngestBatch`, detection enrichment, debounce, cross-detection rules |
| `/internal/vision/detectors.go` | `Detector` interface, `RawDetection`, model version tracking |
| `/internal/vision/storage.go` | `ClipStore` interface, S3 implementation, presigning, tier transitions |
| `/internal/vision/events.go` | Event types, `RegisterHandlers`, NATS publishing functions |
| `/internal/vision/privacy.go` | `PrivacyConfig`, mask application, face blur coordination |
| `/internal/vision/config.go` | `GlobalConfig`, `CameraConfig`, `ZoneThreshold` |
| `/internal/vision/edge.go` | Edge agent registration, heartbeat, model distribution |
| `/internal/api/vision_handler.go` | HTTP handler for all `/api/v1/vision/` endpoints |
| `/migrations/021_vision.sql` | Complete database migration |

### What This Design Does NOT Include (Deliberately)

**No ML model training pipeline**: The system consumes ONNX models but does not train them. Model training happens offline and models are uploaded to S3. The edge agent pulls the latest model on heartbeat. This keeps the runtime system focused on inference and event processing.

**No facial recognition or individual identification**: The system counts people and detects objects (gloves, hairnets, boxes) but does not identify who anyone is. Buddy-punch detection works by correlating clock-in timestamps with camera observation of person count near the time clock, not by identifying faces. This is a deliberate privacy and legal liability constraint.

**No continuous video streaming through the cloud**: All live video stays on the local network or goes peer-to-peer via WebRTC. The cloud stores only event-triggered clips and key frame snapshots. This reduces bandwidth costs by orders of magnitude and keeps latency acceptable.


# Part 2: AI/ML Computer Vision Pipeline

## 1. Model Architecture

### 1.1 Person Detection and Multi-Object Tracking (MOT)

**Model Choice: YOLOv8m + BoT-SORT**

| Parameter | Value |
|---|---|
| Detection model | YOLOv8m (medium) -- best accuracy/speed tradeoff for person class |
| Tracker | BoT-SORT (combines motion + appearance ReID) |
| Input resolution | 640x640 (native YOLO input tile) |
| Camera feed resolution | 1080p decoded, downscaled to 640x640 for inference |
| Target FPS | 15 fps per camera (sufficient for tracking, saves 50% compute vs 30fps) |
| GPU memory | ~220 MB (FP16 TensorRT engine) |
| Latency per frame | ~12 ms on Jetson Orin NX, ~6 ms on Orin AGX |

**Why YOLOv8m over alternatives:**
- RT-DETR is stronger on small objects but 2x heavier; person detection at kitchen distances (2-8m) does not need it.
- YOLOv8n is too inaccurate for partial occlusions behind counters/equipment.
- YOLOv8m at INT8 gives mAP 50.2 on COCO person class -- sufficient.

**BoT-SORT over ByteTrack:**
- ByteTrack is lighter but loses tracks when persons pass behind equipment. BoT-SORT's appearance ReID feature recovers identity after occlusion, critical in cluttered kitchens.
- ReID backbone: OSNet-x0.25 (~8 MB, runs only on new/lost tracks, not every frame).

### 1.2 Pose Estimation (Compliance Monitoring)

**Model Choice: YOLOv8m-pose (17 keypoints) for general pose + MediaPipe Hands for hand-specific tasks**

| Parameter | Value |
|---|---|
| Body pose model | YOLOv8m-pose (top-down, 17 COCO keypoints) |
| Hand detail model | MediaPipe Hands (21 landmarks per hand) -- triggered only in handwash zones |
| Input resolution | 256x192 crop per detected person (body), 224x224 crop per hand ROI |
| Target FPS | 10 fps (pose does not need frame-level granularity) |
| GPU memory | ~180 MB (pose) + ~45 MB (hands, loaded only for sink-zone cameras) |

**Compliance detection logic:**

```
HANDWASHING DETECTION:
  Trigger:    Person bbox overlaps "sink zone" polygon for > 2 seconds
  Pose check: Both wrist keypoints below elbow keypoints AND within sink zone polygon
  Hand model: MediaPipe confirms rubbing motion (relative landmark velocity > threshold)
  Duration:   Timer starts on first qualifying frame, event fires at 20 seconds
  Confidence: Require 80%+ of frames in window to pass all checks

GLOVE DETECTION:
  Model:      Fine-tuned YOLOv8s-cls on hand crops (binary: gloved/ungloved)
  Input:      Hand ROI crops from pose wrist keypoints, padded 40%
  Dataset:    ~5,000 labeled hand crops
  Accuracy:   Target 94%+ recall (prefer false positives over misses)
```

### 1.3 Object Detection (Domain-Specific)

**Model Choice: YOLOv8s fine-tuned on restaurant domain classes**

| Class group | Specific classes | Training data source |
|---|---|---|
| Food safety | uncovered food container, temperature probe, cutting board | Synthetic + manual |
| Waste/cleaning | waste bin (open/closed/overflowing), mop, spray bottle | Open Images + fine-tune |
| Delivery | delivery box, delivery bag, receipt printer output | Custom collection |
| Hazard | fire, smoke, liquid spill, broken glass | Synthetic generation |
| Equipment | oven door (open/closed), fridge door (open/closed), fryer basket | Custom collection |

| Parameter | Value |
|---|---|
| Model | YOLOv8s fine-tuned (~11M params) |
| Input resolution | 640x640 |
| Target FPS | 10 fps |
| GPU memory | ~120 MB (FP16 TensorRT) |
| Classes | ~25 restaurant-domain classes |

**Fire/smoke gets special treatment:** Secondary model (EfficientNet-B0) runs on every frame at 320x320 as a parallel safety net. Adds only ~15 MB and ~3 ms per frame. If either detector fires, event escalates immediately (no smoothing window).

### 1.4 Action Recognition

**Model Choice: SlowFast (R50 backbone) fine-tuned on kitchen activities**

| Parameter | Value |
|---|---|
| Model | SlowFast R50 (slow pathway 4 fps, fast pathway 32 fps) |
| Input | 8-frame slow clip + 32-frame fast clip per person crop |
| Clip duration | 2 seconds of buffered frames |
| GPU memory | ~350 MB (FP16) |
| Classes | prep_chopping, prep_mixing, cleaning_surface, cleaning_floor, carrying_item, slip_fall, standing_idle, phone_use |

**Slip/fall detection (safety-critical):**
- Primary: SlowFast classifies "slip_fall" action.
- Secondary: Rule-based detector using pose keypoints -- if center-of-mass velocity exceeds threshold AND bounding box aspect ratio inverts (width > height), flag as potential fall.
- Either trigger fires an immediate alert.

### 1.5 Occupancy Counting

Reuses person detector from 1.1, applying zone polygons:

```
ZONE DEFINITIONS (per camera, configured in admin UI):
  +--------------------------------------------------+
  | Camera FOV                                        |
  |   +----------+  +----------+  +--------------+   |
  |   | PREP     |  | GRILL    |  | DISH PIT     |   |
  |   | STATION  |  | STATION  |  |              |   |
  |   | cap: 2   |  | cap: 1   |  | cap: 2      |   |
  |   +----------+  +----------+  +--------------+   |
  |   +--------------------------------------------+  |
  |   | WALKWAY (throughput tracking only)          |  |
  |   +--------------------------------------------+  |
  +--------------------------------------------------+
```

No additional GPU cost -- piggybacks on person detector.

### 1.6 Anomaly Detection

**Model Choice: Conv-LSTM autoencoder for scene-level + rule engine for behavioral anomaly**

- Scene anomaly: Conv-LSTM autoencoder trained on "normal hours" footage per camera. High reconstruction error = anomaly.
- Behavioral rules: After-hours person detection, hand in register area without POS transaction event.
- ~80 MB (FP16), runs at 2 fps.

---

## 2. Edge Deployment Architecture

### 2.1 Hardware: NVIDIA Jetson Orin NX 16GB

| Tier | Cameras | Hardware | Cost |
|---|---|---|---|
| Small restaurant (1-4 cameras) | 1x Orin NX 16GB | ~$1,200 |
| Medium restaurant (5-10 cameras) | 2x Orin NX 16GB | ~$2,200 |
| Large / multi-kitchen (11-20) | 1x AGX Orin 64GB + 1x Orin NX | ~$3,500 |

### 2.2 Model Optimization

| Model | Precision | Reason |
|---|---|---|
| YOLOv8m (person det) | INT8 | Person detection tolerates quantization well |
| YOLOv8m-pose | FP16 | Keypoint regression sensitive to quantization |
| YOLOv8s (objects) | INT8 | Classification is quantization-friendly |
| SlowFast R50 | FP16 | Temporal modeling degrades at INT8 |
| Fire classifier | INT8 | Binary classification -- robust |
| Conv-LSTM anomaly | FP16 | Reconstruction quality matters |

### 2.3 Graceful Degradation

```
TIER 1 (GPU < 70%):   All models at target FPS
TIER 2 (GPU 70-85%):  Action recognition drops to top-3 persons, anomaly to 1 fps
TIER 3 (GPU 85-95%):  Object detection drops to 5 fps, action recognition paused
TIER 4 (GPU > 95%):   Only person detection + fire/smoke classifier run
```

Priority queue: fire/smoke > person_detection > pose > action > anomaly

### 2.4 Edge-to-Cloud Sync

```
Events:      gRPC stream (persistent, reconnect with backoff) -- ~1 MB/hour
Clips:       HTTPS PUT to pre-signed S3 URLs -- ~100 MB/hour
Health:      HTTPS POST every 60s
Model pulls: HTTPS GET, poll every 15 min -- ~50 MB/month
TOTAL:       ~100-150 MB/hour per restaurant
```

---

## 3. Training Pipeline

### 3.1 Strategy

- **Phase 1 (Bootstrap):** Partner with 3-5 pilot restaurants, record 8hr/day for 2 weeks, contract annotation (~$20K for 20K frames).
- **Phase 2 (Active Learning):** Deploy initial models, log uncertain frames (confidence 0.3-0.7), weekly annotation batches.
- **Phase 3 (Production Flywheel):** Manager confirmations/rejections feed back as labels. 100 restaurants = ~5,000 new labeled samples/week.

### 3.2 Synthetic Data for Rare Events

| Event | Approach | Volume |
|---|---|---|
| Slip/fall | Unity ragdoll + StyleGAN transfer | 2,000 clips |
| Fire/smoke | Blender fluid simulation + GAN | 5,000 images |
| Broken glass | Diffusion-based generation | 1,000 images |
| Liquid spill | Diffusion inpainting on real floors | 2,000 images |

Target ratio: 70% real + 30% synthetic.

### 3.3 Model Versioning and A/B Deployment

Canary rollout: 10% of edge devices get new model for 7 days. Compare event volume, human-in-loop rejection rate, latency. Auto-promote if equal or better; auto-rollback if regression.

---

## 4. Inference Pipeline

```
PER-CAMERA PIPELINE:

RTSP Decode -> Frame Sample (15fps) -> Preprocess -> STAGE 1: Detection (YOLOv8m, 12ms)
                                                          |
                    +-------------------------------------+---------------------------+
                    |                                     |                           |
            STAGE 2a: Track                    STAGE 2b: Pose              STAGE 2c: Zone Count
            (BoT-SORT)                         (compliance zones only)     (centroid-in-polygon)
                    |                                     |
            STAGE 3: Action                    Compliance Rules Engine
            (SlowFast, 2s clips)               (handwash, gloves)
                    |                                     |
                    +-----------> EVENT AGGREGATOR <-------+
                                  (temporal smoothing, dedup, confidence)
                                        |
                                  EVENT DISPATCHER
                                  (priority queue, rate limiting, clip extraction)

PARALLEL: Fire/smoke classifier runs independently on every frame (3ms, bypasses all smoothing)
```

### Latency Budget Per Frame (amortized)

| Component | Time | Notes |
|---|---|---|
| RTSP decode (NVDEC) | 2 ms | Hardware decoder |
| Resize + normalize | 1 ms | CUDA kernel |
| YOLOv8m inference | 12 ms | TensorRT INT8 |
| Tracking (BoT-SORT) | 3 ms | CPU |
| Pose estimation | ~3 ms | Amortized (only compliance zones) |
| Action recognition | ~2 ms | Amortized (every 2s per person) |
| Fire classifier | 3 ms | Every frame, parallel |
| Post-processing | 2 ms | CPU |
| **Total amortized** | **~26 ms** | Well within 66ms budget (15fps) |

---

## 5. Accuracy and Confidence

### Thresholds Per Event Type

| Event Type | Detection | Event Fire | Rationale |
|---|---|---|---|
| Fire/smoke | 0.30 | 0.30 | Never miss a fire |
| Slip/fall | 0.50 | 0.60 | Serious but not life-threatening |
| Compliance | 0.60 | 0.80 | Higher to prevent alert fatigue |
| Occupancy | 0.45 | N/A | Errors average out |

### False Positive Management

- **Level 1 (Model):** NMS, confidence thresholds, temporal smoothing
- **Level 2 (Context):** Zone-aware filtering, schedule-aware, cross-reference
- **Level 3 (Human-in-loop):** Manager review for mid-confidence events; CONFIRM/DISMISS/NOT SURE
- **Alert fatigue prevention:** Max 20 alerts/camera/hour, auto-raise thresholds if dismiss rate > 40%

---

## 6. Privacy Pipeline

### Face Anonymization (runs BEFORE any other processing)

```
RTSP Frame -> Face Detector (YOLOv8n-face, 2ms) -> Gaussian Blur (sigma>=20, bbox+20%) -> Anonymized frame to pipeline
```

- Face blurring in RTSP decode thread, BEFORE frame enters any buffer
- No unblurred frame ever written to disk or transmitted
- Nightly verification job: run face detection on stored clips, quarantine if any face detected

### Zone-Level Privacy

```
FULL_MASK:           Black rectangle, zero processing (restrooms, break rooms)
FACE_BLUR_NO_TRACK:  Blur + counting only, no persistent IDs (dining areas)
FACE_BLUR_TRACKED:   Blur + full tracking + pose + action (kitchen, staff consent required)
FACE_BLUR_MINIMAL:   Blur + person detection + fire/safety only
```

### Biometric Data Prohibition

- NO facial recognition / face embedding storage -- ever
- NO gait recognition or biometric identification
- Track IDs are integer counters, reset when person exits frame
- ReID features computed transiently in GPU memory, never written to disk

---

## 7. Performance Budgets

### Per Camera

~39.2% GPU utilization per camera on Orin NX. With dynamic batching, 5-6 cameras fit on a single device at ~82% utilization.

### Network

| Direction | Bandwidth |
|---|---|
| Camera -> Edge (local) | 3 Mbps per camera |
| Edge -> Cloud (WAN) | 0.25 Mbps per camera |
| Total for 4 cameras | 12 Mbps local, 1 Mbps WAN |

### Cost Per Restaurant

| Component | One-Time | Monthly |
|---|---|---|
| Edge hardware | $1,200 | - |
| IP cameras (x4) | $800 | - |
| PoE switch + cabling | $200 | - |
| Cloud compute | - | $50-80 |
| Cloud storage | - | $10-20 |
| Edge power | - | ~$5 |
| **Total first year** | **$2,350** | **$780/yr** |



---

# Part 3: Security & Privacy Architecture

## 1. Threat Model

### Trust Boundaries

| Boundary | From | To | Controls |
|---|---|---|---|
| TB-1 | Camera | Edge Appliance | RTSPS, VLAN isolation, mutual auth |
| TB-2 | Edge Appliance | Cloud Ingest | mTLS (SPIFFE identity), WireGuard |
| TB-3 | Cloud Ingest | S3 Storage | IAM + encryption, tenant-scoped keys |
| TB-4 | Cloud Ingest | NATS JetStream | mTLS, ACLs per tenant subject |
| TB-5 | API Gateway | Dashboard User | OIDC/SAML + RBAC + audit |
| TB-6 | Edge Appliance | Local Network | VLAN isolation, no lateral movement |

### STRIDE Analysis

| Threat | Component | Severity | Mitigation |
|---|---|---|---|
| Spoofing | Camera feed injection | Critical | Mutual auth, stream signing, tamper detection |
| Tampering | Video clips in S3 | High | WORM-mode storage, integrity checksums |
| Repudiation | Video access without trail | High | Mandatory audit logging, no bypass path |
| Info Disclosure | Raw faces stored/transmitted | Critical | Edge-side anonymization before storage/transmission |
| Info Disclosure | Video of payment terminals | Critical | PCI privacy zones, automatic card number detection |
| DoS | Camera flood / stream overload | Medium | Rate limiting at edge, circuit breakers |
| Elevation | Staff accessing live feeds | High | RBAC with CCTV-specific roles, MFA enforcement |
| Elevation | Tenant A accessing Tenant B video | Critical | PostgreSQL RLS, S3 bucket policies, NATS ACLs |

---

## 2. Privacy by Design

### Face Anonymization

**Core Principle:** Faces are anonymized at the edge BEFORE any frame leaves the on-premise appliance. The cloud never receives identifiable facial data.

Pipeline: Privacy Zone Mask -> Face Detection (SCRFD/RetinaFace) -> Gaussian Blur (sigma>=20, bbox+20%) -> CV Inference on anonymized frame.

**Critical Controls:**
- Face detection runs BEFORE any other CV model
- Raw frames exist only in edge RAM, never written to disk or transmitted
- Anonymization is a mandatory pipeline stage that cannot be bypassed via configuration
- Nightly verification job samples stored clips and quarantines any detected faces

### Biometric Data Prohibition

| Data Type | Status | Enforcement |
|---|---|---|
| Facial geometry / faceprints | PROHIBITED | No embedding extraction models deployed |
| Facial recognition templates | PROHIBITED | No FR models in approved model registry |
| Gait analysis biometrics | PROHIBITED | Pose outputs skeleton keypoints only |
| Re-identification across cameras | PROHIBITED | No cross-camera tracking |
| Voice biometrics | PROHIBITED | Audio channels stripped at edge |

**Technical enforcement:** Edge model registry only permits signed models. Quarterly model audit verifies output types.

### Employee Consent Workflow

1. Location owner enables CCTV -> system generates jurisdiction-specific notices
2. Physical signage requirements displayed (printable PDF generated)
3. Owner acknowledges signage placement
4. Each employee receives in-app notification explaining monitoring scope
5. Consent collected where required (BIPA, CCPA, GDPR)
6. Cameras activate ONLY after all prerequisites complete

---

## 3. Data Classification & Retention

| Tier | Data Type | Classification | Retention |
|---|---|---|---|
| T1 | Raw video frames | RESTRICTED - EPHEMERAL | Edge RAM only, never persisted |
| T2 | Anonymized video clips | CONFIDENTIAL | 14 days default (configurable 7-30) |
| T3 | Detection events | CONFIDENTIAL | 90 days default (configurable 30-180) |
| T4 | Aggregated metrics | INTERNAL | Indefinite (no PII) |
| T5 | Model inference metadata | INTERNAL | 365 days |
| T6 | Audit logs | COMPLIANCE | 7 years, immutable |

### Right to Erasure

Since faces are anonymized at edge, most stored clips don't contain identifiable data. If an individual IS identifiable by context:
- Option A: Delete entire clip
- Option B: Further redact the individual
- Option C: Flag for deletion when legal hold lifts
- Response within 30 days (GDPR) / 45 days (CCPA)

---

## 4. Access Control

### CCTV-Specific RBAC

| Role | Permissions | Constraints |
|---|---|---|
| cctv_viewer | Live feeds, recent clips (24h), events | MFA required, location-scoped, 30min timeout |
| cctv_reviewer | All clips, export (watermarked), annotate events | MFA, exports logged, reason required |
| cctv_admin | Camera/zone/rule management, retention config | MFA, changes require confirmation |
| cctv_compliance | Mandatory privacy zones, legal holds, audit logs | FireLine internal only |

**No role grants access to raw (un-anonymized) video. This permission does not exist in the system.**

### Video Access Controls

- All video access logged: user_id, camera_id, clip_id, access_type, timestamp, source_ip
- Live: WebRTC with DTLS-SRTP, session-bound token
- Clips: Pre-signed S3 URL (5 min TTL, single-use)
- All delivered video has invisible watermark with viewer identity
- Employee anomaly subjects cannot view attached clips (request through HR)

---

## 5. Regulatory Compliance

### Compliance Matrix

| Requirement | BIPA (IL) | CCPA/CPRA (CA) | GDPR (EU) | PCI DSS v4 |
|---|---|---|---|---|
| No biometric storage | FireLine stores NONE | FireLine stores NONE | FireLine stores NONE | N/A |
| Consent | Notice required | Notice at collection | Legitimate interest + DPIA | N/A |
| Right to erasure | N/A | Supported | Supported | N/A |
| Cameras near payment | N/A | N/A | N/A | MUST NOT capture card numbers |
| Employee notice | Recommended | Required | Required | N/A |
| Breach notification | N/A | 72 hours (AG) | 72 hours (DPA) | Immediate to acquirer |

### PCI DSS Specific Controls

- Mandatory privacy zones over all POS terminal screens and card readers
- Secondary CV model detects card-like rectangles near POS locations
- Audio stripped at edge (customers reading card numbers aloud)

---

## 6. Network Security

### Architecture

```
Restaurant LAN:
  VLAN 10 (POS/Payment) -- VLAN 20 (CCTV/Camera) -- VLAN 30 (Guest WiFi)
  NO ROUTING BETWEEN VLANs
  
  Edge device bridges VLAN 20 (cameras) to WireGuard tunnel (cloud)
  Cameras have NO internet access (DNS blocked, default route null)
  Cameras cannot communicate with each other (port isolation)
```

### Edge Device Hardening

- Read-only root filesystem with dm-verity
- Secure Boot + measured boot via TPM 2.0
- LUKS2 encrypted data partition, key sealed to TPM PCR values
- Outbound: WireGuard tunnel to cloud ONLY
- Inbound: RTSP from camera VLAN ONLY
- No SSH by default (requires physical button + Vault-issued certificate)
- Physical tamper switch triggers key zeroization
- Signed heartbeat to cloud every 60 seconds

### Camera Credential Management

- Unique 32+ char password per camera, stored in HashiCorp Vault
- Default manufacturer credentials changed during provisioning
- Rotated every 90 days by automated job
- Camera admin interface accessible only from edge device IP
- UPnP, ONVIF discovery, Telnet/SSH disabled on cameras

---

## 7. Audit & Accountability

### Audit Log Integrity

- Append-only PostgreSQL table with row-level immutability trigger
- Real-time replication to S3 with Object Lock (WORM mode)
- Hash chaining (SHA-256 of previous event) for tamper evidence
- Hourly automated chain integrity verification
- INSERT-only DB permission for writes, SELECT-only for reads
- No role can delete or modify audit logs

### Automated Compliance Reports

- **Daily:** Camera health, anonymization verification, heartbeat status
- **Weekly:** Video access by role/location, privacy zone changes, erasure request status
- **Monthly:** Retention compliance, credential rotation, firmware updates, consent coverage
- **Annual:** Full compliance audit, DPIA reviews, pen test results

---

## 8. Incident Response

### Camera Compromise

1. Edge auto-stops ingesting from compromised camera
2. Alert to location owner + security team
3. Review logs for malicious frames; quarantine affected clips
4. Camera factory reset with new credentials; replace if firmware compromised

### Edge Device Theft

1. Cloud immediately revokes: SPIFFE identity, WireGuard key, Vault tokens, NATS credentials
2. Encrypted disk (LUKS2 + TPM) means data is unreadable without the device's TPM
3. Generate new credentials for all connected cameras
4. Deploy replacement device, re-provision from fleet GitRepo (5 min)

### Video Data Breach

1. Assess scope (clips are face-anonymized, reducing severity)
2. Revoke compromised credentials, force password reset
3. Notify per regulations: GDPR 72h, CCPA "without unreasonable delay", PCI immediate
4. Root cause analysis, remediation plan, executive summary

---

## 9. Key Architectural Decisions

| Decision | Rationale |
|---|---|
| Anonymize at edge, never in cloud | Raw faces never leave premises; reduces cloud breach blast radius |
| No facial recognition at all | BIPA private right of action ($1K-$5K/violation); no legitimate business need |
| Audio stripped at edge | Wiretapping laws vary; audio adds legal complexity with no operational value |
| 14-day default clip retention | Balances investigation needs with privacy minimization |
| Hardware TPM on edge devices | Stolen disk is unreadable; key bound to hardware |
| Hash-chained audit logs | Tamper-evident without blockchain complexity |



---

# Part 4: Frontend UX Design

## 1. New Pages

### Camera Management (`/cameras`)

```
+---------------------------------------------------------------+
| Header: "Camera Management"            [+ Add Camera] button  |
+---------------------------------------------------------------+
| Filter bar: [All Zones v] [Status: All v]  [Search...]        |
+---------------------------------------------------------------+
| Camera Grid (1col mobile, 2col md, 3col lg)                   |
| +------------------+  +------------------+  +------------------+
| | [Thumbnail]      |  | [Thumbnail]      |  | [Thumbnail]      |
| | Kitchen Cam 1    |  | Dining NW        |  | Entrance         |
| | Zone: Grill      |  | Zone: Dining     |  | Zone: Entrance   |
| | Status: Online   |  | Status: Online   |  | Status: Offline  |
| | AI: Occ, Station |  | AI: Occ, Queue   |  | AI: Queue        |
| | [Configure] [View]  | [Configure] [View]  | [Configure]      |
| +------------------+  +------------------+  +------------------+
```

**Configure** opens side panel with: camera name, zone assignment, stream URL, AI feature toggles, **Privacy Zone Editor** (drag-to-draw rectangles on camera preview), and delete button.

### Live Operations View (`/live`)

```
+-------------------------------------------+  +-----------+
|          Live Camera Feed                 |  | Station    |
|    (HLS player with AI overlay canvas)    |  | Board      |
|    [bounding boxes for people]            |  | Grill 1 [G]|
|    [zone highlights semi-transparent]     |  | Prep A  [Y]|
|    [occupancy count overlay top-right]    |  | Fry     [G]|
+-------------------------------------------+  | Expo    [R]|
| Occupancy Gauges        | Queue Length: 6 |  | Dish    [G]|
| [Dining: 42/80] [Kitchen: 8/12]  Wait ~4m|  +-----------+
+-------------------------------------------+
| Recent Compliance Events (scrolling, last 10)              |
| 2:34pm Handwash OK - Prep | 2:31pm Glove Miss - Grill    |
+------------------------------------------------------------+
```

Video: HLS via `hls.js` (low-latency mode, 2-4s latency). AI overlay on `<canvas>` at 5fps. ONE live stream at a time; camera switcher swaps source.

### Compliance Dashboard (`/compliance`)

```
+---------------------------------------------------------------+
| [Overall: 87%] [Handwash: 92%] [Gloves: 84%] [Hair: 91%]    |
+---------------------------------------------------------------+
| Hourly Compliance Chart (Recharts)    | Violation Breakdown   |
| - Handwash line (blue)                | (pie/donut chart)     |
| - Gloves line (amber)                 |                       |
| - Hair nets line (green)              |                       |
+---------------------------------------------------------------+
| Violations Timeline (scrollable list)                         |
| [!] 2:31pm Glove violation - Grill   [View Clip] [Dismiss]  |
| [*] 1:45pm Handwash miss - Prep      [View Clip] [Dismiss]  |
+---------------------------------------------------------------+
```

### Video Evidence Viewer (`/evidence/:clipId`)

```
+---------------------------------------------------------------+
|              Video Player (HLS + AI overlay)                  |
+---------------------------------------------------------------+
| Timeline scrubber with AI annotation markers                  |
+---------------------------------------------------------------+
| Clip Metadata              | Linked Context                   |
| Camera: Kitchen Cam 1      | Anomaly: Cash variance...        |
| Time: Apr 6, 2:31pm        | Compliance: Glove miss           |
| Duration: 45s               | [Go to anomaly detail]           |
+---------------------------------------------------------------+
| AI Annotations List (click to seek video to timestamp)        |
| 0:03 - Person entering zone                                  |
| 0:12 - Glove not detected on left hand                       |
+---------------------------------------------------------------+
```

---

## 2. Integration into Existing Pages

### Dashboard -- New Widgets

- **Occupancy Widget:** Real-time dining/kitchen occupancy with mini progress bars (WebSocket-driven)
- **Compliance Score Card:** Today's score (0-100) with violation count, links to `/compliance`

### Labor Page -- Station Occupancy Overlay

Background heatmap lane on existing shift timeline showing station utilization (green/amber/red) from `useOccupancyHistory()`.

### Operations Page -- Bottleneck Indicators

New "Kitchen Bottlenecks" section with per-station cards showing: severity, cycle time vs target, backed-up order count. Border color by severity.

### Customer Page -- Dine-in Occupancy Trends

Recharts area chart: occupancy over time (hourly, 7 days). Overlaid line for estimated wait time.

### Intelligence Page -- Video Evidence

Anomaly detail modal gains "VIDEO EVIDENCE" section with thumbnail, [Play Clip], [Open Full Viewer]. Inline `VideoClipPlayer` in modal.

### Portfolio Page -- Cross-Location Visual Comparison

New "Visual Operations" section: horizontal bar chart comparing kitchen throughput/utilization per location. Click to navigate to that location's Live View.

---

## 3. Real-time Components

### WebSocket Architecture

```
Browser -- wss://api.fireline.io/ws/cctv?location=<id>&token=<jwt>
  |
  +--> occupancy_update   --> cctvStore.occupancy
  +--> station_update     --> cctvStore.stations
  +--> compliance_event   --> cctvStore.complianceEvents (cap at 50)
  +--> cctv_alert         --> cctvStore.alerts + toast notification
  +--> queue_update       --> cctvStore.queueMetrics
  +--> camera_status      --> cctvStore.cameraStatuses
```

### Key Components

| Component | Purpose |
|---|---|
| `CameraFeed` | HLS player + AI overlay canvas (5fps redraws) |
| `OccupancyGauge` | Circular SVG gauge with animated transitions |
| `StationBoard` | Vertical status pills (green/yellow/red/gray) |
| `ComplianceTimeline` | Scrollable violation event list |
| `VideoClipPlayer` | HLS player + annotation overlay + scrubber |
| `ZoneEditor` | Privacy zone draw/edit tool on camera preview |
| `CCTVAlertToaster` | Toast notifications (critical: stay, warning: auto-dismiss 15s) |

---

## 4. Privacy UX

### Privacy Zone Editor

Canvas overlay on camera snapshot. Click-drag to draw rectangles. Corner handles for resize. Label input per zone. Delete button. All coordinates stored as 0-1 normalized values. Touch support for tablet.

### Audit Indicator

Persistent header badge when viewing live feeds: `[Eye] Viewing audited | Session: 4m 32s`

### Staff Consent Management

DataTable showing: Name, Status (Consented/Pending/Declined), Date. [Send Reminder] [Export Report] actions.

---

## 5. Mobile/Tablet

### Staff Tablet (React Native)

- **NO camera feeds** for staff
- Compliance reminders as push notification cards: "Handwash Reminder -- 45 min since last recorded handwash at Prep. [Mark Complete] [Snooze 10m]"
- Simplified Station Status Board for shift leads (colored pills, no video)

### Manager Mobile (Web, responsive)

- Live View: single camera full-width, station board as horizontal scroll below
- Compliance: KPIs stack 2x2, chart full width
- Camera Management: single column grid, config as full-screen modal

---

## 6. Performance Strategy

1. **Single active HLS stream** at a time. Camera switching destroys previous `hls.js` instance.
2. **Thumbnail grid** for overview (static JPEGs refreshed every 10s).
3. **Canvas overlay at 5fps** via `requestAnimationFrame` with timestamp check.
4. **WebSocket updates at 1s intervals**, Zustand store updates debounced.
5. **Code splitting**: all CCTV pages lazy-loaded. `hls.js` (~200KB) loaded only when needed.
6. **HLS buffer capped at 10s**. Cleanup on unmount releases media source buffers.

---

## 7. Component Architecture

### New File Structure

```
web/src/
  stores/cctv.ts                    # Zustand store for real-time CCTV data
  hooks/useCCTV.ts                  # React Query hooks for REST endpoints
  hooks/useCCTVWebSocket.ts         # WebSocket connection manager
  pages/
    CameraManagementPage.tsx
    LiveOperationsPage.tsx
    ComplianceDashboardPage.tsx
    VideoEvidencePage.tsx
  components/cctv/
    CameraFeed.tsx                  OccupancyGauge.tsx
    StationBoard.tsx                ComplianceTimeline.tsx
    ComplianceChart.tsx             HeatmapOverlay.tsx
    VideoClipPlayer.tsx             VideoAnnotationList.tsx
    ZoneEditor.tsx                  CCTVAlertToaster.tsx
    CameraGrid.tsx                  CameraConfigPanel.tsx
    BottleneckCard.tsx              QueueIndicator.tsx
    ConsentTable.tsx                AuditIndicator.tsx
    OccupancyWidget.tsx             ComplianceScoreCard.tsx
    PortfolioCCTVRow.tsx

tablet/components/
    ComplianceReminder.tsx
    StationStatusList.tsx
```

### Routes (added to App.tsx)

```typescript
<Route path="cameras" element={<CameraManagementPage />} />
<Route path="live" element={<LiveOperationsPage />} />
<Route path="compliance" element={<ComplianceDashboardPage />} />
<Route path="evidence/:clipId" element={<VideoEvidencePage />} />
```

### Nav Items (after Intelligence, before Alerts)

```typescript
{ to: '/cameras', label: 'Cameras', icon: Video },
{ to: '/live', label: 'Live View', icon: MonitorPlay },
{ to: '/compliance', label: 'Compliance', icon: ShieldCheck },
```



---

# Part 5: Infrastructure & Reliability

## 1. Edge Computing

### Hardware: NVIDIA Jetson Orin NX 16GB

Per-location BOM:
- 1x Jetson Orin NX 16GB + Seeed reComputer J4012: $599
- 1x 256GB NVMe SSD: $35
- 1x Managed PoE switch (Ubiquiti USW-Lite-8-PoE): $109
- 2-4x IP cameras (Hikvision 4MP RTSP): $120 each
- 1x UPS (APC BX600): $65
- 1x Enclosure + cables: $50
- **Total (2 cameras): ~$1,098 | (4 cameras): ~$1,338**

### OS: Ubuntu 22.04 + JetPack 6.x + k3s

k3s gives Kubernetes-compatible workload management at ~512MB RAM overhead. Enables GitOps via Fleet.

### Fleet Management: Rancher Fleet

Hub-and-spoke model for managing 100+ edge k3s clusters:

```
Cloud (Rancher Fleet Manager)
  +-- Fleet GitRepo (github.com/fireline/edge-fleet)
  |     +-- base/         # common k3s manifests
  |     +-- overlays/     # per-location camera configs
  |     +-- clusters/     # auto-registered
  +-- System Upgrade Controller (k3s, OS, models)
```

**Provisioning:** Boot pre-flashed image -> cloud-init -> joins Tailscale -> registers with Fleet -> applies manifests -> camera discovery -> inference starts in ~5 minutes.

### Offline Resilience

| Internet Status | Behavior |
|---|---|
| Online | Events to cloud NATS, clips to S3 |
| Degraded (>50% PL) | Events buffer in local NATS JetStream (NVMe) |
| Offline | Full local operation, 48h buffer for events + clips |
| Reconnect | Auto-drain: events replayed, clips uploaded (most recent first) |

Local NATS JetStream: 256MB mem, 20GB file store (~48h of events at peak).

---

## 2. Cloud Architecture

### Services (ECS Fargate)

| Service | Purpose | Resources | Scaling |
|---|---|---|---|
| cctv-event-aggregator | Consume events, write PostgreSQL, trigger alerts | 512 CPU / 1024 MB | 2-10 tasks |
| cctv-clip-manager | Presigned URLs, metadata indexing, S3 lifecycle | 256 CPU / 512 MB | 2-4 tasks |
| cctv-analytics | Batch analytics: waste trends, accuracy tracking | 1024 CPU / 2048 MB | 1-3 tasks |
| cctv-fleet-proxy | WebSocket proxy, camera config push, health | 512 CPU / 1024 MB | 2-4 tasks |

**No cloud-side GPU inference.** All inference at edge: $599 one-time vs $1+/hr cloud GPU.

### Autoscaling by Fleet Size

| Locations | Events/day | Clips/day | Aggregator Tasks | Clip Manager |
|---|---|---|---|---|
| 10 | ~5,000 | ~500 | 2 | 2 |
| 50 | ~25,000 | ~2,500 | 3 | 2 |
| 200 | ~100,000 | ~10,000 | 5 | 3 |
| 1,000 | ~500,000 | ~50,000 | 10 | 4 |

---

## 3. Video Pipeline

### S3 Storage Lifecycle

| Tier | Days | Storage Class | Access |
|---|---|---|---|
| Hot | 0-7 | Standard | Fast dashboard playback |
| Warm | 7-30 | Standard-IA | Millisecond retrieval |
| Cold | 30-90 | Glacier IR | Minutes retrieval |
| Archive | 90-365 | Deep Archive | Hours retrieval |
| Delete | 365+ | Deleted | - |

**Bucket key:** `clips/{org_id}/{location_id}/{camera_id}/{YYYY-MM-DD}/{event_id}.mp4`

### CDN: CloudFront

Signed URLs (5 min TTL) for clip playback. Browser streams from edge cache, S3 origin fetch on miss.

---

## 4. Networking

### Tailscale (site-to-cloud)

Handles NAT traversal automatically, zero network config at site, ACLs for device-level access. $6/device/month.

```json
ACL Policy:
  edge-device -> cloud-services: ports 4222 (NATS), 443 (HTTPS)
  fleet-admin -> edge-device: ports 6443 (k8s), 22 (SSH)
  cctv-proxy -> edge-device: port 8554 (live view)
```

### Camera VLAN Isolation

```
VLAN 10 (Camera): 192.168.10.0/24 -- cameras + edge device eth0
VLAN 1 (Restaurant LAN): 192.168.1.0/24 -- edge device eth1, POS, guest WiFi
NO routing between VLANs. Edge device bridges the two.
```

### mTLS via SPIFFE/SPIRE

Edge workloads get SPIFFE identities: `spiffe://fireline.io/edge/{location_id}/inference`. NATS connections use mTLS with SVIDs. Server extracts location_id for authorization.

---

## 5. Observability

### Edge Metrics (pushed via Prometheus remote_write)

- **Hardware:** GPU temp, GPU util%, CPU util%, memory, disk, uptime
- **Inference:** FPS per camera, latency p50/p95/p99, model version, detection counts
- **Stream:** RTSP status, frame drops, reconnects, decode FPS
- **Upload:** Clips pending, upload latency, failures, NATS connection status

### Cloud Metrics

- Events received/processed, processing latency, E2E latency
- Clips uploaded, upload latency, storage bytes/cost
- Active locations/cameras, fleet devices online/total

### Alerting Rules

| Alert | Condition | Severity |
|---|---|---|
| CameraOffline | Stream down > 5min | warning |
| EdgeDeviceUnresponsive | No metrics > 3min | critical |
| InferenceFPSDegraded | FPS < 10 for 2min | warning |
| GPUTemperatureHigh | GPU > 85C for 1min | critical |
| EdgeDiskPressure | Disk > 85% for 5min | warning |
| EventProcessingLag | NATS pending > 5000 for 5min | warning |
| EventLatencySLOBreach | <95% within 30s for 5min | critical |
| ClipUploadFailureRate | >5% failures for 10min | warning |

### Grafana Dashboards

- **Fleet Overview:** Device map, camera uptime (30d), events/hour, FPS heatmap, storage growth
- **Per-Location:** GPU metrics, per-camera FPS, detection counts, upload queue, NATS status, disk usage

---

## 6. SLOs

| SLO | SLI | Target | Error Budget (30d) |
|---|---|---|---|
| Event E2E latency | % events delivered < 30s | 95% | 36 hours |
| Camera uptime | % time streaming | 99.5% | 3.6 hours/camera |
| Edge device availability | Inference running + metrics reporting | 99.9% | 43 min/device |
| Clip upload success | % clips uploaded within 1 hour | 99% | 7.2 hours |

### Failure Modes

| Failure | Detection | Recovery |
|---|---|---|
| Camera hardware failure | Stream status = 0 | Spare shipped next-day |
| Edge device crash | Metrics stop | Hardware watchdog reboots, k3s restarts pods |
| Network partition | Tailscale status = 0 | Edge operates autonomously, 48h buffer |
| NATS cloud down | Connection status = 0 across fleet | Edge JetStream buffers; cloud NATS is 3-node HA |
| Bad model deploy | Confidence < 0.3 sustained | Auto-rollback via canary metrics |
| Power outage | All metrics stop | UPS gives 10min; device boots clean on restore |

---

## 7. Cost Model

### Per-Location Monthly

| Component | One-Time | Monthly |
|---|---|---|
| Edge hardware (amortized 3yr) | - | $31 |
| IP cameras (amortized 3yr) | - | $7 |
| Tailscale | - | $6 |
| Edge power | - | $5 |
| S3 storage | - | $4 |
| CloudFront egress | - | $2 |
| Cloud compute (amortized) | - | $3 |
| NATS cloud (amortized) | - | $2 |
| Prometheus storage | - | $1 |
| **Total** | - | **~$61/mo** |

### By Fleet Size

| Locations | Edge HW (one-time) | Monthly | Per-Location |
|---|---|---|---|
| 10 | $11K | $810 | $81 |
| 50 | $55K | $3,350 | $67 |
| 200 | $220K | $12,200 | $61 |
| 1,000 | $1.1M | $57,500 | $58 |

**ROI:** Cost per waste event detected at 200 locations: ~$0.12/event. If each event drives $5 of waste reduction, the system pays for itself at 40:1.

### Optimization Levers

| Lever | Savings | Tradeoff |
|---|---|---|
| Resolution 4MP -> 1080p | -15% BW | Slightly less accuracy at distance |
| Inference FPS 15 -> 10 | -30% GPU | 100ms slower detection |
| Clip retention 365d -> 90d | -60% S3 | Lose historical evidence |
| Thumbnail instead of clip | -90% S3/BW | No video evidence |

---

## 8. Deployment & Updates

### Edge OS: A/B Partitions

```
nvme0n1p1: 512MB EFI boot (shared)
nvme0n1p2: 40GB Root A (active)
nvme0n1p3: 40GB Root B (standby)
nvme0n1p4: 8GB  Persistent config
nvme0n1p5: 167GB Data (clips, NATS, models)
```

Update flow: download new rootfs to standby -> verify checksum -> update GRUB -> reboot -> health check passes (mark active) or fails (revert, reboot).

### ML Model Canary Rollout

```
Phase 1: 5% of devices (5 locations)   -- 24h soak
Phase 2: 25% of devices (25 locations) -- 24h soak
Phase 3: 100% of devices               -- auto-promote if SLOs met
```

**Promotion criteria:** inference_fps >= 12, confidence_avg >= previous, false_positive_rate <= previous + 5%, gpu_memory <= previous + 10%.

**Rollback triggers:** FPS < 8 for 5min on 3+ devices, GPU temp > 90, crash loop (3 restarts in 10min).

### Zero-Downtime Camera Reconfiguration

1. Fleet GitRepo updated with new camera config
2. Fleet syncs ConfigMap to edge k3s
3. Camera manager detects change, gracefully stops affected pipeline
4. Flushes in-progress clip, publishes "camera.reconfiguring"
5. Starts new pipeline, waits for first successful decode
6. Publishes "camera.online"
7. Other cameras continue uninterrupted (~5s gap for reconfigured camera)
