#pragma once
#include <iostream>
#include <fstream>
#include <sstream>
#include <ctime>
#include <cstdarg>
#include <iomanip>
#include <mutex>

class Logger {
public:
	enum Level { DEBUG, INFO, WARN, FATAL };

	static void Write(Level level, const char* format, ...) {
		//static std::mutex logMutex; // 声明全局锁
		//std::lock_guard<std::mutex> lock(logMutex); // 获取锁
		//std::string levelStr;
		//switch (level) {
		//case DEBUG: levelStr = "DEBUG"; break;
		//case INFO: levelStr = "INFO"; break;
		//case WARN: levelStr = "WARN"; break;
		//case FATAL: levelStr = "FATAL"; break;
		//}

		//std::va_list args;
		//va_start(args, format);
		//std::string message = FormatMessage(format, args);
		//va_end(args);

		//std::stringstream ss;
		//ss << "[" << GetTimestamp() << "][" << levelStr << "] " << message;
		//std::string logEntry = ss.str();

		//std::cout << logEntry << std::endl; // 输出到控制台

		//std::ofstream logFile("d:\\mytlog.txt", std::ios::app); // 输出到文件
		//if (logFile.is_open()) {
		//	logFile << logEntry << std::endl;
		//	logFile.close();
		//}
	}

private:

	static std::string GetTimestamp() {
		std::time_t now = std::time(nullptr);
		struct tm localTime;
#ifdef _WIN32
		localtime_s(&localTime, &now);
#else
		localtime_r(&now, &localTime);
#endif
		char buffer[20];
		std::strftime(buffer, sizeof(buffer), "%Y-%m-%d %H:%M:%S", &localTime);
		return std::string(buffer);
	}

	static std::string FormatMessage(const char* format, va_list args) {
		int size = std::vsnprintf(nullptr, 0, format, args) + 1; // 计算所需的缓冲区大小
		if (size <= 0) {
			return std::string();
		}

		std::unique_ptr<char[]> buffer(new char[size]);
		std::vsnprintf(buffer.get(), size, format, args); // 格式化消息
		return std::string(buffer.get(), buffer.get() + size - 1); // 去掉末尾的空字符
	}
};