/*
 * SPDX-FileCopyrightText: 2024 M5Stack Technology CO LTD
 *
 * SPDX-License-Identifier: MIT
 */
#include "ai_detection_node.h"
#include <signal.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <unistd.h>
#include <fstream>
#include <stdexcept>
#include <iostream>

using namespace StackFlows;

int main_exit_flag = 0;

static void signal_handler(int signal_no) {
    std::cout << "Received signal " << signal_no << ", shutting down..." << std::endl;
    main_exit_flag = 1;
}

int main(int argc, char* argv[]) {
    // Setup signal handlers
    signal(SIGINT, signal_handler);
    signal(SIGTERM, signal_handler);
    
    std::cout << "Starting AI Detection Node..." << std::endl;
    
    try {
        // Create AI detection node
        AIDetectionNode detection_node("ai-detection");
        
        std::cout << "AI Detection Node initialized successfully" << std::endl;
        std::cout << "Waiting for detection requests..." << std::endl;
        
        // Main loop
        while (!main_exit_flag) {
            sleep(1);
        }
        
        std::cout << "AI Detection Node shutting down..." << std::endl;
        
    } catch (const std::exception& e) {
        std::cerr << "Error: " << e.what() << std::endl;
        return 1;
    } catch (...) {
        std::cerr << "Unknown error occurred" << std::endl;
        return 1;
    }
    
    return 0;
}
