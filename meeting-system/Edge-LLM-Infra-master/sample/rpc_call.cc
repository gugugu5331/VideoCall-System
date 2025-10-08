#include <iostream>
#include "pzmq.hpp"
#include "pzmq_data.h"
#include <string>

using namespace StackFlows;

int main(int argc, char* argv[])
{
    pzmq _rpc("test");

    _rpc.call_rpc_action("fun1", "call fun1_", [](pzmq* self, const std::shared_ptr<pzmq_data>& msg) {
        std::cout << "Response from fun1: " << msg->string() << std::endl;
    });

    _rpc.call_rpc_action("fun2", "call fun2_", [](pzmq* self, const std::shared_ptr<pzmq_data>& msg) {
        std::cout << "Response from fun2: " << msg->string() << std::endl;
    });

    return 0;
}