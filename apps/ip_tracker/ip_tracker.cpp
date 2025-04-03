#include <iostream>
#include <fstream>
#include <vector>
#include <string>
#include <sstream>
#include <stdexcept>
#include <ifaddrs.h>
#include <arpa/inet.h>
#include <curl/curl.h>
#include <netdb.h> // Added to resolve NI_MAXHOST and NI_NUMERICHOST

using namespace std;

// Helper function to fetch public IP
size_t WriteCallback(void* contents, size_t size, size_t nmemb, string* userp) {
    userp->append((char*)contents, size * nmemb);
    return size * nmemb;
}

string getPublicIP(const string& url) {
    CURL* curl;
    CURLcode res;
    string readBuffer;

    curl = curl_easy_init();
    if (curl) {
        curl_easy_setopt(curl, CURLOPT_URL, url.c_str());
        curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, WriteCallback);
        curl_easy_setopt(curl, CURLOPT_WRITEDATA, &readBuffer);
        res = curl_easy_perform(curl);
        curl_easy_cleanup(curl);

        if (res != CURLE_OK) {
            throw runtime_error("Failed to fetch public IP: " + string(curl_easy_strerror(res)));
        }
    } else {
        throw runtime_error("Failed to initialize CURL");
    }

    return readBuffer;
}

// Helper function to fetch local IPs
void getLocalIPs(vector<string>& ipv4s, vector<string>& ipv6s) {
    struct ifaddrs* ifaddr;
    if (getifaddrs(&ifaddr) == -1) {
        throw runtime_error("Failed to get network interfaces");
    }

    for (struct ifaddrs* ifa = ifaddr; ifa != nullptr; ifa = ifa->ifa_next) {
        if (ifa->ifa_addr == nullptr) continue;

        int family = ifa->ifa_addr->sa_family;
        char host[NI_MAXHOST]; // Resolved undeclared identifier

        if (family == AF_INET) { // IPv4
            if (getnameinfo(ifa->ifa_addr, sizeof(struct sockaddr_in), host, NI_MAXHOST, nullptr, 0, NI_NUMERICHOST) == 0) {
                ipv4s.push_back(host);
            }
        } else if (family == AF_INET6) { // IPv6
            if (getnameinfo(ifa->ifa_addr, sizeof(struct sockaddr_in6), host, NI_MAXHOST, nullptr, 0, NI_NUMERICHOST) == 0) {
                ipv6s.push_back(host);
            }
        }
    }

    freeifaddrs(ifaddr);
}

int main(int argc, char* argv[]) {
    if (argc < 2) {
        cout << "Usage: ./ip_tracker [IPv4|IPv6|both]" << endl;
        return 1;
    }

    string mode = argv[1];
    vector<string> ipv4s, ipv6s;

    try {
        getLocalIPs(ipv4s, ipv6s);
    } catch (const exception& e) {
        cerr << "Error fetching local IPs: " << e.what() << endl;
        return 1;
    }

    ofstream file("/tmp/ip.txt");
    if (!file.is_open()) {
        cerr << "Error creating file" << endl;
        return 1;
    }

    try {
        if (mode == "IPv4" || mode == "both") {
            file << "Local IPv4:\n";
            for (const auto& ip : ipv4s) {
                file << "    " << ip << "\n";
            }

            try {
                string publicIPv4 = getPublicIP("http://4.ipw.cn");
                file << "Global IPv4:\n    " << publicIPv4 << "\n";
            } catch (const exception& e) {
                file << "Get global IPv4 error: " << e.what() << "\n";
            }
        }

        if (mode == "IPv6" || mode == "both") {
            file << "Local IPv6:\n";
            for (const auto& ip : ipv6s) {
                file << "    " << ip << "\n";
            }

            try {
                string publicIPv6 = getPublicIP("http://6.ipw.cn");
                file << "Global IPv6:\n    " << publicIPv6 << "\n";
            } catch (const exception& e) {
                file << "Get global IPv6 error: " << e.what() << "\n";
            }
        }
    } catch (const exception& e) {
        cerr << "Error: " << e.what() << endl;
        return 1;
    }

    file.close();
    return 0;
}
