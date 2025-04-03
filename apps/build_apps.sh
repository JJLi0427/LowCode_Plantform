#!/bin/bash

if [ -f ip_tracker/ip_tracker.cpp ]; then
    mkdir -p ip_tracker/bin
    if pkg-config --exists libcurl; then
        g++ -std=c++11 -o ip_tracker/bin/ip_tracker ip_tracker/ip_tracker.cpp -lcurl
    else
        echo "lcurl lib not found."
    fi
else
    echo "ip_tracker.cpp not found."
fi

if [ -f image_converter/image_converter.cpp ]; then
    mkdir -p image_converter/bin
    if pkg-config --exists opencv4; then
        g++ -std=c++11 -o image_converter/bin/image_converter image_converter/image_converter.cpp `pkg-config --cflags --libs opencv4`
    else
        echo "opencv lib not found."
    fi
else
    echo "image_converter.cpp not found."
fi

if [ -f dbuser_manager/dbuser_manager.cpp ]; then
    mkdir -p apps/dbuser_manager/bin
    if pkg-config --exists mysqlclient; then
        g++ -std=c++11 -o apps/dbuser_manager/bin/dbuser_manager apps/dbuser_manager/dbuser_manager.cpp -lmysqlclient
    else
        echo "mysql lib not found."
    fi
else
    echo "dbuser_manager.cpp not found."
fi