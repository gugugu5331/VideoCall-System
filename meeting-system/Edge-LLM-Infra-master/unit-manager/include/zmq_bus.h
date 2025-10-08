/*
 * SPDX-FileCopyrightText: 2024 M5Stack Technology CO LTD
 *
 * SPDX-License-Identifier: MIT
 */
#pragma once

#include "pzmq.hpp"
#include <vector>
#include "unit_data.h"

using namespace StackFlows;

int zmq_bus_publisher_push(const std::string &work_id, const std::string &json_str);
void zmq_com_send(int com_id, const std::string &out_str);

class zmq_bus_com
{
protected:
    std::string _zmq_url;
    int exit_flage;
    int err_count;
    int _port;
    std::string json_str_;
    int json_str_flage_;

public:
    std::unique_ptr<pzmq> user_chennal_;
    zmq_bus_com();
    void work(const std::string &zmq_url_format, int port);
    void stop();
    void select_json_str(const std::string &json_src, std::function<void(const std::string &)> out_fun);
    virtual void on_data(const std::string &data);
    virtual void send_data(const std::string &data);
    ~zmq_bus_com();
};