#include "pzmq.hpp"
#include "pzmq_data.h"
#include <iostream>
#include <string>
#include <unistd.h>

using namespace StackFlows;

void message_handler(pzmq* zmq_obj, const std::shared_ptr<pzmq_data>& data)
{
    std::cout << "Received: " << data->string() << std::endl;
}

int main(int argc, char* argv[])
{
    try {
        // 1. 创建SUB套接字
        pzmq zsub_("ipc:///tmp/5001.socket", ZMQ_SUB, message_handler);
        std::cout << "Subscriber started. Waiting for messages..." << std::endl;
        while (true) {
            sleep(1);
        }

    } catch (const std::exception& e) {
        std::cerr << "Error: " << e.what() << std::endl;
        return 1;
    }
    return 0;
}