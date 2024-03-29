// @ts-check
const initMediaFiles = /** @type {string[]} */ (["{{ StringsJoin .MediaFiles "\", \"" }}"]);
const transitionVideoPath = /** @type {string} */("{{ .TransitionVideo }}");
const playOnlyOne = /** @type {boolean} */ ({{ .PlayOnlyOne }});
const loopFirstVideo = /** @type {boolean} */ ({{ .LoopFirstVideo }});
const hashKey = /** @type {string} */("{{ .HashKey }}");
// @ts-ignore
const isOBS = !!(window?.obsstudio?.pluginVersion);
let isTransition = true; // set true for init on purpose


/**
 * shuffleArr takes in an array and returns a new array in random order
 * @template T
 * @param {T[]} originalArray any array to be shuffled
 * @return {T[]} shuffled array
 */
function shuffleArr(originalArray) {
  let a = [...originalArray]
  let j, x, i;
  for (i = a.length - 1; i > 0; i--) {
    j = Math.floor(Math.random() * (i + 1));
    x = a[i];
    a[i] = a[j];
    a[j] = x;
  }
  return a;
}

/**
 * storePlaylistState takes in a playlist and stores it into localstorage
 * @param {string[]} state
 * @returns {void}
 */
function storePlaylistState(state) {
  localStorage.setItem(`playlist-${hashKey}`, JSON.stringify(state));
}

/**
 * getNewPlaylist creates a newly randomize list of files and stores it in
 *  localstorage
 * @returns {string[]} a new playlist
 */
function getNewPlaylist() {
  const playlist = shuffleArr(initMediaFiles);
  storePlaylistState(playlist);
  return playlist;
}

/**
 * getPlaylist will get the playlist state form localstorage or create a new one
 * @returns {string[]} current playlist state
 */
function getPlaylist() {
  let playlist = [];
  playlist = JSON.parse(localStorage.getItem(`playlist-${hashKey}`) || '[]');
  if (!playlist?.length || typeof playlist.pop === 'undefined') {
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
function getNextPlaylistItem(returnEncoded = true) {
  const playlist = getPlaylist();
  let mediaItem = playlist.pop() || '';

  console.debug({
    lastPlayed: localStorage.getItem(`lastPlayed-${hashKey}`),
    mediaItem,
    wasLastPlayed: localStorage.getItem(`lastPlayed-${hashKey}`) === mediaItem
  });

  // check if we played this mediaItem last run
  if (localStorage.getItem(`lastPlayed-${hashKey}`) === mediaItem) {
    // moves the repeated item to the end so its not skipped entirely
    storePlaylistState([mediaItem].concat(playlist));
    mediaItem = getNextPlaylistItem(false);
  }

  if (returnEncoded) {
    const parts = mediaItem.split('/');
    // @ts-ignore
    const file = encodeURIComponent(parts.pop());
    const path = `${encodeURI(parts.join('/'))}`;
    mediaItem = `${path}/${file}`;
  }
  return mediaItem;
}

/**
 * playNext is the core function of this project and handles the loading and
 *   playing of the alternating video players
 * @param {HTMLMediaElement} currentPlayer currently playing video player
 * @param {HTMLMediaElement} nextPlayer the next video player to be played
 * @returns {void}
 */
function playNext(currentPlayer, nextPlayer) {
  const currentMp4Source = /** @type {HTMLSourceElement} */(currentPlayer.getElementsByClassName('mp4Source')[0]);
  const nextMp4Source  =  /** @type {HTMLSourceElement} */(nextPlayer.getElementsByClassName('mp4Source')[0]);
  const currentVideo = currentMp4Source.getAttribute('src');

  currentPlayer.load();
  if (currentVideo && currentVideo !== transitionVideoPath) {
    localStorage.setItem(`lastPlayed-${hashKey}`, currentVideo);
  }

  let videoSrc = localStorage.getItem(`lastPlayed-${hashKey}`) || '';
  if (!loopFirstVideo && (!transitionVideoPath || !isTransition)) {
    videoSrc = getNextPlaylistItem();
    console.debug(`next video: ${videoSrc}`);
  }

  // TODO: we can use this opacity to crossfade between mediaFiles
  currentPlayer.style['z-index'] = 1;
  currentPlayer.style['opacity'] = '1';
  nextPlayer.style['z-index'] = 0;
  nextPlayer.style['opacity'] = '0';

  if (transitionVideoPath && transitionVideoPath !== '' && isTransition) {
    videoSrc = transitionVideoPath;
    isTransition = false;
  } else {
    isTransition = true;
  }

  nextMp4Source.src = ''
  nextMp4Source.removeAttribute('src'); // empty source


  if (videoSrc) {
    nextMp4Source.setAttribute('src', videoSrc);
    nextPlayer.load();
  }

  if (playOnlyOne) {
    // Remove videos after playing once
    currentPlayer.onended = () => {
      currentMp4Source.src = '';
      currentMp4Source.removeAttribute('src');
      nextMp4Source.src = '';
      nextMp4Source.removeAttribute('src');
    };
  }


  /**
 * handleLoadedData event listener for data to load will play the video on HAVE_CURRENT_DATA
 * https://developer.mozilla.org/en-US/docs/Web/API/HTMLMediaElement/readyState
 * @returns {void}
 */
  function handleLoadedData() {
    if (currentPlayer.readyState >= 3) {
      currentPlayer.play();
      currentPlayer.removeEventListener('loadeddata', handleLoadedData, false);
    }
   }

   if (currentPlayer.readyState >= 3) {
     currentPlayer.play();
   } else {
     currentPlayer.addEventListener('loadeddata', handleLoadedData, false);
   }
}
