/*
 * SPDX-FileCopyrightText: 2024 M5Stack Technology CO LTD
 *
 * SPDX-License-Identifier: MIT
 */
#pragma once
#include <string>
#include <cstring>
#include <unordered_map>
#include <list>
#include <vector>
#include <functional>
#include "pzmq.hpp"
#include <memory>
#define WORK_ID_NONE -100

namespace StackFlows {
std::string sample_json_str_get(const std::string &json_str, const std::string &json_key);
int sample_get_work_id_num(const std::string &work_id);
std::string sample_get_work_id_name(const std::string &work_id);
std::string sample_get_work_id(int work_id_num, const std::string &unit_name);
bool decode_stream(const std::string &in, std::string &out, std::unordered_map<int, std::string> &stream_buff);

std::string unit_call(const std::string &unit_name, const std::string &unit_action, const std::string &data);
void unit_call(const std::string &unit_name, const std::string &unit_action, const std::string &data, std::function<void(const std::shared_ptr<StackFlows::pzmq_data> &)> callback);

bool file_exists(const std::string& filePath);
void unicode_to_utf8(unsigned int codepoint, char *output, int *length);
};  // namespace StackFlows
