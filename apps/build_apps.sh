#!/bin/bash

STATIC_FLAG=""
if [ "$1" = "-static" ]; then
    echo "Using static compilation"
    STATIC_FLAG="-static"
fi

if [ -f apps/ip_tracker/ip_tracker.cpp ]; then
    mkdir -p apps/ip_tracker/bin
    g++ -std=c++11 ${STATIC_FLAG} -o apps/ip_tracker/bin/ip_tracker apps/ip_tracker/ip_tracker.cpp -lcurl
else
    echo "ip_tracker.cpp not found."
fi

if [ -f apps/image_converter/image_converter.cpp ]; then
    mkdir -p apps/image_converter/bin
    g++ -std=c++11 ${STATIC_FLAG} -o apps/image_converter/bin/image_converter apps/image_converter/image_converter.cpp `pkg-config --cflags --libs opencv4`
else
    echo "image_converter.cpp not found."
fi

if [ -f apps/dbuser_manager/dbuser_manager.cpp ]; then
    mkdir -p apps/dbuser_manager/bin
    g++ -std=c++11 ${STATIC_FLAG} -o apps/dbuser_manager/bin/dbuser_manager apps/dbuser_manager/dbuser_manager.cpp -lmysqlclient
else
    echo "dbuser_manager.cpp not found."
fi