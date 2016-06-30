.PHONY: all

KOMPOSE_ENVS := \
	-e OS_PLATFORM_ARG \
	-e OS_ARCH_ARG \
	-e TESTDIRS \
	-e TESTFLAGS \
	-e TESTVERBOSE

BIND_DIR := bundles

default: binary

all:
	./script/make.sh

binary:
	./script/make.sh binary

clean:
	./script/make.sh clean

