-- ----------------------------
-- Enable TimescaleDB extension
-- ----------------------------
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

-- ----------------------------
-- Sequence for alert_types
-- ----------------------------
CREATE SEQUENCE IF NOT EXISTS alert_types_id_seq
  INCREMENT 1
  START 1
  MINVALUE 1
  MAXVALUE 2147483647
  CACHE 1;

-- ----------------------------
-- Table structure for alert_types
-- ----------------------------
CREATE TABLE IF NOT EXISTS "public"."alert_types" (
  "id" int4 NOT NULL DEFAULT nextval('alert_types_id_seq'::regclass),
  "name" varchar(50) COLLATE "pg_catalog"."default" NOT NULL,
  "icon" varchar(50) COLLATE "pg_catalog"."default",
  "color" varchar(20) COLLATE "pg_catalog"."default",
  "description" text COLLATE "pg_catalog"."default",
  "created_at" timestamptz(6) DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz(6) DEFAULT CURRENT_TIMESTAMP,
  "display_name" varchar(50) COLLATE "pg_catalog"."default",
  "severity" varchar(255) COLLATE "pg_catalog"."default",
  PRIMARY KEY ("id")
);

-- ----------------------------
-- Insert default alert types
-- ----------------------------
INSERT INTO "public"."alert_types" ("id", "name", "icon", "color", "description", "created_at", "updated_at", "display_name", "severity")
VALUES
(1, 'restricted', 'warning-circle', '#ff4d4f', 'Person detected in restricted area', '2025-06-24 05:05:22.839556+07', '2025-06-24 05:05:22.839556+07', 'Area Restriction', 'high'),
(2, 'fall-detection', 'warning-circle', '#ff4d4f', 'Person fall detected', '2025-06-24 05:05:22.856952+07', '2025-06-24 05:05:22.856952+07', 'Fall Detection', 'high'),
(3, 'loitering', 'warning-circle', '#ff4d4f', 'Extended loitering detected', '2025-06-24 05:05:22.866992+07', '2025-06-24 05:05:22.866992+07', 'Loitering Detection', 'high'),
(4, 'hazardous-area', 'warning-circle', '#ff4d4f', 'Person detected in hazardous area', '2025-08-04 09:44:08.436178+07', '2025-08-04 09:44:08.436178+07', 'Hazardous Area', 'high'),
(5, 'personal-protective-equipment', 'exclamation-circlewarning-circle', '#1890ff', 'Personal protective equipment violation detected', '2025-08-04 10:11:29.819417+07', '2025-08-04 10:11:29.819417+07', 'Personal Protective Equipment', 'high')
ON CONFLICT (id) DO NOTHING;

-- ----------------------------
-- Table structure for alerts
-- ----------------------------
CREATE TABLE IF NOT EXISTS "public"."alerts" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "alert_type_id" int4,
  "camera_id" int4,
  "message" text COLLATE "pg_catalog"."default" NOT NULL,
  "severity" varchar(20) COLLATE "pg_catalog"."default" NOT NULL,
  "image_url" varchar(255) COLLATE "pg_catalog"."default",
  "is_active" bool DEFAULT true,
  "detected_at" timestamptz(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "resolved_at" timestamptz(6),
  "resolved_by" varchar(100) COLLATE "pg_catalog"."default",
  "resolution_note" text COLLATE "pg_catalog"."default",
  "created_at" timestamptz(6) DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz(6) DEFAULT CURRENT_TIMESTAMP
);

-- Convert alerts to TimescaleDB hypertable (if not already)
SELECT create_hypertable('alerts', 'detected_at', if_not_exists => TRUE);

-- ----------------------------
-- Sequence for cameras
-- ----------------------------
CREATE SEQUENCE IF NOT EXISTS cameras_id_seq
  INCREMENT 1
  START 1
  MINVALUE 1
  MAXVALUE 2147483647
  CACHE 1;

-- ----------------------------
-- Table structure for cameras
-- ----------------------------
CREATE TABLE IF NOT EXISTS "public"."cameras" (
  "id" int4 NOT NULL DEFAULT nextval('cameras_id_seq'::regclass),
  "name" varchar(100) COLLATE "pg_catalog"."default" NOT NULL,
  "ip_address" varchar(15) COLLATE "pg_catalog"."default",
  "location" varchar(100) COLLATE "pg_catalog"."default",
  "status" varchar(20) COLLATE "pg_catalog"."default" DEFAULT 'active'::character varying,
  "created_at" timestamptz(6) DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz(6) DEFAULT CURRENT_TIMESTAMP,
  "ws_url" varchar(255) COLLATE "pg_catalog"."default"
);

-- ----------------------------
-- Table structure for face_recognitions
-- ----------------------------
CREATE TABLE IF NOT EXISTS "public"."face_recognitions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "camera_id" int8 NOT NULL,
  "image_path" varchar(255) COLLATE "pg_catalog"."default",
  "object_name" varchar(255) COLLATE "pg_catalog"."default",
  "detected_at" timestamptz(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "created_at" timestamptz(6) DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz(6) DEFAULT CURRENT_TIMESTAMP
);

-- Convert face_recognitions to TimescaleDB hypertable (if not already)
SELECT create_hypertable('face_recognitions', 'detected_at', if_not_exists => TRUE);

-- ----------------------------
-- Table structure for people_counts
-- ----------------------------
CREATE TABLE IF NOT EXISTS "public"."people_counts" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "camera_id" int4 NOT NULL,
  "timestamp" timestamptz(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "male_count" int4 DEFAULT 0,
  "female_count" int4 DEFAULT 0,
  "child_count" int4 DEFAULT 0,
  "adult_count" int4 DEFAULT 0,
  "elderly_count" int4 DEFAULT 0,
  "total_count" int4 GENERATED ALWAYS AS (
(male_count + female_count)
) STORED,
  "created_at" timestamptz(6) DEFAULT CURRENT_TIMESTAMP
);

-- Convert people_counts to TimescaleDB hypertable (if not already)
SELECT create_hypertable('people_counts', 'timestamp', if_not_exists => TRUE);

-- ----------------------------
-- Table structure for vehicle_counts
-- ----------------------------
CREATE TABLE IF NOT EXISTS "public"."vehicle_counts" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "cctv_id" int4 NOT NULL,
  "timestamp" timestamptz(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "in_count_car" int4 DEFAULT 0,
  "in_count_truck" int4 DEFAULT 0,
  "in_count_people" int4 DEFAULT 0,
  "out_count" int4 DEFAULT 0,
  "device_timestamp" timestamptz(6),
  "device_timestamp_utc" float8,
  "total_in_count" int4 GENERATED ALWAYS AS (
((in_count_car + in_count_truck) + in_count_people)
) STORED,
  "total_vehicle_in_count" int4 GENERATED ALWAYS AS (
(in_count_car + in_count_truck)
) STORED,
  "created_at" timestamptz(6) DEFAULT CURRENT_TIMESTAMP
);

-- Convert vehicle_counts to TimescaleDB hypertable (if not already)
SELECT create_hypertable('vehicle_counts', 'timestamp', if_not_exists => TRUE);

-- ----------------------------
-- TimescaleDB Compression Policies
-- ----------------------------

-- Add compression policies for time-series tables (if not already set)
DO $$
BEGIN
  -- Set compression for people_counts
  ALTER TABLE people_counts SET (timescaledb.compress = true);
  BEGIN
    PERFORM add_compression_policy('people_counts', INTERVAL '7 days', if_not_exists => TRUE);
  EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'Compression policy for people_counts already exists or could not be created';
  END;

  -- Set compression for vehicle_counts
  ALTER TABLE vehicle_counts SET (timescaledb.compress = true);
  BEGIN
    PERFORM add_compression_policy('vehicle_counts', INTERVAL '7 days', if_not_exists => TRUE);
  EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'Compression policy for vehicle_counts already exists or could not be created';
  END;

  -- Set compression for alerts
  ALTER TABLE alerts SET (timescaledb.compress = true);
  BEGIN
    PERFORM add_compression_policy('alerts', INTERVAL '30 days', if_not_exists => TRUE);
  EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'Compression policy for alerts already exists or could not be created';
  END;

  -- Set compression for face_recognitions
  ALTER TABLE face_recognitions SET (timescaledb.compress = true);
  BEGIN
    PERFORM add_compression_policy('face_recognitions', INTERVAL '14 days', if_not_exists => TRUE);
  EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'Compression policy for face_recognitions already exists or could not be created';
  END;
END $$;

-- ----------------------------
-- TimescaleDB Retention Policies
-- ----------------------------

-- Add retention policies to automatically remove old data (if not already set)
DO $$
BEGIN
  BEGIN
    PERFORM add_retention_policy('people_counts', INTERVAL '90 days', if_not_exists => TRUE);
  EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'Retention policy for people_counts already exists or could not be created';
  END;

  BEGIN
    PERFORM add_retention_policy('vehicle_counts', INTERVAL '90 days', if_not_exists => TRUE);
  EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'Retention policy for vehicle_counts already exists or could not be created';
  END;

  BEGIN
    PERFORM add_retention_policy('alerts', INTERVAL '365 days', if_not_exists => TRUE);
  EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'Retention policy for alerts already exists or could not be created';
  END;

  BEGIN
    PERFORM add_retention_policy('face_recognitions', INTERVAL '180 days', if_not_exists => TRUE);
  EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'Retention policy for face_recognitions already exists or could not be created';
  END;
END $$;

-- ----------------------------
-- Indexes for TimescaleDB tables
-- ----------------------------

-- Add indexes to improve query performance (if not already exist)
DO $$
BEGIN
  -- Create indexes for people_counts
  IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_people_counts_camera_id') THEN
    CREATE INDEX idx_people_counts_camera_id ON people_counts(camera_id);
  END IF;

  -- Create indexes for vehicle_counts
  IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_vehicle_counts_cctv_id') THEN
    CREATE INDEX idx_vehicle_counts_cctv_id ON vehicle_counts(cctv_id);
  END IF;

  -- Create indexes for alerts
  IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_alerts_camera_id') THEN
    CREATE INDEX idx_alerts_camera_id ON alerts(camera_id);
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_alerts_alert_type_id') THEN
    CREATE INDEX idx_alerts_alert_type_id ON alerts(alert_type_id);
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_alerts_is_active') THEN
    CREATE INDEX idx_alerts_is_active ON alerts(is_active);
  END IF;

  -- Create indexes for face_recognitions
  IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_face_recognitions_camera_id') THEN
    CREATE INDEX idx_face_recognitions_camera_id ON face_recognitions(camera_id);
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_face_recognitions_object_name') THEN
    CREATE INDEX idx_face_recognitions_object_name ON face_recognitions(object_name);
  END IF;
END $$;

-- ----------------------------
-- TimescaleDB Continuous Aggregates
-- ----------------------------

-- Create continuous aggregates for common analytical queries (if not already exist)
DO $$
BEGIN
  -- Create people_counts_hourly if not exists
  IF NOT EXISTS (SELECT 1 FROM pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace WHERE c.relname = 'people_counts_hourly' AND n.nspname = 'public') THEN
    -- Hourly people count aggregates
    CREATE MATERIALIZED VIEW people_counts_hourly
    WITH (timescaledb.continuous) AS
    SELECT
      time_bucket('1 hour', timestamp) AS bucket,
      camera_id,
      SUM(in_count) AS total_in_count,
      SUM(out_count) AS total_out_count,
      AVG(in_count) AS avg_in_count,
      AVG(out_count) AS avg_out_count,
      MAX(in_count) AS max_in_count,
      MAX(out_count) AS max_out_count
    FROM people_counts
    GROUP BY bucket, camera_id;

    -- Add refresh policy
    BEGIN
      PERFORM add_continuous_aggregate_policy('people_counts_hourly',
        start_offset => INTERVAL '3 days',
        end_offset => INTERVAL '1 hour',
        schedule_interval => INTERVAL '1 hour',
        if_not_exists => TRUE);
    EXCEPTION WHEN OTHERS THEN
      RAISE NOTICE 'Continuous aggregate policy for people_counts_hourly already exists or could not be created';
    END;
  END IF;

  -- Create vehicle_counts_hourly if not exists
  IF NOT EXISTS (SELECT 1 FROM pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace WHERE c.relname = 'vehicle_counts_hourly' AND n.nspname = 'public') THEN
    -- Hourly vehicle count aggregates
    CREATE MATERIALIZED VIEW vehicle_counts_hourly
    WITH (timescaledb.continuous) AS
    SELECT
      time_bucket('1 hour', timestamp) AS bucket,
      cctv_id,
      SUM(in_count_car) AS total_cars,
      SUM(in_count_truck) AS total_trucks,
      SUM(total_vehicle_in_count) AS total_vehicles,
      AVG(total_vehicle_in_count) AS avg_vehicles
    FROM vehicle_counts
    GROUP BY bucket, cctv_id;

    -- Add refresh policy
    BEGIN
      PERFORM add_continuous_aggregate_policy('vehicle_counts_hourly',
        start_offset => INTERVAL '3 days',
        end_offset => INTERVAL '1 hour',
        schedule_interval => INTERVAL '1 hour',
        if_not_exists => TRUE);
    EXCEPTION WHEN OTHERS THEN
      RAISE NOTICE 'Continuous aggregate policy for vehicle_counts_hourly already exists or could not be created';
    END;
  END IF;

  -- Create alerts_daily if not exists
  IF NOT EXISTS (SELECT 1 FROM pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace WHERE c.relname = 'alerts_daily' AND n.nspname = 'public') THEN
    -- Daily alert aggregates
    CREATE MATERIALIZED VIEW alerts_daily
    WITH (timescaledb.continuous) AS
    SELECT
      time_bucket('1 day', detected_at) AS bucket,
      camera_id,
      alert_type_id,
      severity,
      COUNT(*) AS alert_count
    FROM alerts
    GROUP BY bucket, camera_id, alert_type_id, severity;

    -- Add refresh policy
    BEGIN
      PERFORM add_continuous_aggregate_policy('alerts_daily',
        start_offset => INTERVAL '30 days',
        end_offset => INTERVAL '1 day',
        schedule_interval => INTERVAL '1 day',
        if_not_exists => TRUE);
    EXCEPTION WHEN OTHERS THEN
      RAISE NOTICE 'Continuous aggregate policy for alerts_daily already exists or could not be created';
    END;
  END IF;
END $$;

-- ----------------------------
-- TimescaleDB Job Scheduling
-- ----------------------------

-- Create jobs to vacuum tables periodically (if not already exist)
DO $$
DECLARE
  job_id INTEGER;
BEGIN
  -- Check if job for people_counts exists
  SELECT id INTO job_id FROM timescaledb_information.jobs
  WHERE proc_name = 'add_job' AND proc_schema = 'public'
  AND config->>'command' = 'VACUUM ANALYZE people_counts;' LIMIT 1;

  IF job_id IS NULL THEN
    PERFORM add_job('VACUUM ANALYZE people_counts;', '1 day');
  END IF;

  -- Check if job for vehicle_counts exists
  SELECT id INTO job_id FROM timescaledb_information.jobs
  WHERE proc_name = 'add_job' AND proc_schema = 'public'
  AND config->>'command' = 'VACUUM ANALYZE vehicle_counts;' LIMIT 1;

  IF job_id IS NULL THEN
    PERFORM add_job('VACUUM ANALYZE vehicle_counts;', '1 day');
  END IF;

  -- Check if job for alerts exists
  SELECT id INTO job_id FROM timescaledb_information.jobs
  WHERE proc_name = 'add_job' AND proc_schema = 'public'
  AND config->>'command' = 'VACUUM ANALYZE alerts;' LIMIT 1;

  IF job_id IS NULL THEN
    PERFORM add_job('VACUUM ANALYZE alerts;', '1 day');
  END IF;
END $$;

-- ----------------------------
-- TimescaleDB Configuration Notes
-- ----------------------------
-- This database is configured with TimescaleDB for time-series data optimization:
-- 1. Tables converted to hypertables: people_counts, vehicle_counts, alerts, face_recognitions
-- 2. Compression policies: Data older than 7-30 days is automatically compressed
-- 3. Retention policies: Data is automatically removed after 90-365 days
-- 4. Continuous aggregates: Pre-computed views for faster analytics queries
-- 5. Indexes: Optimized for common query patterns
-- 6. Scheduled jobs: Regular maintenance tasks for optimal performance
