#!/usr/bin/env bash

ffmpeg -r 30 -i snap.%05d.png -vcodec libx264 -preset veryslow -crf 18 timelapse.mp4