FROM  ubuntu:20.04

ARG DEBIAN_FRONTEND=noninteractive
ENV TZ=Asia/Shanghai

SHELL ["/bin/bash", "-c"]

RUN apt-get clean && \
    apt-get autoclean
COPY apt/sources.list /etc/apt/

RUN apt-get update && \
    apt-get install -y \
    libssl-dev gcc g++ make gdb \
    curl \
    build-essential \
    libboost-all-dev \
    vim \
    libzmq3-dev \
    libgoogle-glog-dev \
    cmake \
    libbsd-dev

COPY install/eventpp /tmp/install/eventpp
RUN /tmp/install/eventpp/install_eventpp.sh 

COPY install/simdjson /tmp/install/simdjson
RUN /tmp/install/simdjson/install_simdjson.sh 








