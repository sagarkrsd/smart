# Copyright 2018 The OpenEBS Authors.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#    http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


# Build the smart image.

build: vet fmt smart

PACKAGES = $(shell go list ./... | grep -v '/vendor/')

vet:
	go list ./... | grep -v "./vendor/*" | xargs go vet

fmt:
	find . -type f -name "*.go" | grep -v "./vendor/*" | xargs gofmt -s -w -l

test: 	vet fmt
	@echo "--> Running go test" ;
	@go test $(PACKAGES)

header:
	@echo "----------------------------"
	@echo "--> smart       "
	@echo "----------------------------"
	@echo

smart: header
	@echo '--> Building binary...'
	@CTLNAME=$(shell go build cmd/main.go)
	@echo '--> Built binary.'
	@echo

.PHONY: build