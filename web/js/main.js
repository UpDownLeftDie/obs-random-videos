// @ts-check
// @ts-ignore
const initMediaFiles = /** @type {string[]} */ (["{{ StringsJoin .MediaFiles "\", \"" }}"]);
const transitionVideoPath = /** @type {string} */("{{ .TransitionVideo }}");
const playOnlyOne = /** @type {boolean} */ ({{ .PlayOnlyOne }});
const loopFirstVideo = /** @type {boolean} */ ({{ .LoopFirstVideo }});
const hashKey = /** @type {string} */("{{ .HashKey }}");
// @ts-ignore
const isOBS = !!(window?.obsstudio?.pluginVersion);

/**
 * VideoPlayerManager handles the dual-player system for seamless video playback
 */
class VideoPlayerManager {
  /**
   * @param {string} player1Id - ID of the first video element
   * @param {string} player2Id - ID of the second video element
   */
  constructor(player1Id, player2Id) {
    /** @type {HTMLVideoElement} */
    this.player1 = /** @type {HTMLVideoElement} */ (document.getElementById(player1Id));
    /** @type {HTMLVideoElement} */
    this.player2 = /** @type {HTMLVideoElement} */ (document.getElementById(player2Id));
    /** @type {boolean} */
    this.isTransition = true; // set true for init on purpose
    /** @type {PlaylistManager} */
    this.playlistManager = new PlaylistManager(initMediaFiles, hashKey);

    this.setupEventListeners();
  }

  /**
   * Set up event listeners for both players
   */
  setupEventListeners() {
    // Player 1 ended event
    this.player1.addEventListener(
      'ended',
      () => {
        if (!loopFirstVideo) {
          this.playlistManager.progressPlaylist();
        }
        this.playNext(this.player2, this.player1);
        // Preload the next video after this one starts playing
        setTimeout(() => this.preloadNextVideo(), 1000);
      },
      { passive: true }
    );

    // Player 2 ended event
    this.player2.addEventListener(
      'ended',
      () => {
        if (!loopFirstVideo) {
          this.playlistManager.progressPlaylist();
        }
        this.playNext(this.player1, this.player2);
        // Preload the next video after this one starts playing
        setTimeout(() => this.preloadNextVideo(), 1000);
      },
      { passive: true }
    );

    // Add error handling for both players
    this.setupErrorHandling(this.player1);
    this.setupErrorHandling(this.player2);

    // Add controls for non-OBS environments
    if (!isOBS) {
      this.player1.setAttribute('controls', 'true');
      this.player2.setAttribute('controls', 'true');
    }
  }

  /**
   * Set up error handling for a player
   * @param {HTMLVideoElement} player - The video player element
   */
  setupErrorHandling(player) {
    const sources = player.querySelectorAll('source');
    sources.forEach(source => {
      source.addEventListener('error', this.handleMediaError);
      source.addEventListener('stalled', this.handleMediaError);
      source.addEventListener('suspend', this.handleMediaError);
      source.addEventListener('waiting', this.handleMediaError);
    });
  }

  /**
   * Handle media errors
   * @param {ErrorEvent | Event} event - The error event
   */
  handleMediaError(event) {
    const { type, target } = event;
    const videoSrc =
      // @ts-ignore
      /** @type {string} */ (target?.src) || 'UNKNOWN SOURCE FILE ERROR';

    switch (type) {
      case 'error':
        const errorStr = `Error loading: ${videoSrc}`;
        console.error(errorStr);
        if (isOBS) break;

        const errorText = document.createTextNode(errorStr);
        const error = document.createElement('h1');
        error.setAttribute('id', 'error');
        error.appendChild(errorText);
        document.body.innerHTML = '';
        document.body.appendChild(error);
        break;
      case 'stalled':
      case 'suspended':
      case 'waiting':
      default:
        console.error(`Error loading: [TYPE: ${type}] ${videoSrc}`);
    }
  }

  /**
   * Preload the next video in the playlist
   */
  preloadNextVideo() {
    if (playOnlyOne) return; // Don't preload if we're only playing one video

    // Get the next video but don't remove it from the playlist yet
    const nextVideoSrc = this.playlistManager.getNextItem(true);
    if (!nextVideoSrc) return;

    // Create a preload link for the browser to fetch the video in advance
    const preloadLink = document.createElement('link');
    preloadLink.rel = 'preload';
    preloadLink.as = 'video';
    preloadLink.href = nextVideoSrc;

    // Remove any existing preload links to avoid duplicates
    const existingPreloads = document.querySelectorAll('link[rel="preload"][as="video"]');
    existingPreloads.forEach(link => link.remove());

    // Add the new preload link
    document.head.appendChild(preloadLink);
    console.debug('Preloading next video:', nextVideoSrc);
  }

  /**
   * Play the next video
   * @param {HTMLVideoElement} currentPlayer - The currently playing video element
   * @param {HTMLVideoElement} nextPlayer - The next video element to play
   */
  playNext(currentPlayer, nextPlayer) {
    try {
      const currentMp4Source = /** @type {HTMLSourceElement} */(currentPlayer.querySelector('source'));
      const nextMp4Source = /** @type {HTMLSourceElement} */(nextPlayer.querySelector('source'));
      const currentVideo = currentMp4Source.getAttribute('src');

      // Load the current player
      currentPlayer.load();

      // Store the last played video if it's not the transition video
      if (currentVideo && currentVideo !== transitionVideoPath) {
        this.playlistManager.setLastPlayed(currentVideo);
      }

      let videoSrc = this.playlistManager.getLastPlayed() || '';
      // Get the next video if not looping the first video and not in transition mode
      if (!loopFirstVideo && (!transitionVideoPath || !this.isTransition)) {
        videoSrc = this.playlistManager.getNextItem();
        console.debug(`next video: ${videoSrc}`);
      }

      // Set z-index for proper video layering
      // TODO: we can use this opacity to crossfade between mediaFiles
      currentPlayer.style.zIndex = '1';
      currentPlayer.style.opacity = '1';
      nextPlayer.style.zIndex = '0';
      nextPlayer.style.opacity = '0';

      // Handle transition video logic
      if (transitionVideoPath && transitionVideoPath !== '' && this.isTransition) {
        videoSrc = transitionVideoPath;
        this.isTransition = false;
      } else {
        this.isTransition = true;
      }

      // Clear the next player's source
      nextMp4Source.src = '';
      nextMp4Source.removeAttribute('src'); // empty source

      // Set the new video source if available
      if (videoSrc) {
        nextMp4Source.setAttribute('src', videoSrc);
        nextPlayer.load();
      }

      // Handle play-once mode
      if (playOnlyOne) {
        // Remove videos after playing once
        currentPlayer.onended = () => {
          currentMp4Source.src = '';
          currentMp4Source.removeAttribute('src');
          nextMp4Source.src = '';
          nextMp4Source.removeAttribute('src');
        };
      }

      // Play the video if it's ready, otherwise wait for it to load
      this.playVideoWhenReady(currentPlayer);

    } catch (error) {
      console.error('Error in playNext function:', error);
    }
  }

  /**
   * Play the video when it's ready
   * @param {HTMLVideoElement} player - The video player element
   */
  playVideoWhenReady(player) {
    const handleLoadedData = () => {
      if (player.readyState >= 3) {
        player.play().catch(err => console.error('Error playing video:', err));
        player.removeEventListener('loadeddata', handleLoadedData, false);
      }
    };

    if (player.readyState >= 3) {
      player.play().catch(err => console.error('Error playing video:', err));
    } else {
      player.addEventListener('loadeddata', handleLoadedData, false);
    }
  }

  /**
   * Initialize the player with the first video
   */
  initialize() {
    const mp4Source = /** @type {HTMLSourceElement} */(this.player1.querySelector('source'));
    let video = this.playlistManager.getNextItem();
    // have to move the state forward after getting the first video
    this.playlistManager.progressPlaylist();

    mp4Source.setAttribute('src', video);
    this.playNext(this.player1, this.player2);
  }
}

/**
 * PlaylistManager handles the playlist state and operations
 */
class PlaylistManager {
  /**
   * @param {string[]} initialPlaylist - The initial list of media files
   * @param {string} hashKey - The hash key for localStorage
   */
  constructor(initialPlaylist, hashKey) {
    /** @type {string[]} */
    this.initialPlaylist = initialPlaylist;
    /** @type {string} */
    this.hashKey = hashKey;
  }

  /**
   * Get the current playlist from localStorage or create a new one
   * @returns {string[]} The current playlist
   */
  getPlaylist() {
    let playlist = [];
    playlist = JSON.parse(localStorage.getItem(`playlist-${this.hashKey}`) || '[]');
    if (!playlist?.length || typeof playlist.pop === 'undefined') {
      playlist = this.createNewPlaylist();
    }
    return playlist;
  }

  /**
   * Create a new shuffled playlist
   * @returns {string[]} A new shuffled playlist
   */
  createNewPlaylist() {
    const playlist = this.shuffleArray(this.initialPlaylist);
    this.storePlaylist(playlist);
    return playlist;
  }

  /**
   * Store the playlist in localStorage
   * @param {string[]} playlist - The playlist to store
   */
  storePlaylist(playlist) {
    localStorage.setItem(`playlist-${this.hashKey}`, JSON.stringify(playlist));
  }

  /**
   * Progress the playlist by removing the last item
   */
  progressPlaylist() {
    const playlist = this.getPlaylist();
    playlist.pop();
    this.storePlaylist(playlist);
  }

  /**
   * Get the next item in the playlist
   * @param {boolean} [returnEncoded=true] - Whether to return the encoded path
   * @returns {string} The next item in the playlist
   */
  getNextItem(returnEncoded = true) {
    const playlist = this.getPlaylist();
    let mediaItem = playlist.pop() || '';

    console.debug({
      lastPlayed: this.getLastPlayed(),
      mediaItem,
      wasLastPlayed: this.getLastPlayed() === mediaItem
    });

    // check if we played this mediaItem last run
    if (this.getLastPlayed() === mediaItem) {
      // moves the repeated item to the end so its not skipped entirely
      this.storePlaylist([mediaItem].concat(playlist));
      mediaItem = this.getNextItem(false);
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
   * Set the last played item
   * @param {string} videoSrc - The video source URL
   */
  setLastPlayed(videoSrc) {
    localStorage.setItem(`lastPlayed-${this.hashKey}`, videoSrc);
  }

  /**
   * Get the last played item
   * @returns {string} The last played item
   */
  getLastPlayed() {
    return localStorage.getItem(`lastPlayed-${this.hashKey}`) || '';
  }

  /**
   * Shuffle an array using the Fisher-Yates algorithm
   * @template T
   * @param {T[]} originalArray - The array to shuffle
   * @returns {T[]} The shuffled array
   */
  shuffleArray(originalArray) {
    const array = [...originalArray];
    let currentIndex = array.length;
    let randomIndex, temp;

    // While there remain elements to shuffle
    while (currentIndex > 0) {
      // Pick a remaining element
      randomIndex = Math.floor(Math.random() * currentIndex);
      currentIndex--;

      // And swap it with the current element
      temp = array[currentIndex];
      array[currentIndex] = array[randomIndex];
      array[randomIndex] = temp;
    }

    return array;
  }
}

// Export the classes for use in other files
// @ts-ignore
window.VideoPlayerManager = VideoPlayerManager;
// @ts-ignore
window.PlaylistManager = PlaylistManager;
