/*
 * AI Inference Node Main Entry
 * Supports ASR, Emotion Detection, and Synthesis Detection
 */
#include "StackFlow.h"
#include "channel.h"
#include "asr_task.h"
#include "whisper_asr_task.h"
#include "emotion_task.h"
#include "synthesis_task.h"
#include <signal.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <unistd.h>
#include <fstream>
#include <stdexcept>
#include <iostream>

using namespace StackFlows;
using namespace AIInference;
using json = nlohmann::json;

int main_exit_flage = 0;
static void __sigint(int iSigNo)
{
    main_exit_flage = 1;
}

// AI Inference Node class
class ai_inference : public StackFlow
{
private:
    int task_count_;
    std::unordered_map<int, std::shared_ptr<BaseTask>> ai_tasks_;

public:
    ai_inference() : StackFlow("llm")
    {
        std::cerr << "[Constructor] ai_inference() START" << std::endl;
        task_count_ = 20; // Support up to 10 concurrent tasks
        std::cerr << "[Constructor] task_count_ = " << task_count_ << std::endl;
        std::cerr << "[Constructor] ai_inference() END" << std::endl;
    }

    void task_output(const std::weak_ptr<BaseTask> task_obj_weak,
                     const std::weak_ptr<llm_channel_obj> llm_channel_weak, 
                     const std::string &data, bool finish)
    {
        auto task_obj = task_obj_weak.lock();
        auto llm_channel = llm_channel_weak.lock();
        if (!(task_obj && llm_channel))
        {
            return;
        }
        
        if (llm_channel->enstream_)
        {
            static int count = 0;
            nlohmann::json data_body;
            data_body["index"] = count++;
            data_body["delta"] = data;
            if (!finish)
                data_body["delta"] = data;
            else
                data_body["delta"] = std::string("");
            data_body["finish"] = finish;
            if (finish)
                count = 0;

            llm_channel->send(task_obj->response_format_, data_body, LLM_NO_ERROR);
        }
        else if (finish)
        {
            llm_channel->send(task_obj->response_format_, data, LLM_NO_ERROR);
        }
    }

    void task_user_data(const std::weak_ptr<BaseTask> task_obj_weak,
                        const std::weak_ptr<llm_channel_obj> llm_channel_weak, 
                        const std::string &object,
                        const std::string &data)
    {
        nlohmann::json error_body;
        auto task_obj = task_obj_weak.lock();
        auto llm_channel = llm_channel_weak.lock();
        if (!(task_obj && llm_channel))
        {
            error_body["code"] = -11;
            error_body["message"] = "Task run failed.";
            send("None", "None", error_body, unit_name_);
            return;
        }
        if (data.empty() || (data == "None"))
        {
            error_body["code"] = -24;
            error_body["message"] = "The inference data is empty.";
            send("None", "None", error_body, unit_name_);
            return;
        }
        
        const std::string *next_data = &data;
        std::string tmp_msg;
        if (object.find("stream") != std::string::npos)
        {
            static std::unordered_map<int, std::string> stream_buff;
            try
            {
                if (decode_stream(data, tmp_msg, stream_buff)) {
                    return;
                };
            }
            catch (...)
            {
                stream_buff.clear();
                error_body["code"] = -25;
                error_body["message"] = "Stream data index error.";
                send("None", "None", error_body, unit_name_);
                return;
            }
            next_data = &tmp_msg;
        }

        task_obj->inference((*next_data));
    }

    int setup(const std::string &work_id, const std::string &object, const std::string &data) override
    {
        std::cerr << "[AI Inference] === SETUP START ===" << std::endl;
        std::cerr << "[AI Inference] work_id=" << work_id << std::endl;
        std::cerr << "[AI Inference] object=" << object << std::endl;
        std::cerr << "[AI Inference] data=" << data << std::endl;

        nlohmann::json error_body;
        std::cerr << "[AI Inference] llm_task_channel_.size()=" << llm_task_channel_.size() << std::endl;
        std::cerr << "[AI Inference] task_count_=" << task_count_ << std::endl;

        if ((llm_task_channel_.size() - 1) >= task_count_)
        {
            std::cerr << "[AI Inference] Task full!" << std::endl;
            error_body["code"] = -21;
            error_body["message"] = "task full";
            send("None", "None", error_body, unit_name_);
            return -1;
        }

        std::cerr << "[AI Inference] Extracting work_id_num..." << std::endl;
        int work_id_num = sample_get_work_id_num(work_id);
        std::cerr << "[AI Inference] work_id_num=" << work_id_num << std::endl;

        std::cerr << "[AI Inference] Getting channel (using work_id string)..." << std::endl;
        auto llm_channel = get_channel(work_id);  // Use string like test node does
        std::cerr << "[AI Inference] Channel obtained successfully" << std::endl;

        nlohmann::json config_body;
        try
        {
            config_body = nlohmann::json::parse(data);
        }
        catch (...)
        {
            error_body["code"] = -2;
            error_body["message"] = "json format error.";
            send("None", "None", error_body, unit_name_);
            return -2;
        }

        // Determine task type based on model name
        std::string model_name = config_body.at("model");
        std::shared_ptr<BaseTask> task_obj;
        
        if (model_name.find("whisper") != std::string::npos) {
            task_obj = std::make_shared<WhisperASRTask>(work_id);
            std::cout << "Creating Whisper ASR task for work_id: " << work_id << std::endl;
        }
        else if (model_name.find("asr") != std::string::npos ||
                 model_name.find("speech") != std::string::npos) {
            task_obj = std::make_shared<ASRTask>(work_id);
            std::cout << "Creating ASR task for work_id: " << work_id << std::endl;
        }
        else if (model_name.find("emotion") != std::string::npos || 
                 model_name.find("sentiment") != std::string::npos) {
            task_obj = std::make_shared<EmotionTask>(work_id);
            std::cout << "Creating Emotion Detection task for work_id: " << work_id << std::endl;
        }
        else if (model_name.find("synthesis") != std::string::npos || 
                 model_name.find("deepfake") != std::string::npos ||
                 model_name.find("fake") != std::string::npos) {
            task_obj = std::make_shared<SynthesisTask>(work_id);
            std::cout << "Creating Synthesis Detection task for work_id: " << work_id << std::endl;
        }
        else {
            // Default to ASR task
            task_obj = std::make_shared<ASRTask>(work_id);
            std::cout << "Creating default ASR task for work_id: " << work_id << std::endl;
        }

        int ret = task_obj->load_model(config_body);
        if (ret == 0)
        {
            llm_channel->set_output(true);
            llm_channel->set_stream(task_obj->enstream_);
            task_obj->set_output(std::bind(&ai_inference::task_output, this,
                                          std::weak_ptr<BaseTask>(task_obj),
                                          std::weak_ptr<llm_channel_obj>(llm_channel),
                                          std::placeholders::_1,
                                          std::placeholders::_2));
            llm_channel->subscriber_work_id(
                "",
                std::bind(&ai_inference::task_user_data, this,
                         std::weak_ptr<BaseTask>(task_obj),
                         std::weak_ptr<llm_channel_obj>(llm_channel),
                         std::placeholders::_1,
                         std::placeholders::_2));
            ai_tasks_[work_id_num] = task_obj;
            task_obj->start();
            send("None", "None", LLM_NO_ERROR, work_id);

            return 0;
        }
        else
        {
            nlohmann::json load_error_body;
            load_error_body["code"] = -5;
            load_error_body["message"] = "Model loading failed.";
            send("None", "None", load_error_body, unit_name_);
            return -1;
        }
    }

    void taskinfo(const std::string &work_id, const std::string &object, const std::string &data) override
    {
        nlohmann::json req_body;
        int work_id_num = sample_get_work_id_num(work_id);
        if (WORK_ID_NONE == work_id_num)
        {
            std::vector<std::string> task_list;
            std::transform(llm_task_channel_.begin(), llm_task_channel_.end(), 
                          std::back_inserter(task_list),
                          [](const auto task_channel)
                          { return task_channel.second->work_id_; });
            req_body = task_list;
            send("llm.tasklist", req_body, LLM_NO_ERROR, work_id);
        }
        else
        {
            if (ai_tasks_.find(work_id_num) == ai_tasks_.end())
            {
                req_body["code"] = -6;
                req_body["message"] = "Unit Does Not Exist";
                send("None", "None", req_body, work_id);
                return;
            }
            auto task_obj = ai_tasks_[work_id_num];
            req_body["model"] = task_obj->model_;
            req_body["response_format"] = task_obj->response_format_;
            req_body["enoutput"] = task_obj->enoutput_;
            req_body["inputs"] = task_obj->inputs_;
            send("llm.taskinfo", req_body, LLM_NO_ERROR, work_id);
        }
    }

    int exit(const std::string &work_id, const std::string &object, const std::string &data) override
    {
        std::cerr << "[AI Inference] === EXIT START ===" << std::endl;
        std::cerr << "[AI Inference] work_id=" << work_id << std::endl;

        nlohmann::json error_body;
        int work_id_num = sample_get_work_id_num(work_id);
        std::cerr << "[AI Inference] work_id_num=" << work_id_num << std::endl;

        if (ai_tasks_.find(work_id_num) == ai_tasks_.end())
        {
            std::cerr << "[AI Inference] ERROR: Unit does not exist" << std::endl;
            error_body["code"] = -6;
            error_body["message"] = "Unit Does Not Exist";
            send("None", "None", error_body, work_id);
            return -1;
        }

        std::cerr << "[AI Inference] Stopping task..." << std::endl;
        ai_tasks_[work_id_num]->stop();

        std::cerr << "[AI Inference] Erasing task from map..." << std::endl;
        ai_tasks_.erase(work_id_num);

        // NOTE: Do NOT stop channel or send response here!
        // The base class StackFlow::exit() will handle cleanup and call sys_release_unit()
        // Stopping the channel here causes send() to block

        std::cerr << "[AI Inference] === EXIT END (returning 0) ===" << std::endl;
        return 0;
    }

    ~ai_inference()
    {
        while (1)
        {
            auto iteam = ai_tasks_.begin();
            if (iteam == ai_tasks_.end())
            {
                break;
            }
            iteam->second->stop();
            get_channel(iteam->first)->stop_subscriber("");
            iteam->second.reset();
            ai_tasks_.erase(iteam->first);
        }
    }
};

int main(int argc, char *argv[])
{
    // Disable stdout buffering for immediate output
    setbuf(stdout, NULL);
    setbuf(stderr, NULL);

    signal(SIGTERM, __sigint);
    signal(SIGINT, __sigint);
    mkdir("/tmp/llm", 0777);
    
    std::cout << "========================================" << std::endl;
    std::cout << "AI Inference Node Starting..." << std::endl;
    std::cout << "========================================" << std::endl;
    std::cout << "Unit name: llm" << std::endl;
    std::cout << "Supported models:" << std::endl;
    std::cout << "  - ASR (Automatic Speech Recognition)" << std::endl;
    std::cout << "  - Emotion Detection" << std::endl;
    std::cout << "  - Synthesis Detection (Deepfake)" << std::endl;
    std::cout << "========================================" << std::endl;

    std::cerr << "[main] Before creating ai_inference object" << std::endl;
    std::cout << "Initializing AI Inference node..." << std::endl;
    ai_inference ai_node;
    std::cerr << "[main] After creating ai_inference object" << std::endl;
    std::cout << "AI Inference node initialized successfully!" << std::endl;
    std::cout << "Node is ready to accept requests..." << std::endl;
    std::cout << "========================================" << std::endl;

    std::cerr << "[main] Entering main loop" << std::endl;
    while (!main_exit_flage)
    {
        sleep(1);
    }
    std::cerr << "[main] Exiting main loop" << std::endl;

    std::cout << "AI Inference Node shutting down..." << std::endl;
    return 0;
}

