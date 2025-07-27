#pragma once

#include <string>
#include <vector>
#include <chrono>
#include <memory>
#include <functional>

namespace ffmpeg_detection {

// 时间测量工具
class Timer {
public:
    Timer();
    void start();
    void stop();
    int64_t elapsed_ms() const;
    double elapsed_seconds() const;
    void reset();

private:
    std::chrono::high_resolution_clock::time_point start_time_;
    std::chrono::high_resolution_clock::time_point end_time_;
    bool is_running_;
};

// 日志级别
enum class LogLevel {
    DEBUG,
    INFO,
    WARNING,
    ERROR,
    FATAL
};

// 日志工具
class Logger {
public:
    static Logger& get_instance();
    
    void set_level(LogLevel level);
    void set_output_file(const std::string& filename);
    
    void debug(const std::string& message);
    void info(const std::string& message);
    void warning(const std::string& message);
    void error(const std::string& message);
    void fatal(const std::string& message);
    
    template<typename... Args>
    void debug(const std::string& format, Args... args) {
        log(LogLevel::DEBUG, format_message(format, args...));
    }
    
    template<typename... Args>
    void info(const std::string& format, Args... args) {
        log(LogLevel::INFO, format_message(format, args...));
    }
    
    template<typename... Args>
    void warning(const std::string& format, Args... args) {
        log(LogLevel::WARNING, format_message(format, args...));
    }
    
    template<typename... Args>
    void error(const std::string& format, Args... args) {
        log(LogLevel::ERROR, format_message(format, args...));
    }
    
    template<typename... Args>
    void fatal(const std::string& format, Args... args) {
        log(LogLevel::FATAL, format_message(format, args...));
    }

private:
    Logger();
    ~Logger();
    void log(LogLevel level, const std::string& message);
    std::string format_message(const std::string& format);
    template<typename T, typename... Args>
    std::string format_message(const std::string& format, T value, Args... args);
    
    LogLevel level_;
    std::string output_file_;
    std::mutex mutex_;
};

// 文件工具
class FileUtils {
public:
    static bool file_exists(const std::string& path);
    static bool directory_exists(const std::string& path);
    static bool create_directory(const std::string& path);
    static std::vector<std::string> list_files(const std::string& directory, const std::string& extension = "");
    static std::string get_file_extension(const std::string& filename);
    static std::string get_filename_without_extension(const std::string& filename);
    static std::string get_directory(const std::string& path);
    static int64_t get_file_size(const std::string& path);
    static bool copy_file(const std::string& src, const std::string& dst);
    static bool delete_file(const std::string& path);
    static std::vector<uint8_t> read_binary_file(const std::string& path);
    static bool write_binary_file(const std::string& path, const std::vector<uint8_t>& data);
    static std::string read_text_file(const std::string& path);
    static bool write_text_file(const std::string& path, const std::string& content);
};

// 字符串工具
class StringUtils {
public:
    static std::vector<std::string> split(const std::string& str, char delimiter);
    static std::string join(const std::vector<std::string>& strings, const std::string& delimiter);
    static std::string trim(const std::string& str);
    static std::string to_lower(const std::string& str);
    static std::string to_upper(const std::string& str);
    static bool starts_with(const std::string& str, const std::string& prefix);
    static bool ends_with(const std::string& str, const std::string& suffix);
    static std::string replace(const std::string& str, const std::string& from, const std::string& to);
    static bool contains(const std::string& str, const std::string& substr);
    static std::string format_bytes(int64_t bytes);
    static std::string format_duration(int64_t milliseconds);
    static std::string format_percentage(double value, int precision = 2);
};

// 数学工具
class MathUtils {
public:
    static double mean(const std::vector<double>& values);
    static double variance(const std::vector<double>& values);
    static double standard_deviation(const std::vector<double>& values);
    static double median(const std::vector<double>& values);
    static double min(const std::vector<double>& values);
    static double max(const std::vector<double>& values);
    static std::vector<double> normalize(const std::vector<double>& values);
    static std::vector<double> softmax(const std::vector<double>& values);
    static double sigmoid(double x);
    static double relu(double x);
    static double tanh(double x);
    static double clamp(double value, double min_val, double max_val);
    static int clamp(int value, int min_val, int max_val);
    static bool is_nan(double value);
    static bool is_inf(double value);
    static double round_to_precision(double value, int precision);
};

// 内存工具
class MemoryUtils {
public:
    static int64_t get_peak_memory_usage_mb();
    static int64_t get_current_memory_usage_mb();
    static int64_t get_available_memory_mb();
    static double get_memory_usage_percentage();
    static void print_memory_info();
    static bool check_memory_available(int64_t required_mb);
};

// 线程工具
class ThreadUtils {
public:
    static int get_cpu_count();
    static void set_thread_affinity(std::thread& thread, int cpu_id);
    static void set_thread_priority(std::thread& thread, int priority);
    static void sleep_ms(int milliseconds);
    static void sleep_us(int microseconds);
    static std::string get_thread_id();
};

// 配置工具
class ConfigUtils {
public:
    static bool load_config(const std::string& filename, std::unordered_map<std::string, std::string>& config);
    static bool save_config(const std::string& filename, const std::unordered_map<std::string, std::string>& config);
    static std::string get_config_value(const std::unordered_map<std::string, std::string>& config, 
                                       const std::string& key, const std::string& default_value = "");
    static int get_config_value_int(const std::unordered_map<std::string, std::string>& config, 
                                   const std::string& key, int default_value = 0);
    static double get_config_value_double(const std::unordered_map<std::string, std::string>& config, 
                                         const std::string& key, double default_value = 0.0);
    static bool get_config_value_bool(const std::unordered_map<std::string, std::string>& config, 
                                     const std::string& key, bool default_value = false);
};

// 性能监控
class PerformanceMonitor {
public:
    PerformanceMonitor(const std::string& name);
    ~PerformanceMonitor();
    
    void start();
    void stop();
    void reset();
    
    int64_t get_total_time_ms() const;
    int64_t get_average_time_ms() const;
    int64_t get_min_time_ms() const;
    int64_t get_max_time_ms() const;
    int64_t get_call_count() const;
    double get_throughput_fps() const;
    
    void print_stats() const;

private:
    std::string name_;
    std::chrono::high_resolution_clock::time_point start_time_;
    std::vector<int64_t> times_;
    mutable std::mutex mutex_;
};

// 宏定义
#define LOG_DEBUG(...) Logger::get_instance().debug(__VA_ARGS__)
#define LOG_INFO(...) Logger::get_instance().info(__VA_ARGS__)
#define LOG_WARNING(...) Logger::get_instance().warning(__VA_ARGS__)
#define LOG_ERROR(...) Logger::get_instance().error(__VA_ARGS__)
#define LOG_FATAL(...) Logger::get_instance().fatal(__VA_ARGS__)

#define PERF_MONITOR(name) PerformanceMonitor perf_monitor(name)
#define PERF_START() perf_monitor.start()
#define PERF_STOP() perf_monitor.stop()

} // namespace ffmpeg_detection 