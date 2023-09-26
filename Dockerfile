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

FROM golang:1.21.0-alpine AS builder

RUN apk add --update --no-cache make

WORKDIR /go/src/tf2d2

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN make build

FROM alpine:3.18.3

ARG TARGET_OS=linux
ARG TARGET_ARCH=amd64

COPY --from=builder /go/src/tf2d2/bin/${TARGET_OS}-${TARGET_ARCH}/tf2d2 /usr/local/bin/

ENTRYPOINT ["tf2d2"]
