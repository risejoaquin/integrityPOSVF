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

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -o bin/posd-arm64 ./cmd/posd

build-all: build-pos build-builder build-linux-arm64

package: build-all
	mkdir -p dist/integritypos/bin
	cp bin/posd dist/integritypos/bin/
	cp bin/posd-arm64 dist/integritypos/bin/
	cp bin/builder dist/integritypos/bin/
	cp -r assets dist/integritypos/
	cp -r migrations dist/integritypos/
	cp config.example.yaml dist/integritypos/
	tar -czvf integritypos-release.tar.gz -C dist integritypos
	rm -rf dist
