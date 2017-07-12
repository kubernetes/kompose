# Copyright 2016 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


# This Dockefile creates image where all Kompose tests can be run

FROM registry.centos.org/centos/centos:7

RUN yum install -y epel-release
RUN yum install -y golang make git jq &&\
    yum -y clean all

ENV GOPATH="/opt/go" \
    # KOMPOSE_TMP_SRC is where kompose source will be mounted from host
    KOMPOSE_TMP_SRC="/opt/tmp/kompose" 

ENV PATH="$PATH:$GOPATH/bin" \
    # KOMPOSE_SRC is where kompose source will be copied when container starts (by run.sh script)
    # this is to ensure that we won't write anything to host volume mount
    KOMPOSE_SRC="$GOPATH/src/github.com/kubernetes/kompose"

RUN go get github.com/Masterminds/glide &&\
    go get github.com/sgotti/glide-vc &&\
    go get github.com/golang/lint/golint

WORKDIR $KOMPOSE_SRC
# This image can be run as any user
RUN chmod -R ugo+rw $GOPATH


COPY run.sh /opt/tools/
ENTRYPOINT ["/opt/tools/run.sh"]



