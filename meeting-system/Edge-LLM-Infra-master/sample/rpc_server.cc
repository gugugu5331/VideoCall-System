#include <iostream>
#include "pzmq.hpp"
#include "pzmq_data.h"
#include <string>

using namespace StackFlows;

std::string fun1_(pzmq* self, const std::shared_ptr<pzmq_data>& msg) {
    std::string raw_msg = msg->string();
    std::cout << "fun1 received: " << raw_msg << std::endl;
    return "hello fun1";
}

std::string fun2_(pzmq* self, const std::shared_ptr<pzmq_data>& msg) {
    std::string raw_msg = msg->string();
    std::cout << "fun2 received: " << raw_msg << std::endl;
    return "hello fun2";
}

int main(int argc, char *argv[]) {
    pzmq _rpc("test");
    _rpc.register_rpc_action("fun1", fun1_);
    _rpc.register_rpc_action("fun2", fun2_);
    
    while(1) {
        sleep(1);
    }
    return 0;
}