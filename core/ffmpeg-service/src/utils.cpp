#include "utils.h"
#include <iostream>
#include <fstream>
#include <sstream>
#include <algorithm>
#include <cctype>
#include <filesystem>
#include <thread>
#include <chrono>
#include <iomanip>
#include <cstring>

#ifdef _WIN32
#include <windows.h>
#include <psapi.h>
#else
#include <unistd.h>
#include <sys/resource.h>
#include <sys/sysinfo.h>
#endif

namespace ffmpeg_detection {

// Timer 实现
Timer::Timer() : is_running_(false) {}

void Timer::start() {
    start_time_ = std::chrono::high_resolution_clock::now();
    is_running_ = true;
}

void Timer::stop() {
    if (is_running_) {
        end_time_ = std::chrono::high_resolution_clock::now();
        is_running_ = false;
    }
}

int64_t Timer::elapsed_ms() const {
    auto end = is_running_ ? std::chrono::high_resolution_clock::now() : end_time_;
    return std::chrono::duration_cast<std::chrono::milliseconds>(end - start_time_).count();
}

double Timer::elapsed_seconds() const {
    auto end = is_running_ ? std::chrono::high_resolution_clock::now() : end_time_;
    return std::chrono::duration_cast<std::chrono::duration<double>>(end - start_time_).count();
}

void Timer::reset() {
    is_running_ = false;
}

// Logger 实现
Logger::Logger() : level_(LogLevel::INFO) {}

Logger::~Logger() {
    if (!output_file_.empty()) {
        // 关闭输出文件
    }
}

Logger& Logger::get_instance() {
    static Logger instance;
    return instance;
}

void Logger::set_level(LogLevel level) {
    level_ = level;
}

void Logger::set_output_file(const std::string& filename) {
    output_file_ = filename;
}

void Logger::debug(const std::string& message) {
    log(LogLevel::DEBUG, message);
}

void Logger::info(const std::string& message) {
    log(LogLevel::INFO, message);
}

void Logger::warning(const std::string& message) {
    log(LogLevel::WARNING, message);
}

void Logger::error(const std::string& message) {
    log(LogLevel::ERROR, message);
}

void Logger::fatal(const std::string& message) {
    log(LogLevel::FATAL, message);
}

void Logger::log(LogLevel level, const std::string& message) {
    if (level < level_) {
        return;
    }
    
    std::lock_guard<std::mutex> lock(mutex_);
    
    auto now = std::chrono::system_clock::now();
    auto time_t = std::chrono::system_clock::to_time_t(now);
    auto ms = std::chrono::duration_cast<std::chrono::milliseconds>(now.time_since_epoch()) % 1000;
    
    std::stringstream ss;
    ss << std::put_time(std::localtime(&time_t), "%Y-%m-%d %H:%M:%S");
    ss << "." << std::setfill('0') << std::setw(3) << ms.count() << " ";
    
    switch (level) {
        case LogLevel::DEBUG:   ss << "[DEBUG] "; break;
        case LogLevel::INFO:    ss << "[INFO] "; break;
        case LogLevel::WARNING: ss << "[WARN] "; break;
        case LogLevel::ERROR:   ss << "[ERROR] "; break;
        case LogLevel::FATAL:   ss << "[FATAL] "; break;
    }
    
    ss << message << std::endl;
    
    // 输出到控制台
    std::cout << ss.str();
    
    // 输出到文件
    if (!output_file_.empty()) {
        std::ofstream file(output_file_, std::ios::app);
        if (file.is_open()) {
            file << ss.str();
            file.close();
        }
    }
}

std::string Logger::format_message(const std::string& format) {
    return format;
}

template<typename T, typename... Args>
std::string Logger::format_message(const std::string& format, T value, Args... args) {
    size_t pos = format.find("{}");
    if (pos == std::string::npos) {
        return format;
    }
    
    std::stringstream ss;
    ss << value;
    
    std::string result = format.substr(0, pos) + ss.str() + format.substr(pos + 2);
    return format_message(result, args...);
}

// FileUtils 实现
bool FileUtils::file_exists(const std::string& path) {
    return std::filesystem::exists(path);
}

bool FileUtils::directory_exists(const std::string& path) {
    return std::filesystem::exists(path) && std::filesystem::is_directory(path);
}

bool FileUtils::create_directory(const std::string& path) {
    try {
        return std::filesystem::create_directories(path);
    } catch (const std::exception& e) {
        LOG_ERROR("创建目录失败: %s", e.what());
        return false;
    }
}

std::vector<std::string> FileUtils::list_files(const std::string& directory, const std::string& extension) {
    std::vector<std::string> files;
    
    try {
        for (const auto& entry : std::filesystem::directory_iterator(directory)) {
            if (entry.is_regular_file()) {
                std::string filename = entry.path().filename().string();
                if (extension.empty() || StringUtils::ends_with(filename, extension)) {
                    files.push_back(entry.path().string());
                }
            }
        }
    } catch (const std::exception& e) {
        LOG_ERROR("列出文件失败: %s", e.what());
    }
    
    return files;
}

std::string FileUtils::get_file_extension(const std::string& filename) {
    size_t pos = filename.find_last_of('.');
    if (pos != std::string::npos) {
        return filename.substr(pos + 1);
    }
    return "";
}

std::string FileUtils::get_filename_without_extension(const std::string& filename) {
    size_t pos = filename.find_last_of('.');
    if (pos != std::string::npos) {
        return filename.substr(0, pos);
    }
    return filename;
}

std::string FileUtils::get_directory(const std::string& path) {
    return std::filesystem::path(path).parent_path().string();
}

int64_t FileUtils::get_file_size(const std::string& path) {
    try {
        return std::filesystem::file_size(path);
    } catch (const std::exception& e) {
        LOG_ERROR("获取文件大小失败: %s", e.what());
        return -1;
    }
}

bool FileUtils::copy_file(const std::string& src, const std::string& dst) {
    try {
        std::filesystem::copy_file(src, dst, std::filesystem::copy_options::overwrite_existing);
        return true;
    } catch (const std::exception& e) {
        LOG_ERROR("复制文件失败: %s", e.what());
        return false;
    }
}

bool FileUtils::delete_file(const std::string& path) {
    try {
        return std::filesystem::remove(path);
    } catch (const std::exception& e) {
        LOG_ERROR("删除文件失败: %s", e.what());
        return false;
    }
}

std::vector<uint8_t> FileUtils::read_binary_file(const std::string& path) {
    std::vector<uint8_t> data;
    
    try {
        std::ifstream file(path, std::ios::binary);
        if (file.is_open()) {
            file.seekg(0, std::ios::end);
            size_t size = file.tellg();
            file.seekg(0, std::ios::beg);
            
            data.resize(size);
            file.read(reinterpret_cast<char*>(data.data()), size);
            file.close();
        }
    } catch (const std::exception& e) {
        LOG_ERROR("读取二进制文件失败: %s", e.what());
    }
    
    return data;
}

bool FileUtils::write_binary_file(const std::string& path, const std::vector<uint8_t>& data) {
    try {
        std::ofstream file(path, std::ios::binary);
        if (file.is_open()) {
            file.write(reinterpret_cast<const char*>(data.data()), data.size());
            file.close();
            return true;
        }
    } catch (const std::exception& e) {
        LOG_ERROR("写入二进制文件失败: %s", e.what());
    }
    
    return false;
}

std::string FileUtils::read_text_file(const std::string& path) {
    std::string content;
    
    try {
        std::ifstream file(path);
        if (file.is_open()) {
            std::stringstream buffer;
            buffer << file.rdbuf();
            content = buffer.str();
            file.close();
        }
    } catch (const std::exception& e) {
        LOG_ERROR("读取文本文件失败: %s", e.what());
    }
    
    return content;
}

bool FileUtils::write_text_file(const std::string& path, const std::string& content) {
    try {
        std::ofstream file(path);
        if (file.is_open()) {
            file << content;
            file.close();
            return true;
        }
    } catch (const std::exception& e) {
        LOG_ERROR("写入文本文件失败: %s", e.what());
    }
    
    return false;
}

// StringUtils 实现
std::vector<std::string> StringUtils::split(const std::string& str, char delimiter) {
    std::vector<std::string> tokens;
    std::stringstream ss(str);
    std::string token;
    
    while (std::getline(ss, token, delimiter)) {
        tokens.push_back(token);
    }
    
    return tokens;
}

std::string StringUtils::join(const std::vector<std::string>& strings, const std::string& delimiter) {
    if (strings.empty()) {
        return "";
    }
    
    std::string result = strings[0];
    for (size_t i = 1; i < strings.size(); i++) {
        result += delimiter + strings[i];
    }
    
    return result;
}

std::string StringUtils::trim(const std::string& str) {
    size_t start = str.find_first_not_of(" \t\n\r");
    if (start == std::string::npos) {
        return "";
    }
    
    size_t end = str.find_last_not_of(" \t\n\r");
    return str.substr(start, end - start + 1);
}

std::string StringUtils::to_lower(const std::string& str) {
    std::string result = str;
    std::transform(result.begin(), result.end(), result.begin(), ::tolower);
    return result;
}

std::string StringUtils::to_upper(const std::string& str) {
    std::string result = str;
    std::transform(result.begin(), result.end(), result.begin(), ::toupper);
    return result;
}

bool StringUtils::starts_with(const std::string& str, const std::string& prefix) {
    if (str.length() < prefix.length()) {
        return false;
    }
    return str.compare(0, prefix.length(), prefix) == 0;
}

bool StringUtils::ends_with(const std::string& str, const std::string& suffix) {
    if (str.length() < suffix.length()) {
        return false;
    }
    return str.compare(str.length() - suffix.length(), suffix.length(), suffix) == 0;
}

std::string StringUtils::replace(const std::string& str, const std::string& from, const std::string& to) {
    std::string result = str;
    size_t pos = 0;
    while ((pos = result.find(from, pos)) != std::string::npos) {
        result.replace(pos, from.length(), to);
        pos += to.length();
    }
    return result;
}

bool StringUtils::contains(const std::string& str, const std::string& substr) {
    return str.find(substr) != std::string::npos;
}

std::string StringUtils::format_bytes(int64_t bytes) {
    const char* units[] = {"B", "KB", "MB", "GB", "TB"};
    int unit_index = 0;
    double size = static_cast<double>(bytes);
    
    while (size >= 1024.0 && unit_index < 4) {
        size /= 1024.0;
        unit_index++;
    }
    
    std::stringstream ss;
    ss << std::fixed << std::setprecision(2) << size << " " << units[unit_index];
    return ss.str();
}

std::string StringUtils::format_duration(int64_t milliseconds) {
    int64_t seconds = milliseconds / 1000;
    int64_t minutes = seconds / 60;
    int64_t hours = minutes / 60;
    
    seconds %= 60;
    minutes %= 60;
    
    std::stringstream ss;
    if (hours > 0) {
        ss << hours << "h " << minutes << "m " << seconds << "s";
    } else if (minutes > 0) {
        ss << minutes << "m " << seconds << "s";
    } else {
        ss << seconds << "s";
    }
    
    return ss.str();
}

std::string StringUtils::format_percentage(double value, int precision) {
    std::stringstream ss;
    ss << std::fixed << std::setprecision(precision) << (value * 100.0) << "%";
    return ss.str();
}

// MathUtils 实现
double MathUtils::mean(const std::vector<double>& values) {
    if (values.empty()) {
        return 0.0;
    }
    
    double sum = 0.0;
    for (double value : values) {
        sum += value;
    }
    return sum / values.size();
}

double MathUtils::variance(const std::vector<double>& values) {
    if (values.empty()) {
        return 0.0;
    }
    
    double avg = mean(values);
    double sum_sq_diff = 0.0;
    
    for (double value : values) {
        double diff = value - avg;
        sum_sq_diff += diff * diff;
    }
    
    return sum_sq_diff / values.size();
}

double MathUtils::standard_deviation(const std::vector<double>& values) {
    return std::sqrt(variance(values));
}

double MathUtils::median(const std::vector<double>& values) {
    if (values.empty()) {
        return 0.0;
    }
    
    std::vector<double> sorted_values = values;
    std::sort(sorted_values.begin(), sorted_values.end());
    
    size_t size = sorted_values.size();
    if (size % 2 == 0) {
        return (sorted_values[size / 2 - 1] + sorted_values[size / 2]) / 2.0;
    } else {
        return sorted_values[size / 2];
    }
}

double MathUtils::min(const std::vector<double>& values) {
    if (values.empty()) {
        return 0.0;
    }
    return *std::min_element(values.begin(), values.end());
}

double MathUtils::max(const std::vector<double>& values) {
    if (values.empty()) {
        return 0.0;
    }
    return *std::max_element(values.begin(), values.end());
}

std::vector<double> MathUtils::normalize(const std::vector<double>& values) {
    if (values.empty()) {
        return values;
    }
    
    double min_val = min(values);
    double max_val = max(values);
    double range = max_val - min_val;
    
    if (range == 0.0) {
        return std::vector<double>(values.size(), 0.5);
    }
    
    std::vector<double> normalized;
    normalized.reserve(values.size());
    
    for (double value : values) {
        normalized.push_back((value - min_val) / range);
    }
    
    return normalized;
}

std::vector<double> MathUtils::softmax(const std::vector<double>& values) {
    if (values.empty()) {
        return values;
    }
    
    // 找到最大值以避免数值溢出
    double max_val = max(values);
    
    std::vector<double> exp_values;
    exp_values.reserve(values.size());
    
    double sum_exp = 0.0;
    for (double value : values) {
        double exp_val = std::exp(value - max_val);
        exp_values.push_back(exp_val);
        sum_exp += exp_val;
    }
    
    std::vector<double> softmax_values;
    softmax_values.reserve(values.size());
    
    for (double exp_val : exp_values) {
        softmax_values.push_back(exp_val / sum_exp);
    }
    
    return softmax_values;
}

double MathUtils::sigmoid(double x) {
    return 1.0 / (1.0 + std::exp(-x));
}

double MathUtils::relu(double x) {
    return std::max(0.0, x);
}

double MathUtils::tanh(double x) {
    return std::tanh(x);
}

double MathUtils::clamp(double value, double min_val, double max_val) {
    return std::max(min_val, std::min(max_val, value));
}

int MathUtils::clamp(int value, int min_val, int max_val) {
    return std::max(min_val, std::min(max_val, value));
}

bool MathUtils::is_nan(double value) {
    return std::isnan(value);
}

bool MathUtils::is_inf(double value) {
    return std::isinf(value);
}

double MathUtils::round_to_precision(double value, int precision) {
    double factor = std::pow(10.0, precision);
    return std::round(value * factor) / factor;
}

// MemoryUtils 实现
int64_t MemoryUtils::get_peak_memory_usage_mb() {
#ifdef _WIN32
    PROCESS_MEMORY_COUNTERS_EX pmc;
    if (GetProcessMemoryInfo(GetCurrentProcess(), (PROCESS_MEMORY_COUNTERS*)&pmc, sizeof(pmc))) {
        return pmc.PeakWorkingSetSize / (1024 * 1024);
    }
#else
    struct rusage rusage;
    if (getrusage(RUSAGE_SELF, &rusage) == 0) {
        return rusage.ru_maxrss / 1024; // Linux returns KB
    }
#endif
    return 0;
}

int64_t MemoryUtils::get_current_memory_usage_mb() {
#ifdef _WIN32
    PROCESS_MEMORY_COUNTERS_EX pmc;
    if (GetProcessMemoryInfo(GetCurrentProcess(), (PROCESS_MEMORY_COUNTERS*)&pmc, sizeof(pmc))) {
        return pmc.WorkingSetSize / (1024 * 1024);
    }
#else
    FILE* file = fopen("/proc/self/status", "r");
    if (file) {
        char line[128];
        while (fgets(line, 128, file) != NULL) {
            if (strncmp(line, "VmRSS:", 6) == 0) {
                int rss;
                sscanf(line, "VmRSS: %d", &rss);
                fclose(file);
                return rss / 1024; // Convert KB to MB
            }
        }
        fclose(file);
    }
#endif
    return 0;
}

int64_t MemoryUtils::get_available_memory_mb() {
#ifdef _WIN32
    MEMORYSTATUSEX memInfo;
    memInfo.dwLength = sizeof(MEMORYSTATUSEX);
    if (GlobalMemoryStatusEx(&memInfo)) {
        return memInfo.ullAvailPhys / (1024 * 1024);
    }
#else
    struct sysinfo si;
    if (sysinfo(&si) == 0) {
        return (si.freeram * si.mem_unit) / (1024 * 1024);
    }
#endif
    return 0;
}

double MemoryUtils::get_memory_usage_percentage() {
    int64_t total = get_available_memory_mb() + get_current_memory_usage_mb();
    if (total == 0) {
        return 0.0;
    }
    return static_cast<double>(get_current_memory_usage_mb()) / total * 100.0;
}

void MemoryUtils::print_memory_info() {
    LOG_INFO("内存使用情况 - 当前: %s, 峰值: %s, 可用: %s, 使用率: %.1f%%",
             StringUtils::format_bytes(get_current_memory_usage_mb() * 1024 * 1024).c_str(),
             StringUtils::format_bytes(get_peak_memory_usage_mb() * 1024 * 1024).c_str(),
             StringUtils::format_bytes(get_available_memory_mb() * 1024 * 1024).c_str(),
             get_memory_usage_percentage());
}

bool MemoryUtils::check_memory_available(int64_t required_mb) {
    return get_available_memory_mb() >= required_mb;
}

// ThreadUtils 实现
int ThreadUtils::get_cpu_count() {
    return std::thread::hardware_concurrency();
}

void ThreadUtils::set_thread_affinity(std::thread& thread, int cpu_id) {
    // 平台特定的线程亲和性设置
    // 这里简化处理
}

void ThreadUtils::set_thread_priority(std::thread& thread, int priority) {
    // 平台特定的线程优先级设置
    // 这里简化处理
}

void ThreadUtils::sleep_ms(int milliseconds) {
    std::this_thread::sleep_for(std::chrono::milliseconds(milliseconds));
}

void ThreadUtils::sleep_us(int microseconds) {
    std::this_thread::sleep_for(std::chrono::microseconds(microseconds));
}

std::string ThreadUtils::get_thread_id() {
    std::stringstream ss;
    ss << std::this_thread::get_id();
    return ss.str();
}

// ConfigUtils 实现
bool ConfigUtils::load_config(const std::string& filename, std::unordered_map<std::string, std::string>& config) {
    std::string content = FileUtils::read_text_file(filename);
    if (content.empty()) {
        return false;
    }
    
    std::stringstream ss(content);
    std::string line;
    
    while (std::getline(ss, line)) {
        line = StringUtils::trim(line);
        
        // 跳过注释和空行
        if (line.empty() || line[0] == '#' || line[0] == ';') {
            continue;
        }
        
        size_t pos = line.find('=');
        if (pos != std::string::npos) {
            std::string key = StringUtils::trim(line.substr(0, pos));
            std::string value = StringUtils::trim(line.substr(pos + 1));
            
            // 移除引号
            if (value.length() >= 2 && value[0] == '"' && value[value.length() - 1] == '"') {
                value = value.substr(1, value.length() - 2);
            }
            
            config[key] = value;
        }
    }
    
    return true;
}

bool ConfigUtils::save_config(const std::string& filename, const std::unordered_map<std::string, std::string>& config) {
    std::stringstream ss;
    
    for (const auto& pair : config) {
        ss << pair.first << " = " << pair.second << std::endl;
    }
    
    return FileUtils::write_text_file(filename, ss.str());
}

std::string ConfigUtils::get_config_value(const std::unordered_map<std::string, std::string>& config,
                                         const std::string& key, const std::string& default_value) {
    auto it = config.find(key);
    return (it != config.end()) ? it->second : default_value;
}

int ConfigUtils::get_config_value_int(const std::unordered_map<std::string, std::string>& config,
                                     const std::string& key, int default_value) {
    auto it = config.find(key);
    if (it != config.end()) {
        try {
            return std::stoi(it->second);
        } catch (const std::exception& e) {
            LOG_WARNING("无法解析配置值 %s: %s", key.c_str(), e.what());
        }
    }
    return default_value;
}

double ConfigUtils::get_config_value_double(const std::unordered_map<std::string, std::string>& config,
                                           const std::string& key, double default_value) {
    auto it = config.find(key);
    if (it != config.end()) {
        try {
            return std::stod(it->second);
        } catch (const std::exception& e) {
            LOG_WARNING("无法解析配置值 %s: %s", key.c_str(), e.what());
        }
    }
    return default_value;
}

bool ConfigUtils::get_config_value_bool(const std::unordered_map<std::string, std::string>& config,
                                       const std::string& key, bool default_value) {
    auto it = config.find(key);
    if (it != config.end()) {
        std::string value = StringUtils::to_lower(it->second);
        return (value == "true" || value == "1" || value == "yes" || value == "on");
    }
    return default_value;
}

// PerformanceMonitor 实现
PerformanceMonitor::PerformanceMonitor(const std::string& name) : name_(name) {}

PerformanceMonitor::~PerformanceMonitor() {
    print_stats();
}

void PerformanceMonitor::start() {
    start_time_ = std::chrono::high_resolution_clock::now();
}

void PerformanceMonitor::stop() {
    auto end_time = std::chrono::high_resolution_clock::now();
    auto duration = std::chrono::duration_cast<std::chrono::milliseconds>(end_time - start_time_).count();
    
    std::lock_guard<std::mutex> lock(mutex_);
    times_.push_back(duration);
}

void PerformanceMonitor::reset() {
    std::lock_guard<std::mutex> lock(mutex_);
    times_.clear();
}

int64_t PerformanceMonitor::get_total_time_ms() const {
    std::lock_guard<std::mutex> lock(mutex_);
    int64_t total = 0;
    for (int64_t time : times_) {
        total += time;
    }
    return total;
}

int64_t PerformanceMonitor::get_average_time_ms() const {
    std::lock_guard<std::mutex> lock(mutex_);
    if (times_.empty()) {
        return 0;
    }
    return get_total_time_ms() / times_.size();
}

int64_t PerformanceMonitor::get_min_time_ms() const {
    std::lock_guard<std::mutex> lock(mutex_);
    if (times_.empty()) {
        return 0;
    }
    return *std::min_element(times_.begin(), times_.end());
}

int64_t PerformanceMonitor::get_max_time_ms() const {
    std::lock_guard<std::mutex> lock(mutex_);
    if (times_.empty()) {
        return 0;
    }
    return *std::max_element(times_.begin(), times_.end());
}

int64_t PerformanceMonitor::get_call_count() const {
    std::lock_guard<std::mutex> lock(mutex_);
    return times_.size();
}

double PerformanceMonitor::get_throughput_fps() const {
    int64_t total_time_ms = get_total_time_ms();
    if (total_time_ms == 0) {
        return 0.0;
    }
    return static_cast<double>(get_call_count()) / (total_time_ms / 1000.0);
}

void PerformanceMonitor::print_stats() const {
    LOG_INFO("性能统计 [%s] - 调用次数: %lld, 总时间: %s, 平均时间: %lldms, 最小: %lldms, 最大: %lldms, 吞吐量: %.2f fps",
             name_.c_str(),
             get_call_count(),
             StringUtils::format_duration(get_total_time_ms()).c_str(),
             get_average_time_ms(),
             get_min_time_ms(),
             get_max_time_ms(),
             get_throughput_fps());
}

} // namespace ffmpeg_detection 