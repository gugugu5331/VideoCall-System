#include "unit_data.h"
unit_data::unit_data()
{
}

void unit_data::init_zmq(const std::string &url)
{
    inference_url = url;
    user_inference_chennal_ = std::make_unique<pzmq>(inference_url, ZMQ_PUB);
}

void unit_data::send_msg(const std::string &json_str)
{
    user_inference_chennal_->send_data(json_str);
}

unit_data::~unit_data()
{
    user_inference_chennal_.reset();
}