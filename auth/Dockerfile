
# Utilise une image officielle de Go
FROM golang:1.24


# Crée un dossier dans le container
WORKDIR /app

# Copie le go.mod et go.sum en premier (si utilisés)
COPY go.mod go.sum ./

# Télécharge les dépendances
RUN go mod download

# Copie le reste de l'application
COPY . .

# Compile le programme
RUN go build -o auth-service .

# Expose le port 5001 pour l'authentification
EXPOSE 5001

# Commande pour lancer le microservice
CMD ["./auth-service"]