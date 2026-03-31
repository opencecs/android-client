// libmytrpc.cpp : 定义 DLL 应用程序的导出函数。
//

#include "stdafx.h"
#include "libmytrpc.h"
#include "RCore.h"
#include "PkgData.h"
#include "string.h"
#include "stdlib.h"
#include "Logger.h"
#include <iostream>
#include <share.h>

//#define RPORT	10050
#define RPORT	10005

#define CMD_CHECK_LIVE			100
#define CMD_TAKECAPTURE			102
#define CMD_TAKECAPCMPRESS	    103
#define CMD_TOUCHACTION		    104
#define CMD_KEYEVENT		    105
#define CMD_INPUTTEXT		    106
#define CMD_RUNSTOAPP		    107
#define CMD_DUMPNODE		    108
#define CMD_GETROTATE		    109
#define CMD_EXEC				110
#define CMD_STARTVIDEO	    	111
#define CMD_STOPVIDEO	    	112
#define CMD_AUDIOSTREAM         113
#define CMD_USENEWNODE		    114

#define TOUCH_EV_DOWN			1
#define TOUCH_EV_UP				2
#define TOUCH_EV_MOVE			3
#define TOUCH_EV_TAP			4

typedef struct PeerInfo{
	int port;
	string ip;
}PeerInfo;

static pthread_mutex_t g_client_mutex = PTHREAD_MUTEX_INITIALIZER;
static map<long,std::shared_ptr<RPeer>> g_clients;
static long g_base_cnt = 1L;
static void (*onVideoStream)(int rot,void* data,int len) = NULL;
static void(*onAudioStream)(void* data, int len) = NULL;
static 
RResult OnSysCall(RPeer* peer,int cmd,void* data,int len){
	RResult ret;
	if (cmd == CMD_STARTVIDEO) {
		if (onVideoStream) {
			BYTE* ptr = (BYTE*)data;
			onVideoStream(ptr[0], ptr + 8, len - 8);
		}
	}
	else if (cmd == CMD_AUDIOSTREAM) {
		if (onAudioStream) {
			onAudioStream(data,len);
		}
	}
	return ret;
};

static void fOnClose(RPeer* peer){
	
}

static RResult* mSysCall(std::shared_ptr<RPeer> peer, int handle, int needRet, int method, void* data, int len, int timeout = -1) {
	Logger::Write(Logger::INFO, "调用命令开始(%d %d)", handle,method);
	RResult* ret = peer->SysCall(needRet, method, data, len, timeout);
	Logger::Write(Logger::INFO, "调用命令结束(%d %d)", handle, method);
	return ret;
}


static std::shared_ptr<RPeer> checkPeerLive(long handle, std::shared_ptr<RPeer> peer){
	if(peer && peer->IsClosed()){
		PeerInfo* info = (PeerInfo*)peer->GetData();
		string ip = info->ip.c_str();
		int port = info->port;
		std::shared_ptr<RPeer> newpeer =  RPeerManager::getInstance()->connectService(ip.c_str(), info->port, OnSysCall, fOnClose, 5);
		if(newpeer){
			int bLive = 0;
			RResult* pRet = mSysCall(newpeer,handle,1,CMD_CHECK_LIVE,NULL,0,1000);
			if(pRet){
				if(pRet->data){
					free(pRet->data);
					bLive = 1;
				}
				delete pRet;
			}
			if(!bLive){
				newpeer->Close();
				newpeer = NULL;
			}
		}
		if(newpeer){
			peer->Close();
			pthread_mutex_lock(&g_client_mutex);
			map<long, std::shared_ptr<RPeer>>::iterator it = g_clients.find(handle);
			if(it != g_clients.end()){
				g_clients.erase(it);
			}
			PeerInfo* p = new PeerInfo();
			p->ip = ip;
			p->port = port;
			newpeer->SetData(p);
			g_clients[handle] = newpeer;
			pthread_mutex_unlock(&g_client_mutex);
			return newpeer;
		}else{
			return peer;
		}
	}else{
		return peer;
	}
}

static std::shared_ptr<RPeer> obtainPeer(long handle){
	std::shared_ptr<RPeer> peer = NULL;
	Logger::Write(Logger::INFO, "检测(%d) 是否活着开始", handle);
	pthread_mutex_lock(&g_client_mutex);
	map<long, std::shared_ptr<RPeer>>::iterator it = g_clients.find(handle);
	if(it != g_clients.end()){
		peer = it->second;
		pthread_mutex_unlock(&g_client_mutex);
		peer = checkPeerLive(handle,peer);
	}else{
		pthread_mutex_unlock(&g_client_mutex);
	}
	if (peer && !peer->IsClosed()) {
		Logger::Write(Logger::INFO, "检测(%d) 句柄还活着", handle);
	}
	else {
		Logger::Write(Logger::INFO, "检测(%d) 句柄已死", handle);
	}
	
	return peer;
}

static int touchAction(long handle,int id,int ev,int x,int y){
	int ret = 0;
	std::shared_ptr<RPeer> peer = obtainPeer(handle);
	if(peer){
	   if(!peer->IsClosed()){
			PkgData pkg;
			pkg.WrapInt(id);
			pkg.WrapInt(ev);
			pkg.WrapInt(x);
			pkg.WrapInt(y);
			int outLen = 0;
			void* outData = pkg.GetPtr(outLen);
			RResult* pRet = mSysCall(peer,handle,0,CMD_TOUCHACTION,outData,outLen);
			if(pRet){
				if(pRet->data){
					free(pRet->data);
				}
				delete pRet;
			}
			ret = 1;
	   }
	}
	return ret;
}

static int runOrStopApp(long handle,int isRun,const char* pack){
		int ret = 0;
		std::shared_ptr<RPeer> peer = obtainPeer(handle);
		if(peer){
		   if(pack && strlen(pack) && !peer->IsClosed()){
				PkgData pkg;
				pkg.WrapInt(isRun);
				pkg.WrapData((void*)pack,strlen(pack));
				int outLen = 0;
				void* outData = pkg.GetPtr(outLen);
				RResult* pRet = mSysCall(peer, handle, 0,CMD_RUNSTOAPP,outData,outLen);
				if(pRet){
					if(pRet->data){
						free(pRet->data);
					}
					ret = 1;
					delete pRet;
				}
		   }
		}
		return ret;
}

void sleepMilliseconds(int milliseconds);
unsigned long long getCurrentMilliseconds();

extern "C"{
	LIBMYTRPC_API long MYAPI openDevice(const char* ip,int port,long timeout){
		Logger::Write(Logger::INFO, "openDevice(%s:%d) 开始",ip,port);
		std::shared_ptr<RPeer> peer = RPeerManager::getInstance()->connectService(ip,port,OnSysCall,fOnClose,timeout);
		if(peer){
			int bLive = 0;
			RResult* pRet = mSysCall(peer,-1,1,CMD_CHECK_LIVE,NULL,0,1000);
			pthread_mutex_lock(&g_client_mutex);
			if(pRet){
				if(pRet->data){
					free(pRet->data);
					bLive = 1;
				}
				delete pRet;
			}
			if(!bLive){
				peer->Close();
				peer = NULL;
				pthread_mutex_unlock(&g_client_mutex);
				Logger::Write(Logger::INFO, "openDevice(%s:%d) 失败", ip, port);
				return 0;
			}
			long ret = g_base_cnt++;
			PeerInfo* info = new PeerInfo();
			info->ip = ip;
			info->port = port;
			peer->SetData(info);
			g_clients[ret] = peer;
			Logger::Write(Logger::INFO, "openDevice(%s:%d) 成功句柄是%d", ip, port,ret);
			pthread_mutex_unlock(&g_client_mutex);
			return ret;
		}	
		return 0;
	}


	LIBMYTRPC_API int MYAPI closeDevice(long handle){
		int ret = 0;
		pthread_mutex_lock(&g_client_mutex);
		Logger::Write(Logger::INFO, "closeDevice(%d)", handle);
		map<long, std::shared_ptr<RPeer>>::iterator it = g_clients.find(handle);
		if(it != g_clients.end()){
			std::shared_ptr<RPeer> peer = it->second;
			g_clients.erase(it);
			pthread_mutex_unlock(&g_client_mutex);
			if(peer){
				peer->Close();
			}
			ret = 1;
		}
		else {
			pthread_mutex_unlock(&g_client_mutex);
		}
		Logger::Write(Logger::INFO, "closeDevice(%d) 结束", handle);
		return ret;
	}

	
	LIBMYTRPC_API int MYAPI checkLive(long handle){
		int ret = 0;
		std::shared_ptr<RPeer> peer = NULL;
		Logger::Write(Logger::INFO, "checkLive(%d) 开始", handle);
		pthread_mutex_lock(&g_client_mutex);
		map<long, std::shared_ptr<RPeer>>::iterator it = g_clients.find(handle);
		if(it != g_clients.end()){
			peer = it->second;
			pthread_mutex_unlock(&g_client_mutex);
			peer = checkPeerLive(handle,peer);
		}
		else {
			pthread_mutex_unlock(&g_client_mutex);
		}
		if(peer && !peer->IsClosed()){
			ret = 1;
			Logger::Write(Logger::INFO, "checkLive(%d) 还活着", handle);
		}
		else {
			Logger::Write(Logger::INFO, "checkLive(%d) 死了", handle);
		}
		
		return ret;
	}

	LIBMYTRPC_API BYTE* MYAPI takeCaptrueEx(long handle,int l,int t,int r,int b,int* w,int* h,int* stride){
		BYTE* pData = 0;
		std::shared_ptr<RPeer> peer = obtainPeer(handle);
		if(peer){
			if(!peer->IsClosed()){
				PkgData pkgarg;
				pkgarg.WrapInt(l);
				pkgarg.WrapInt(t);
				pkgarg.WrapInt(r);
				pkgarg.WrapInt(b);
				int outLen = 0;
				void* outData = pkgarg.GetPtr(outLen);
				RResult* pRet = mSysCall(peer,handle,1,CMD_TAKECAPTURE,outData,outLen);
				if(pRet){
					if(pRet->data){
						PkgData pkg(pRet->data,pRet->len);
						int tmpW = 0;
						int tmpH = 0;
						int tmpStride = 0;
						pkg.UnWrapInt(tmpW);
						pkg.UnWrapInt(tmpH);
						pkg.UnWrapInt(tmpStride);
						int dataLen = 0;
						void* p = (BYTE*)pkg.UnWrapData(dataLen);
						if(p){
							pData = (BYTE*)malloc(dataLen);
							memcpy(pData,p,dataLen);
						}
						if(w) *w = tmpW;
						if(h) *h = tmpH;
						if(stride) *stride = tmpStride;
						free(pRet->data);
					}
					delete pRet;
				}
			}
		}
		return pData;
	}

	LIBMYTRPC_API BYTE* MYAPI takeCaptrue(long handle,int* w,int* h,int* stride){
		BYTE* pData = 0;
		std::shared_ptr<RPeer> peer = obtainPeer(handle);
		if(peer){
			if(!peer->IsClosed()){
				PkgData pkgarg;
				pkgarg.WrapInt(0);
				pkgarg.WrapInt(0);
				pkgarg.WrapInt(0);
				pkgarg.WrapInt(0);
				int outLen = 0;
				void* outData = pkgarg.GetPtr(outLen);
				RResult* pRet = mSysCall(peer,handle,1,CMD_TAKECAPTURE,outData,outLen);
				if(pRet){
					if(pRet->data){
						PkgData pkg(pRet->data,pRet->len);
						int tmpW = 0;
						int tmpH = 0;
						int tmpStride = 0;
						pkg.UnWrapInt(tmpW);
						pkg.UnWrapInt(tmpH);
						pkg.UnWrapInt(tmpStride);
						int dataLen = 0;
						void* p = (BYTE*)pkg.UnWrapData(dataLen);
						if(p){
							pData = (BYTE*)malloc(dataLen);
							memcpy(pData,p,dataLen);
						}
						if(w) *w = tmpW;
						if(h) *h = tmpH;
						if(stride) *stride = tmpStride;
						free(pRet->data);
					}
					delete pRet;
				}
			}
		}
		return pData;
	}

	LIBMYTRPC_API BYTE* MYAPI takeCaptrueCompressEx(long handle,int l,int t,int r,int b,int type,int quality,int* len){
		BYTE* pData = 0;
		std::shared_ptr<RPeer> peer = obtainPeer(handle);
		if(peer){
			if(!peer->IsClosed()){
				PkgData pkg;
				pkg.WrapInt(type);
				pkg.WrapInt(quality);
				pkg.WrapInt(l);
				pkg.WrapInt(t);
				pkg.WrapInt(r);
				pkg.WrapInt(b);
				int outLen = 0;
				void* outData = pkg.GetPtr(outLen);
				RResult* pRet = mSysCall(peer, handle,1,CMD_TAKECAPCMPRESS,outData,outLen);
				if(pRet){
					if(pRet->data){
						pData = (BYTE*)pRet->data;
						if(len) *len = pRet->len;
					}
					delete pRet;
				}
			}
		}
		return pData;
	}

	LIBMYTRPC_API BYTE* MYAPI takeCaptrueCompress(long handle,int type,int quality,int* len){
		BYTE* pData = 0;
		//char buffer[1024] = {0};
		//sprintf(buffer,"msg %ld %d %d %p",handle,type,quality,len);
		//MessageBox(NULL,buffer,"",MB_OK);
		std::shared_ptr<RPeer> peer = obtainPeer(handle);
		if(peer){
			if(!peer->IsClosed()){
				PkgData pkg;
				pkg.WrapInt(type);
				pkg.WrapInt(quality);
				pkg.WrapInt(0);
				pkg.WrapInt(0);
				pkg.WrapInt(0);
				pkg.WrapInt(0);
				int outLen = 0;
				void* outData = pkg.GetPtr(outLen);
				RResult* pRet = mSysCall(peer, handle, 1,CMD_TAKECAPCMPRESS,outData,outLen);
				if(pRet){
					if(pRet->data){
						pData = (BYTE*)pRet->data;
						if(len) *len = pRet->len;
					}
					delete pRet;
				}
			}
		}
		return pData;
	}

	LIBMYTRPC_API char* MYAPI dumpNodeXml(long handle,int bDumpAll){
		char* pData = 0;
		std::shared_ptr<RPeer> peer = obtainPeer(handle);
		if(peer){
			if(!peer->IsClosed()){
				PkgData pkg;
				pkg.WrapInt(bDumpAll);
				int outLen = 0;
				void* outData = pkg.GetPtr(outLen);
				RResult* pRet = mSysCall(peer, handle, 1,CMD_DUMPNODE,outData,outLen,45000);
				if(pRet){
					if(pRet->data){
						pData = (char*)pRet->data;
					}
					delete pRet;
				}
			}
		}
		return pData;
	}

	LIBMYTRPC_API char* MYAPI dumpNodeXmlEx(long handle, int useNewMode, int timeout) {
		char* pData = 0;
		std::shared_ptr<RPeer> peer = obtainPeer(handle);
		if (peer) {
			if (!peer->IsClosed()) {
				{
					PkgData pkg;
					if (useNewMode)
						pkg.WrapInt(1);
					else
						pkg.WrapInt(0);
					int outLen = 0;
					void* outData = pkg.GetPtr(outLen);
					RResult* pRet = mSysCall(peer, handle, 1, CMD_USENEWNODE, outData, outLen);
					if (pRet) {
						if (pRet->data) {
							free(pRet->data);
						}
						delete pRet;
					}
				}
				PkgData pkg;
				pkg.WrapInt(1);
				int outLen = 0;
				void* outData = pkg.GetPtr(outLen);
				RResult* pRet = mSysCall(peer, handle, 1, CMD_DUMPNODE, outData, outLen, timeout);
				if (pRet) {
					if (pRet->data) {
						pData = (char*)pRet->data;
					}
					delete pRet;
				}
			}
			else {
				printf("nul\n");
			}
		}
		else {
			printf("nul1\n");
		}
		return pData;
	}

	LIBMYTRPC_API char* MYAPI execCmd(long handle,int bWaitForExit,const char* cmdline){
		char* pData = 0;
		std::shared_ptr<RPeer> peer = obtainPeer(handle);
		if(peer){
			if(!peer->IsClosed()){
				PkgData pkg;
				pkg.WrapInt(bWaitForExit);
				pkg.WrapData((void*)cmdline,strlen(cmdline));
				int outLen = 0;
				void* outData = pkg.GetPtr(outLen);
				RResult* pRet = mSysCall(peer, handle, 1,CMD_EXEC,outData,outLen);
				if(pRet){
					if(pRet->data){
						pData = (char*)pRet->data;
					}
					delete pRet;
				}
			}
		}
		return pData;
	}

	LIBMYTRPC_API int MYAPI keyPress(long handle,int code){
		int ret = 0;
		std::shared_ptr<RPeer> peer = obtainPeer(handle);
		if(peer){
		   if(!peer->IsClosed()){
				PkgData pkg;
				pkg.WrapInt(1);//事件类型
				pkg.WrapInt(2);//弹起获知释放
				pkg.WrapInt(code);
				int outLen = 0;
				void* outData = pkg.GetPtr(outLen);
				RResult* pRet = mSysCall(peer, handle, 1,CMD_KEYEVENT,outData,outLen);
				if(pRet){
					if(pRet->data){
						free(pRet->data);
					}
					ret = 1;
					delete pRet;
				}
		   }
		}
		return ret;
	}

	LIBMYTRPC_API int MYAPI useNewNodeMode(long handle, bool use) {
		int ret = 0;
		std::shared_ptr<RPeer> peer = obtainPeer(handle);
		if (peer) {
			if (!peer->IsClosed()) {
				PkgData pkg;
				if(use)
				   pkg.WrapInt(1);
				else
				   pkg.WrapInt(0);
				int outLen = 0;
				void* outData = pkg.GetPtr(outLen);
				RResult* pRet = mSysCall(peer, handle, 1, CMD_USENEWNODE, outData, outLen);
				if (pRet) {
					if (pRet->data) {
						free(pRet->data);
					}
					ret = 1;
					delete pRet;
				}
			}
		}
		return ret;
	}

	LIBMYTRPC_API int  MYAPI sendText(long handle,const char* text)
	{
		int ret = 0;
		std::shared_ptr<RPeer> peer = obtainPeer(handle);
		if(peer){
		   if(text && strlen(text) && !peer->IsClosed()){
				PkgData pkg;
				pkg.WrapData((void*)text,strlen(text));
				int outLen = 0;
				void* outData = pkg.GetPtr(outLen);
				RResult* pRet = mSysCall(peer, handle, 1,CMD_INPUTTEXT,outData,outLen);
				if(pRet){
					if(pRet->data){
						free(pRet->data);
					}
					ret = 1;
					delete pRet;
				}
		   }
		}
		return ret;
	}

	LIBMYTRPC_API int MYAPI getDisplayRotate(long handle){
		int ret = 0;
		std::shared_ptr<RPeer> peer = obtainPeer(handle);
		if(peer){
		   if(!peer->IsClosed()){
				RResult* pRet = mSysCall(peer, handle, 1,CMD_GETROTATE,NULL,0);
				if(pRet){
					if(pRet->data){
						PkgData pkg(pRet->data,pRet->len);
						pkg.UnWrapInt(ret);
						free(pRet->data);
					}
					delete pRet;
				}
		   }
		}
		return ret;
	}


	LIBMYTRPC_API int  MYAPI  openApp(long handle,const char* pkg)
	{
		return runOrStopApp(handle,1,pkg);
	}

	LIBMYTRPC_API int  MYAPI  stopApp(long handle,const char* pkg)
	{
		return runOrStopApp(handle,0,pkg);
	}

	LIBMYTRPC_API int MYAPI touchClick(long handle,int id,int x,int y){
		 return touchAction(handle,id,TOUCH_EV_TAP,x,y);
	}

	LIBMYTRPC_API int MYAPI touchDown(long handle,int id,int x,int y){

		return touchAction(handle,id,TOUCH_EV_DOWN,x,y);
	}

	LIBMYTRPC_API int MYAPI touchUp(long handle,int id,int x,int y){

		return touchAction(handle,id,TOUCH_EV_UP,x,y);
	}

	LIBMYTRPC_API int MYAPI touchMove(long handle,int id,int x,int y){

		return touchAction(handle,id,TOUCH_EV_MOVE,x,y);
	}

	static void* swpie_thread(void* arg){
		long* ptr = (long*)arg;
		int id = ptr[0];
		int x0 = ptr[1];
		int y0 = ptr[2];
		int x1 = ptr[3];
		int y1 = ptr[4];
		long millis = ptr[5];
		long handle = ptr[6];
		int len = millis;
		if(len < 50){
			len = 50;
		}
		if(len == 0) len = 1;
		float stepx = (x1 - x0) * 1.0f / len;
		float stepy = (y1 - y0) * 1.0f / len;
		bool ret = touchAction(handle,id, TOUCH_EV_DOWN, x0, y0);
		if (1) {
			for(int i = 0;i < len;i++){
				float x = x0 + (i + 1) * stepx;
				float y = y0 + (i + 1) * stepy;
				long long start = getCurrentMilliseconds();
				ret = touchAction(handle,id, TOUCH_EV_MOVE, static_cast<int>(x), static_cast<int>(y));
				if(i > 0 && i % 20 == 0){
					long long end = getCurrentMilliseconds();
					long long duration = (end - start);
					if(duration < 20){
						sleepMilliseconds(10LL - duration);
					}
				}
			}
			touchAction(handle,id, TOUCH_EV_MOVE, x1, y1);
			touchAction(handle,id, TOUCH_EV_UP, x1, y1);
		}
		delete[] ptr;
		return NULL;
	}

	LIBMYTRPC_API void MYAPI swipe(long handle,int id,int x0, int y0, int x1, int y1, long millis,bool async) {
		long* arg = new long[7];
		if(!arg){
			return;
		}
		arg[0] = id;
		arg[1] = x0;
		arg[2] = y0;
		arg[3] = x1;
		arg[4] = y1;
		arg[5] = millis;
		arg[6] = handle;
		if(async){
			pthread_t tid;
			pthread_create(&tid,NULL,swpie_thread,arg);
			pthread_detach(tid);
		}else{
			swpie_thread(arg);
		}
		
	}

	LIBMYTRPC_API int MYAPI startVideoStream(long handle,int w,int h,int bitrate, void (*cb)(int rot,void* data,int len),
		   void (*audio_cb)(void*,int)){
		int ret = 0;
		std::shared_ptr<RPeer> peer = obtainPeer(handle);
		if(peer){
		   if(!peer->IsClosed()){
			    onVideoStream = cb;
				onAudioStream = audio_cb;
			    PkgData pkgarg;
				pkgarg.WrapInt(w);
				pkgarg.WrapInt(h);
				pkgarg.WrapInt(bitrate);
				int outLen = 0;
				void* outData = pkgarg.GetPtr(outLen);
				RResult* pRet = mSysCall(peer, handle, 1,CMD_STARTVIDEO,outData,outLen);
				if(pRet){
					if(pRet->data){
						PkgData pkg(pRet->data,pRet->len);
						pkg.UnWrapInt(ret);
						free(pRet->data);
					}
					delete pRet;
				}
		   }
		}
		return ret;
	}

	LIBMYTRPC_API int MYAPI stopVideoStream(long handle){
		int ret = 0;
		std::shared_ptr<RPeer> peer = obtainPeer(handle);
		if(peer){
		   if(!peer->IsClosed()){
				RResult* pRet = mSysCall(peer, handle, 1,CMD_STOPVIDEO,NULL,0);
				if(pRet){
					delete pRet;
					ret = 1;
				}
		   }
		}
		return ret;
	}

	LIBMYTRPC_API void MYAPI  freeRpcPtr(void* data)
	{
		if(data){
			free(data);
		}
	}

	LIBMYTRPC_API int  MYAPI  getVersion(){
		return 10;
	}
}
