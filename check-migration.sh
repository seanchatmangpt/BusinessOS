#!/bin/bash
# Quick script to check if migration 052 was applied

# Check if onboarding_completed column exists in user table
echo "Checking if onboarding_completed column exists..."
echo "Please run this with your DATABASE_URL:"
echo ""
echo "psql \$DATABASE_URL -c \"SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'user' AND column_name = 'onboarding_completed';\""
echo ""
echo "Expected output:"
echo "      column_name       | data_type"
echo "------------------------+-----------"
echo " onboarding_completed   | boolean"
echo ""
echo "If you see no rows, the migration was NOT applied."
