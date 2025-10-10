
// Utilisation de Postgres localement avec Docker
# 1) Démarre un Postgres 16 local sur 5432
docker run --name pg \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=monitoring_database \
  -p 5432:5432 -d postgres:16

# 2) Teste
docker ps
psql -h localhost -p 5432 -U postgres -d monitoring_database
