#include <mysql/mysql.h>
#include <iostream>
#include <stdexcept>
#include <string>
#include <cstdlib>
#include <fstream>

class UserPermissionManager {
private:
    MYSQL* conn;
    std::ofstream logFile;

public:
    UserPermissionManager(const std::string& host, unsigned int port, 
        const std::string& user, const std::string& password
    ) : logFile("/tmp/dbuser_manager.log", std::ios::out) {
        if (!logFile.is_open()) {
            throw std::runtime_error("Failed to open log file");
        }
        conn = mysql_init(nullptr);
        if (!conn) {
            Log("MySQL initialization failed");
            throw std::runtime_error("MySQL initialization failed");
        }
        if (!mysql_real_connect(conn, host.c_str(), user.c_str(), password.c_str(), nullptr, port, nullptr, 0)) {
            Log(std::string("MySQL connection failed: ") + mysql_error(conn));
            throw std::runtime_error(mysql_error(conn));
        }
        Log("MySQL connection established successfully");
    }

    ~UserPermissionManager() {
        if (conn) {
            mysql_close(conn);
            Log("MySQL connection closed");
        }
        if (logFile.is_open()) {
            logFile.close();
        }
    }

    void AddUser(const std::string& username, const std::string& password) {
        std::string query = "CREATE USER '" + username + "'@'%' IDENTIFIED BY '" + password + "';";
        ExecuteQuery(query, "User " + username + " added successfully.");
    }

    void DeleteUser(const std::string& username) {
        std::string query = "DROP USER '" + username + "'@'%';";
        ExecuteQuery(query, "User " + username + " deleted successfully.");
    }

    void GrantPermission(const std::string& username, const std::string& dbName, 
        const std::string& tableName, const std::string& permission
    ) {
        std::string query = "GRANT " + permission + " ON " + dbName + "." + tableName + " TO '" + username + "'@'%';";
        ExecuteQuery(query, "Granted " + permission + " permission on " + dbName + "." + tableName + " to user " + username + ".");
    }

    void RevokePermission(const std::string& username, const std::string& dbName, 
        const std::string& tableName, const std::string& permission
    ) {
        std::string query = "REVOKE " + permission + " ON " + dbName + "." + tableName + " FROM '" + username + "'@'%';";
        ExecuteQuery(query, "Revoked " + permission + " permission on " + dbName + "." + tableName + " from user " + username + ".");
    }

    void QueryUserPermissions(const std::string& username) {
        std::string query = "SHOW GRANTS FOR '" + username + "'@'%';";
        Log("Executing query: " + query);
        if (mysql_query(conn, query.c_str())) {
            std::string errorMessage = mysql_error(conn);
            Log("Query failed: " + errorMessage);
            throw std::runtime_error(errorMessage);
        }

        MYSQL_RES* result = mysql_store_result(conn);
        if (!result) {
            std::string errorMessage = mysql_error(conn);
            Log("Failed to retrieve query result: " + errorMessage);
            throw std::runtime_error(errorMessage);
        }

        MYSQL_ROW row;
        std::cout << "Permissions for user '" << username << "':" << std::endl;
        while ((row = mysql_fetch_row(result))) {
            std::cout << row[0] << std::endl;
            Log("Permission: " + std::string(row[0]));
        }

        mysql_free_result(result);
    }

private:
    void ExecuteQuery(const std::string& query, const std::string& successMessage) {
        Log("Executing query: " + query);
        if (mysql_query(conn, query.c_str())) {
            std::string errorMessage = mysql_error(conn);
            Log("Query failed: " + errorMessage);
            throw std::runtime_error(errorMessage);
        }
        Log(successMessage);
        std::cout << successMessage << std::endl;
    }

    void Log(const std::string& message) {
        if (logFile.is_open()) {
            logFile << message << std::endl;
        }
    }
};

int main(int argc, char* argv[]) {
    if (argc < 6) {
        std::cerr << "Usage: <host> <port> <user> <password> <op> [additional arguments]" << std::endl;
        return 1;
    }

    try {
        std::string host = argv[1];
        unsigned int port = std::stoi(argv[2]);
        std::string user = argv[3];
        std::string password = argv[4];
        
        UserPermissionManager manager(host, port, user, password);

        int op = std::stoi(argv[5]);
        switch (op) {
            case 1: // Add User
                if (argc < 8) {
                    std::cerr << "Usage for Add User: 1 <username> <password>" << std::endl;
                    return 1;
                }
                {
                    std::string username = argv[6];
                    std::string userPassword = argv[7];
                    manager.AddUser(username, userPassword);
                }
                break;
            case 2: // Delete User
                if (argc < 7) {
                    std::cerr << "Usage for Delete User: 2 <username>" << std::endl;
                    return 1;
                }
                {
                    std::string username = argv[6];
                    manager.DeleteUser(username);
                }
                break;
            case 3: // Grant Permission
                if (argc < 10) {
                    std::cerr << "Usage for Grant Permission: 3 <username> <dbName> <tableName> <permission>" << std::endl;
                    return 1;
                }
                {
                    std::string username = argv[6];
                    std::string dbName = argv[7];
                    std::string tableName = argv[8];
                    std::string permission = argv[9];
                    manager.GrantPermission(username, dbName, tableName, permission);
                }
                break;
            case 4: // Revoke Permission
                if (argc < 10) {
                    std::cerr << "Usage for Revoke Permission: 4 <username> <dbName> <tableName> <permission>" << std::endl;
                    return 1;
                }
                {
                    std::string username = argv[6];
                    std::string dbName = argv[7];
                    std::string tableName = argv[8];
                    std::string permission = argv[9];
                    manager.RevokePermission(username, dbName, tableName, permission);
                }
                break;
            case 5: // Query User Permissions
                if (argc < 7) {
                    std::cerr << "Usage for Query User Permissions: 5 <username>" << std::endl;
                    return 1;
                }
                {
                    std::string username = argv[6];
                    manager.QueryUserPermissions(username);
                }
                break;
            default:
                std::cerr << "Invalid operation" << std::endl;
                std::cerr << "op: 1 - Add User, 2 - Delete User, 3 - Grant Permission, 4 - Revoke Permission, 5 - Query User Permissions" << std::endl;
                return 1;
        }

    } catch (const std::exception& ex) {
        std::cerr << "Error: " << ex.what() << std::endl;
    }

    return 0;
}
