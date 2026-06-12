#!/bin/bash

# Script de test global pour la plateforme BI Analytics

echo "🚀 Démarrage de la vérification de la plateforme..."

# 1. Vérification de Docker et Docker Compose
if ! [ -x "$(command -v docker-compose)" ]; then
  echo "❌ Erreur : docker-compose n'est pas installé." >&2
  exit 1
fi

# 2. Vérification du fichier .env
if [ ! -f .env ]; then
  echo "⚠️  Fichier .env manquant. Création à partir de .env.example..."
  cp .env.example .env
fi

# 3. Lancement de l'infrastructure
echo "📦 Lancement de l'infrastructure (Postgres, Valkey, Maildev)..."
docker-compose up -d valkey postgres_meta maildev

echo "⏳ Attente du démarrage des services (5s)..."
sleep 5

# 4. Tests des services Go
echo "🔍 Vérification de gateway-auth-go (Compilation)..."
cd gateway-auth-go && go mod tidy && go build -o /dev/null ./cmd/main.go && echo "✅ gateway-auth-go : OK" || echo "❌ gateway-auth-go : Échec"
cd ..

echo "🔍 Vérification de metadata-semantic-go (Compilation)..."
cd metadata-semantic-go && go mod tidy && go build -o /dev/null ./cmd/main.go && echo "✅ metadata-semantic-go : OK" || echo "❌ metadata-semantic-go : Échec"
cd ..

# 5. Tests des services Rust
echo "🔍 Vérification de query-engine-rust (Check syntaxe)..."
cd query-engine-rust && cargo check && echo "✅ query-engine-rust : OK" || echo "❌ query-engine-rust : Échec"
cd ..

# 6. Vérification de la connectivité Valkey (via docker)
echo "🔍 Vérification de la connectivité Valkey..."
docker exec bi_valkey_broker valkey-cli ping | grep -q PONG && echo "✅ Valkey : OK" || echo "❌ Valkey : Échec"

echo "🏁 Vérification terminée."
echo "Pour lancer l'application complète, utilisez : docker-compose up --build"
