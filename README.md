# OBS Random Video Player

Play videos in random order!

_Perfect for BRB screens!_

Notice: the VLC Source now supports "Shuffle playlist" option.

## [Download](https://github.com/UpDownLeftDie/obs-random-videos/releases/latest)

### [Discord Support](https://discord.gg/zxYsjpxaxN)

## Instructions

1. Download the latest executable for your system from above
2. Place the executable into the folder where your video/audio files you want to use are located
3. Run the file and follow the prompts for your settings
4. A new file called `obs-random-videos.html` will be created in the same folder
5. Add a new `Browser Source` to OBS
   1. Check "Local File" and select the new `obs-random-videos.html` you just created.
   2. **Copy this path**
   3. **Uncheck "Local file" and set the url to `file://<paste the path you copied>`**
      1. Windows example: `file://C:/obs-videos-folder/obs-random-videos.html`
      2. Mac/Linux example: `file:///obs-videos-folder/obs-random-videos.html` (note the `///`)
   4. Set the width and height to your full canvas resolution
   5. Check "Shutdown source when not visible"
   6. If you want a new shuffle each time: check "Refresh browser when scene becomes active"
   7. Leave "Page permissions" set to "Read access to OBS status information"
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

- VLC Source now supports "Shuffle playlist"
  - This may be a more stable alternative to this project
  - Mac's with M1 chip: Install the Intel version of VLC to use the VLC plugin
- Set `Refresh browser when scene becomes active`
- Video resolutions should match your canvas aspect ratio
- **Autoplay only works in OBS!**
  - To test in your browser you must hit the Play button first
- Pro tip: webm videos support transparency (convert mov to webm to save on file size)

## Stuck? Or nothing happening?

  1. Try hitting `Refresh cache of current page` in OBS
  2. [Enable remote debugging](https://github.com/crowbartools/Firebot/wiki/Troubleshooting-Firebot-Overlay-issues-in-OBS-Studio) and open the page for the browser source
     1. Open Chrome Dev tools
     2. `Application` tab
     3. Make sure `Local and session storage` box is CHECKED
     4. Click `Clear site data`
  3. If issues persist: join the [Discord server](https://discord.gg/zxYsjpxaxN)

## TODO

- Audio-only or video-only modes
- Option for cross-fading between videos
- Option to HTML background color
  - Good for videos with weird aspect ratio
  - Good for audio files
