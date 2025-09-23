// 删除原有hv头文件，添加：
#pragma once

#include "network/TcpServer.h"
#include "network/EventLoop.h"
#include <any>
#include <atomic>

#include "zmq_bus.h"
#include "network/TcpConnection.h"

class TcpSession : public zmq_bus_com
{
public:
    explicit TcpSession(const network::TcpConnectionPtr &conn)
        : conn_(conn) {}

    void send_data(const std::string &data) override
    {
        printf("zmq_bus_com::send_data : send:%s\n", data.c_str());
        network::Buffer *buf = new network::Buffer;
        buf->append(data.c_str(), data.size());
        conn_->send(buf); 
    }

    network::TcpConnectionPtr conn_;
};