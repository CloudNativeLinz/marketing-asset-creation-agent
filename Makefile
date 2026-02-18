.PHONY: clean test run

run:
	go run . -p prompts/speaking-prompt.md -i assets/thispersondoesnotexist.jpg -s 1024x1024 

clean:
	rm -f assets/*_generated.*

test:
	go test -v -timeout 5m ./...
