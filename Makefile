.PHONY: bootstrap doctor build web test lint audit smoke serve clean

bootstrap:
	bash ./scripts/bootstrap-ubuntu.sh

doctor:
	bash ./scripts/doctor.sh

web:
	cd web && npm install && npm run build
	rm -rf internal/ui/dist && mkdir -p internal/ui/dist
	cp -r web/dist/* internal/ui/dist/

build: web
	go build -o bin/platformforge ./cmd/platformforge

test:
	go test ./...
	cd web && npm test

integration:
	go test ./tests/integration/...

e2e:
	bash ./tests/e2e/run-e2e.sh

lint:
	bash ./scripts/lint.sh

audit:
	bash ./scripts/audit.sh

smoke:
	bash ./tests/smoke/wsl-smoke.sh

serve: build
	./bin/platformforge serve

clean:
	rm -rf bin/ web/dist web/node_modules internal/ui/dist
