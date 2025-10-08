#!/bin/bash

set -ex

export TZ=Asia/Shanghai

LLM_HOME_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
echo "MONITOR_HOME_DIR: ${LLM_HOME_DIR}"

apt-get clean
apt-get autoclean
apt-get update

apt-get install -y \
    libssl-dev gcc g++ make \
    curl \
    git \
    gdb \
    openssh-server \
    build-essential \
    libboost-all-dev \
    net-tools \
    vim \
    libzmq3-dev \
    libgoogle-glog-dev \
    cmake \
    libbsd-dev

# 安装eventpp
if [ -d "${LLM_HOME_DIR}/docker/build/install/eventpp" ]; then
    cd ${LLM_HOME_DIR}/docker/build/install/eventpp
    ./install_eventpp.sh || { echo "eventpp安装失败"; exit 1; }
else
    echo "警告：未找到eventpp安装目录 !!!"
    exit 1;
fi

# 安装simdjson
if [ -d "${LLM_HOME_DIR}/docker/build/install/simdjson" ]; then
    cd ${LLM_HOME_DIR}/docker/build/install/simdjson
    ./install_simdjson.sh || { echo "simdjson安装失败"; exit 1; }
else
    echo "警告：未找到simdjson安装目录 !!!"
    exit 1;
fi

echo "所有组件安装完成！"