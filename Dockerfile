FROM debian:bookworm-slim

WORKDIR /app

# Instala o wget no Debian
RUN apt-get update && apt-get install -y wget && rm -rf /var/lib/apt/lists/*

# Baixa e configura o binário
RUN wget -O pormade-email-manager https://github.com/viniggjpormade/pormade-email-manager/releases/latest/download/pormade-email-manager && \
    chmod +x pormade-email-manager && \
	mv pormade-email-manager /usr/local/bin/pormade-email-manager

ENTRYPOINT ["pormade-email-manager"]