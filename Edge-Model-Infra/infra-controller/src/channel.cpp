#include "channel.h"
#include "sample_log.h"
#include <iostream>

using namespace StackFlows;

llm_channel_obj::llm_channel_obj(const std::string &_publisher_url, const std::string &inference_url,
                                 const std::string &unit_name)
    : unit_name_(unit_name), inference_url_(inference_url), publisher_url_(_publisher_url)
{
    // publisher_url_ = _publisher_url;
    zmq_url_index_ = -1000;
    zmq_[-1] = std::make_shared<pzmq>(publisher_url_, ZMQ_PUB);
    zmq_[-2].reset();
}

llm_channel_obj::~llm_channel_obj()
{
    std::cout << "llm_channel_obj 析构" << std::endl;
}

void llm_channel_obj::subscriber_event_call(const std::function<void(const std::string &, const std::string &)> &call,
                                            pzmq *_pzmq, const std::shared_ptr<pzmq_data> &raw)
{
    auto _raw = raw->string();
    const char *user_inference_flage_str = "\"action\"";
    std::size_t pos = _raw.find(user_inference_flage_str);
    while (true)
    {
        if (pos == std::string::npos)
        {
            break;
        }
        else if ((pos > 0) && (_raw[pos - 1] != '\\'))
        {
            std::string zmq_com = sample_json_str_get(_raw, "zmq_com");
            if (!zmq_com.empty())
                set_push_url(zmq_com);
            request_id_ = sample_json_str_get(_raw, "request_id");
            work_id_ = sample_json_str_get(_raw, "work_id");
            break;
        }
        pos = _raw.find(user_inference_flage_str, pos + sizeof(user_inference_flage_str));
    }
    call(sample_json_str_get(_raw, "object"), sample_json_str_get(_raw, "data"));
}
void message_handler(pzmq *zmq_obj, const std::shared_ptr<pzmq_data> &data)
{
    std::cout << "Received: " << data->string() << std::endl;
}

int llm_channel_obj::subscriber_work_id(const std::string &work_id,
                                        const std::function<void(const std::string &, const std::string &)> &call)
{
    int id_num;
    std::string subscriber_url;
    std::regex pattern(R"((\w+)\.(\d+))");
    std::smatch matches;
    if ((!work_id.empty()) && std::regex_match(work_id, matches, pattern))
    {
        if (matches.size() == 3)
        {
            // std::string part1 = matches[1].str();
            id_num = std::stoi(matches[2].str());
            std::string input_url_name = work_id + ".out_port";
            std::string input_url = unit_call("sys", "sql_select", input_url_name);
            if (input_url.empty())
            {
                return -1;
            }
            subscriber_url = input_url;
        }
    }
    else
    {
        id_num = 0;
        subscriber_url = inference_url_;
    }

    zmq_[id_num] = std::make_shared<pzmq>(
        subscriber_url, ZMQ_SUB,
        std::bind(&llm_channel_obj::subscriber_event_call, this, call, std::placeholders::_1, std::placeholders::_2));
    return 0;
}

void llm_channel_obj::stop_subscriber_work_id(const std::string &work_id)
{
    int id_num;
    std::regex pattern(R"((\w+)\.(\d+))");
    std::smatch matches;
    if (std::regex_match(work_id, matches, pattern))
    {
        if (matches.size() == 3)
        {
            // std::string part1 = matches[1].str();
            id_num = std::stoi(matches[2].str());
        }
    }
    else
    {
        id_num = 0;
    }
    if (zmq_.find(id_num) != zmq_.end())
        zmq_.erase(id_num);
}

void llm_channel_obj::subscriber(const std::string &zmq_url, const pzmq::msg_callback_fun &call)
{
    zmq_url_map_[zmq_url] = zmq_url_index_--;
    zmq_[zmq_url_map_[zmq_url]] = std::make_shared<pzmq>(zmq_url, ZMQ_SUB, call);
}

void llm_channel_obj::stop_subscriber(const std::string &zmq_url)
{
    if (zmq_url.empty())
    {
        zmq_.clear();
        zmq_url_map_.clear();
    }
    else if (zmq_url_map_.find(zmq_url) != zmq_url_map_.end())
    {
        zmq_.erase(zmq_url_map_[zmq_url]);
        zmq_url_map_.erase(zmq_url);
    }
}

int llm_channel_obj::send_raw_to_pub(const std::string &raw)
{
    return zmq_[-1]->send_data(raw);
}

int llm_channel_obj::send_raw_to_usr(const std::string &raw)
{
    if (zmq_[-2])
    {
        return zmq_[-2]->send_data(raw);
    }
    else
    {
        return -1;
    }
}

void llm_channel_obj::set_push_url(const std::string &url)
{
    if (output_url_ != url)
    {
        output_url_ = url;
        zmq_[-2].reset(new pzmq(output_url_, ZMQ_PUSH));
    }
}

void llm_channel_obj::cear_push_url()
{
    zmq_[-2].reset();
}

int llm_channel_obj::send_raw_for_url(const std::string &zmq_url, const std::string &raw)
{
    pzmq _zmq(zmq_url, ZMQ_PUSH);
    return _zmq.send_data(raw);
}
