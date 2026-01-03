
/*
 * SPDX-FileCopyrightText: 2024 M5Stack Technology CO LTD
 *
 * SPDX-License-Identifier: MIT
 */
#include "zmq_bus.h"

#include "all.h"
#include <stdbool.h>
#include <functional>
#include <cstring>
#include <StackFlowUtil.h>
#if defined(__ARM_NEON) || defined(__ARM_NEON__)
#include <arm_neon.h>
#endif

#ifdef ENABLE_BSON
#include <bson/bson.h>
#endif

using namespace StackFlows;

zmq_bus_com::zmq_bus_com()
{
    exit_flage = 1;
    err_count = 0;
    json_str_flage_ = 0;
}

void zmq_bus_com::work(const std::string &zmq_url_format, int port)
{
    _port = port;
    exit_flage = 1;
    std::string ports = std::to_string(port);
    std::vector<char> buff(zmq_url_format.length() + ports.length(), 0);
    sprintf((char *)buff.data(), zmq_url_format.c_str(), port);
    _zmq_url = std::string((char *)buff.data());
    user_chennal_ =
        std::make_unique<pzmq>(_zmq_url, ZMQ_PULL, [this](pzmq *_pzmq, const std::shared_ptr<pzmq_data> &data) {
            this->send_data(data->string());
        });
}

void zmq_bus_com::stop()
{
    exit_flage = 0;
    user_chennal_.reset();
}

void zmq_bus_com::on_data(const std::string &data)
{
    std::cout << "on_data:" << data << std::endl;

    unit_action_match(_port, data);
}

void zmq_bus_com::send_data(const std::string &data)
{
}


zmq_bus_com::~zmq_bus_com()
{
    if (exit_flage)
    {
        stop();
    }
}

int zmq_bus_publisher_push(const std::string &work_id, const std::string &json_str)
{
    ALOGW("zmq_bus_publisher_push json_str:%s", json_str.c_str());

    if (work_id.empty())
    {
        ALOGW("work_id is empty");
        return -1;
    }
    unit_data *unit_p = NULL;
    SAFE_READING(unit_p, unit_data *, work_id);
    if (unit_p)
        unit_p->send_msg(json_str);
    ALOGW("zmq_bus_publisher_push work_id:%s", work_id.c_str());

    else
    {
        ALOGW("zmq_bus_publisher_push failed, not have work_id:%s", work_id.c_str());
        return -1;
    }
    return 0;
}

void *usr_context;

void zmq_com_send(int com_id, const std::string &out_str)
{
    char zmq_push_url[128];
    sprintf(zmq_push_url, zmq_c_format.c_str(), com_id);
    pzmq _zmq(zmq_push_url, ZMQ_PUSH);
    std::string out = out_str + "\n";
    _zmq.send_data(out);
}


void zmq_bus_com::select_json_str(const std::string &json_src, std::function<void(const std::string &)> out_fun)
{
    // The TCP side is a byte stream: one recv() may contain a partial JSON line
    // or multiple newline-delimited JSON requests. unit_action_match expects a
    // single JSON object each time, so we must do framing here.
    //
    // Protocol: client sends one JSON object per line, terminated by '\n'.
    json_str_.append(json_src);

    // Guard against unbounded growth if the sender is buggy or the connection
    // gets desynced.
    constexpr size_t kMaxBufferedBytes = 8 * 1024 * 1024; // 8MB
    if (json_str_.size() > kMaxBufferedBytes) {
        json_str_.clear();
        json_str_flage_ = 0;
        return;
    }

    size_t pos = 0;
    while (true) {
        auto nl = json_str_.find('\n', pos);
        if (nl == std::string::npos) {
            break;
        }

        std::string line = json_str_.substr(pos, nl - pos);
        if (!line.empty() && line.back() == '\r') {
            line.pop_back();
        }
        if (!line.empty()) {
            out_fun(line);
        }
        pos = nl + 1;
    }

    if (pos > 0) {
        json_str_.erase(0, pos);
    }
}
