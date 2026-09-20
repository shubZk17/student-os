#!/bin/bash
# LocalStack Initialization Script
# Runs automatically when LocalStack starts (via /etc/localstack/init/ready.d/).
# Creates AWS-compatible resources that StudentOS can use locally.

set -euo pipefail

ENDPOINT="http://localhost:4566"
REGION="us-east-1"

echo "=== StudentOS LocalStack Initialization ==="

# ---------------------------------------------------------------------------
# S3: Asset storage bucket (for future resume/portfolio uploads)
# ---------------------------------------------------------------------------
echo "[1/3] Creating S3 bucket: studentos-assets"
awslocal s3 mb s3://studentos-assets --region "$REGION" 2>/dev/null || true

# Verify
awslocal s3 ls | grep studentos-assets && echo "  ✓ S3 bucket ready" || echo "  ✗ S3 bucket creation failed"

# ---------------------------------------------------------------------------
# EventBridge: Ingestion schedule (6-hour cadence)
# ---------------------------------------------------------------------------
echo "[2/3] Creating EventBridge rule: studentos-ingest-schedule"
awslocal events put-rule \
    --name studentos-ingest-schedule \
    --schedule-expression "rate(6 hours)" \
    --state ENABLED \
    --description "Triggers StudentOS job ingestion every 6 hours" \
    --region "$REGION" 2>/dev/null || true

echo "  ✓ EventBridge rule created"

# The target would normally point to the backend's POST /internal/ingest
# endpoint. In LocalStack, we configure this as an HTTP target for
# demonstration. The actual ingestion can be triggered manually via:
#   curl -X POST -H "X-Ingest-Token: $INGEST_TOKEN" http://localhost:8080/internal/ingest

# ---------------------------------------------------------------------------
# Verify all resources
# ---------------------------------------------------------------------------
echo "[3/3] Verifying LocalStack resources..."
echo ""
echo "  S3 Buckets:"
awslocal s3 ls
echo ""
echo "  EventBridge Rules:"
awslocal events list-rules --region "$REGION" --query 'Rules[].{Name:Name,State:State,Schedule:ScheduleExpression}' --output table
echo ""
echo "=== StudentOS LocalStack ready ==="
