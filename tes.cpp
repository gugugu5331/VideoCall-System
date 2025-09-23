#include <thread>
#include <mutex>
#include <condition_variable>
#include <queue>
#include <List>
#include <functional>
#include <optional>
#include <memory>
#include <stdexcept>

template<typename K, typename V>
class ConcurrentCache {
public：
    explicit ConcurrentCache(size_t capacit):capacit_(capacity){
        if(capacity == 0){
            throw std:invalid_argument("Capacity must be  greater than 0");
        }
    }

    std::optionanl<Value> get(const Key& key){
        std::lock_guard<std::mutex> lock(mutex_);

        auto it = cache_map_.find(key);
        if(it ==  cache_map_.end()){
           return std::nullopt;
        }
        lru_list_.splice(lru_list_.begin(), lru_list_,it->second.second);
        return it->second.first;
    }

    
    //插入（拷贝语义）
    void put(const Key& key, const Value& value){
        std::lock_guard<std::mutex> lock(mutex_);

        auto it = cache_map_.find(key);
        if(it != cache_map_.end()){
            it -> second.first = value;
            lru_list_..splice(lru_list_.begin(), lru_list_, it->second.second);
            return ;
        }
        if(cache_map_.size() >= capacity_){
            auto lru_key = lru_list_.back();
            cache_map_.erase(lru_key);
            lru_list_.pop_back();

        }
        lru_list_.push_front(key);
        cache_map_[key] = {value, lru_list_.begin()};
    }

    //插入（移动语义）
    void put(Key&& key, Value&& value){
        std::lock_guard<std::mutex> lock(mutex_);

        auto it = cache_map_.find(key);
        if(it != cache_map_.end()){
            it -> second.first = value;
            lru_list_.splice(lru_list_.begin(), lru_list_, it->second.second);
            return ;
        }
        if(cache_map_.size() >= capacity_){
            auto lru_key = lru_list_.back();
            cache_map_.erase(lru_key);
            lru_list_.pop_back();

        }
        lru_list_.push_front(std::move(key));
        cache_map_[lru_list_.front()] = {std::move(value), lru_list_.begin()};
    }
    
private:
    size_t capacity_;
    std::unordered_map<key, std::pair<Value, typename std::list<Key>::iterator>> cache_map_;
    std::list<Key> lru_list;
    mutable std::mutex mutex_;
};