#!/bin/bash

if [ -f ip_tracker/ip_tracker.cpp ]; then
    mkdir -p ip_tracker/bin
    g++ -std=c++11 -o ip_tracker/bin/ip_tracker ip_tracker/ip_tracker.cpp -lcurl
else
    echo "ip_tracker.cpp not found."
fi

if [ -f image_converter/image_converter.cpp ]; then
    mkdir -p image_converter/bin
    g++ -std=c++11 -o image_converter/bin/image_converter image_converter/image_converter.cpp `pkg-config --cflags --libs opencv4`
else
    echo "image_converter.cpp not found."
fi

if [ -f dbuser_manager/dbuser_manager.cpp ]; then
    mkdir -p dbuser_manager/bin
    g++ -std=c++11 -o dbuser_manager/bin/dbuser_manager dbuser_manager/dbuser_manager.cpp -lmysqlclient
else
    echo "dbuser_manager.cpp not found."
fi