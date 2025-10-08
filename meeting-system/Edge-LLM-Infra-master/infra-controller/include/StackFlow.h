/*
 * SPDX-FileCopyrightText: 2024 M5Stack Technology CO LTD
 *
 * SPDX-License-Identifier: MIT
 */
#pragma once

// #define __cplusplus 1

#include <semaphore.h>
#include <unistd.h>
#include <iostream>

#include <string>
#include <list>
#include <functional>
#include <unordered_map>
#include <mutex>
#include <eventpp/eventqueue.h>
#include <thread>
#include <memory>
#include "json.hpp"
#include <regex>
#include "pzmq.hpp"
#include "StackFlowUtil.h"
#include "channel.h"
namespace StackFlows {
class StackFlow {
public:
    typedef enum {
        EVENT_NONE = 0,
        EVENT_SETUP,
        EVENT_EXIT,
        EVENT_PAUSE,
        EVENT_TASKINFO,
    } LOCAL_EVENT;

    std::string unit_name_;
    std::string request_id_;
    std::string out_zmq_url_;

    std::atomic<bool> exit_flage_;
    std::atomic<int> status_;

    // 线程安全的事件队列，存储事件类型
    eventpp::EventQueue<int, void(const std::shared_ptr<void> &)> event_queue_;
    std::unique_ptr<std::thread> even_loop_thread_;

    std::unique_ptr<pzmq> rpc_ctx_;

    std::unordered_map<int, std::shared_ptr<llm_channel_obj>> llm_task_channel_;

    StackFlow(const std::string &unit_name);
    void even_loop();
    void _none_event(const std::shared_ptr<void> &arg);

    template <typename T>
    std::shared_ptr<llm_channel_obj> get_channel(T workid)
    {
        int _work_id_num;
        if constexpr (std::is_same<T, int>::value) {
            _work_id_num = workid;
        } else if constexpr (std::is_same<T, std::string>::value) {
            _work_id_num = sample_get_work_id_num(workid);
        } else {
            return nullptr;
        }
        return llm_task_channel_.at(_work_id_num);
    }

    std::string _rpc_setup(pzmq *_pzmq, const std::shared_ptr<pzmq_data> &data);
    void _setup(const std::shared_ptr<void> &arg)
    {
        std::shared_ptr<pzmq_data> originalPtr = std::static_pointer_cast<pzmq_data>(arg);
        // data->get_param(0), data->get_param(1)
        std::string zmq_url = originalPtr->get_param(0);
        std::string data    = originalPtr->get_param(1);

        request_id_  = sample_json_str_get(data, "request_id");
        out_zmq_url_ = zmq_url;
        if (status_.load()) setup(zmq_url, data);
    };
    virtual int setup(const std::string &zmq_url, const std::string &raw);
    virtual int setup(const std::string &work_id, const std::string &object, const std::string &data);

    std::string _rpc_exit(pzmq *_pzmq, const std::shared_ptr<pzmq_data> &data);
    void _exit(const std::shared_ptr<void> &arg)
    {
        std::shared_ptr<pzmq_data> originalPtr = std::static_pointer_cast<pzmq_data>(arg);
        std::string zmq_url                    = originalPtr->get_param(0);
        std::string data                       = originalPtr->get_param(1);
        request_id_                            = sample_json_str_get(data, "request_id");
        out_zmq_url_                           = zmq_url;
        if (status_.load()) exit(zmq_url, data);
    }
    virtual int exit(const std::string &zmq_url, const std::string &raw);
    virtual int exit(const std::string &work_id, const std::string &object, const std::string &data);

    std::string _rpc_pause(pzmq *_pzmq, const std::shared_ptr<pzmq_data> &data);
    void _pause(const std::shared_ptr<void> &arg)
    {
        std::shared_ptr<pzmq_data> originalPtr = std::static_pointer_cast<pzmq_data>(arg);
        std::string zmq_url                    = originalPtr->get_param(0);
        std::string data                       = originalPtr->get_param(0);
        request_id_                            = sample_json_str_get(data, "request_id");
        out_zmq_url_                           = zmq_url;
        if (status_.load()) pause(zmq_url, data);
    }
    virtual void pause(const std::string &zmq_url, const std::string &raw);
    virtual void pause(const std::string &work_id, const std::string &object, const std::string &data);

    std::string _rpc_taskinfo(pzmq *_pzmq, const std::shared_ptr<pzmq_data> &data);
    void _taskinfo(const std::shared_ptr<void> &arg)
    {
        std::shared_ptr<pzmq_data> originalPtr = std::static_pointer_cast<pzmq_data>(arg);
        // data->get_param(0), data->get_param(1)
        std::string zmq_url = originalPtr->get_param(0);
        std::string data    = originalPtr->get_param(1);
        request_id_         = sample_json_str_get(data, "request_id");
        out_zmq_url_        = zmq_url;
        if (status_.load()) taskinfo(zmq_url, data);
    }
    virtual void taskinfo(const std::string &zmq_url, const std::string &raw);
    virtual void taskinfo(const std::string &work_id, const std::string &object, const std::string &data);

    int send(const std::string &object, const nlohmann::json &data, const std::string &error_msg,
             const std::string &work_id, const std::string &zmq_url = "")
    {
        nlohmann::json out_body;
        out_body["request_id"] = request_id_;
        out_body["work_id"]    = work_id;
        out_body["created"]    = time(NULL);
        out_body["object"]     = object;
        out_body["data"]       = data;
        if (error_msg.empty()) {
            out_body["error"]["code"]    = 0;
            out_body["error"]["message"] = "";
        } else
            out_body["error"] = error_msg;

        if (zmq_url.empty()) {
            pzmq _zmq(out_zmq_url_, ZMQ_PUSH);
            std::string out = out_body.dump();
            out += "\n";
            return _zmq.send_data(out);
        } else {
            pzmq _zmq(zmq_url, ZMQ_PUSH);
            std::string out = out_body.dump();
            out += "\n";
            return _zmq.send_data(out);
        }
    }

    std::string sys_sql_select(const std::string &key);
    void sys_sql_set(const std::string &key, const std::string &val);
    void sys_sql_unset(const std::string &key);
    int sys_register_unit(const std::string &unit_name);
    template <typename T>
    bool sys_release_unit(T workid)
    {
        std::string _work_id;
        int _work_id_num;
        if constexpr (std::is_same<T, int>::value) {
            _work_id     = sample_get_work_id(workid, unit_name_);
            _work_id_num = workid;
        } else if constexpr (std::is_same<T, std::string>::value) {
            _work_id     = workid;
            _work_id_num = sample_get_work_id_num(workid);
        } else {
            return false;
        }
        pzmq _call("sys");
        _call.call_rpc_action("release_unit", _work_id, [](pzmq *_pzmq, const std::shared_ptr<pzmq_data> &data) {});
        llm_task_channel_[_work_id_num].reset();
        llm_task_channel_.erase(_work_id_num);
        return false;
    }
    bool sys_release_unit(int work_id_num, const std::string &work_id);
    ~StackFlow();
};
};  // namespace StackFlows