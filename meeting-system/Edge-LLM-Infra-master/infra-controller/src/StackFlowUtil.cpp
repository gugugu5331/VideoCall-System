/*
 * SPDX-FileCopyrightText: 2024 M5Stack Technology CO LTD
 *
 * SPDX-License-Identifier: MIT
 */
#include "StackFlowUtil.h"
#include <vector>
#include <glob.h>
#include <fstream>
#include "pzmq.hpp"
#if defined(__ARM_NEON) || defined(__ARM_NEON__)
#include <arm_neon.h>
#endif

// #include <iconv.h>

std::string StackFlows::sample_json_str_get(const std::string &json_str, const std::string &json_key)
{
    std::string key_val;
    std::string format_val;
    // ALOGD("json_str: %s json_key:%s\n", json_str.c_str(), json_key.c_str());
    std::string find_key = "\"" + json_key + "\"";
    int subs_start       = json_str.find(find_key);
    if (subs_start == std::string::npos) {
        return key_val;
    }
    int status    = 0;
    char last_c   = '\0';
    int obj_flage = 0;
    for (auto c : json_str.substr(subs_start + find_key.length())) {
        switch (status) {
            case 0: {
                switch (c) {
                    case '"': {
                        status = 100;
                    } break;
                    case '{': {
                        key_val.push_back(c);
                        obj_flage = 1;
                        status    = 10;
                    } break;
                    case ':':
                        obj_flage = 1;
                        break;
                    case ',':
                    case '}': {
                        obj_flage = 0;
                        status    = -1;
                    } break;
                    case ' ':
                        break;
                    default: {
                        if (obj_flage) {
                            key_val.push_back(c);
                        }
                    } break;
                }
            } break;
            case 10: {
                key_val.push_back(c);
                if (c == '{') {
                    obj_flage++;
                }
                if (c == '}') {
                    obj_flage--;
                }
                if (obj_flage == 0) {
                    if (!key_val.empty()) {
                        status = -1;
                    }
                }
            } break;
            case 100: {
                if ((c == '"') && (last_c != '\\')) {
                    obj_flage = 0;
                    status    = -1;
                } else {
                    key_val.push_back(c);
                }
            } break;
            default:
                break;
        }
        last_c = c;
    }
    if (obj_flage != 0) {
        key_val.clear();
    }
    // ALOGD("key_val:%s\n", key_val.c_str());
    return key_val;
}

int StackFlows::sample_get_work_id_num(const std::string &work_id)
{
    int a = work_id.find(".");
    if ((a == std::string::npos) || (a == work_id.length() - 1)) {
        return WORK_ID_NONE;
    }
    return std::stoi(work_id.substr(a + 1));
}

std::string StackFlows::sample_get_work_id_name(const std::string &work_id)
{
    int a = work_id.find(".");
    if (a == std::string::npos) {
        return work_id;
    } else {
        return work_id.substr(0, a);
    }
}

std::string StackFlows::sample_get_work_id(int work_id_num, const std::string &unit_name)
{
    return unit_name + "." + std::to_string(work_id_num);
}

void StackFlows::unicode_to_utf8(unsigned int codepoint, char *output, int *length) {
    if (codepoint <= 0x7F) {
        output[0] = codepoint & 0x7F;
        *length = 1;
    } else if (codepoint <= 0x7FF) {
        output[0] = 0xC0 | ((codepoint >> 6) & 0x1F);
        output[1] = 0x80 | (codepoint & 0x3F);
        *length = 2;
    } else if (codepoint <= 0xFFFF) {
        output[0] = 0xE0 | ((codepoint >> 12) & 0x0F);
        output[1] = 0x80 | ((codepoint >> 6) & 0x3F);
        output[2] = 0x80 | (codepoint & 0x3F);
        *length = 3;
    } else if (codepoint <= 0x10FFFF) {
        output[0] = 0xF0 | ((codepoint >> 18) & 0x07);
        output[1] = 0x80 | ((codepoint >> 12) & 0x3F);
        output[2] = 0x80 | ((codepoint >> 6) & 0x3F);
        output[3] = 0x80 | (codepoint & 0x3F);
        *length = 4;
    } else {
        *length = 0;
    }
}

// clang-format on
bool StackFlows::decode_stream(const std::string &in, std::string &out,
                               std::unordered_map<int, std::string> &stream_buff)
{
    int index          = std::stoi(StackFlows::sample_json_str_get(in, "index"));
    std::string finish = StackFlows::sample_json_str_get(in, "finish");
    stream_buff[index] = StackFlows::sample_json_str_get(in, "delta");
    // sample find flage: false:true
    if (finish.find("f") == std::string::npos) {
        for (size_t i = 0; i < stream_buff.size(); i++) {
            out += stream_buff.at(i);
        }
        stream_buff.clear();
        return false;
    }
    return true;
}

std::string StackFlows::unit_call(const std::string &unit_name, const std::string &unit_action, const std::string &data)
{
    std::string value;
    pzmq _call(unit_name);
    _call.call_rpc_action(unit_action, data, [&value](pzmq *_pzmq, const std::shared_ptr<pzmq_data> &raw) { value = raw->string(); });
    return value;
}

void StackFlows::unit_call(const std::string &unit_name, const std::string &unit_action, const std::string &data, std::function<void(const std::shared_ptr<StackFlows::pzmq_data> &)> callback)
{
    std::string value;
    StackFlows::pzmq _call(unit_name);
    _call.call_rpc_action(unit_action, data, [callback](StackFlows::pzmq *_pzmq, const std::shared_ptr<StackFlows::pzmq_data> &raw) { callback(raw); });
}

bool StackFlows::file_exists(const std::string &filePath)
{
    std::ifstream file(filePath);
    return file.good();
}