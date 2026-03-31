#include "stdafx.h"
#include "json/json.h"
#include  "libmytrpc.h"
#include  "pthread.h"
#include "rapidxml.hpp"
#include <string>
#include <regex>
using namespace std;

#ifdef WIN32
#include <windows.h>
void sleepMilliseconds(int milliseconds) {
    Sleep(milliseconds);
}

unsigned long long getCurrentMilliseconds() {
    SYSTEMTIME systemTime;
    GetSystemTime(&systemTime);
    FILETIME fileTime;
    SystemTimeToFileTime(&systemTime, &fileTime);
    ULARGE_INTEGER largeInteger;
    largeInteger.LowPart = fileTime.dwLowDateTime;
    largeInteger.HighPart = fileTime.dwHighDateTime;
    return largeInteger.QuadPart / 10000ULL;
}
#else
#include <unistd.h>
#include <sys/time.h>

void sleepMilliseconds(int milliseconds) {
    usleep(milliseconds * 1000);
}

unsigned long long getCurrentMilliseconds() {
    struct timeval timeVal;
    gettimeofday(&timeVal, NULL);
    unsigned long long milliseconds = timeVal.tv_sec * 1000LL + timeVal.tv_usec / 1000;
    return milliseconds;
}
#endif

typedef Json::Writer JsonWriter;
typedef Json::Reader JsonReader;
typedef Json::Value  JsonValue;

static pthread_mutex_t g_node_mutex = PTHREAD_MUTEX_INITIALIZER;

static bool equalWith(const std::string& str, const std::string& prefix) {
	if (str.length() != prefix.length()) {
		return false;
	}
	return str == prefix;
}

static bool startsWith(const std::string& str, const std::string& prefix) {
    if (str.length() < prefix.length()) {
        return false;
    }
    return str.substr(0, prefix.length()) == prefix;
}

static bool endsWith(const std::string& str, const std::string& suffix) {
    if (str.length() < suffix.length()) {
        return false;
    }

    return str.substr(str.length() - suffix.length()) == suffix;
}

static bool contains(const std::string& str, const std::string& substr1) {
	size_t pos = str.find(substr1);
	bool r =  (pos != std::string::npos);
    return r;
}

static bool matcher(const std::string& str, const std::string& substr1) {
    std::regex pattern(substr1);
	if (std::regex_match(str, pattern)) {
        return true;
    }
    return false;
}

class NodeBounds{
public:
	int l;
	int t;
	int r;
	int b;
	NodeBounds():l(0),t(0),r(0),b(0){
	}
	NodeBounds(int l,int t,int r,int b){
		this->l = l;
		this->t = t;
		this->r = r;
		this->b = b;
	}

	bool IsInRect(int x, int y) {
		if (x >= l && y >= t && x <= r && y <= b)
			return true;
		return false;
	}

	bool isEqual(NodeBounds& rc) {
		if (rc.l == l && rc.t == t && rc.r == r && rc.b == b){
			return true;
		}
		return false;
	}

	bool IsInRect(NodeBounds& rc) {
		if (IsInRect(rc.l, rc.t) &&
			IsInRect(rc.r, rc.b)) {
			return true;
		}
		return false;
	}
};

class NodeObject{
public:
	std::string mId;
	std::string mText;
	std::string mDesc;
	std::string mPackage;
	std::string mClz;
	int	mIndex;
	int mDepth;
	int mDrawingOrder;
	bool mCheckable;
	bool mChecked;
	bool mClickable;
	bool mEnable;
	bool mFocusable;
	bool mFocused;
	bool mScrollable;
	bool mLongClickable;
	bool mPassword;
	bool mSelected;
	bool mVisible;
	NodeBounds mBounds;
	rapidxml::xml_node<>* mNode;
	std::vector<NodeObject*> m_childs;
	int  mIsRoot;
	int  mRef;
	NodeObject*  mRoot;
	NodeObject*  mParent;
	DLONG mHandle;
	NodeObject(rapidxml::xml_node<>* node):mIndex(-1),mDepth(-1),mDrawingOrder(-1),
	mCheckable(false),mChecked(false),mClickable(false),mEnable(false),mFocusable(false),
	mFocused(false),mScrollable(false),mLongClickable(false),mPassword(false),
	mSelected(false),mVisible(false),mNode(node),mIsRoot(0),mRef(0),mRoot(NULL),mParent(NULL),mHandle(0){
		if(node){
			for (rapidxml::xml_attribute<>* attr = node->first_attribute(); attr; attr = attr->next_attribute()) {
				string name = attr->name();
				char* value = attr->value();
				if (!value) continue;
				if (name == "index") {
					mIndex = atoi(attr->value());
				}
				else if (name == "depth") {
					mDepth = atoi(attr->value());
				}
				else if (name == "drawingorder") {
					mDrawingOrder = atoi(attr->value());
				}
				else if (name == "text") {
					mText = value;
				}
				else if (name == "resource-id") {
					mId = value;
				}
				else if (name == "class") {
					mClz = value;
				}
				else if (name == "package") {
					mPackage = value;
				}
				else if (name == "content-desc") {
					mDesc = value;
				}
				else if (name == "checkable") {
					if (!strcmp(value, "true"))
						mCheckable = true;
					else mCheckable = false;
				}
				else if (name == "checked") {
					if (!strcmp(value, "true"))
						mChecked = true;
					else mChecked = false;
				}
				else if (name == "clickable") {
					if (!strcmp(value, "true"))
						mClickable = true;
					else mClickable = false;
				}
				else if (name == "enabled") {
					if (!strcmp(value, "true"))
						mEnable = true;
					else mEnable = false;
				}
				else if (name == "focusable") {
					if (!strcmp(value, "true"))
						mFocusable = true;
					else mFocusable = false;
				}
				else if (name == "focused") {
					if (!strcmp(value, "true"))
						mFocused = true;
					else mFocused = false;
				}
				else if (name == "scrollable") {
					if (!strcmp(value, "true"))
						mScrollable = true;
					else mScrollable = false;
				}
				else if (name == "long-clickable") {
					if (!strcmp(value, "true"))
						mLongClickable = true;
					else mLongClickable = false;
				}
				else if (name == "password") {
					if (!strcmp(value, "true"))
						mPassword = true;
					else mPassword = false;
				}
				else if (name == "selected") {
					if (!strcmp(value, "true"))
						mSelected = true;
					else mSelected = false;
				}
				else if (name == "visible") {
					if (!strcmp(value, "true"))
						mVisible = true;
					else mVisible = false;
				}
				else if (name == "bounds") {
					int l = 0;
					int t = 0;
					int r = 0;
					int b = 0;
					if (4 == sscanf(value, "[%d,%d][%d,%d]", &l, &t, &r, &b)) {
						mBounds.l = l;
						mBounds.t = t;
						mBounds.r = r;
						mBounds.b = b;
					}
				}
			}
		}else{
			mIsRoot = 1;
		}
	}

	JsonValue toJson(){
		JsonValue v;
		JsonValue bounds;
		v["text"] = mText;
		v["id"] = mId;
		v["desc"] = this->mDesc;
		v["package"] = this->mPackage;
		v["class"] = this->mClz;
		v["index"] = this->mIndex;
		v["checked"] = this->mChecked;
		v["checkable"] = this->mCheckable;
		v["clickable"] = this->mClickable;
		v["enabled"] = this->mEnable;
		v["focusable"] = this->mFocusable;
		v["focused"] = this->mFocused;
		v["scrollable"] = this->mScrollable;
		v["longclickable"] = this->mLongClickable;
		v["password"] = this->mPassword;
		v["selected"] = this->mSelected;
		v["visible"] = this->mVisible;
		bounds["l"] = mBounds.l;
		bounds["t"] = mBounds.t;
		bounds["r"] = mBounds.r;
		bounds["b"] = mBounds.b;
		v["visible"] = this->mVisible;
		v["bounds"] = bounds;
		return v;
	}

	~NodeObject(){
		if (!m_childs.empty()) {
			vector<NodeObject*>::iterator it = m_childs.begin();
			while (it != m_childs.end()) {
				NodeObject* filter = *it;
				if (filter) {
					delete filter;
				}
				it++;
			}
			m_childs.clear();
		}
	}
};

class NodeFinder{
public:
	virtual bool isInclude(NodeObject* obj){
		return false;
	}
};

static pthread_mutex_t g_sel_mutex = PTHREAD_MUTEX_INITIALIZER;

class NodeSelector{
public:
	NodeSelector():mHandle(0){}
	~NodeSelector(){
		clear();
	}
	void clear(){
		if (!finders.empty()) {
			vector<NodeFinder*>::iterator it = finders.begin();
			while (it != finders.end()) {
				NodeFinder* filter = *it;
				if (filter) {
					delete filter;
				}
				it++;
			}
			finders.clear();
		}
	}
	DLONG mHandle;
	vector<NodeFinder*> finders;
	map<DLONG,int>       mResultCache;
};

static map<DLONG,NodeSelector*> g_selCaches;
static map<DLONG,int>           g_resultCaches;
static pthread_mutex_t g_selcache_mutex = PTHREAD_MUTEX_INITIALIZER;


static NodeObject* obtainNode(DLONG handle){
	 char* str = dumpNodeXml(handle,true);
	 if(str){
		 std::string xml = str;
		 rapidxml::xml_document<char> xml_doc;
		 NodeObject* pRoot = NULL;
		 try {
			 std::vector<char> data(xml.begin(), xml.end());
			 data.emplace_back('\0');
			 xml_doc.parse<0>(data.data());
			 rapidxml::xml_node<>* p = xml_doc.first_node();
			 rapidxml::xml_node<>* rootNode = NULL;
			 if (p && (rootNode = p->first_node("node")) != NULL) {
				 rapidxml::xml_node<>* pc = rootNode = p->first_node("node");
				 if(pRoot == NULL){
					pRoot = new NodeObject(NULL);
					pRoot->mRoot = pRoot;
					pRoot->mHandle = handle;
					
				 }
				 for(;pc;pc = pc->next_sibling()){
					 std::stack<NodeObject*> uiobjStack;
					 NodeObject* item = new NodeObject(pc);
					 uiobjStack.push(item);
					 item->mRoot = pRoot;
					 item->mHandle = handle;
					 item->mParent = pRoot;
					 pRoot->m_childs.push_back(item);
					 while (!uiobjStack.empty()) {
						 NodeObject* currentNode = uiobjStack.top();
						 uiobjStack.pop();
						 if (currentNode) {
							 for (rapidxml::xml_node<>* childNode 
								 = currentNode->mNode->first_node();
								 childNode; childNode = childNode->next_sibling()) {
								 NodeObject* item = new NodeObject(childNode);
								 if(item){
									 item->mRoot = pRoot;
									 item->mHandle = handle;
									 item->mParent = currentNode;
									 currentNode->m_childs.push_back(item);
									 uiobjStack.push(item);
								 }
							 }
						 }
					 }
				 }
			 }
			 free(str);
			 return pRoot;
		 }
		 catch (rapidxml::parse_error e) {
		 }
		 if(pRoot){
			delete pRoot;
		 }
		 free(str);
	 }
	 return NULL;
}

static int find(NodeObject* root,NodeSelector* sel,int limit,vector<NodeObject*>& arr) {
	if (root) {
		std::stack<NodeObject*> nodestack;
		nodestack.push(root);
		while (!nodestack.empty()) {
			NodeObject* obj = nodestack.top();
			nodestack.pop();
			if (obj) {
				std::vector<NodeObject*> vList = obj->m_childs;
				int cnt = vList.size();
				for (int i = cnt-1; i >= 0; i--) {
					nodestack.push(vList[i]);
				}
				bool bFind = true;
				if (!sel->finders.empty()) {
					vector<NodeFinder*>::iterator it = sel->finders.begin();
					while (it != sel->finders.end()) {
						NodeFinder* filter = *it;
						bool r = filter->isInclude(obj);
						if (!r) {
							bFind = false;
							break;
						}
						it++;
					}
				}
				if (bFind) {
					arr.push_back(obj);
					if (arr.size() >= limit) {
						break;
					}
				}
			}
		}
	}
	return arr.size();
}

NodeSelector* obtainSelector(DLONG handle){
	NodeSelector* ret = NULL;
	pthread_mutex_lock(&g_selcache_mutex);
	map<DLONG,NodeSelector*>::iterator it = 
		g_selCaches.find(handle);
	if(it != g_selCaches.end()){
		ret = it->second;
	}
	pthread_mutex_unlock(&g_selcache_mutex);
	return ret;
}

#define EQ_FINDER(ID) \
class ID##NodeFinder:public NodeFinder{ \
public:\
	std::string mStr;\
	ID##NodeFinder(const char* str){\
		if(str){\
			mStr = str;\
		}\
	}\
	virtual ~ID##NodeFinder(){} \
	virtual bool isInclude(NodeObject* obj){\
		if(equalWith(mStr,obj->m##ID)) return true; \
		return false; \
	}\
};

#define STARTWITH_FINDER(ID) \
class ID##StartWithNodeFinder:public NodeFinder{ \
public:\
	std::string mStr;\
	ID##StartWithNodeFinder(const char* str){\
		if(str){\
			mStr = str;\
		}\
	}\
	virtual ~ID##StartWithNodeFinder(){} \
	virtual bool isInclude(NodeObject* obj){\
		if(startsWith(obj->m##ID,mStr)) return true; \
		return false; \
	}\
};

#define ENDWITH_FINDER(ID) \
class ID##EndsWithNodeFinder:public NodeFinder{ \
public:\
	std::string mStr;\
	ID##EndsWithNodeFinder(const char* str){\
		if(str){\
			mStr = str;\
		}\
	}\
	virtual ~ID##EndsWithNodeFinder(){} \
	virtual bool isInclude(NodeObject* obj){\
		if(endsWith(obj->m##ID,mStr)) return true; \
		return false; \
	}\
};

#define CONTAINWITH_FINDER(ID) \
class ID##ContainWithNodeFinder:public NodeFinder{ \
public:\
	std::string mStr;\
	ID##ContainWithNodeFinder(const char* str){\
		if(str){\
			mStr = str;\
		}\
	}\
	virtual ~ID##ContainWithNodeFinder(){} \
	virtual bool isInclude(NodeObject* obj){\
		if(contains(obj->m##ID,mStr)) return true; \
		return false; \
	}\
};

#define MATCH_FINDER(ID) \
class ID##MatchNodeFinder:public NodeFinder{ \
public:\
	std::string mStr;\
	ID##MatchNodeFinder(const char* str){\
		if(str){\
			mStr = str;\
		}\
	}\
	virtual ~ID##MatchNodeFinder(){} \
	virtual bool isInclude(NodeObject* obj){\
		if(matcher(obj->m##ID,mStr)) return true; \
		return false; \
	}\
};

#define BOOLEAN_FINDER(ID) \
class ID##BoolNodeFinder:public NodeFinder{ \
public:\
	bool bVal;\
	ID##BoolNodeFinder(bool r){\
			bVal = r;\
	}\
	virtual bool isInclude(NodeObject* obj){\
	if(obj->m##ID == bVal) return true; \
		return false; \
	}\
};

#define INTEGER_FINDER(ID) \
class ID##IntegerNodeFinder:public NodeFinder{ \
public:\
	int iVal;\
	ID##IntegerNodeFinder(int r){\
			iVal = r;\
	}\
	virtual bool isInclude(NodeObject* obj){\
	if(obj->m##ID == iVal) return true; \
		return false; \
	}\
};


#define SEL(ID) \
	EQ_FINDER(ID) \
	STARTWITH_FINDER(ID) \
	ENDWITH_FINDER(ID) \
	CONTAINWITH_FINDER(ID) \
	MATCH_FINDER(ID) \
	LIBMYTRPC_API void ID##Equal(DLONG handle, const char* str) {\
		NodeSelector* selector = obtainSelector(handle); \
		if(selector){\
			selector->finders.push_back(new ID##NodeFinder(str));				\
		}\
	}\
	LIBMYTRPC_API void ID##StartWith(DLONG handle, const char* str) {\
		NodeSelector* selector = obtainSelector(handle); \
		if(selector){\
			selector->finders.push_back(new ID##StartWithNodeFinder(str));				\
		}\
	}\
	LIBMYTRPC_API void ID##EndWith(DLONG handle, const char* str) {\
		NodeSelector* selector = obtainSelector(handle); \
		if(selector){\
			selector->finders.push_back(new ID##EndsWithNodeFinder(str));				\
		}\
	}\
	LIBMYTRPC_API void ID##ContainWith(DLONG handle, const char* str) {\
		NodeSelector* selector = obtainSelector(handle); \
		if(selector){\
			selector->finders.push_back(new ID##ContainWithNodeFinder(str));				\
		}\
	}\
	LIBMYTRPC_API void ID##MatchWith(DLONG handle, const char* str) {\
		NodeSelector* selector = obtainSelector(handle); \
		if(selector){\
			selector->finders.push_back(new ID##MatchNodeFinder(str));				\
		}\
	}

#define BOOLSEL(ID) \
	BOOLEAN_FINDER(ID) \
	LIBMYTRPC_API void ID(DLONG handle, bool r) {\
		NodeSelector* selector = obtainSelector(handle); \
		if(selector){\
			selector->finders.push_back(new ID##BoolNodeFinder(r));				\
		}\
	}

#define INTEGERSEL(ID) \
	INTEGER_FINDER(ID) \
	LIBMYTRPC_API void ID(DLONG handle, int r) {\
		NodeSelector* selector = obtainSelector(handle); \
		if(selector){\
			selector->finders.push_back(new ID##IntegerNodeFinder(r));				\
		}\
	}

class BoundsNodeFinder :public NodeFinder {
public:
	BoundsNodeFinder(NodeBounds& rc) :m_rc(rc) {}
	virtual bool isInclude(NodeObject* node) {
		if (node->mBounds.isEqual(m_rc)) {
			return true;
		}
		return false;
	}

private:
	NodeBounds m_rc;
};

class BoundsInsideNodeFinder :public NodeFinder {
public:
	BoundsInsideNodeFinder(NodeBounds& rc) :m_rc(rc) {}
	virtual ~BoundsInsideNodeFinder() {};
	virtual bool isInclude(NodeObject* node) {
		if (m_rc.IsInRect(node->mBounds)) {
			return true;
		}
		return false;
	}
	 
private:
	NodeBounds m_rc;
};

static int findNodesImp(DLONG sel,int cnt,vector<NodeObject*>& vList,int timeout){
	DLONG current = getCurrentMilliseconds();
	NodeSelector* selector = obtainSelector(sel);
	if(selector){
		while (true) {
			NodeObject* pRoot = obtainNode(selector->mHandle);
			if(pRoot){
				int n = find(pRoot,selector, cnt, vList); 
				if (n > 0) {
					return n;
				}
				delete pRoot;
			}
			if (timeout != -1) {
				DLONG now = getCurrentMilliseconds();
				if ((now - current) >= timeout) {
					break; 
				}
			}
			sleepMilliseconds(50);
		}
	}
	return vList.size();
}

class NodeResult{
public:
	vector<NodeObject*> vList;
	NodeResult(){
	}
	~NodeResult(){
		if(vList.size() > 0){
			if(vList.at(0)->mRoot){
				delete vList.at(0)->mRoot;
			}
		}
	}
};

extern "C"{
	SEL(Id)
	SEL(Text)
	SEL(Clz)
	SEL(Package)
	SEL(Desc)
	BOOLSEL(Enable)
	BOOLSEL(Checkable)
	BOOLSEL(Clickable)
	BOOLSEL(Focusable)
	BOOLSEL(Focused)
	BOOLSEL(Scrollable)
	BOOLSEL(LongClickable)
	BOOLSEL(Password)
	BOOLSEL(Selected)
	BOOLSEL(Visible)
	INTEGERSEL(Index)

	LIBMYTRPC_API void BoundsInside(DLONG handle, int l,int t,int r,int b) {
		NodeSelector* selector = obtainSelector(handle);
		if(selector){
			NodeBounds bounds(l,t,r,b);
			BoundsInsideNodeFinder* finder = new BoundsInsideNodeFinder(bounds);
			selector->finders.push_back(finder);			
		}
	}

	LIBMYTRPC_API void BoundsEqual(DLONG handle, int l,int t,int r,int b) {
		NodeSelector* selector = obtainSelector(handle);
		if(selector){
			NodeBounds bounds(l,t,r,b);
			BoundsNodeFinder* finder = new BoundsNodeFinder(bounds);
			selector->finders.push_back(finder);			
		}
	}

	LIBMYTRPC_API DLONG newSelector(DLONG handle){
		//findOf(handle,sel,1,timeout);
		pthread_mutex_lock(&g_selcache_mutex);
		NodeSelector* sel = new NodeSelector();
		sel->mHandle = handle;
		g_selCaches[(DLONG)sel] = sel;
		pthread_mutex_unlock(&g_selcache_mutex);
		return (DLONG)sel;
	}

	LIBMYTRPC_API void clearSelector(DLONG handle){
		//findOf(handle,sel,1,timeout);
		pthread_mutex_lock(&g_selcache_mutex);
		map<DLONG,NodeSelector*>::iterator it = 
			g_selCaches.find(handle);
		if(it != g_selCaches.end()){
			it->second->clear();
		}
		pthread_mutex_unlock(&g_selcache_mutex);
	}

	LIBMYTRPC_API void freeSelector(DLONG handle){
		//findOf(handle,sel,1,timeout);
		pthread_mutex_lock(&g_selcache_mutex);
		map<DLONG,NodeSelector*>::iterator it = 
			g_selCaches.find(handle);
		if(it != g_selCaches.end()){
			delete it->second;
			g_selCaches.erase(it);
		}
		pthread_mutex_unlock(&g_selcache_mutex);
	}

	LIBMYTRPC_API DLONG findNodes(DLONG sel,int maxCntRet,int timeout){
		NodeResult* r = new NodeResult();
		if(!r){
			return 0L;
		}
		//char log[1024] = {0};
		//sprintf(log,"findNodes begin=>%lld %d",(DLONG)sel,sizeof(DLONG));
		//MessageBox(::GetDesktopWindow(),log,"",MB_OK);
		int n = 
			findNodesImp(sel,maxCntRet,r->vList,timeout);
		if(n > 0){
			pthread_mutex_lock(&g_node_mutex);
			g_resultCaches[(DLONG)r] = 1;
			pthread_mutex_unlock(&g_node_mutex);
		}else{
			delete r;
			return 0L;
		}
		//char log[1024] = {0};
		//sprintf(log,"findNodes ret=>%lld %d",(DLONG)r,sizeof(DLONG));
		//MessageBox(::GetDesktopWindow(),log,"",MB_OK);
		return (DLONG)r;//0x0000012e80d380e0
	}

	LIBMYTRPC_API DLONG getNodesSize(DLONG nodes){
		//char log[1024] = {0};
		//sprintf(log,"getNodesSize ret=>%ld",nodes);
		//MessageBox(::GetDesktopWindow(),log,"",MB_OK); 
		int size = 0;
		pthread_mutex_lock(&g_node_mutex);
		map<DLONG,int>::iterator it = 
			g_resultCaches.find(nodes);
		if(it != g_resultCaches.end()){
			DLONG hd = it->first;
			NodeResult* r = (NodeResult*)nodes;
			size = r->vList.size();
		}
		pthread_mutex_unlock(&g_node_mutex);
		//sprintf(log,"getNodesSize end ret %ld=>%ld",nodes,size);
		//MessageBox(::GetDesktopWindow(),log,"",MB_OK);
		return size;
	}

	LIBMYTRPC_API DLONG getNodeByIndex(DLONG nodes,int index){
		DLONG ret = 0;
		pthread_mutex_lock(&g_node_mutex);
		map<DLONG,int>::iterator it = 
			g_resultCaches.find(nodes);
		if(it != g_resultCaches.end()){
			NodeResult* r = (NodeResult*)it->first;
			if(r->vList.size() > index && index >= 0){
				ret = (DLONG)r->vList.at(index);
			}
		}
		pthread_mutex_unlock(&g_node_mutex);
		return ret;
	}

	LIBMYTRPC_API char* getNodeJson(DLONG node){
		char* ret = NULL;
		NodeObject* obj = (NodeObject*)node;
		if(obj){
			ret = strdup(obj->toJson().toStyledString().c_str());
		}
		return ret;
	}

	LIBMYTRPC_API char* getNodeText(DLONG node){
		char* ret = NULL;
		NodeObject* obj = (NodeObject*)node;
		if(obj){
			ret = strdup(obj->mText.c_str());
		}
		return ret;
	}

	LIBMYTRPC_API char* getNodeDesc(DLONG node){
		char* ret = NULL;
		NodeObject* obj = (NodeObject*)node;
		if(obj){
			ret = strdup(obj->mDesc.c_str());
		}
		return ret;
	}

	LIBMYTRPC_API char* getNodePackage(DLONG node){
		char* ret = NULL;
		NodeObject* obj = (NodeObject*)node;
		if(obj){
			ret = strdup(obj->mPackage.c_str());
		}
		return ret;
	}

	LIBMYTRPC_API char* getNodeClass(DLONG node){
		char* ret = NULL;
		NodeObject* obj = (NodeObject*)node;
		if(obj){
			ret = strdup(obj->mClz.c_str());
		}
		return ret;
	}

	LIBMYTRPC_API char* getNodeId(DLONG node){
		char* ret = NULL;
		NodeObject* obj = (NodeObject*)node;
		if(obj){
			ret = strdup(obj->mId.c_str());
		}
		return ret;
	}

	LIBMYTRPC_API int getNodeNound(DLONG node,int* l,int* t,int* r,int* b){
		char* ret = NULL;
		NodeObject* obj = (NodeObject*)node;
		if(obj){
			if(l) *l = obj->mBounds.l;
			if(t) *t = obj->mBounds.t;
			if(r) *r = obj->mBounds.r;
			if(b) *b = obj->mBounds.b;
			return 1;
		}
		return 0;
	}

	LIBMYTRPC_API int getNodeNoundCenter(DLONG node,int* x,int* y){
		char* ret = NULL;
		NodeObject* obj = (NodeObject*)node;
		if(obj){
			if(x) *x = (obj->mBounds.l + obj->mBounds.r) / 2;
			if(y) *y = (obj->mBounds.t + obj->mBounds.b) / 2;
			return 1;
		}
		return 0;
	}

    LIBMYTRPC_API int clickNode(DLONG node){
		int ret = 0;
		NodeObject* obj = (NodeObject*)node;
		if(obj){
			int x = (obj->mBounds.l + obj->mBounds.r) / 2;
			int y = (obj->mBounds.t + obj->mBounds.b) / 2;
			touchClick(obj->mHandle,8,x,y);
			ret = 1;
		}
		return ret;
	}

	 LIBMYTRPC_API int longClickNode(DLONG node){
		int ret = 0;
		NodeObject* obj = (NodeObject*)node;
		if(obj){
			int x = (obj->mBounds.l + obj->mBounds.r) / 2;
			int y = (obj->mBounds.t + obj->mBounds.b) / 2;
			touchDown(obj->mHandle,8,x,y);
			sleepMilliseconds(3000);
			touchUp(obj->mHandle,8,x,y);
			ret = 1;
		}
		return ret;
	}

	LIBMYTRPC_API DLONG  getNodeParent(DLONG node){
		DLONG ret = 0;
		NodeObject* obj = (NodeObject*)node;
		if(obj){
			return (DLONG)obj->mParent;
		}
		return ret;
	}


	LIBMYTRPC_API DLONG getNodeChildCount(DLONG node){
		DLONG ret = 0;
		NodeObject* obj = (NodeObject*)node;
		if(obj){
			ret = obj->m_childs.size();
		}
		return ret;
	}

	LIBMYTRPC_API DLONG  getNodeChild(DLONG node,int index){
		DLONG ret = 0;
		NodeObject* obj = (NodeObject*)node;
		if(obj){
			DLONG size = obj->m_childs.size();
			if(index >= 0 && index < size){
				return (DLONG)obj->m_childs.at(index);
			}
		}
		return ret;
	}

	LIBMYTRPC_API void freeNodes(DLONG nodes){
		if(nodes == 0){
			return;
		}
		pthread_mutex_lock(&g_node_mutex);
		map<DLONG,int>::iterator it = 
			g_resultCaches.find(nodes);
		if(it != g_resultCaches.end()){
			NodeResult* r = (NodeResult*)it->first;
			g_resultCaches.erase(it);
			delete r;
		}
		pthread_mutex_unlock(&g_node_mutex);
	}
}