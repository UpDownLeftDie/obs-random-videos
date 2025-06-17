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
      2. Mac/Linux example: `file:///obs-videos-folder/obs-random-videos.html` **(note the `///` instead of `//`)**
   4. Set the width and height to your full canvas resolution
   5. Check "Shutdown source when not visible"
   6. If you want a new shuffle each time: check "Refresh browser when scene becomes active"
   7. Leave "Page permissions" set to "Read access to OBS status information"
6. **Rerun the executable when you add or remove videos** or want to change any settings

### Supported files

#### Video

- mp4
- m4v
- webm
- mpeg4
- mov

#### Audio

- mp3
- ogg
- aac

## Project Structure

The project is organized into the following directories:

- `internal/` - Internal packages not meant for external use
  - `media/` - Media file handling functions (file detection, path handling)
  - `ui/` - User interface and template functions (prompts, HTML generation)
- `web/` - Web assets
  - `js/` - JavaScript files for the HTML player
    - `main.js` - Contains the VideoPlayerManager and PlaylistManager classes
    - `body.js` - Initializes the player
  - `templates/` - HTML templates

### JavaScript Architecture

The JavaScript code uses an object-oriented approach with two main classes:

1. **VideoPlayerManager**: Handles the dual-player system for seamless video playback
   - Manages the two video players
   - Handles transitions between videos
   - Provides error handling
   - Implements preloading for better performance

2. **PlaylistManager**: Manages the playlist state and operations
   - Handles shuffling of videos
   - Manages localStorage for persistence
   - Provides methods for getting the next video

## Notes

- VLC Source now supports "Shuffle playlist"
  - This may be a more stable alternative to this project
  - Mac's with M1 chip: Install the Intel version of VLC to use the VLC plugin
- Set `Refresh browser when scene becomes active`
- Video resolutions should match your canvas aspect ratio
- **Autoplay only works in OBS!**
  - To test in your browser you must hit the Play button first
- Pro tip: webm videos support transparency (convert mov to webm to save on file size)

## Troubleshooting

If you're experiencing issues with the player:

1. Try hitting `Refresh cache of current page` in OBS
2. **In OBS go to `Scene Collection` then `New`**
   1. **Add only your random video browser source and see if this works**
   2. **If it works, then theres an external issue.** I have no idea what is causing this so, if you have any info on this us know in the [discussions](https://github.com/UpDownLeftDie/obs-random-videos/discussions)
3. [Enable remote debugging](https://github.com/crowbartools/Firebot/wiki/Troubleshooting-Firebot-Overlay-issues-in-OBS-Studio) and open the page for the browser source
   1. Open Chrome Dev tools
   2. `Application` tab
   3. Make sure `Local and session storage` box is CHECKED
   4. Click `Clear site data`
4. Check if your video files are in a supported format and can be played in a browser
5. Ensure your video files don't have special characters in their filenames
6. If issues persist: use [GitHub Discussions](https://github.com/UpDownLeftDie/obs-random-videos/discussions/categories/q-a) or join the [Discord server](https://discord.gg/zxYsjpxaxN)

## TODO

- [x] Support for multiple video formats
- [x] Support for audio files
- [x] Switch to [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- [ ] Audio-only or video-only modes
- [ ] Option for cross-fading between videos
- [ ] Option to set HTML background color
  - Good for videos with weird aspect ratio
  - Good for audio files
- [ ] Switch to [Wails](https://wails.io/) UI for a graphical interface
- [ ] Add support for image files (with configurable display duration)
- [ ] Improve error handling and user feedback
