# Copyright © 2023 The tf2d2 Authors

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

GOOS 	?= $(shell go env GOOS)
GOARCH	?= $(shell go env GOARCH)

default: build

build:
	go build -o ./bin/$(GOOS)-$(GOARCH)/tf2d2

local-release:
	goreleaser release --clean --skip-publish --skip-docker --skip-validate --snapshot

lint:
	golangci-lint run ./...

test:
	go test -v -covermode=atomic -coverprofile=coverage.out ./...

pre-commit:
	pre-commit run --all-files

local-run:
	./bin/$(GOOS)-$(GOARCH)/tf2d2 $(ARGS)
