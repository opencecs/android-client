#pragma once
#include <string>
using namespace std;

class PkgData
{
public:
	PkgData(int isFree = 1);
	PkgData(void* data,int size);
	~PkgData(void);
	int   WrapInt(int data);
	int   UnWrapInt(int& ret);
	int   WrapData(void* data,int len);
	void* UnWrapData(int& len);
	void* GetPtr(int& len);
	string UnWrapStr();
	void   Rewind();
private:
	int m_offset;
	int m_size;
	int m_isFree;
	unsigned char* m_ptr;
};
