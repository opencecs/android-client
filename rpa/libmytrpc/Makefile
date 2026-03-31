CROSS =
CC = $(CROSS)gcc
CXX = $(CROSS)g++
AR=$(CROSS)ar
RANLIB= $(CROSS)ranlib
STRIP=$(CROSS)strip
DEBUG = -g -O2
CFLAGS = $(DEBUG) -Wall -c -fPIC
RM = rm -rf
CP = cp
SRCS = $(wildcard ./*.cpp)
OBJS = $(patsubst %.cpp, %.o, $(SRCS))
OUTDIR = $(PWD)/out

HEADER_PATH = -I$(OUTDIR)/include
LIB_PATH =

LIBS = 

TARGET = libmytrpc.a

$(TARGET) : $(OBJS)
	$(AR) -r $(TARGET) *.o

$(OBJS):%.o : %.cpp
	$(CXX) $(CFLAGS) $< -o $@ $(HEADER_PATH)

clean:
	$(RM) $(TARGET) *.o
