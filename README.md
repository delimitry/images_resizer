# images_resizer
A tool for downsizing images in Go (golang).

Description:
------------

Supported formats of images: `jpeg`, `png`, `gif`.  
Uses simple average color from 2x2 pixel block for smoothing.

Inside Go source file there are 3 functions:  
`resizeImages`, `asyncResizeImages` and `workersPoolResizeImages`

1. `resizeImages` - simply resizes images one by one. The slowest function. Consumes the least amount of memory.
2. `asyncResizeImages` - uses a goroutine for each image, i.e. spawns `len(imageFiles)` goroutines. This is the fastest (according to my profiling), but most memory consuming function.
3. `workersPoolResizeImages` - this function uses pool of workers to run several `workersNum` concurrent `resizeImage` functions. It is just a little bit slower than `asyncResizeImages` function, but consumes limited amount of memory (depends on images' size).

Now the tool uses `workersPoolResizeImages` function.

### Usage
```
Usage: images_resizer [options]

  -d string
        Directory with images to scale (default ".")
  -f float
        Scaling factor value from (0.0 to 1.0) (default 0.5)
```
And the output:
```
user@user:~$ ./images_resizer -d ~/images/ -f 0.5
--------------------------------------------------------------------------------
Resizing IMAG0123.JPEG (worker 4)
Resizing image.png (worker 0)
Resizing screen.gif (worker 2)
Resizing test.png (worker 3)
Resizing af12d3b95576a7.jpg (worker 1)
test.png OK [0]
af12d3b95576a7.jpg OK [1]
screen.gif OK [2]
image.png OK [3]
IMAG0123.JPEG OK [4]
--------------------------------------------------------------------------------
Resize images time: 4.143626s
```

License:
--------
Released under [The MIT License](https://github.com/delimitry/images_resizer/blob/master/LICENSE).
