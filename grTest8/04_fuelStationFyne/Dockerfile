
FROM postgres:16

ENV POSTGRES_USER=postgres
ENV POSTGRES_DB=fuelstation

EXPOSE 5432

COPY migrations/ /docker-entrypoint-initdb.d/