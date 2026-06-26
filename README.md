# IntegrityPOS

IntegrityPOS by SolidBit es un sistema de punto de venta fijo y de grado industrial, diseñado con arquitectura Clean Code, minimalismo y resiliencia operativa. Escrito puramente en Go (Golang) y usando PostgreSQL.

## Requisitos

- **Go**: 1.22+
- **PostgreSQL**: 15+

## Estructura del Proyecto

- `cmd/posd/`: Punto de entrada del binario principal (servidor POS).
- `cmd/builder/`: Herramienta de empaquetado/build.
- `internal/`: Lógica interna (handlers, services, repositories, models).
- `assets/`: Plantillas HTML y archivos estáticos.
- `migrations/`: Scripts de migración SQL para PostgreSQL.

## Compilación

Para compilar el binario principal:

```bash
go build -o posd.exe ./cmd/posd
```

Para compilar el builder:

```bash
go build -o builder.exe ./cmd/builder
```

## Ejecución en Desarrollo

1. Clona el repositorio.
2. Asegúrate de tener PostgreSQL ejecutándose.
3. Copia `config.example.yaml` a `config.yaml` y configura los detalles de tu base de datos (DSN).
4. Ejecuta las migraciones en tu base de datos (o confía en el flag automático si está habilitado).
5. Ejecuta el servidor:

```bash
./posd.exe -config config.yaml
```

La aplicación web estará disponible por defecto en el puerto configurado (ej: `http://localhost:8080/pos`).

## Ejecución de Pruebas

El sistema incluye pruebas unitarias y pruebas de integración usando mocks manuales y PostgreSQL.

Para correr las pruebas unitarias:

```bash
go test ./internal/service/...
```

Para correr las pruebas de integración (requiere una base de datos de prueba separada):

```bash
INTEGRATION_TESTS=1 go test ./internal/repository/...
```

**Nota sobre Integración**:
Las pruebas de integración buscan conectar a `postgres://localhost:5432/integritypos_test?sslmode=disable`. Asegúrate de crear esta base de datos antes de ejecutarlas:
```sql
CREATE DATABASE integritypos_test;
```

## Licencia

Propietaria - Solidbit. Todos los derechos reservados.
