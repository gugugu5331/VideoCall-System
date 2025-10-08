/**
 * SPDX-FileCopyrightText: 2024 M5Stack Technology CO LTD
 *
 * SPDX-License-Identifier: MIT
 */
#include "pzmq_data.h"

namespace StackFlows {

pzmq_data::pzmq_data() {
    zmq_msg_init(&msg);
}

pzmq_data::~pzmq_data() {
    zmq_msg_close(&msg);
}

std::shared_ptr<std::string> pzmq_data::get_string() {
    auto len = zmq_msg_size(&msg);
    return std::make_shared<std::string>((const char*)zmq_msg_data(&msg), len);
}

std::string pzmq_data::string() {
    auto len = zmq_msg_size(&msg);
    return std::string((const char*)zmq_msg_data(&msg), len);
}

void* pzmq_data::data() {
    return zmq_msg_data(&msg);
}

size_t pzmq_data::size() {
    return zmq_msg_size(&msg);
}

zmq_msg_t* pzmq_data::get() {
    return &msg;
}

std::string pzmq_data::get_param(int index, const std::string& idata) {
    const char* data = nullptr;
    int size = 0;
    
    if (idata.length() > 0) {
        data = idata.c_str();
        size = static_cast<int>(idata.length());
    } else {
        data = static_cast<const char*>(zmq_msg_data(&msg));
        size = static_cast<int>(zmq_msg_size(&msg));
    }

    if ((index % 2) == 0) {
        return std::string(data + 1, static_cast<size_t>(data[0]));
    } else {
        return std::string(data + data[0] + 1, static_cast<size_t>(size - data[0] - 1));
    }
}

std::string pzmq_data::set_param(std::string param0, std::string param1) {
    std::string data = " " + param0 + param1;
    data[0] = static_cast<char>(param0.length());
    return data;
}

} // namespace StackFlows