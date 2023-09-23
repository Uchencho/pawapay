OUTPUT = main

.PHONY: test
test:
	go test -failfast ./...

.PHONY: clean
clean:
	rm -f $(OUTPUT)

build-local: 
	go build -o $(OUTPUT) ./client/main.go

run: build-local
	@echo ">> Running application ..."
	./$(OUTPUT)
