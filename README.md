# integrityPOS by solidbit

integrityPOS es un sistema de punto de venta fijo y de grado industrial, diseñado con un enfoque Local-First, resiliencia operativa y pragmatismo absoluto. Desarrollado en Go (Golang) y PostgreSQL, sin dependencias de frameworks pesados o ORMs, interactuando directamente con el hardware y ofreciendo una interfaz web ultra rápida servida desde el propio binario.

## Requisitos

- **Go:** 1.22+
- **Base de Datos:** PostgreSQL 15+
- **Sistema Operativo:** Linux (ej. Raspberry Pi OS) o Windows (Desarrollo y pruebas)

## Estructura del Proyecto

El proyecto está diseñado usando Clean Architecture simplificada:

- `cmd/`: Puntos de entrada para el servidor POS (`posd`) y el generador de instaladores (`builder`).
- `internal/`: Lógica de negocio (servicios), acceso a datos (repositorios), manejadores de API y Web (handlers), y builder de instaladores.
- `internal/module/`: Sistema de módulos extensibles (mesas, fidelización, etc.).
- `assets/`: Plantillas HTML y recursos estáticos para el POS, segmentado por temas/verticales.
- `migrations/`: Scripts SQL puros para la creación de esquemas.

## Instrucciones para Compilar

```bash
# Compilar el servidor POS
go build -o posd ./cmd/posd

# Compilar el generador de instaladores
go build -o builder ./cmd/builder
```

## Instrucciones para Ejecutar en Desarrollo

1. **Base de Datos:** Asegúrate de tener PostgreSQL corriendo y crea la base de datos `integritypos`.
2. **Configuración:** Renombra o copia `.env.example` o ajusta el `config.yaml` con la URL de conexión a tu DB y credenciales necesarias.
3. **Migraciones:** El sistema ejecuta automáticamente los scripts en la carpeta `migrations/` al iniciar si las tablas no existen.
4. **Ejecutar:**

```bash
./posd
# o en Windows: posd.exe
```

La aplicación estará disponible en `http://localhost:8080`.
El panel de administración está en `http://localhost:8080/admin` (Autenticación requerida vía Bearer token en API o cabeceras para vistas web).

## Cómo usar el Builder

El Builder permite generar versiones personalizadas de integrityPOS para diferentes verticales (Restaurant, Retail, etc.) empaquetadas con módulos específicos.

1. Ejecuta el builder:
   ```bash
   ./builder
   ```
2. Abre tu navegador en `http://localhost:8081`.
3. Selecciona la vertical, los módulos requeridos y haz clic en "Generar Instalador".
4. Se descargará un archivo `.tar.gz` con el binario compilado (si estás en Windows generará `.exe`), las plantillas personalizadas, y un script de instalación (`install.sh` o `install.bat`).

## Pruebas

### Unitarias
Se prueban los servicios sin depender de la base de datos, usando mocks internos en Go puro.
```bash
go test ./internal/service/...
```

### Integración
Requieren una base de datos de pruebas real.
```bash
# Asegúrate de tener una base de datos 'integritypos_test' en PostgreSQL
export INTEGRATION_TESTS=1
export TEST_DB_DSN="postgres://postgres:postgres@localhost:5432/integritypos_test?sslmode=disable"
go test ./internal/repository/...
```

## Despliegue en Raspberry Pi (Linux)

1. Transfiere el instalador (`.tar.gz`) generado por el builder a la Raspberry Pi.
2. Descomprímelo: `tar -xzf integritypos_installer.tar.gz`
3. Ejecuta el script de instalación con permisos de superusuario:
   ```bash
   sudo bash install.sh
   ```
4. Sigue las instrucciones y ajusta el `config.yaml` final.

## Licencia

Propietaria - SolidBit
