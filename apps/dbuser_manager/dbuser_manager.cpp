#include <mysql/mysql.h>
#include <iostream>
#include <stdexcept>
#include <string>

class UserPermissionManager {
private:
    MYSQL* conn;

public:
    UserPermissionManager(const std::string& host, const std::string& user, const std::string& password, unsigned int port) {
        conn = mysql_init(nullptr);
        if (!conn) {
            throw std::runtime_error("MySQL initialization failed");
        }
        if (!mysql_real_connect(conn, host.c_str(), user.c_str(), password.c_str(), nullptr, port, nullptr, 0)) {
            throw std::runtime_error(mysql_error(conn));
        }
    }

    ~UserPermissionManager() {
        if (conn) {
            mysql_close(conn);
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

    void GrantPermission(const std::string& username, const std::string& dbName, const std::string& tableName, const std::string& permission) {
        std::string query = "GRANT " + permission + " ON " + dbName + "." + tableName + " TO '" + username + "'@'%';";
        ExecuteQuery(query, "Granted " + permission + " permission on " + dbName + "." + tableName + " to user " + username + ".");
    }

    void RevokePermission(const std::string& username, const std::string& dbName, const std::string& tableName, const std::string& permission) {
        std::string query = "REVOKE " + permission + " ON " + dbName + "." + tableName + " FROM '" + username + "'@'%';";
        ExecuteQuery(query, "Revoked " + permission + " permission on " + dbName + "." + tableName + " from user " + username + ".");
    }

private:
    void ExecuteQuery(const std::string& query, const std::string& successMessage) {
        if (mysql_query(conn, query.c_str())) {
            throw std::runtime_error(mysql_error(conn));
        }
        std::cout << successMessage << std::endl;
    }
};

int main() {
    try {
        std::string host = "127.0.0.1";
        std::string user = "root";
        std::string password = "Lijia0427.";
        unsigned int port = 3306;

        UserPermissionManager manager(host, user, password, port);

        manager.AddUser("testuser", "testpassword");
        manager.GrantPermission("testuser", "testdb", "*", "SELECT, INSERT");
        manager.RevokePermission("testuser", "testdb", "*", "INSERT");
        manager.DeleteUser("testuser");

    } catch (const std::exception& ex) {
        std::cerr << "Error: " << ex.what() << std::endl;
    }

    return 0;
}
