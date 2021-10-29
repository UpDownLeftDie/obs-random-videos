// @ts-check
const mediaFolder = /** @type {string} */ ("{{ .MediaFolder }}");
const initMediaFiles = /** @type {string[]} */ (["{{ StringsJoin .MediaFiles "\", \"" }}"]);
const transitionVideo = /** @type {string} */("{{ .TransitionVideo }}");
const playOnlyOne = /** @type {boolean} */ ({{ .PlayOnlyOne }});
const loopFirstVideo = /** @type {boolean} */ ({{ .LoopFirstVideo }});
const transitionVideoPath = /** @type {string} */ (
  `${mediaFolder}${transitionVideo}`
);

let isTransition = true;

/**
 * shuffleArr takes in an array and returns a new array in random order
 * @param {any[]} a any array to be shuffled
 * @return {any[]} shuffled array
 */
function shuffleArr(a) {
  var j, x, i;
  for (i = a.length - 1; i > 0; i--) {
    j = Math.floor(Math.random() * (i + 1));
    x = a[i];
    a[i] = a[j];
    a[j] = x;
  }
  return a;
}

/**
 * prependFolderToFiles simply adds a folder path to a list of files and returns
 *  the new array
 * @param {string} folder folder path, must end with a trailing slash
 * @param {string[]} files array of file names
 * @returns {string[]} new array with full path to files
 */
function prependFolderToFiles(folder, files) {
  return files.map((file) => `${folder}${file}`);
}

/**
 * storePlaylistState takes in a playlist and stores it into localstorage
 * @param {string[]} state
 * @returns {void}
 */
function storePlaylistState(state) {
  localStorage.setItem('playlist', JSON.stringify(state));
}

/**
 * getNewPlaylist creates a newly randomize list of files and stores it in
 *  localstorage
 * @returns {string[]} a new playlist
 */
function getNewPlaylist() {
  const playlist = prependFolderToFiles(
    mediaFolder,
    shuffleArr(initMediaFiles),
  );
  storePlaylistState(playlist);
  return playlist;
}

/**
 * getPlaylist will get the playlist state form localstorage or create a new one
 * @returns {string[]} current playlist state
 */
function getPlaylist() {
  let playlist = [];
  try {
    playlist = JSON.parse(localStorage.getItem('playlist'));
  } catch {
    console.log('playlist empty!');
  }
  if (!playlist || playlist.length === 0) {
    playlist = getNewPlaylist();
  }
  return playlist;
}

/**
 * progressPlaylistState removes the last item from the playlist and stores the
 *  updated version in localstorage
 * @returns {void}
 */
function progressPlaylistState() {
  const playlist = getPlaylist();
  playlist.pop();
  storePlaylistState(playlist);
}

/**
 * getNextPlaylistItem returns the next item in the playlist unless it matches
 *  the last thing played then it moves that item to the end and returns the
 *  next item after that
 * @returns {string} the next item in the playlist
 */
function getNextPlaylistItem() {
  const playlist = getPlaylist();
  let mediaItem = playlist.pop();

  // check if we played this mediaItem last run
  console.log({ lastPlayed: localStorage.getItem('lastPlayed'), mediaItem });
  if (localStorage.getItem('lastPlayed') === mediaItem) {
    // moves the repeated item to the end so its not skipped entirely
    storePlaylistState([mediaItem].concat(playlist));
    mediaItem = getNextPlaylistItem();
  }
  return mediaItem;
}

/**
 * playNext is the core function of this project and handles the loading and
 *   playing of the alternating video players
 * @param {HTMLMediaElement} player currently playing video player
 * @param {HTMLMediaElement} nextPlayer the next video player to be played
 * @returns {void}
 */
function playNext(player, nextPlayer) {
  const currentMp4Source = player.getElementsByClassName('mp4Source')[0];
  const nextMp4Source = nextPlayer.getElementsByClassName('mp4Source')[0];
  const currentVideo = currentMp4Source.getAttribute('src');
  if (currentVideo !== transitionVideoPath) {
    localStorage.setItem('lastPlayed', currentVideo);
  }

  let video = localStorage.getItem('lastPlayed');
  if (!loopFirstVideo && (!transitionVideo || !isTransition)) {
    video = getNextPlaylistItem();
    console.log(`next video: ${video}`);
  }

  // TODO: we can use this opacity to crossfade between mediaFiles
  player.style['z-index'] = 1;
  player.style['opacity'] = '1';
  nextPlayer.style['z-index'] = 0;
  nextPlayer.style['opacity'] = '0';

  if (transitionVideo && transitionVideo !== '' && isTransition) {
    video = transitionVideoPath;
    isTransition = false;
  } else {
    isTransition = true;
  }
  nextMp4Source.setAttribute('src', video);
  nextPlayer.load();
  nextPlayer.pause();

  if (playOnlyOne) {
    // Remove videos after playing once
    player.onended = () => {
      currentMp4Source.removeAttribute('src');
      nextMp4Source.removeAttribute('src');
      player.load();
      nextPlayer.load();
    };
  }

  player.play();
}
