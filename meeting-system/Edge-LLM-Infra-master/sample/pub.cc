#include "pzmq.hpp"
#include "pzmq_data.h"
#include <iostream>
#include <string>
#include <unistd.h>  // for sleep()

using namespace StackFlows;

int main(int argc, char* argv[])
{
    try {
        // 1. 创建PUB套接字
        pzmq zpush_("ipc:///tmp/5001.socket", ZMQ_PUB);

        // 2. 等待连接建立（PUB-SUB模式需要）
        std::cout << "Publisher started. Waiting for subscribers..." << std::endl;
        sleep(1);  // 重要：给订阅者时间连接

        // 3. 发送消息
        int count = 0;
        while (true) {
            std::string msg = "Message " + std::to_string(++count);
            zpush_.send_data(msg);
            std::cout << "Sent: " << msg << std::endl;
            sleep(1);  // 每秒发送一条
        }

    } catch (const std::exception& e) {
        std::cerr << "Error: " << e.what() << std::endl;
        return 1;
    }
    return 0;
}