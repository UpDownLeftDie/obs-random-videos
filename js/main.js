const mediaFolder = "{{ .MediaFolder }}";
const initMediaFiles =  ["{{ StringsJoin .MediaFiles "\", \"" }}"];
const transitionVideo = "{{ .TransitionVideo }}";
const playOnlyOne = {{ .PlayOnlyOne }};
const loopFirstVideo = {{ .LoopFirstVideo }};
const transitionVideoPath =  `${mediaFolder}${transitionVideo}`

const mediaFiles = shuffleArr(initMediaFiles);
let count = 0;
let isTransition = true;

function shuffleArr(a) {
  var j, x, i;
  for (i = a.length - 1; i > 0; i--) {
    j = Math.floor(Math.random() * (i + 1));
    x = a[i];
    a[i] = a[j];
    a[j] = x;
  }


  return a.map(b => `${mediaFolder}${b}`);
}

function playNext(player, nextPlayer) {
  const nextMp4Source = nextPlayer.getElementsByClassName("mp4Source")[0];
  const currentMp4Source = player.getElementsByClassName('mp4Source')[0];
  if(!loopFirstVideo) {
    if (!transitionVideo || !isTransition) {
      count++;
      if (count > mediaFiles.length - 1) count = 0;
    }
  }

  if (playOnlyOne) {
    if (count > 1) {
      // Remove video after playing once
      currentMp4Source.removeAttribute('src');
      nextMp4Source.removeAttribute('src');
      player.load();
      nextPlayer.load();
    }
  } else {
    // TODO: we can use this opacity to crossfade between mediaFiles
    player.style["z-index"] = 1;
    player.style["opacity"] = 1;
    nextPlayer.style["z-index"] = 0;
    nextPlayer.style["opacity"] = 0;
    let video = `${mediaFiles[count]}`;

    if (transitionVideo && transitionVideo !== "" && isTransition) {
      video = transitionVideoPath;
      isTransition = false;
    } else {
      isTransition = true;
    }
    nextMp4Source.setAttribute("src", video);
    nextPlayer.load();
    nextPlayer.pause();

    const currentVideo = currentMp4Source.getAttribute("src");
    if (currentVideo !== transitionVideoPath) {
      localStorage.setItem("lastPlayed", currentVideo);
    }
    player.play();
  }
}