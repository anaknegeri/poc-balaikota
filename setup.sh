#!/bin/bash
# File ini berisi instruksi untuk menjalankan semua migrasi

# Buat struktur folder yang diperlukan
mkdir -p db/init db/migrations

# Salin file migrasi ke folder migrasi
echo "Menyalin file migrasi..."
cp 001_initial_schema.sql db/migrations/
cp 002_seed_alert_types.sql db/migrations/
cp 003_rollback.sql db/migrations/

# Buat script init db menjadi executable
chmod +x db/init/01-init-db.sh

echo "Setup database migrasi selesai!"
echo "Jalankan 'docker-compose up' untuk memulai aplikasi dengan database yang sudah diinisialisasi."
