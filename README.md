# OBS Random Video Player

Play videos in random order!

_Perfect for BRB screens!_

## [Download](https://github.com/UpDownLeftDie/obs-random-videos/releases/latest)

### [Chat Support](https://www.guilded.gg/i/pWrnl1qk?cid=377fa0a0-d505-4616-92cc-00d1d9f4e925&intent=forum)

## Instructions

1. Download the latest executable for your system from above
2. Place the executable into the folder where your video/audio files you want to use are located
3. Run the file and follow the prompts for your settings
4. A new file called `obs-random-videos.html` will be created in the same folder
5. Add a new `Browser Source` to OBS
   1. Set it to "Local File" and select the new `obs-random-videos.html` you just created
   2. Set the width and height to your full canvas resolution
   3. Check "Shutdown source when not visible"
   4. If you want a new shuffle each time: check "Refresh browser when scene becomes active"
6. Rerun the executable to change any settings

### Supported files

#### Video

- mp4
- m4v
- webm
- mpeg4

#### Audio

- mp3
- ogg
- aac

## Notes

- Set `Refresh browser when scene becomes active`
- Video resolutions should match your canvas aspect ratio
- **Autoplay only works in OBS!**
  - ~~If you want to test this in your browser:~~ Not working in v3.3.0
    1. ~~Open the `obs-random-videos.html` with your browser~~
    2. ~~Right-click on the image and click "Show controls"~~
    3. ~~Hit the play button~~
- Pro tip: webm videos support transparency (convert mov to webm to save on file size)

## Stuck? Or nothing happening?

  1. Try hitting `Refresh cache of current page` in OBS **at least** as many time as you have videos in your playlist
  2. [Enable remote debugging](https://github.com/crowbartools/Firebot/wiki/Troubleshooting-Firebot-Overlay-issues-in-OBS-Studio) and open the page for the browser source
     1. Open Chrome Dev tools
     2. `Application` tab
     3. Make sure `Local and session storage` box is CHECKED
     4. Click `Clear site data`
  3. If issues persist: join the [Guilded server](https://www.guilded.gg/i/pWrnl1qk?cid=377fa0a0-d505-4616-92cc-00d1d9f4e925&intent=forum)

## TODO

- Audio-only or video-only modes
- Option for cross-fading between videos
- Option to HTML background color
  - Good for videos with weird aspect ratio
  - Good for audio files
