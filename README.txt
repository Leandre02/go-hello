
// Utilisation de Postgres localement avec Docker
# 1) Démarre un Postgres 16 local sur 5432
docker run --name pg \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=monitoring_database \
  -p 5432:5432 -d postgres:16

# 2) Teste
docker ps
psql -h localhost -p 5432 -U postgres -d monitoring_database


A savoir :
Image Docker = modèle (recette figée).

Fichier immuable, en couches (layers).

Contient tout pour lancer une app: OS minimal, runtime, dépendances, ton binaire.

On la construit (docker build -t monapp:1.0 .), on la stocke (local ou registry), on ne la “modifie” pas en live.

Conteneur = instance qui tourne (copie vivante de l’image).

Process isolé créé à partir d’une image.

A un système de fichiers en écriture au-dessus de l’image (writable layer), garde de l’état (logs, fichiers… si non montés en volume).

On le lance/arrête (docker run, docker stop), on peut en lancer plusieurs depuis la même image.

Analogie rapide

Image = classe (définition).

Conteneur = objet (instance en exécution).

Exemples concrets
# Construire l'image
docker build -t api:1.0 .

# Lister les images
docker images

# Créer + démarrer un conteneur depuis l’image
docker run -d --name api1 -p 8080:8080 api:1.0

# Lancer un deuxième conteneur depuis la même image
docker run -d --name api2 -p 8081:8080 api:1.0

# Lister les conteneurs en cours
docker ps

Points clés à retenir

Image = read-only, conteneur = image + couche writable.

Les données persistantes → volumes (-v), pas dans la couche du conteneur.

Les images vivent dans un registry (Docker Hub, GHCR…).

Un conteneur meurt, l’image reste; tu peux en relancer un autre identique en 1 commande.



komle@ObstinateMage:~/cloud/go-hello$ go build -o /home/komle/cloud/go-hello/tmp/app ./src/cmd/server
komle@ObstinateMage:~/cloud/go-hello$ docker build -t go-hello:latest -f dockerfile .
[+] Building 22.4s (16/16) FINISHED                                                                                                                              docker:default
 => [internal] load build definition from dockerfile                                                                                                                       0.0s
 => => transferring dockerfile: 567B                                                                                                                                       0.0s
 => [internal] load metadata for docker.io/library/golang:1.24                                                                                                             0.8s
 => [internal] load metadata for gcr.io/distroless/static-debian12:latest                                                                                                  0.6s
 => [internal] load .dockerignore                                                                                                                                          0.0s
 => => transferring context: 163B                                                                                                                                          0.0s
 => [build 1/7] FROM docker.io/library/golang:1.24@sha256:273d4e65baa782dbe293c9192d600b72b17d415c1429e16bed99efcc5e61efb8                                                 0.1s
 => => resolve docker.io/library/golang:1.24@sha256:273d4e65baa782dbe293c9192d600b72b17d415c1429e16bed99efcc5e61efb8                                                       0.1s
 => [stage-1 1/3] FROM gcr.io/distroless/static-debian12:latest@sha256:87bce11be0af225e4ca761c40babb06d6d559f5767fbf7dc3c47f0f1a466b92c                                    0.1s
 => => resolve gcr.io/distroless/static-debian12:latest@sha256:87bce11be0af225e4ca761c40babb06d6d559f5767fbf7dc3c47f0f1a466b92c                                            0.1s
 => [internal] load build context                                                                                                                                          0.0s
 => => transferring context: 5.76kB                                                                                                                                        0.0s
 => CACHED [build 2/7] WORKDIR /app                                                                                                                                        0.0s
 => CACHED [build 3/7] COPY go.mod ./                                                                                                                                      0.0s
 => CACHED [build 4/7] COPY go.sum ./                                                                                                                                      0.0s
 => CACHED [build 5/7] RUN go mod download                                                                                                                                 0.0s
 => [build 6/7] COPY . .                                                                                                                                                   0.1s
 => [build 7/7] RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./src/cmd/server                                                                                20.5s
 => CACHED [stage-1 2/3] WORKDIR /app                                                                                                                                      0.0s
 => CACHED [stage-1 3/3] COPY --from=build /app/app /app/app                                                                                                               0.0s
 => exporting to image                                                                                                                                                     0.4s
 => => exporting layers                                                                                                                                                    0.0s
 => => exporting manifest sha256:8ae876a113fb67688e0438340450b76c53c521d500973209c923642bd847b8c5                                                                          0.0s
 => => exporting config sha256:1b5078552c04189df8373b241989c002c859c281114026132dcb92ca042988df                                                                            0.0s
 => => exporting attestation manifest sha256:704e3ffac991d2d6a864dec33fd4b51959c34baa7a8e15fba5292f87c7befeb4                                                              0.0s
 => => exporting manifest list sha256:afa0591dc1b31c2a4fa0df1f3316de154ada425e35304ec864ab61582d4759bf                                                                     0.0s
 => => naming to docker.io/library/go-hello:latest                                                                                                                         0.0s
 => => unpacking to docker.io/library/go-hello:latest                                                                                                                      0.2s
komle@ObstinateMage:~/cloud/go-hello$ docker build -t go-hello:latest -f dockerfile .
[+] Building 0.9s (16/16) FINISHED                                                                                                                               docker:default
 => [internal] load build definition from dockerfile                                                                                                                       0.0s
 => => transferring dockerfile: 567B                                                                                                                                       0.0s
 => [internal] load metadata for docker.io/library/golang:1.24                                                                                                             0.4s
 => [internal] load metadata for gcr.io/distroless/static-debian12:latest                                                                                                  0.3s
 => [internal] load .dockerignore                                                                                                                                          0.0s
 => => transferring context: 163B                                                                                                                                          0.0s
 => [build 1/7] FROM docker.io/library/golang:1.24@sha256:273d4e65baa782dbe293c9192d600b72b17d415c1429e16bed99efcc5e61efb8                                                 0.1s
 => => resolve docker.io/library/golang:1.24@sha256:273d4e65baa782dbe293c9192d600b72b17d415c1429e16bed99efcc5e61efb8                                                       0.0s
 => [internal] load build context                                                                                                                                          0.0s
 => => transferring context: 1.19kB                                                                                                                                        0.0s
 => [stage-1 1/3] FROM gcr.io/distroless/static-debian12:latest@sha256:87bce11be0af225e4ca761c40babb06d6d559f5767fbf7dc3c47f0f1a466b92c                                    0.1s
 => => resolve gcr.io/distroless/static-debian12:latest@sha256:87bce11be0af225e4ca761c40babb06d6d559f5767fbf7dc3c47f0f1a466b92c                                            0.0s
 => CACHED [stage-1 2/3] WORKDIR /app                                                                                                                                      0.0s
 => CACHED [build 2/7] WORKDIR /app                                                                                                                                        0.0s
 => CACHED [build 3/7] COPY go.mod ./                                                                                                                                      0.0s
 => CACHED [build 4/7] COPY go.sum ./                                                                                                                                      0.0s
 => CACHED [build 5/7] RUN go mod download                                                                                                                                 0.0s
 => CACHED [build 6/7] COPY . .                                                                                                                                            0.0s
 => CACHED [build 7/7] RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./src/cmd/server                                                                          0.0s
 => CACHED [stage-1 3/3] COPY --from=build /app/app /app/app                                                                                                               0.0s
 => exporting to image                                                                                                                                                     0.2s
 => => exporting layers                                                                                                                                                    0.0s
 => => exporting manifest sha256:8ae876a113fb67688e0438340450b76c53c521d500973209c923642bd847b8c5                                                                          0.0s
 => => exporting config sha256:1b5078552c04189df8373b241989c002c859c281114026132dcb92ca042988df                                                                            0.0s
 => => exporting attestation manifest sha256:277b2c70de42fb54a24e460d03355821d3110658363e74672dc02c5397b3a663                                                              0.0s
 => => exporting manifest list sha256:cdc0048d7a1a142bc7f5726c11d59907fa1ee7eae92b0eb6984fd1c128a965f3                                                                     0.0s
 => => naming to docker.io/library/go-hello:latest                                                                                                                         0.0s
 => => unpacking to docker.io/library/go-hello:latest                                                                                                                      0.0s
komle@ObstinateMage:~/cloud/go-hello$ # Assure-toi que Postgres écoute sur 127.0.0.1:5432
docker run --rm --network host \
  -e DATABASE_URL="postgres://postgres:postgres@localhost:5432/monitoring_database?sslmode=disable" \
  go-hello:latest
