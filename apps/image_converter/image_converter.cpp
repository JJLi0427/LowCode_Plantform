#include <opencv2/opencv.hpp>
#include <iostream>
#include <string>

void print_usage() {
    std::cout << "Please enter a from & into path width & height [scale or tile]\n";
    std::cout << "Supported Image Type: PNG JPEG BMP TIFF\n";
    std::cout << "[Usage]: image format convert: [exe] iImageFile.jpg oImageFile.png 0 0 none\n";
    std::cout << "[Usage]: image scale  convert: [exe] iImageFile.jpg oImageFile.png 640 480 scale\n";
    std::cout << "[Usage]: image tile   convert: [exe] iImageFile.jpg oImageFile.png 2640 1480 tile\n";
}

int main(int argc, char* argv[]) {
    if (argc != 6) {
        print_usage();
        return 1;
    }

    std::string from = argv[1];
    std::string into = argv[2];
    int width = std::stoi(argv[3]);
    int height = std::stoi(argv[4]);
    std::string opt = argv[5];

    cv::Mat image = cv::imread(from);
    if (image.empty()) {
        std::cerr << "Error: Could not open or find the image.\n";
        return 1;
    }

    if (width == 0 || height == 0) {
        // Format conversion
        if (!cv::imwrite(into, image)) {
            std::cerr << "Error: Could not save the image.\n";
            return 1;
        }
    } else if (opt == "scale") {
        // Scaling
        std::cout << "scale:::\n";
        cv::Mat scaled;
        cv::resize(image, scaled, cv::Size(width, height), 0, 0, cv::INTER_LINEAR);
        if (!cv::imwrite(into, scaled)) {
            std::cerr << "Error: Could not save the scaled image.\n";
            return 1;
        }
    } else if (opt == "tile") {
        // Tiling
        std::cout << "tile:::\n";
        cv::Mat tiled(height, width, image.type());
        for (int y = 0; y < height; y += image.rows) {
            for (int x = 0; x < width; x += image.cols) {
                cv::Rect roi(x, y, std::min(image.cols, width - x), std::min(image.rows, height - y));
                image(cv::Rect(0, 0, roi.width, roi.height)).copyTo(tiled(roi));
            }
        }
        if (!cv::imwrite(into, tiled)) {
            std::cerr << "Error: Could not save the tiled image.\n";
            return 1;
        }
    } else {
        std::cerr << "Error: Invalid option. Use 'scale' or 'tile'.\n";
        return 1;
    }

    return 0;
}
