.PHONY: build-pos build-builder test test-integration run-dev run-builder lint clean

# Construye el binario del POS
build-pos:
	go build -o bin/posd ./cmd/posd

# Construye el binario del Builder
build-builder:
	go build -o bin/builder ./cmd/builder

# Ejecuta todas las pruebas unitarias
test:
	go test ./internal/service/... ./internal/module/... ./internal/handler/...

# Ejecuta pruebas de integración (requiere BD de test)
test-integration:
	INTEGRATION_TESTS=1 go test ./internal/repository/...

# Inicia el servidor POS en modo desarrollo
run-dev:
	go run ./cmd/posd

# Inicia el Builder en modo desarrollo
run-builder:
	go run ./cmd/builder

# Analiza el código con go vet
lint:
	go vet ./...

# Limpia binarios
clean:
	rm -rf bin/
