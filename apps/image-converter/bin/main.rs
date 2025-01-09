use image::imageops::FilterType;
use image::RgbaImage;
use std::env;
use std::path::Path;
use std::ffi::OsString;

fn main() {
    let (from, into, width, height, opt) = if env::args_os().count() == 6 {
        (
            env::args_os().nth(1).unwrap(),
            env::args_os().nth(2).unwrap(),
            env::args_os().nth(3).unwrap(),
            env::args_os().nth(4).unwrap(),
            env::args_os().nth(5)
        )
    } else {
        println!("Please enter a from & into path width & height [scale or tile]");
        println!("Supported Image Type: PNG JPEG GIF BMP ICO TIFF WebP AVIF PNM");
        println!("[Usage]: image format convert: [exe] iImageFile.jpg oImageFile.png 0 0 none");
        println!("[Usage]: image scale  convert: [exe] iImageFile.jpg oImageFile.png 640 480 scale");
        println!("[Usage]: image tile   convert: [exe] iImageFile.jpg oImageFile.png 2640 1480 tile");
        std::process::exit(1);
    };
    
    let sw = width.into_string();
    let sh = height.into_string();
    let optscale = OsString::from("scale");
    let w = sw.expect("REASON").parse::<u32>().unwrap();
    let h = sh.expect("REASON").parse::<u32>().unwrap();

    if w == 0 || h == 0 {
        let im = image::open(&Path::new(&from)).unwrap();
        im.save(&Path::new(&into)).unwrap();
    }
    else if opt == Some(optscale) {
        println!("scale:::");
        let scale = image::open(&Path::new(&from)).unwrap();
        let scaled = scale.resize(w, h, FilterType::Triangle);
        scaled.save(&Path::new(&into)).unwrap();
    }else{
        println!("tile:::");
        let tile = image::open(&Path::new(&from)).unwrap();

        let mut img = RgbaImage::new(w,h);
        image::imageops::tile(&mut img, &tile);
        img.save(&Path::new(&into)).unwrap();
    }

}
