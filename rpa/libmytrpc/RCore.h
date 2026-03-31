#ifndef RPEER_H
#define RPEER_H
#include <pthread.h>
#include <map>
#include <memory>
#include <mutex>
#include <condition_variable>
#include <memory>

using namespace std;
class RPeerManager;
typedef struct RResult{
	int status;
	void* data;
	int len;
public:RResult():status(0),data(NULL),len(0){
	}
}RResult;

class RPeer;
class RPeerPtr {
public:
	RPeerPtr(std::shared_ptr<RPeer> ptr) :m_ptr(ptr) {
	}
	~RPeerPtr() {
	}
	std::shared_ptr<RPeer> getPtr() {
		return m_ptr;
	}
private:
	std::shared_ptr<RPeer> m_ptr;
};

class RPeer{
 friend class RPeerManager;
protected:
	static void* OnThreadPeer(void* lp);
	static void* OnThreadSysCall(void* lp);
	RPeer(int fd,RResult (*fOnSysCall)(RPeer*,int,void*,int),void (*fOnClose)(RPeer*));
	void OnRun(RPeerPtr* ptr);
	std::shared_ptr<RPeer> Start();
	void OnSysCall(void* lp);
	void SysCallRet(int session,void* result,int len);
public:
	~RPeer();
	void Close();
	bool IsClosed();
	void SetData(void* ptr);
	void* GetData();
	RResult* SysCall(int needRet,int method,void* data,int len,int timeout = -1);
private:
	pthread_mutex_t m_lock;
	pthread_mutex_t m_lock_hash;
	pthread_mutex_t m_send_mutex;
    pthread_cond_t  m_send_cond;
	bool m_bRun;
	int  m_fd;
	int  m_sessionId;
	map<int,RResult> m_ret_hash;
	void (*m_fOnClose)(RPeer*);
	RResult (*m_pOnSysCall)(RPeer*,int,void*,int);
	void* m_ptr;
};

class RPeerManager{
protected:
	static RPeerManager* m_pInstance;
	void (*m_cb)(RPeer* peer);
	int m_port;
	void (*m_fOnClose)(RPeer*);
	RResult (*m_pOnSysCall)(RPeer*,int,void*,int);
	RPeerManager();
	~RPeerManager();
	void OnRun();
	static void* OnThreadRun(void* lp);
public:
	void registerService(int port,void (*cb)(RPeer* peer),RResult (*OnSysCall)(RPeer*,int,void*,int),void (*fOnClose)(RPeer*));
	std::shared_ptr<RPeer> connectService(const char* ip,int port,RResult (*OnSysCall)(RPeer*,int,void*,int),void (*fOnClose)(RPeer*),long timeout);
	static RPeerManager* getInstance();

};
#endif