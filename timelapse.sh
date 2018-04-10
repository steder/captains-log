#!/usr/bin/env bash

ffmpeg -r 30 -s 800x600 -i snap_%05d.jpg -vcodec libx264 -pix_fmt yuv420p -preset veryslow -crf 18 timelapse.mp4