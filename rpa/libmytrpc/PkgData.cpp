#include "stdafx.h"
#include "PkgData.h"
#include "stdlib.h"
#include <string.h>
#ifdef WIN32
#include <winsock.h>
#pragma comment(lib,"ws2_32")
#else
#include <arpa/inet.h>
#endif

#define MAGIC	0x55667788

#define GETINT(P) ((P[0] & 0x000000ff) | ((P[1] << 8) & 0x0000ff00) | ((P[2] << 16) & 0x00ff0000) | ((P[3] << 24) & 0xff000000))
#define SETINT(P,N) \
		P[0] = (unsigned char)(N & 0xff); \
		P[1] = (unsigned char)((N >> 8) & 0xff); \
		P[2] = (unsigned char)((N >> 16) & 0xff); \
		P[3] = (unsigned char)((N >> 24) & 0xff);

PkgData::PkgData(int isFree){
    m_offset = sizeof(int);
    m_size = m_offset;
    m_ptr = (unsigned char*)malloc(sizeof(int));
    //*(int*)m_ptr = htonl(MAGIC);
    SETINT(m_ptr,MAGIC);
    m_isFree = isFree;
}

PkgData::PkgData(void* data,int size){
    m_ptr = (unsigned char*)data;
    m_size = size;
    m_offset = sizeof(int);
    m_isFree = 0;
}

PkgData::~PkgData(void){
    if(m_isFree){
        free(m_ptr);
    }
}

void PkgData::Rewind(){
	m_offset = sizeof(int);
}

string PkgData::UnWrapStr(){
	int len = 0;
	void* ptr = UnWrapData(len);
	if(ptr){
		char* str = (char*)malloc(len + 1);
		if(str){
			memcpy(str,ptr,len);
			str[len] = 0;
			string ret = str;
			free(str);
			return ret;
		}
	}
	return "";
}

int PkgData::WrapData(void* data,int len){
    if(len <= 0)
        return -1;
    int size = m_size + len + sizeof(int);
    unsigned char* ptr = (unsigned char*)realloc(m_ptr,size);
    if(ptr != NULL){
        m_ptr = ptr;
        ptr += m_offset;
        //*(int*)ptr = htonl(len);
        SETINT(ptr,len);
        ptr += sizeof(int);
        memcpy(ptr,data,len);
        m_offset += (len + sizeof(int));
        m_size = m_offset;
        return 0;
    }
    return -1;
}

int PkgData::WrapInt(int data){
    //data = htonl(data);
    unsigned  char ptr[5] = {0};
    SETINT(ptr,data);
    return WrapData((void*)ptr,sizeof(int));
}

void* PkgData::GetPtr(int& len){
    len = m_size;
    return m_ptr;
}

int PkgData::UnWrapInt(int& ret){
    int len = 0;
    unsigned  char* ptr = (unsigned  char*)UnWrapData(len);
    if(ptr){
        //ret = htonl(*(int*)ptr);
        ret = GETINT(ptr);
        return 0;
    }else if(ptr){
        m_offset -= (len + sizeof(int));
    }
    return -1;
}

void* PkgData::UnWrapData(int& len){
    if(!m_ptr && m_size < sizeof(int))
        return NULL;
    if(m_offset >= m_size){
        len = 0;
        return NULL;
    }
    //int magic = ntohl(*(int*)m_ptr);
    int magic = GETINT(m_ptr);
    if(magic != MAGIC){
        return NULL;
    }
    unsigned char* ptr = m_ptr + m_offset;
    //len = ntohl((*int*)ptr);
    len = GETINT(ptr);
    ptr += sizeof(int);
    m_offset += (len + sizeof(int));
    return ptr;
}