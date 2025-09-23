/*
 * SPDX-FileCopyrightText: 2024 M5Stack Technology CO LTD
 *
 * SPDX-License-Identifier: MIT
 */
#pragma once
#include <pthread.h>
#include "sample_log.h"

#include <map>
#include <string>
#include <memory>
#include <stdexcept>
#include <utility>

#include <any>
#include <unordered_map>

extern std::unordered_map<std::string, std::any> key_sql;
extern pthread_spinlock_t key_sql_lock;

#define SAFE_READING(_val, _type, _key)               \
    do {                                              \
        pthread_spin_lock(&key_sql_lock);             \
        try {                                         \
            _val = std::any_cast<_type>(key_sql.at(_key)); \
        } catch (...) {                               \
        }                                             \
        pthread_spin_unlock(&key_sql_lock);           \
    } while (0)

#define SAFE_SETTING(_key, _val)            \
    do {                                    \
        pthread_spin_lock(&key_sql_lock);   \
        try {                               \
            key_sql[_key] = _val;           \
        } catch (...) {                     \
        }                                   \
        pthread_spin_unlock(&key_sql_lock); \
    } while (0)

#define SAFE_ERASE(_key)                    \
    do {                                    \
        pthread_spin_lock(&key_sql_lock);   \
        try {                               \
            key_sql.erase(_key);            \
        } catch (...) {                     \
        }                                   \
        pthread_spin_unlock(&key_sql_lock); \
    } while (0)

void load_default_config();
void unit_action_match(int com_id, const std::string &json_str);

extern std::string zmq_s_format;
extern std::string zmq_c_format;
extern int main_exit_flage;
