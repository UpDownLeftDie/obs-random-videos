# OBS Random Video Player

Play videos in random order!

*Perfect for BRB screens!*

## Instructions

1. Copy the `index.html` file or its code
2. Edit the `videosFolder` to point to the folder of videos you want
   1. You can set this to `""` if you're placing the html file in the same folder as your videos
3. Add all the videos from that folder to the `videosList`
4. Set a `transitionVideo` or set it to `""` to disable it
5. If you only want one video to play set `playOnlyOne` to `true`
6. Add a new `Browser Source` to OBS
   1. Set it to "Local File" and select `index.html` you just edited
   2. Set the width and height to your full canvas resolution
   3. Check "Shutdown source when not visible"
   4. If you want a new shuffle each time: check "Refresh browser when scene becomes active"

## Notes

* Set `Refresh browser when scene becomes active` to randomize on each load
* mp3 file should work too
* **Autoplay only works in OBS!**
  * If you want to test this in your browser:
    1. Open the `index.html` with your browser
    2. Right-click on the image and click "Show controls"
    3. Hit the play button
